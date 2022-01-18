// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// __          __              _
// \ \        / /             (_)
//  \ \  /\  / /_ _ _ __ _ __  _ _ __   __ _
//   \ \/  \/ / _` | '__| '_ \| | '_ \ / _` |
//    \  /\  / (_| | |  | | | | | | | | (_| | : This file is generated, do not edit it.
//     \/  \/ \__,_|_|  |_| |_|_|_| |_|\__, |
//                                      __/ |
//                                     |___/

package rest_model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// APIError api error
//
// swagger:model apiError
type APIError struct {

	// args
	Args *APIErrorArgs `json:"args,omitempty"`

	// cause
	Cause *APIErrorCause `json:"cause,omitempty"`

	// cause message
	CauseMessage string `json:"causeMessage,omitempty"`

	// code
	Code string `json:"code,omitempty"`

	// data
	Data interface{} `json:"data,omitempty"`

	// message
	Message string `json:"message,omitempty"`

	// request Id
	RequestID string `json:"requestId,omitempty"`
}

// Validate validates this api error
func (m *APIError) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateArgs(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCause(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *APIError) validateArgs(formats strfmt.Registry) error {
	if swag.IsZero(m.Args) { // not required
		return nil
	}

	if m.Args != nil {
		if err := m.Args.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("args")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("args")
			}
			return err
		}
	}

	return nil
}

func (m *APIError) validateCause(formats strfmt.Registry) error {
	if swag.IsZero(m.Cause) { // not required
		return nil
	}

	if m.Cause != nil {
		if err := m.Cause.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("cause")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("cause")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this api error based on the context it is used
func (m *APIError) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateArgs(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateCause(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *APIError) contextValidateArgs(ctx context.Context, formats strfmt.Registry) error {

	if m.Args != nil {
		if err := m.Args.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("args")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("args")
			}
			return err
		}
	}

	return nil
}

func (m *APIError) contextValidateCause(ctx context.Context, formats strfmt.Registry) error {

	if m.Cause != nil {
		if err := m.Cause.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("cause")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("cause")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *APIError) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *APIError) UnmarshalBinary(b []byte) error {
	var res APIError
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}