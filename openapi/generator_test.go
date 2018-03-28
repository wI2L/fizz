package openapi

import (
	"reflect"
	"testing"

	"github.com/loopfz/gadgeto/tonic"
	"github.com/stretchr/testify/assert"
)

var genConfig = &SpecGenConfig{
	ValidatorTag:      tonic.ValidationTag,
	PathLocationTag:   tonic.PathTag,
	QueryLocationTag:  tonic.QueryTag,
	HeaderLocationTag: tonic.HeaderTag,
	EnumTag:           tonic.EnumTag,
	DefaultTag:        tonic.DefaultTag,
}

// TestSchemaFromPrimitiveType tests that a schema
// can be created given a primitive input type.
func TestSchemaFromPrimitiveType(t *testing.T) {
	rt := reflect.TypeOf

	g, err := NewGenerator(genConfig)
	if err != nil {
		t.Error(err)
	}
	// Use a pointer to primitive type to test
	// pointer dereference and property nullable.
	schema := g.newSchemaFromType(rt(new(int64)))

	// Ensure it is an inlined schema before
	// accessing properties for assertions.
	if schema.Schema == nil {
		t.Error("expected an inlined schema, got a schema reference")
	}
	assert.Equal(t, "integer", schema.Type)
	assert.Equal(t, "int64", schema.Format)
	assert.True(t, schema.Nullable)
}

// TestSchemaFromUnsupportedType tests that a schema
// cannot be created given an unsupported input type.
func TestSchemaFromUnsupportedType(t *testing.T) {
	rt := reflect.TypeOf

	g, err := NewGenerator(genConfig)
	if err != nil {
		t.Error(err)
	}
	// Test with nil input.
	schema := g.newSchemaFromType(nil)
	assert.Nil(t, schema)

	// Test with unsupported input.
	schema = g.newSchemaFromType(rt(func() {}))
	assert.Nil(t, schema)
	assert.Len(t, g.Errors(), 1)
}
