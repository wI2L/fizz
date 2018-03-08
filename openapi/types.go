package openapi

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// DataType represents a primitive type.
type DataType int

// Type constants.
const (
	TypeInteger DataType = iota
	TypeLong
	TypeFloat
	TypeDouble
	TypeString
	TypeByte
	TypeBinary
	TypeBoolean
	TypeDate
	TypeDateTime
	TypePassword

	TypeUnsupported

	// TypeComplex represents non-primitive types like
	// Go struct, for which a schema must be generated.
	TypeComplex
)

// Type returns the type corresponding to the DataType.
func (dt DataType) Type() string {
	if 0 <= dt && dt < DataType(len(types)) {
		return types[dt]
	}
	return ""
}

// Format returns the format corresponding to the DataType.
func (dt DataType) Format() string {
	if 0 <= dt && dt < DataType(len(formats)) {
		return formats[dt]
	}
	return ""
}

var (
	tofTime      = reflect.TypeOf(time.Time{})
	tofByteSlice = reflect.TypeOf([]byte{})
)

// DataTypeFromGo returns an OpenAPI data type
// from a Golang value.
func DataTypeFromGo(t reflect.Type) DataType {
	// Dereference any pointer.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.AssignableTo(tofTime) {
		return TypeDateTime
	}
	if t.AssignableTo(tofByteSlice) {
		return TypeByte
	}
	// Switch over primitive types.
	switch t.Kind() {
	case reflect.Int64, reflect.Uint64:
		return TypeLong
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return TypeInteger
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return TypeInteger
	case reflect.Float32:
		return TypeFloat
	case reflect.Float64:
		return TypeDouble
	case reflect.Bool:
		return TypeBoolean
	case reflect.String:
		return TypeString
		// Handle known unsupported types.
	case reflect.Func, reflect.Chan, reflect.Uintptr, reflect.UnsafePointer, reflect.Interface, reflect.Complex64, reflect.Complex128:
		return TypeUnsupported
	}
	// Complex types:
	// * reflect.Struct
	// * reflect.Map
	// * reflect.Slice
	// * reflect.Array
	return TypeComplex
}

func stringToType(val string, t reflect.Type) (interface{}, error) {
	switch t.Kind() {
	case reflect.Bool:
		return strconv.ParseBool(val)
	case reflect.String:
		return val, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(val, 10, t.Bits())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.ParseUint(val, 10, t.Bits())
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(val, t.Bits())
	}
	// Compare to.
	if t.AssignableTo(tofTime) {
		return time.Parse(time.RFC3339, val)
	}
	if t.AssignableTo(tofByteSlice) {
		return time.ParseDuration(val)
	}
	return nil, fmt.Errorf("unknown type %s", t.String())
}

var types = [...]string{
	TypeInteger:  "integer",
	TypeLong:     "integer",
	TypeFloat:    "number",
	TypeDouble:   "number",
	TypeString:   "string",
	TypeByte:     "string",
	TypeBinary:   "string",
	TypeBoolean:  "boolean",
	TypeDate:     "string",
	TypeDateTime: "string",
	TypePassword: "string",
	TypeComplex:  "string",
}

var formats = [...]string{
	TypeInteger:  "int32",
	TypeLong:     "int64",
	TypeFloat:    "float",
	TypeDouble:   "double",
	TypeString:   "",
	TypeByte:     "byte",
	TypeBinary:   "binary",
	TypeBoolean:  "",
	TypeDate:     "date",
	TypeDateTime: "date-time",
	TypePassword: "password",
	TypeComplex:  "",
}
