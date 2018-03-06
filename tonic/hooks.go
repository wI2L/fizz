package tonic

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var defaultMediaType = "application/json"

// BindHook is the hook called by the wrapping gin-handler when
// binding an incoming request to the tonic-handler's input object.
type BindHook func(*gin.Context, interface{}) error

// RenderHook is the last hook called by the wrapping gin-handler
// before returning. It takes the Gin context, the HTTP status code
// and the response payload as parameters.
// Its role is to render the payload to the client to the
// proper format.
type RenderHook func(*gin.Context, int, interface{})

// ErrorHook lets you interpret errors returned by your handlers.
// After analysis, the hook should return a suitable http status code
// and and error payload.
// This lets you deeply inspect custom error types.
type ErrorHook func(*gin.Context, error) (int, interface{})

// DefaultErrorHook is the default error hook.
// It returns a StatusBadRequest with a payload containing
// the error message.
func DefaultErrorHook(c *gin.Context, e error) (int, interface{}) {
	return http.StatusBadRequest, gin.H{
		"error": e.Error(),
	}
}

// DefaultBindingHook is the default binding hook.
// It uses Gin JSON binding to bind the body parameters of the request
// to the input object of the handler.
// Ir teturns an error if Gin binding fails.
func DefaultBindingHook(c *gin.Context, i interface{}) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	if c.Request.ContentLength == 0 || c.Request.Method == http.MethodGet {
		return nil
	}
	if err := c.ShouldBindWith(i, binding.JSON); err != nil && err != io.EOF {
		return fmt.Errorf("error parsing request body: %s", err.Error())
	}
	return nil
}

// DefaultRenderHook is the default render hook.
// It marshals the payload to JSON, or returns an empty body if the payload is nil.
// If Gin is running in debug mode, the marshalled JSON is indented.
func DefaultRenderHook(c *gin.Context, status int, payload interface{}) {
	if payload != nil {
		if gin.IsDebugging() {
			c.IndentedJSON(status, payload)
		} else {
			c.JSON(status, payload)
		}
	} else {
		c.String(status, "")
	}
}

// MediaType returns the current media type (MIME)
// used by the actual render hook.
func MediaType() string {
	return defaultMediaType
}

// SetErrorHook sets the given hook as the
// default error handling hook.
func SetErrorHook(eh ErrorHook) {
	if eh != nil {
		errorHook = eh
	}
}

// SetBindHook sets the given hook as the
// default binding hook.
func SetBindHook(bh BindHook) {
	if bh != nil {
		bindHook = bh
	}
}

// SetRenderHook sets the given hook as the default
// rendering hook. The media type is used to generate
// the OpenAPI specification.
func SetRenderHook(rh RenderHook, mediaType string) {
	if rh != nil {
		renderHook = rh
	}
	if mediaType == "" {
		defaultMediaType = "*/*"
	}
}
