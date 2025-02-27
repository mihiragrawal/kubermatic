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

// UnbindUserFromRoleBindingReader is a Reader for the UnbindUserFromRoleBinding structure.
type UnbindUserFromRoleBindingReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UnbindUserFromRoleBindingReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUnbindUserFromRoleBindingOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewUnbindUserFromRoleBindingUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewUnbindUserFromRoleBindingForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewUnbindUserFromRoleBindingDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewUnbindUserFromRoleBindingOK creates a UnbindUserFromRoleBindingOK with default headers values
func NewUnbindUserFromRoleBindingOK() *UnbindUserFromRoleBindingOK {
	return &UnbindUserFromRoleBindingOK{}
}

/* UnbindUserFromRoleBindingOK describes a response with status code 200, with default header values.

RoleBinding
*/
type UnbindUserFromRoleBindingOK struct {
	Payload *models.RoleBinding
}

func (o *UnbindUserFromRoleBindingOK) Error() string {
	return fmt.Sprintf("[DELETE /api/v1/projects/{project_id}/dc/{dc}/clusters/{cluster_id}/roles/{namespace}/{role_id}/bindings][%d] unbindUserFromRoleBindingOK  %+v", 200, o.Payload)
}
func (o *UnbindUserFromRoleBindingOK) GetPayload() *models.RoleBinding {
	return o.Payload
}

func (o *UnbindUserFromRoleBindingOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RoleBinding)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUnbindUserFromRoleBindingUnauthorized creates a UnbindUserFromRoleBindingUnauthorized with default headers values
func NewUnbindUserFromRoleBindingUnauthorized() *UnbindUserFromRoleBindingUnauthorized {
	return &UnbindUserFromRoleBindingUnauthorized{}
}

/* UnbindUserFromRoleBindingUnauthorized describes a response with status code 401, with default header values.

EmptyResponse is a empty response
*/
type UnbindUserFromRoleBindingUnauthorized struct {
}

func (o *UnbindUserFromRoleBindingUnauthorized) Error() string {
	return fmt.Sprintf("[DELETE /api/v1/projects/{project_id}/dc/{dc}/clusters/{cluster_id}/roles/{namespace}/{role_id}/bindings][%d] unbindUserFromRoleBindingUnauthorized ", 401)
}

func (o *UnbindUserFromRoleBindingUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewUnbindUserFromRoleBindingForbidden creates a UnbindUserFromRoleBindingForbidden with default headers values
func NewUnbindUserFromRoleBindingForbidden() *UnbindUserFromRoleBindingForbidden {
	return &UnbindUserFromRoleBindingForbidden{}
}

/* UnbindUserFromRoleBindingForbidden describes a response with status code 403, with default header values.

EmptyResponse is a empty response
*/
type UnbindUserFromRoleBindingForbidden struct {
}

func (o *UnbindUserFromRoleBindingForbidden) Error() string {
	return fmt.Sprintf("[DELETE /api/v1/projects/{project_id}/dc/{dc}/clusters/{cluster_id}/roles/{namespace}/{role_id}/bindings][%d] unbindUserFromRoleBindingForbidden ", 403)
}

func (o *UnbindUserFromRoleBindingForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewUnbindUserFromRoleBindingDefault creates a UnbindUserFromRoleBindingDefault with default headers values
func NewUnbindUserFromRoleBindingDefault(code int) *UnbindUserFromRoleBindingDefault {
	return &UnbindUserFromRoleBindingDefault{
		_statusCode: code,
	}
}

/* UnbindUserFromRoleBindingDefault describes a response with status code -1, with default header values.

errorResponse
*/
type UnbindUserFromRoleBindingDefault struct {
	_statusCode int

	Payload *models.ErrorResponse
}

// Code gets the status code for the unbind user from role binding default response
func (o *UnbindUserFromRoleBindingDefault) Code() int {
	return o._statusCode
}

func (o *UnbindUserFromRoleBindingDefault) Error() string {
	return fmt.Sprintf("[DELETE /api/v1/projects/{project_id}/dc/{dc}/clusters/{cluster_id}/roles/{namespace}/{role_id}/bindings][%d] unbindUserFromRoleBinding default  %+v", o._statusCode, o.Payload)
}
func (o *UnbindUserFromRoleBindingDefault) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *UnbindUserFromRoleBindingDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
