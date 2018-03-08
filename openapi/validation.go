package openapi

import "reflect"

// setSchemaLen sets the given len to the appropriate
// schema field based on the given type.
func setSchemaLen(schema *Schema, len int, t reflect.Type) {
	if isNumber(t) {
		schema.Minimum = len
		schema.Maximum = len
	} else if isString(t) {
		if len >= 0 {
			schema.MinLength = len
			schema.MaxLength = len
		}
	} else if isMap(t) {
		if len >= 0 {
			schema.MinProperties = len
			schema.MaxProperties = len
		}
	} else if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		if len >= 0 {
			schema.MinItems = len
			schema.MaxItems = len
		}
	}
}

// setSchemaMax sets the given maximum to the appropriate
// schema field based on the given type.
func setSchemaMax(schema *Schema, max int, t reflect.Type) {
	if isNumber(t) {
		schema.Maximum = max
	} else if isString(t) {
		if max >= 0 {
			schema.MaxLength = max
		}
	} else if isMap(t) {
		if max >= 0 {
			schema.MaxProperties = max
		}
	} else if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		if max >= 0 {
			schema.MaxItems = max
		}
	}
}

// setSchemaMin sets the given minimum to the appropriate
// schema field based on the given type.
func setSchemaMin(schema *Schema, min int, t reflect.Type) {
	if isNumber(t) {
		schema.Minimum = min
	} else if isString(t) {
		if min >= 0 {
			schema.MinLength = min
		}
	} else if isMap(t) {
		if min >= 0 {
			schema.MinProperties = min
		}
	} else if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		if min >= 0 {
			schema.MinItems = min
		}
	}
}

// setSchemaEq sets the given equals value to the appropriate
// schema field based on the given type.
func setSchemaEq(schema *Schema, eq int, t reflect.Type) {
	// For numbers and strings, equals tag would translate
	// to the `const` property of the JSON Validation spec
	// but OpenAPI doesn't support it.
	if isMap(t) {
		if eq >= 0 {
			schema.MinProperties = eq
			schema.MaxProperties = eq
		}
	} else if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		if eq >= 0 {
			schema.MinItems = eq
			schema.MaxItems = eq
		}
	}
}

// isString returns whether the given reflect type represents a string.
func isString(typ reflect.Type) bool { return typ.Kind() == reflect.String }

// isMap returns whether the given reflect type represents a string.
func isMap(typ reflect.Type) bool { return typ.Kind() == reflect.Map }

// isNumber returns whether the given reflect type
// represents a number.
func isNumber(typ reflect.Type) bool {
	switch typ.Kind() {
	case
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64:
		return true
	}
	return false
}
