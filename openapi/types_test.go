package openapi

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

// TestInternalDataType tests that the internal
// data types have correct types and formats.
func TestInternalDataType(t *testing.T) {
	type formatType struct {
		Type   string
		Format string
	}
	tests := map[InternalDataType]formatType{
		TypeFloat:       {"number", "float"},
		TypeDouble:      {"number", "double"},
		TypeInteger:     {"integer", "int32"},
		TypeLong:        {"integer", "int64"},
		TypeString:      {"string", ""},
		TypeByte:        {"string", "byte"},
		TypeBinary:      {"string", "binary"},
		TypeDate:        {"string", "date"},
		TypeDateTime:    {"string", "date-time"},
		TypeDuration:    {"string", "duration"},
		TypePassword:    {"string", "password"},
		TypeComplex:     {"string", ""},
		TypeBoolean:     {"boolean", ""},
		TypeUnsupported: {},
	}
	for dt, ft := range tests {
		format, typ := dt.Format(), dt.Type()
		if format != ft.Format {
			t.Errorf("expected format %s for data type %s, got %s", ft.Format, dt.String(), format)
		}
		if typ != ft.Type {
			t.Errorf("expected type %s for data type %s, got %s", ft.Type, dt.String(), typ)
		}
	}
}

// TestPrimitiveGoTypeToDataType tests that a Go
// primitive type can be converted to a DataType..
func TestPrimitiveGoTypeToDataType(t *testing.T) {
	tests := map[interface{}]DataType{
		int(0):       TypeInteger,
		int8(1):      TypeInteger,
		int16(2):     TypeInteger,
		int32(3):     TypeInteger,
		int64(4):     TypeLong,
		new(int):     TypeInteger,
		new(int64):   TypeLong,
		uint(0):      TypeInteger,
		uint8(1):     TypeInteger,
		uint16(2):    TypeInteger,
		uint32(3):    TypeInteger,
		uint64(4):    TypeLong,
		new(uint):    TypeInteger,
		new(uint64):  TypeLong,
		float32(1.0): TypeFloat,
		float64(2.0): TypeDouble,
		new(float32): TypeFloat,
		new(float64): TypeDouble,
		true:         TypeBoolean,
		false:        TypeBoolean,
		new(bool):    TypeBoolean,
		"string1":    TypeString,
		`string2`:    TypeString,
		new(string):  TypeString,
	}
	for i, dt := range tests {
		tof := reflect.TypeOf(i)
		dtt := DataTypeFromType(tof)
		if dtt != dt {
			t.Errorf("expected data type %s for Go type %s, got %s", dt, tof.String(), dtt)
		}
	}
}

// TestComplexGoTypeToInternalDataType tests that a complex
// Go type can be converted to a InternalDataType.
func TestComplexGoTypeToDataType(t *testing.T) {
	tests := map[reflect.Type]InternalDataType{
		rt([]byte{}):                 TypeByte,
		rt(time.Now()):               TypeDateTime,
		rt(5 * time.Second):          TypeDuration,
		rt(struct{}{}):               TypeComplex,
		rt([]string{}):               TypeComplex,
		rt([6]string{}):              TypeComplex,
		rt([...]string{}):            TypeComplex,
		rt(map[string]string{}):      TypeComplex,
		rt(map[int]string{}):         TypeComplex,
		rt(func() {}):                TypeUnsupported,
		rt(make(chan int)):           TypeUnsupported,
		rt(uintptr(0xFEED)):          TypeUnsupported,
		rt(unsafe.Pointer(new(int))): TypeUnsupported,
		rt(complex(1.0, float32(0))): TypeUnsupported,
		rt(complex(2.0, float64(0))): TypeUnsupported,
	}
	for tof, dt := range tests {
		dtt := DataTypeFromType(tof)
		if dtt != dt {
			t.Errorf("expected data type %s for Go type %s, got %s", dt, tof.String(), dtt)
		}
	}
}

// TestInternalDataTypeStringer tests that the internal
// data types implements the Stringer interface.
func TestInternalDataTypeStringer(t *testing.T) {
	for _, dt := range []InternalDataType{
		TypeFloat,
		TypeDouble,
		TypeInteger,
		TypeLong,
		TypeString,
		TypeByte,
		TypeBinary,
		TypeDate,
		TypeDateTime,
		TypeDuration,
		TypePassword,
		TypeComplex,
		TypeBoolean,
		TypeUnsupported,
	} {
		s, ss := dt.String(), datatypes[dt]
		if s != ss {
			t.Errorf("expected %s, got %s", ss, s)
		}
	}
	// Invalid DataType constant.
	udt := InternalDataType(255).String()
	if udt != "" {
		t.Errorf("expected invalid data type to have no string representation, got %s", udt)
	}
}

type UUIDv4 struct{}

func (*UUIDv4) Format() string { return "uuid" }
func (*UUIDv4) Type() string   { return "string" }

// TestCustomDataType tests that a custom type
// that implements the DataType interface can be
// used to get a type and format.
func TestCustomDataType(t *testing.T) {
	uuid := &UUIDv4{}

	dt := DataTypeFromType(reflect.TypeOf(uuid))
	assert.NotNil(t, dt)

	if v, ok := dt.(DataType); ok {
		assert.Equal(t, "uuid", v.Format())
		assert.Equal(t, "string", v.Type())
	} else {
		t.Error("expected type to implements DataType interface")
	}
}

// TestStringToTimeType tests that a string can be
// converted to the type of a time.Time.
func TestStringToTimeType(t *testing.T) {
	now := time.Now()
	v, err := stringToType(now.Format(time.RFC3339), tofTime)
	if err != nil {
		t.Error(err)
	}
	vv, ok := v.(time.Time)
	if !ok {
		t.Errorf("expected converted value to be of type %T, got %T", now, v)
	}
	if vv.Equal(now) {
		t.Errorf("expected time to equal %s, got %s", now.String(), vv.String())
	}
}

// TestStringToDurationType tests that a string can be
// converted to the type of a time.Duration.
func TestStringToDurationType(t *testing.T) {
	dur := 60 * time.Minute
	v, err := stringToType(dur.String(), tofDuration)
	if err != nil {
		t.Error(err)
	}
	vv, ok := v.(time.Duration)
	if !ok {
		t.Errorf("expected converted value to be of type %T, got %T", dur, v)
	}
	if vv != dur {
		t.Errorf("expected duration to equal %s, got %s", dur.String(), vv.String())
	}
}

// TestStringToType tests that a string can be
// converted to the kind of a reflect.Type value.
func TestStringToType(t *testing.T) {
	tests := map[string]struct {
		t reflect.Type
		v interface{}
	}{
		"coucou":  {rt(""), "coucou"},
		"0":       {rt(int8(0)), int64(0)},
		"1":       {rt(int16(0)), int64(1)},
		"2":       {rt(int32(0)), int64(2)},
		"3":       {rt(int64(0)), int64(3)},
		"4":       {rt(uint8(4)), uint64(4)},
		"5":       {rt(uint16(5)), uint64(5)},
		"6":       {rt(uint32(6)), uint64(6)},
		"7":       {rt(uint64(7)), uint64(7)},
		"true":    {rt(true), true},
		"f":       {rt(false), false},
		"invalid": {rt(true), false},
		"8":       {rt(float32(8)), float64(8.0)},
		"9.1":     {rt(float64(9.1)), float64(9.1)},
	}
	for from, to := range tests {
		vv, err := stringToType(from, to.t)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(to.v, vv) {
			t.Errorf("expected %s to be converted to %s(%v), got %T(%v)", from, to.t.String(), to.v, vv, vv)
		}
	}
	// Expect error for unknown unsuported types
	// in conversion.
	_, err := stringToType("", reflect.TypeOf([]string{}))
	if err == nil {
		t.Errorf("expected error to be non-nil for unsupported type")
	}
}
