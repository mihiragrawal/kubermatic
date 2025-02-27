// Code generated by go-swagger; DO NOT EDIT.

package project

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

// NewListCNIPluginVersionsForClusterParams creates a new ListCNIPluginVersionsForClusterParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListCNIPluginVersionsForClusterParams() *ListCNIPluginVersionsForClusterParams {
	return &ListCNIPluginVersionsForClusterParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListCNIPluginVersionsForClusterParamsWithTimeout creates a new ListCNIPluginVersionsForClusterParams object
// with the ability to set a timeout on a request.
func NewListCNIPluginVersionsForClusterParamsWithTimeout(timeout time.Duration) *ListCNIPluginVersionsForClusterParams {
	return &ListCNIPluginVersionsForClusterParams{
		timeout: timeout,
	}
}

// NewListCNIPluginVersionsForClusterParamsWithContext creates a new ListCNIPluginVersionsForClusterParams object
// with the ability to set a context for a request.
func NewListCNIPluginVersionsForClusterParamsWithContext(ctx context.Context) *ListCNIPluginVersionsForClusterParams {
	return &ListCNIPluginVersionsForClusterParams{
		Context: ctx,
	}
}

// NewListCNIPluginVersionsForClusterParamsWithHTTPClient creates a new ListCNIPluginVersionsForClusterParams object
// with the ability to set a custom HTTPClient for a request.
func NewListCNIPluginVersionsForClusterParamsWithHTTPClient(client *http.Client) *ListCNIPluginVersionsForClusterParams {
	return &ListCNIPluginVersionsForClusterParams{
		HTTPClient: client,
	}
}

/* ListCNIPluginVersionsForClusterParams contains all the parameters to send to the API endpoint
   for the list c n i plugin versions for cluster operation.

   Typically these are written to a http.Request.
*/
type ListCNIPluginVersionsForClusterParams struct {

	// ClusterID.
	ClusterID string

	// ProjectID.
	ProjectID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list c n i plugin versions for cluster params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListCNIPluginVersionsForClusterParams) WithDefaults() *ListCNIPluginVersionsForClusterParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list c n i plugin versions for cluster params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListCNIPluginVersionsForClusterParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list c n i plugin versions for cluster params
func (o *ListCNIPluginVersionsForClusterParams) WithTimeout(timeout time.Duration) *ListCNIPluginVersionsForClusterParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list c n i plugin versions for cluster params
func (o *ListCNIPluginVersionsForClusterParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list c n i plugin versions for cluster params
func (o *ListCNIPluginVersionsForClusterParams) WithContext(ctx context.Context) *ListCNIPluginVersionsForClusterParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list c n i plugin versions for cluster params
func (o *ListCNIPluginVersionsForClusterParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list c n i plugin versions for cluster params
func (o *ListCNIPluginVersionsForClusterParams) WithHTTPClient(client *http.Client) *ListCNIPluginVersionsForClusterParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list c n i plugin versions for cluster params
func (o *ListCNIPluginVersionsForClusterParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithClusterID adds the clusterID to the list c n i plugin versions for cluster params
func (o *ListCNIPluginVersionsForClusterParams) WithClusterID(clusterID string) *ListCNIPluginVersionsForClusterParams {
	o.SetClusterID(clusterID)
	return o
}

// SetClusterID adds the clusterId to the list c n i plugin versions for cluster params
func (o *ListCNIPluginVersionsForClusterParams) SetClusterID(clusterID string) {
	o.ClusterID = clusterID
}

// WithProjectID adds the projectID to the list c n i plugin versions for cluster params
func (o *ListCNIPluginVersionsForClusterParams) WithProjectID(projectID string) *ListCNIPluginVersionsForClusterParams {
	o.SetProjectID(projectID)
	return o
}

// SetProjectID adds the projectId to the list c n i plugin versions for cluster params
func (o *ListCNIPluginVersionsForClusterParams) SetProjectID(projectID string) {
	o.ProjectID = projectID
}

// WriteToRequest writes these params to a swagger request
func (o *ListCNIPluginVersionsForClusterParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param cluster_id
	if err := r.SetPathParam("cluster_id", o.ClusterID); err != nil {
		return err
	}

	// path param project_id
	if err := r.SetPathParam("project_id", o.ProjectID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
