package tonic

import (
	"reflect"
	"runtime"
	"strings"
)

// HandlerInfo represents a tonic-wrapped handler informations.
type HandlerInfo struct {
	// Handler is the route handler.
	Handler reflect.Value

	// HandlerType is the type of the route handler.
	HandlerType reflect.Type

	// inputType is the type of the input object.
	// This can be nil if the handler use none.
	inputType reflect.Type

	// outputType is the type of the output object.
	// This can be nil if the handler use none.
	outputType reflect.Type

	Status int
}

// InputType returns the input type of the handler.
// If the type is a pointer to a concrete type, it
// is dereferenced.
func (r *HandlerInfo) InputType() reflect.Type {
	if in := r.inputType; in != nil && in.Kind() == reflect.Ptr {
		return in.Elem()
	}
	return r.inputType
}

// OutputType returns the output type of the handler.
// If the type is a pointer to a concrete type, it
// is dereferenced.
func (r *HandlerInfo) OutputType() reflect.Type {
	if out := r.outputType; out != nil && out.Kind() == reflect.Ptr {
		return out.Elem()
	}
	return r.outputType
}

// Name returns the name of the route handler.
func (r *HandlerInfo) Name() string {
	parts := strings.Split(r.NameWithPackage(), ".")
	return parts[len(parts)-1]
}

// NameWithPackage returns the full name of the rout
// handler with its package path.
func (r *HandlerInfo) NameWithPackage() string {
	f := runtime.FuncForPC(r.Handler.Pointer()).Name()
	parts := strings.Split(f, "/")
	return parts[len(parts)-1]
}

// OperationInfo represenst the informations of an operation
// that will be used when generating the OpenAPI specification.
type OperationInfo struct {
	StatusCode        int
	StatusDescription string
	Headers           []*ResponseHeader
	Summary           string
	Description       string
	Deprecated        bool
	Responses         []*OperationReponse
}

// ResponseHeader represents a single header that
// may be returned with an operation response.
type ResponseHeader struct {
	Name        string
	Description string
	Model       interface{}
}

// OperationReponse represents a single response of an
// API operation.
type OperationReponse struct {
	// The response code can be "default"
	// accotding to OAS3.
	Code        string
	Description string
	Model       interface{}
	Headers     []*ResponseHeader
}

// StatusCode sets the default status code of the operation.
func StatusCode(code int) func(*OperationInfo) {
	return func(o *OperationInfo) {
		o.StatusCode = code
	}
}

// StatusDescription sets the default status description of the operation.
func StatusDescription(desc string) func(*OperationInfo) {
	return func(o *OperationInfo) {
		o.StatusDescription = desc
	}
}

// Summary adds a summary to an operation.
func Summary(summary string) func(*OperationInfo) {
	return func(o *OperationInfo) {
		o.Summary = summary
	}
}

// Description adds a description to an operation.
func Description(desc string) func(*OperationInfo) {
	return func(o *OperationInfo) {
		o.Description = desc
	}
}

// Deprecated marks the operation as deprecated.
func Deprecated(deprecated bool) func(*OperationInfo) {
	return func(o *OperationInfo) {
		o.Deprecated = deprecated
	}
}

// Response adds an additional response to the operation.
func Response(code, desc string, model interface{}, headers []*ResponseHeader) func(*OperationInfo) {
	return func(o *OperationInfo) {
		o.Responses = append(o.Responses, &OperationReponse{
			Code:        code,
			Description: desc,
			Model:       model,
			Headers:     headers,
		})
	}
}

// Header adds a header to the operation.
func Header(name, desc string, model interface{}) func(*OperationInfo) {
	return func(o *OperationInfo) {
		o.Headers = append(o.Headers, &ResponseHeader{
			Name:        name,
			Description: desc,
			Model:       model,
		})
	}
}
