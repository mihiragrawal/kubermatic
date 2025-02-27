// Code generated by go-swagger; DO NOT EDIT.

package project

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"k8c.io/kubermatic/v2/pkg/test/e2e/utils/apiclient/models"
)

// AssignSSHKeyToClusterV2Reader is a Reader for the AssignSSHKeyToClusterV2 structure.
type AssignSSHKeyToClusterV2Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *AssignSSHKeyToClusterV2Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewAssignSSHKeyToClusterV2Created()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewAssignSSHKeyToClusterV2Unauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewAssignSSHKeyToClusterV2Forbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewAssignSSHKeyToClusterV2Default(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewAssignSSHKeyToClusterV2Created creates a AssignSSHKeyToClusterV2Created with default headers values
func NewAssignSSHKeyToClusterV2Created() *AssignSSHKeyToClusterV2Created {
	return &AssignSSHKeyToClusterV2Created{}
}

/* AssignSSHKeyToClusterV2Created describes a response with status code 201, with default header values.

SSHKey
*/
type AssignSSHKeyToClusterV2Created struct {
	Payload *models.SSHKey
}

func (o *AssignSSHKeyToClusterV2Created) Error() string {
	return fmt.Sprintf("[PUT /api/v2/projects/{project_id}/clusters/{cluster_id}/sshkeys/{key_id}][%d] assignSshKeyToClusterV2Created  %+v", 201, o.Payload)
}
func (o *AssignSSHKeyToClusterV2Created) GetPayload() *models.SSHKey {
	return o.Payload
}

func (o *AssignSSHKeyToClusterV2Created) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.SSHKey)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAssignSSHKeyToClusterV2Unauthorized creates a AssignSSHKeyToClusterV2Unauthorized with default headers values
func NewAssignSSHKeyToClusterV2Unauthorized() *AssignSSHKeyToClusterV2Unauthorized {
	return &AssignSSHKeyToClusterV2Unauthorized{}
}

/* AssignSSHKeyToClusterV2Unauthorized describes a response with status code 401, with default header values.

EmptyResponse is a empty response
*/
type AssignSSHKeyToClusterV2Unauthorized struct {
}

func (o *AssignSSHKeyToClusterV2Unauthorized) Error() string {
	return fmt.Sprintf("[PUT /api/v2/projects/{project_id}/clusters/{cluster_id}/sshkeys/{key_id}][%d] assignSshKeyToClusterV2Unauthorized ", 401)
}

func (o *AssignSSHKeyToClusterV2Unauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewAssignSSHKeyToClusterV2Forbidden creates a AssignSSHKeyToClusterV2Forbidden with default headers values
func NewAssignSSHKeyToClusterV2Forbidden() *AssignSSHKeyToClusterV2Forbidden {
	return &AssignSSHKeyToClusterV2Forbidden{}
}

/* AssignSSHKeyToClusterV2Forbidden describes a response with status code 403, with default header values.

EmptyResponse is a empty response
*/
type AssignSSHKeyToClusterV2Forbidden struct {
}

func (o *AssignSSHKeyToClusterV2Forbidden) Error() string {
	return fmt.Sprintf("[PUT /api/v2/projects/{project_id}/clusters/{cluster_id}/sshkeys/{key_id}][%d] assignSshKeyToClusterV2Forbidden ", 403)
}

func (o *AssignSSHKeyToClusterV2Forbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewAssignSSHKeyToClusterV2Default creates a AssignSSHKeyToClusterV2Default with default headers values
func NewAssignSSHKeyToClusterV2Default(code int) *AssignSSHKeyToClusterV2Default {
	return &AssignSSHKeyToClusterV2Default{
		_statusCode: code,
	}
}

/* AssignSSHKeyToClusterV2Default describes a response with status code -1, with default header values.

errorResponse
*/
type AssignSSHKeyToClusterV2Default struct {
	_statusCode int

	Payload *models.ErrorResponse
}

// Code gets the status code for the assign SSH key to cluster v2 default response
func (o *AssignSSHKeyToClusterV2Default) Code() int {
	return o._statusCode
}

func (o *AssignSSHKeyToClusterV2Default) Error() string {
	return fmt.Sprintf("[PUT /api/v2/projects/{project_id}/clusters/{cluster_id}/sshkeys/{key_id}][%d] assignSSHKeyToClusterV2 default  %+v", o._statusCode, o.Payload)
}
func (o *AssignSSHKeyToClusterV2Default) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *AssignSSHKeyToClusterV2Default) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
