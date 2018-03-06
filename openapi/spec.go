package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.com/wI2L/fizz/tonic"
)

const (
	version      = "3.0.0"
	anyMediaType = "*/*"
)

// Spec represents an OpenAPI specification.
type Spec struct {
	api *OpenAPI
}

// NewSpec returns a new empty OpenAPI specification.
func NewSpec() *Spec {
	components := &Components{
		Schemas:    make(map[string]*SchemaOrRef),
		Responses:  make(map[string]*ReponseOrRef),
		Parameters: make(map[string]*ParameterOrRef),
		Headers:    make(map[string]*HeaderOrRef),
	}
	return &Spec{
		api: &OpenAPI{
			OpenAPI:    version,
			Info:       &Info{},
			Paths:      make(Paths),
			Components: components,
		},
	}
}

// SetInfo uses the given OpenAPI info for the
// current specification.
func (s *Spec) SetInfo(info *Info) {
	s.api.Info = info
}

// AddTag adds a new tag to the OpenAPI specification.
// If a tag already exists with the same name, it is
// overwritten.
func (s *Spec) AddTag(name, desc string) {
	// Search for an existing tag with the same,
	// and update its description before returning
	// if one is found.
	for _, tag := range s.api.Tags {
		if tag != nil {
			if tag.Name == name {
				tag.Description = desc
				return
			}
		}
	}
	// Add a new tag to the spec.
	s.api.Tags = append(s.api.Tags, &Tag{
		Name:        name,
		Description: desc,
	})
}

// JSON returns the JSON marshaling of the specification.
func (s *Spec) JSON() ([]byte, error) {
	return json.Marshal(s.api)
}

// YAML returns the YAML marshaling of the specification.
func (s *Spec) YAML() ([]byte, error) {
	return yaml.Marshal(s.api)
}

// AddOperation add a new operation to the OpenAPI specification
// using the method and path of the route and the tonic
// handler informations.
func (s *Spec) AddOperation(path, method string, hi *tonic.HandlerInfo, oi *tonic.OperationInfo, tag string) error {
	if hi == nil {
		return errors.New("no handler info")
	}
	it := hi.InputType()
	if it.Kind() == reflect.Ptr {
		it = it.Elem()
	}
	if it.Kind() != reflect.Struct {
		return errors.New("input type is not a struct")
	}
	path = rewritePath(path)

	// If a PathItem does not exists for this
	// path, create a new one.
	item, ok := s.api.Paths[path]
	if !ok {
		item = new(PathItem)
		s.api.Paths[path] = item
	}
	// Create a new operation and set it
	// to the according method of the PathItem.
	op := &Operation{
		ID:          hi.Name(),
		Summary:     oi.Summary,
		Description: oi.Description,
		Deprecated:  oi.Deprecated,
		Responses:   make(Responses),
	}
	if tag != "" {
		op.Tags = append(op.Tags, tag)
	}
	// Operations with methods GET/HEAD/DELETE cannot have a body.
	// Non parameters fields will be ignored.
	allowBody := method != http.MethodGet &&
		method != http.MethodHead &&
		method != http.MethodDelete

	if err := s.setOperationParams(op, it, allowBody); err != nil {
		return err
	}
	ot := hi.OutputType()
	if it.Kind() == reflect.Ptr {
		ot = ot.Elem()
	}
	// Generate the default response from the tonic
	// handler return type.
	if err := s.setOperationResponse(op, ot, strconv.Itoa(hi.Status), tonic.MediaType(), oi.StatusDescription, oi.Headers); err != nil {
		return err
	}
	// Generate additional responses from the operation
	// informations.
	for _, resp := range oi.Responses {
		if resp != nil {
			if err := s.setOperationResponse(op,
				reflect.TypeOf(resp.Model),
				resp.Code,
				tonic.MediaType(),
				resp.Description,
				resp.Headers,
			); err != nil {
				return err
			}
		}
	}
	setOperationBymethod(item, op, method)

	return nil
}

var ginPathParamRe = regexp.MustCompile(`\/:([^\/]*)`)

// rewritePath converts a Gin operation path that use
// colons and asterisks to declare path parameters, to
// an OpenAPI representation that use curly braces.
func rewritePath(path string) string {
	return ginPathParamRe.ReplaceAllString(path, "/{$1}")
}

// setOperationBymethod sets the operation op to the appropriate
// field of item according to the given method.
func setOperationBymethod(item *PathItem, op *Operation, method string) {
	switch method {
	case "GET":
		item.GET = op
	case "PUT":
		item.PUT = op
	case "POST":
		item.POST = op
	case "PATCH":
		item.PATCH = op
	case "HEAD":
		item.HEAD = op
	case "OPTIONS":
		item.OPTIONS = op
	case "TRACE":
		item.TRACE = op
	case "DELETE":
		item.DELETE = op
	}
}

// setOperationResponse adds a response to the operation that
// return the type t with the given media type and status code.
func (s *Spec) setOperationResponse(op *Operation, t reflect.Type, code, mt, desc string, headers []*tonic.ResponseHeader) error {
	if _, ok := op.Responses[code]; ok {
		// A response already exists for this code.
		return fmt.Errorf("response with code %s already exists", code)
	}
	if desc == "" && code != "default" {
		ci, err := strconv.Atoi(code)
		if err != nil {
			return err
		}
		desc = http.StatusText(ci)
	}
	r := &Response{
		Description: desc,
		Content:     make(map[string]*MediaTypeOrRef),
		Headers:     make(map[string]*HeaderOrRef),
	}
	// The response may have no content type specified,
	// in which case we don't assign a schema.
	schema := s.newSchemaFromType(t)
	if schema != nil {
		r.Content[mt] = &MediaTypeOrRef{MediaType: &MediaType{
			Schema: schema,
		}}
	}
	// Assign headers.
	for _, h := range headers {
		if h != nil {
			var sor *SchemaOrRef
			if h.Model == nil {
				// default to string if no type is given.
				sor = &SchemaOrRef{Schema: &Schema{Type: "string"}}
			} else {
				sor = s.newSchemaFromType(reflect.TypeOf(h.Model))
			}
			r.Headers[h.Name] = &HeaderOrRef{Header: &Header{
				Description: h.Description,
				Schema:      sor,
			}}
		}
	}
	op.Responses[code] = &ReponseOrRef{Response: r}

	return nil
}

// setOperationParams adds the fields of the struct type t
// to the given operation.
func (s *Spec) setOperationParams(op *Operation, t reflect.Type, allowBody bool) error {
	if t.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			// ignore unexported fields.
			continue
		}
		sft := sf.Type
		if sft.Kind() == reflect.Ptr {
			sft = sft.Elem()
		}
		// If the struct field is an embedded struct, we recursively
		// use its fields as operations params. This allow developers
		// to reuse their types by composing input models.
		if sf.Anonymous && sft.Kind() == reflect.Struct {
			if err := s.setOperationParams(op, sft, allowBody); err != nil {
				return err
			}
		} else {
			if err := s.addStructFieldToOperation(op, t, i, allowBody); err != nil {
				return err
			}
		}
	}
	// Sort operations parameters by location and name
	// in ascending order.
	paramsOrderedBy(paramyByLocation, paramyByName).Sort(op.Parameters)

	return nil
}

func paramyByName(p1, p2 *ParameterOrRef) bool {
	return p1.Name < p2.Name
}

func paramyByLocation(p1, p2 *ParameterOrRef) bool {
	return locationsOrder[p1.In] < locationsOrder[p2.In]
}

// addStructFieldToOperation add the struct field of the type
// t at index idx to the operation op. A field will be considered
// as a parameter if it has a valid location tag key, or it will
// be treated as part of the request body.
func (s *Spec) addStructFieldToOperation(op *Operation, t reflect.Type, idx int, allowBody bool) error {
	sf := t.Field(idx)

	param, err := s.newParameterFromField(idx, t)
	if err != nil {
		return err
	}
	if param != nil {
		op.Parameters = append(op.Parameters, &ParameterOrRef{
			Parameter: param,
		})
	} else {
		if !allowBody {
			return nil
		}
		// If binding is disabled for this field, don't
		// add it to the request body. This allow using
		// a model type as an operation input while also
		// omitting some fields that are computed by the
		// server.
		if sf.Tag.Get("binding") == "-" {
			return nil
		}
		// The field is not a parameter, add it to
		// the request body.
		if op.RequestBody == nil {
			op.RequestBody = &RequestBody{
				Content: make(map[string]*MediaType),
			}
		}
		// Select the corresponding media type for the
		// given field tag, or default to any type.
		mt := tonic.MediaType()
		if mt == "" {
			mt = anyMediaType
		}
		var schema *Schema

		// Create the media type if no fields
		// have been added yet.
		if _, ok := op.RequestBody.Content[mt]; !ok {
			schema = &Schema{
				Type:       "object",
				Properties: make(map[string]*SchemaOrRef),
			}
			op.RequestBody.Content[mt] = &MediaType{
				Schema: &SchemaOrRef{Schema: schema},
			}
		} else {
			schema = op.RequestBody.Content[mt].Schema.Schema
		}
		fname := fieldNameFromTag(sf, mediaTags[tonic.MediaType()])

		var required bool
		// The required property of a field is not part of its
		// own schema but specified in the parent schema.
		if fname != "" && isStructFieldRequired(sf) {
			required = true
			schema.Required = append(schema.Required, fname)
		}
		schema.Properties[fname] = s.newSchemaFromStructField(sf, required, fname, strings.Title(t.Name()))
	}
	return nil
}

// mediaTags maps media types to well-known
// struct tags sued for marshaling.
var mediaTags = map[string]string{
	"application/json": "json",
	"application/xml":  "xml",
}

// newParameterFromField create a new operation parameter
// from the struct field at index idx in type in. Only the
// parameters of type path, query, header or cookie are concerned.
func (s *Spec) newParameterFromField(idx int, t reflect.Type) (*Parameter, error) {
	field := t.Field(idx)

	location, err := paramLocation(field, t)
	if err != nil {
		return nil, err
	}
	// The parameter location is empty, return nil
	// to indicate that the field is not a parameter.
	if location == "" {
		return nil, nil
	}
	name, required, err := tonic.ParseTagKey(field.Tag.Get(location))
	if err != nil {
		return nil, err
	}
	// Path parameters are aways required.
	if location == "path" {
		required = true
	}
	deprecated, err := strconv.ParseBool(field.Tag.Get("deprecated"))
	if err == nil {
		// Consider invalid values as false.
		deprecated = false
	}
	p := &Parameter{
		Name:        name,
		In:          location,
		Description: field.Tag.Get("description"),
		Required:    required,
		Deprecated:  deprecated,
		Schema:      s.newSchemaFromStructField(field, required, name, strings.Title(t.Name())),
	}
	if field.Type.Kind() == reflect.Bool && location == "query" {
		p.AllowEmptyValue = true
	}
	return p, nil
}

var parameterLocations = []string{"path", "query", "header", "cookie"}

// paramLocation parses the tags of the struct field to extract
// the location of an operation parameter.
func paramLocation(f reflect.StructField, in reflect.Type) (string, error) {
	var c, p int

	has := func(name string, tag reflect.StructTag, i int) {
		if _, ok := tag.Lookup(name); ok {
			c++
			// save name position to extract
			// the value of the unique key.
			p = i
		}
	}
	// Count the number of keys that represents
	// a parameter location from the tag of the
	// struct field.
	for i, n := range parameterLocations {
		has(n, f.Tag, i)
	}
	if c == 0 {
		// This will be considered to be part
		// of the request body.
		return "", nil
	}
	if c > 1 {
		return "", fmt.Errorf("field %s of %s has conflicting parameter location", f.Name, in.Name())
	}
	return parameterLocations[p], nil
}

// newSchemaFromStructField returns a new Schema builded
// from the field's type and its tags.
func (s *Spec) newSchemaFromStructField(sf reflect.StructField, required bool, fname, pname string) *SchemaOrRef {
	sor := s.newSchemaFromType(sf.Type)
	if sor == nil {
		return nil
	}
	// Get the underlying schema, it may be a reference
	// to a component, and update its fields using the
	// informations in the struct field tags.
	schema := s.schemaFromComponents(sor)

	if schema == nil {
		return nil
	}
	// Default value.
	// See section 'Common Mistakes' at
	// https://swagger.io/docs/specification/describing-parameters/
	if d := sf.Tag.Get("default"); d != "" {
		if required {
			logrus.Warnf("field %s of type %s cannot be required and have a default value", fname, pname)
		} else {
			if v, err := stringToType(d, sf.Type); err != nil {
				logrus.Warnf(
					"default value %s of field %s in type %s cannot be converted to field's type: %s",
					d, fname, pname, err,
				)
			} else {
				schema.Default = v
			}
		}
	}
	// Enum.
	es := sf.Tag.Get("enum")
	if es != "" {
		values := strings.Split(es, ",")
		for _, val := range values {
			if v, err := stringToType(val, sf.Type); err != nil {
				logrus.Warnf(
					"enum value %s of field %s in type %s cannot be converted to field's type: %s",
					val, fname, pname, err,
				)
			} else {
				schema.Enum = append(schema.Enum, v)
			}
		}
	}
	// Field description.
	if desc, ok := sf.Tag.Lookup("description"); ok {
		schema.Description = desc
	}
	// Deprecated.
	deprecated, err := strconv.ParseBool(sf.Tag.Get("deprecated"))
	if err == nil {
		// Consider invalid values as false.
		deprecated = false
	}
	schema.Deprecated = deprecated
	// Allow overidding schema properties that were
	// auto inferred manually via tags.
	if t, ok := sf.Tag.Lookup("format"); ok {
		schema.Format = t
	}
	return sor
}

// newSchemaFromType creates a new OpenAPI schema from
// the given reflect type.
func (s *Spec) newSchemaFromType(t reflect.Type) *SchemaOrRef {
	// Dereference pointer.
	if t == nil {
		return nil
	}
	var nullable bool
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		nullable = true
	}
	dt := DataTypeFromGo(t)
	if dt == TypeUnsupported {
		logrus.Warnf("openapi/spec: encountered unsupported type %s", t.Kind().String())
		return nil
	}
	if dt == TypeUnknown {
		switch t.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return s.buildSchemaRecursive(t)
		case reflect.Struct:
			return s.newSchemaFromStruct(t)
		default:
			logrus.Warnf("openapi/spec: encountered unknown type %s", t.Kind().String())
			return nil
		}
	}
	schema := &Schema{
		Type:     dt.Type(),
		Format:   dt.Format(),
		Nullable: nullable,
	}
	return &SchemaOrRef{Schema: schema}
}

// buildSchemaRecursive recursively decomposes the complex
// type t into subsequent schemas.
func (s *Spec) buildSchemaRecursive(t reflect.Type) *SchemaOrRef {
	schema := &Schema{}

	switch t.Kind() {
	case reflect.Ptr:
		return s.buildSchemaRecursive(t.Elem())
	case reflect.Struct:
		return s.newSchemaFromStruct(t)
	case reflect.Map:
		// Map type is considered as a type "object"
		// and should declare underlying items type
		// in additional properties field.
		schema.Type = "object"

		// JSON Schema allow only strings as
		// object key.
		if t.Key().Kind() != reflect.String {
			logrus.Warnf("openapi/spec: encountered type Map with keys of unsupported type %s", t.Key().Kind().String())
			return &SchemaOrRef{Schema: schema}
		}
		schema.AdditionalProperties = s.buildSchemaRecursive(t.Elem())
	case reflect.Slice, reflect.Array:
		// Slice/Array types are considered as a type
		// "array" and should declare underlying items
		// type in items field.
		schema.Type = "array"

		// Go arrays have fixed size.
		if t.Kind() == reflect.Array {
			schema.MinItems = t.Len()
			schema.MaxItems = t.Len()
		}
		schema.Items = s.buildSchemaRecursive(t.Elem())
	default:
		dt := DataTypeFromGo(t)
		schema.Type, schema.Format = dt.Type(), dt.Format()
	}
	return &SchemaOrRef{Schema: schema}
}

// structSchema returns an OpenAPI schema that describe
// the Go struct represented by the type t.
func (s *Spec) newSchemaFromStruct(t reflect.Type) *SchemaOrRef {
	if t.Kind() != reflect.Struct {
		return nil
	}
	name := strings.Title(t.Name())

	// Return an existing schema reference from the
	// API components, if any.
	if name != "" {
		if _, ok := s.api.Components.Schemas[name]; ok {
			return &SchemaOrRef{Reference: &Reference{
				Ref: "#/components/schemas/" + name,
			}}
		}
	}
	schema := s.flattenStructSchema(t, &Schema{
		Properties: make(map[string]*SchemaOrRef),
	})
	schema.Type = "object"

	sor := &SchemaOrRef{Schema: schema}

	// If the type has a name, register the schema
	// with the API specification components and
	// return a relative reference.
	if name != "" {
		s.api.Components.Schemas[name] = sor

		return &SchemaOrRef{Reference: &Reference{
			Ref: "#/components/schemas/" + name,
		}}
	}
	// Return an inlined schema for types with no name.
	return sor
}

// flattenStructSchema recursively flatten the embedded
// fields of the struct type t to the given schema.
func (s *Spec) flattenStructSchema(t reflect.Type, schema *Schema) *Schema {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if f.PkgPath != "" {
			// ignore unexported fields.
			continue
		}
		ft := f.Type
		// Dereference pointer.
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if f.Anonymous && ft != t {
			if ft.Kind() == reflect.Struct {
				schema = s.flattenStructSchema(ft, schema)
				continue
			}
		}
		fname := fieldNameFromTag(f, mediaTags[tonic.MediaType()])

		var required bool
		// The required property of a field is not part of its
		// own schema but specified in the parent schema.
		if fname != "" && isStructFieldRequired(f) {
			required = true
			schema.Required = append(schema.Required, fname)
		}
		// If the type of the field is the same as the type
		// of its the parent, skip the schema generation to
		// avoid an infinite recursive hole.
		if ft == t {
			schema.Properties[fname] = &SchemaOrRef{Reference: &Reference{
				Ref: "#/components/schemas/" + strings.Title(t.Name()),
			}}
		} else {
			schema.Properties[fname] = s.newSchemaFromStructField(f, required, fname, strings.Title(t.Name()))
		}
	}
	return schema
}

// isStructFieldRequired returns whether a struct field
// is required. The information is read from the field
// tag 'binding'.
func isStructFieldRequired(sf reflect.StructField) bool {
	if t, ok := sf.Tag.Lookup("binding"); ok {
		options := strings.Split(t, ",")
		for _, o := range options {
			// As soon as we see a 'dive' or 'keys'
			// options, the following options won't
			// apply to the given field.
			if o == "dive" || o == "keys" {
				return false
			}
			if o == "required" {
				return true
			}
		}
	}
	return false
}

// schemaFromComponents returns either the inlined schema
// in s or the one referenced in the API components.
func (s *Spec) schemaFromComponents(schema *SchemaOrRef) *Schema {
	if schema.Schema != nil && schema.Reference == nil {
		return schema.Schema
	}
	if schema.Reference != nil {
		parts := strings.Split(schema.Reference.Ref, "/")
		if len(parts) == 4 {
			if parts[0] == "#" && // relative ref
				parts[1] == "components" &&
				parts[2] == "schemas" &&
				parts[3] != "" {
				return s.api.Components.Schemas[parts[3]].Schema
			}
		}
	}
	return nil
}

// fieldTagName returns the name of a struct field
// extracted from a serialization tag using its name..
func fieldNameFromTag(sf reflect.StructField, tagName string) string {
	v, ok := sf.Tag.Lookup(tagName)
	if !ok {
		return sf.Name
	}
	parts := strings.Split(strings.TrimSpace(v), ",")
	if len(parts) == 0 {
		return sf.Name
	}
	name := parts[0]
	if name == "" || name == "-" {
		return sf.Name
	}
	return name
}
