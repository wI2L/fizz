package tonic

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// MaxBodyBytes is the maximum allowed size of a request body in bytes.
const MaxBodyBytes = 256 * 1024

const (
	queryTag    = "query"
	pathTag     = "path"
	headerTag   = "header"
	enumTag     = "enum"
	requiredTag = "required"
	defaultTag  = "default"
)

var (
	errorHook  ErrorHook  = DefaultErrorHook
	bindHook   BindHook   = DefaultBindingHook
	renderHook RenderHook = DefaultRenderHook
)

// BindError is an error type returned when tonic fails
// to bind parameters, to differentiate from errors returned
// by the handlers.
type BindError struct {
	message string
	typ     reflect.Type
	field   string
}

// Error implements the builtin error interface for BindError.
func (be BindError) Error() string {
	return fmt.Sprintf(
		"binding error for field %s of type %s: %s",
		be.field,
		be.typ.Name(),
		be.message,
	)
}

// An extractorFunc extracts data from a gin context according to
// parameters specified in a field tag.
type extractor func(*gin.Context, string) (string, []string, error)

// extractQuery is an extractor tgat operated on the query
// parameters of a request.
func extractQuery(c *gin.Context, tag string) (string, []string, error) {
	name, required, err := ParseTagKey(tag)
	if err != nil {
		return "", nil, err
	}
	q := c.Request.URL.Query()[name]

	if required && len(q) == 0 {
		return "", nil, fmt.Errorf("missing query parameter: %s", name)
	}
	return name, q, nil
}

// extractPath is an extractor that operates on the path
// parameters of a request.
func extractPath(c *gin.Context, tag string) (string, []string, error) {
	name, required, err := ParseTagKey(tag)
	if err != nil {
		return "", nil, err
	}
	p := c.Param(name)
	if required && p == "" {
		return "", nil, fmt.Errorf("missing path parameter: %s", name)
	}
	return name, []string{p}, nil
}

// extractHeader is an extractor that operates on the headers
// of a request.
func extractHeader(c *gin.Context, tag string) (string, []string, error) {
	name, required, err := ParseTagKey(tag)
	if err != nil {
		return "", nil, err
	}
	header := c.GetHeader(name)

	if required && header == "" {
		return "", nil, fmt.Errorf("missing header parameter: %s", name)
	}
	return name, []string{header}, nil
}

// ParseTagKey parses the given struct tag key and return the
// name of the field and wether or not it is required.
func ParseTagKey(tag string) (string, bool, error) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return "", false, fmt.Errorf("empty tag")
	}
	name, options := parts[0], parts[1:]

	// Iterate through the tag options to
	// find the required key.
	var required bool
	for _, o := range options {
		o = strings.TrimSpace(o)
		if o == requiredTag {
			required = true
		}
	}
	return name, required, nil
}

// bindStringValue converts and bind the value s
// to the the reflected value v.
func bindStringValue(s string, v reflect.Value) error {
	// Ensure that the reflected value is unaddressable
	// and wasn't obtained by the use of an unexported
	// struct field, or calling a setter will panic.
	if !v.CanSet() {
		return fmt.Errorf("unaddressable value: %v", v)
	}
	i := v.Interface()

	// If the value implements the encoding.TextUnmarshaler
	// interface, bind the returned string representation.
	if unmarshaler, ok := i.(encoding.TextUnmarshaler); ok {
		if err := unmarshaler.UnmarshalText([]byte(s)); err != nil {
			return err
		}
		v.Set(reflect.Indirect(reflect.ValueOf(unmarshaler)))
		return nil
	}
	// Handle time.Duration.
	if _, ok := i.(time.Duration); ok {
		d, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(d))
	}
	// Switch over the kind of the reflected value
	// and convert the string to the proper type.
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetUint(i)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Float32, reflect.Float64:
		i, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetFloat(i)
	default:
		return fmt.Errorf("unsupported parameter type: %v", v.Kind())
	}
	return nil
}
