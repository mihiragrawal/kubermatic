// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// VSphere v sphere
//
// swagger:model VSphere
type VSphere struct {

	// datacenter
	Datacenter string `json:"datacenter,omitempty"`

	// datastore
	Datastore string `json:"datastore,omitempty"`

	// datastore cluster
	DatastoreCluster string `json:"datastoreCluster,omitempty"`

	// enabled
	Enabled bool `json:"enabled,omitempty"`

	// password
	Password string `json:"password,omitempty"`

	// resource pool
	ResourcePool string `json:"resourcePool,omitempty"`

	// username
	Username string `json:"username,omitempty"`

	// VM net name
	VMNetName string `json:"vmNetName,omitempty"`
}

// Validate validates this v sphere
func (m *VSphere) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this v sphere based on context it is used
func (m *VSphere) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VSphere) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VSphere) UnmarshalBinary(b []byte) error {
	var res VSphere
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
