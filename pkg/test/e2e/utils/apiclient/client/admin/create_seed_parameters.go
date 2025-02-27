// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewCreateSeedParams creates a new CreateSeedParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewCreateSeedParams() *CreateSeedParams {
	return &CreateSeedParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewCreateSeedParamsWithTimeout creates a new CreateSeedParams object
// with the ability to set a timeout on a request.
func NewCreateSeedParamsWithTimeout(timeout time.Duration) *CreateSeedParams {
	return &CreateSeedParams{
		timeout: timeout,
	}
}

// NewCreateSeedParamsWithContext creates a new CreateSeedParams object
// with the ability to set a context for a request.
func NewCreateSeedParamsWithContext(ctx context.Context) *CreateSeedParams {
	return &CreateSeedParams{
		Context: ctx,
	}
}

// NewCreateSeedParamsWithHTTPClient creates a new CreateSeedParams object
// with the ability to set a custom HTTPClient for a request.
func NewCreateSeedParamsWithHTTPClient(client *http.Client) *CreateSeedParams {
	return &CreateSeedParams{
		HTTPClient: client,
	}
}

/* CreateSeedParams contains all the parameters to send to the API endpoint
   for the create seed operation.

   Typically these are written to a http.Request.
*/
type CreateSeedParams struct {

	// Body.
	Body CreateSeedBody

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the create seed params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CreateSeedParams) WithDefaults() *CreateSeedParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the create seed params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CreateSeedParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the create seed params
func (o *CreateSeedParams) WithTimeout(timeout time.Duration) *CreateSeedParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the create seed params
func (o *CreateSeedParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the create seed params
func (o *CreateSeedParams) WithContext(ctx context.Context) *CreateSeedParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the create seed params
func (o *CreateSeedParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the create seed params
func (o *CreateSeedParams) WithHTTPClient(client *http.Client) *CreateSeedParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the create seed params
func (o *CreateSeedParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the create seed params
func (o *CreateSeedParams) WithBody(body CreateSeedBody) *CreateSeedParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the create seed params
func (o *CreateSeedParams) SetBody(body CreateSeedBody) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *CreateSeedParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if err := r.SetBodyParam(o.Body); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
