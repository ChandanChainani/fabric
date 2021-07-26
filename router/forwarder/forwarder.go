/*
	Copyright NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package forwarder

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fabric/pb/ctrl_pb"
	"github.com/openziti/fabric/router/xgress"
	"github.com/openziti/fabric/router/xlink"
	"github.com/openziti/fabric/trace"
	"github.com/openziti/foundation/metrics"
	"github.com/openziti/foundation/util/info"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type Forwarder struct {
	circuits        *circuitTable
	destinations    *destinationTable
	faulter         *Faulter
	scanner         *Scanner
	metricsRegistry metrics.UsageRegistry
	traceController trace.Controller
	Options         *Options
	CloseNotify     <-chan struct{}
}

type Destination interface {
	SendPayload(payload *xgress.Payload) error
	SendAcknowledgement(acknowledgement *xgress.Acknowledgement) error
}

type XgressDestination interface {
	Destination
	Unrouted()
	Start()
	IsTerminator() bool
	Label() string
	GetTimeOfLastRxFromLink() int64
}

func NewForwarder(metricsRegistry metrics.UsageRegistry, faulter *Faulter, scanner *Scanner, options *Options, closeNotify <-chan struct{}) *Forwarder {
	f := &Forwarder{
		circuits:        newCircuitTable(),
		destinations:    newDestinationTable(),
		faulter:         faulter,
		scanner:         scanner,
		metricsRegistry: metricsRegistry,
		traceController: trace.NewController(closeNotify),
		Options:         options,
		CloseNotify:     closeNotify,
	}
	f.scanner.setCircuitTable(f.circuits)
	return f
}

func (forwarder *Forwarder) MetricsRegistry() metrics.UsageRegistry {
	return forwarder.metricsRegistry
}

func (forwarder *Forwarder) TraceController() trace.Controller {
	return forwarder.traceController
}

func (forwarder *Forwarder) RegisterDestination(circuitId string, address xgress.Address, destination Destination) {
	forwarder.destinations.addDestination(address, destination)
	forwarder.destinations.linkDestinationToCircuit(circuitId, address)
}

func (forwarder *Forwarder) UnregisterDestinations(circuitId string) {
	if addresses, found := forwarder.destinations.getAddressesForCircuit(circuitId); found {
		for _, address := range addresses {
			if destination, found := forwarder.destinations.getDestination(address); found {
				pfxlog.Logger().Debugf("unregistering destination [@/%v] for [s/%v]", address, circuitId)
				forwarder.destinations.removeDestination(address)
				go destination.(XgressDestination).Unrouted()
			} else {
				pfxlog.Logger().Debugf("no destinations found for [@/%v] for [s/%v]", address, circuitId)
			}
		}
		forwarder.destinations.unlinkCircuit(circuitId)
	} else {
		pfxlog.Logger().Debugf("found no addresses to unregister for [s/%v]", circuitId)
	}
}

func (forwarder *Forwarder) HasDestination(address xgress.Address) bool {
	_, found := forwarder.destinations.getDestination(address)
	return found
}

func (forwarder *Forwarder) RegisterLink(link xlink.Xlink) {
	forwarder.destinations.addDestination(xgress.Address(link.Id().Token), link)
}

func (forwarder *Forwarder) UnregisterLink(link xlink.Xlink) {
	forwarder.destinations.removeDestination(xgress.Address(link.Id().Token))
}

func (forwarder *Forwarder) Route(route *ctrl_pb.Route) {
	circuitId := route.CircuitId
	var circuitFt *forwardTable
	if ft, found := forwarder.circuits.getForwardTable(circuitId); found {
		circuitFt = ft
	} else {
		circuitFt = newForwardTable()
	}
	for _, forward := range route.Forwards {
		circuitFt.setForwardAddress(xgress.Address(forward.SrcAddress), xgress.Address(forward.DstAddress))
	}
	forwarder.circuits.setForwardTable(circuitId, circuitFt)
}

func (forwarder *Forwarder) Unroute(circuitId string, now bool) {
	if now {
		forwarder.circuits.removeForwardTable(circuitId)
		forwarder.EndCircuit(circuitId)
	} else {
		go forwarder.unrouteTimeout(circuitId, forwarder.Options.XgressCloseCheckInterval)
	}
}

func (forwarder *Forwarder) EndCircuit(circuitId string) {
	forwarder.UnregisterDestinations(circuitId)
}

func (forwarder *Forwarder) ForwardPayload(srcAddr xgress.Address, payload *xgress.Payload) error {
	log := pfxlog.ContextLogger(string(srcAddr))

	circuitId := payload.GetCircuitId()
	if forwardTable, found := forwarder.circuits.getForwardTable(circuitId); found {
		if dstAddr, found := forwardTable.getForwardAddress(srcAddr); found {
			if dst, found := forwarder.destinations.getDestination(dstAddr); found {
				if err := dst.SendPayload(payload); err != nil {
					return err
				}
				log.WithFields(payload.GetLoggerFields()).Debugf("=> %s", string(dstAddr))
				return nil
			} else {
				return errors.Errorf("cannot forward payload, no destination for circuit=%v src=%v dst=%v", circuitId, srcAddr, dstAddr)
			}
		} else {
			return errors.Errorf("cannot forward payload, no destination address for circuit=%v src=%v", circuitId, srcAddr)
		}
	} else {
		return errors.Errorf("cannot forward payload, no forward table for circuit=%v src=%v", circuitId, srcAddr)
	}
}

func (forwarder *Forwarder) ForwardAcknowledgement(srcAddr xgress.Address, acknowledgement *xgress.Acknowledgement) error {
	log := pfxlog.ContextLogger(string(srcAddr))

	circuitId := acknowledgement.CircuitId
	if forwardTable, found := forwarder.circuits.getForwardTable(circuitId); found {
		if dstAddr, found := forwardTable.getForwardAddress(srcAddr); found {
			if dst, found := forwarder.destinations.getDestination(dstAddr); found {
				if err := dst.SendAcknowledgement(acknowledgement); err != nil {
					return err
				}
				log.Debugf("=> %s", string(dstAddr))
				return nil

			} else {
				return errors.Errorf("cannot acknowledge, no destination for circuit=%v src=%v dst=%v", circuitId, srcAddr, dstAddr)
			}

		} else {
			return errors.Errorf("cannot acknowledge, no destination address for circuit=%v src=%v", circuitId, srcAddr)
		}

	} else {
		return errors.Errorf("cannot acknowledge, no forward table for circuit=%v src=%v", circuitId, srcAddr)
	}
}

func (forwarder *Forwarder) ReportForwardingFault(circuitId string) {
	if forwarder.faulter != nil {
		forwarder.faulter.report(circuitId)
	} else {
		logrus.Errorf("nil faulter, cannot accept forwarding fault report")
	}
}

func (forwarder *Forwarder) Debug() string {
	return forwarder.circuits.debug() + forwarder.destinations.debug()
}

// unrouteTimeout implements a goroutine to manage route timeout processing. Once a timeout processor has been launched
// for a circuit, it will be checked repeatedly, looking to see if the circuit has crossed the inactivity threshold.
// Once it crosses the inactivity threshold, it gets removed.
//
func (forwarder *Forwarder) unrouteTimeout(circuitId string, interval time.Duration) {
	log := pfxlog.ContextLogger("c/" + circuitId)
	log.Debug("scheduled")
	defer log.Debug("timeout")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if dest := forwarder.getXgressForCircuit(circuitId); dest != nil {
				elapsedDelta := info.NowInMilliseconds() - dest.GetTimeOfLastRxFromLink()
				if (time.Duration(elapsedDelta) * time.Millisecond) >= interval {
					forwarder.circuits.removeForwardTable(circuitId)
					forwarder.EndCircuit(circuitId)
					return
				}
			} else {
				forwarder.circuits.removeForwardTable(circuitId)
				forwarder.EndCircuit(circuitId)
				return
			}
		case <-forwarder.CloseNotify:
			return
		}
	}
}

func (forwarder *Forwarder) getXgressForCircuit(circuitId string) XgressDestination {
	if addresses, found := forwarder.destinations.getAddressesForCircuit(circuitId); found {
		for _, address := range addresses {
			if destination, found := forwarder.destinations.getDestination(address); found {
				return destination.(XgressDestination)
			}
		}
	}
	return nil
}
