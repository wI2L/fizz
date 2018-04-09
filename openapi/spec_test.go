package openapi

import (
	"reflect"
	"testing"
)

// TestYAMLMarshalingRefs tests that spec types
// that contains embedded references  are properly
// marshaled to YAML.
func TestYAMLMarshalingRefs(t *testing.T) {
	// tests for references.
	tests := []interface{}{
		&ParameterOrRef{Reference: &Reference{Ref: ""}},
		&SchemaOrRef{Reference: &Reference{Ref: ""}},
		&ResponseOrRef{Reference: &Reference{Ref: ""}},
		&HeaderOrRef{Reference: &Reference{Ref: ""}},
		&MediaTypeOrRef{Reference: &Reference{Ref: ""}},
		&ExampleOrRef{Reference: &Reference{Ref: ""}},
	}
	for _, i := range tests {
		values := reflect.ValueOf(i).MethodByName("MarshalYAML").Call(nil)

		if len(values) != 2 {
			t.Errorf("expected MarshalYAML to return 2 args, got %d", len(values))
		}
		ret := values[0]

		if !ret.CanInterface() {
			t.Error("cannot get interface for returned type")
		}
		if _, ok := ret.Interface().(*Reference); !ok {
			t.Error("returned type is not a reference")
		}
	}
}

// TestYAMLMarshalingTypes tests that spec types
// that contains embedded types  are properly
// marshaled to YAML.
func TestYAMLMarshalingTypes(t *testing.T) {
	// tests for types.
	tests := map[interface{}]interface{}{
		&ParameterOrRef{Parameter: new(Parameter)}: &Parameter{},
		&SchemaOrRef{Schema: new(Schema)}:          &Schema{},
		&ResponseOrRef{Response: new(Response)}:    &Response{},
		&HeaderOrRef{Header: new(Header)}:          &Header{},
		&MediaTypeOrRef{MediaType: new(MediaType)}: &MediaType{},
		&ExampleOrRef{Example: new(Example)}:       &Example{},
	}
	for i, e := range tests {
		values := reflect.ValueOf(i).MethodByName("MarshalYAML").Call(nil)

		if len(values) != 2 {
			t.Errorf("expected MarshalYAML to return 2 args, got %d", len(values))
		}
		ret := values[0]

		if ret.Type().Kind() != reflect.Interface {
			t.Error("cannot get underlying type of non interface")
		}
		uret := ret.Elem()

		if uret.Type() != reflect.TypeOf(e) {
			t.Errorf("expected type to be %s, got %s", reflect.TypeOf(e).String(), uret.Type().String())
		}
	}
}
