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

	"github.com/loopfz/gadgeto/tonic"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const (
	version        = "3.0.1"
	anyMediaType   = "*/*"
	formatTag      = "format"
	deprecatedTag  = "deprecated"
	descriptionTag = "description"
)

// mediaTags maps media types to well-known
// struct tags used for marshaling.
var mediaTags = map[string]string{
	"application/json": "json",
	"application/xml":  "xml",
}

// Generator is an OpenAPI 3 generator.
type Generator struct {
	api         *OpenAPI
	config      *SpecGenConfig
	schemaTypes map[reflect.Type]struct{}
	errors      []error
}

// NewSpec returns a new empty OpenAPI specification.
func NewSpec(conf *SpecGenConfig) (*Generator, error) {
	if conf == nil {
		return nil, errors.New("missing config")
	}
	components := &Components{
		Schemas:    make(map[string]*SchemaOrRef),
		Responses:  make(map[string]*ReponseOrRef),
		Parameters: make(map[string]*ParameterOrRef),
		Headers:    make(map[string]*HeaderOrRef),
	}
	return &Generator{
		config: conf,
		api: &OpenAPI{
			OpenAPI:    version,
			Info:       &Info{},
			Paths:      make(Paths),
			Components: components,
		},
		schemaTypes: make(map[reflect.Type]struct{}),
	}, nil
}

// SpecGenConfig represents the configuration
// of the spec generation.
type SpecGenConfig struct {
	// Name of the tag used by the validator.v9
	// package. This is used by the spec generator
	// to determine if a field is required.
	ValidatorTag string
	// Name of the tag that represents the path location.
	PathLocationTag string
	// Name of the tag that represents the query location.
	QueryLocationTag string
	// Name of the tag that represents the header location.
	HeaderLocationTag string
	// Name of the that that contains enum values.
	EnumTag string
	// Name of the tag that contains default value.
	DefaultTag string
}

// SetInfo uses the given OpenAPI info for the
// current specification.
func (g *Generator) SetInfo(info *Info) {
	g.api.Info = info
}

// Errors returns the errors thar occurred during
// the generation of the specification.
func (g *Generator) Errors() []error {
	return g.errors
}

// AddTag adds a new tag to the OpenAPI specification.
// If a tag already exists with the same name, it is
// overwritten.
func (g *Generator) AddTag(name, desc string) {
	// Search for an existing tag with the same,
	// and update its description before returning
	// if one is found.
	for _, tag := range g.api.Tags {
		if tag != nil {
			if tag.Name == name {
				tag.Description = desc
				return
			}
		}
	}
	// Add a new tag to the spec.
	g.api.Tags = append(g.api.Tags, &Tag{
		Name:        name,
		Description: desc,
	})
}

// JSON returns the JSON marshaling of the specification.
func (g *Generator) JSON() ([]byte, error) {
	return json.Marshal(g.api)
}

// YAML returns the YAML marshaling of the specification.
func (g *Generator) YAML() ([]byte, error) {
	return yaml.Marshal(g.api)
}

// AddOperation add a new operation to the OpenAPI specification
// using the method and path of the route and the tonic
// handler informations.
func (g *Generator) AddOperation(path, method, tag, id string, in, out reflect.Type, info *OperationInfo) error {
	path = rewritePath(path)

	// If a PathItem does not exists for this
	// path, create a new one.
	item, ok := g.api.Paths[path]
	if !ok {
		item = new(PathItem)
		g.api.Paths[path] = item
	}
	// Create a new operation and set it
	// to the according method of the PathItem.
	op := &Operation{
		ID:          id,
		Summary:     info.Summary,
		Description: info.Description,
		Deprecated:  info.Deprecated,
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

	if in != nil {
		if in.Kind() == reflect.Ptr {
			in = in.Elem()
		}
		if in.Kind() != reflect.Struct {
			return errors.New("input type is not a struct")
		}
		if err := g.setOperationParams(op, in, in, allowBody); err != nil {
			return err
		}
	}
	if out != nil {
		if out.Kind() == reflect.Ptr {
			out = out.Elem()
		}
		// Generate the default response from the tonic
		// handler return type.
		if err := g.setOperationResponse(op, out, strconv.Itoa(info.StatusCode), tonic.MediaType(), info.StatusDescription, info.Headers); err != nil {
			return err
		}
	}
	// Generate additional responses from the operation
	// informations.
	for _, resp := range info.Responses {
		if resp != nil {
			if err := g.setOperationResponse(op,
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
func (g *Generator) setOperationResponse(op *Operation, t reflect.Type, code, mt, desc string, headers []*ResponseHeader) error {
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
	schema := g.newSchemaFromType(t)
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
				sor = g.newSchemaFromType(reflect.TypeOf(h.Model))
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
func (g *Generator) setOperationParams(op *Operation, t, parent reflect.Type, allowBody bool) error {
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
		// to reuse input models using type composition.
		if sf.Anonymous && sft.Kind() == reflect.Struct {
			// If the type of the embedded field is the same as
			// the topmost parent, skip it to avoid an infinite
			// recursive loop.
			if sft == parent {
				g.error(
					"skipped recursive embeding of type %s for parameter %s",
					g.typeName(parent), sf.Name,
				)
			} else if err := g.setOperationParams(op, sft, parent, allowBody); err != nil {
				return err
			}
		} else {
			if err := g.addStructFieldToOperation(op, t, i, allowBody); err != nil {
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
func (g *Generator) addStructFieldToOperation(op *Operation, t reflect.Type, idx int, allowBody bool) error {
	sf := t.Field(idx)

	param, err := g.newParameterFromField(idx, t)
	if err != nil {
		return err
	}
	if param != nil {
		// Check if a parameter with the same name for the
		// same location already exists.
		for _, p := range op.Parameters {
			if p != nil && (p.Name == param.Name) && (p.In == param.In) {
				g.error(
					"openapi/gen: duplicate parameter found in type %s: name=%s, location=%s",
					g.typeName(t), param.Name, param.In,
				)
				return nil
			}
		}
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

		// Check if a field with the same name already exists.
		if _, ok := schema.Properties[fname]; ok {
			g.error(
				"openapi/gen: duplicate request body parameter %s found in type %s",
				fname, g.typeName(t),
			)
			return nil
		}

		var required bool
		// The required property of a field is not part of its
		// own schema but specified in the parent schema.
		if fname != "" && g.isStructFieldRequired(sf) {
			required = true
			schema.Required = append(schema.Required, fname)
		}
		schema.Properties[fname] = g.newSchemaFromStructField(sf, required, fname, g.typeName(t))
	}
	return nil
}

// newParameterFromField create a new operation parameter
// from the struct field at index idx in type in. Only the
// parameters of type path, query, header or cookie are concerned.
func (g *Generator) newParameterFromField(idx int, t reflect.Type) (*Parameter, error) {
	field := t.Field(idx)

	location, err := g.paramLocation(field, t)
	if err != nil {
		return nil, err
	}
	// The parameter location is empty, return nil
	// to indicate that the field is not a parameter.
	if location == "" {
		return nil, nil
	}
	name, err := tonic.ParseTagKey(field.Tag.Get(location))
	if err != nil {
		return nil, err
	}
	required := g.isStructFieldRequired(field)

	// Path parameters are aways required.
	if location == g.config.PathLocationTag {
		required = true
	}
	deprecated, err := strconv.ParseBool(field.Tag.Get(deprecatedTag))
	if err == nil {
		// Consider invalid values as false.
		deprecated = false
	}
	p := &Parameter{
		Name:        name,
		In:          location,
		Description: field.Tag.Get(descriptionTag),
		Required:    required,
		Deprecated:  deprecated,
		Schema:      g.newSchemaFromStructField(field, required, name, g.typeName(t)),
	}
	if field.Type.Kind() == reflect.Bool && location == g.config.QueryLocationTag {
		p.AllowEmptyValue = true
	}
	return p, nil
}

// paramLocation parses the tags of the struct field to extract
// the location of an operation parameter.
func (g *Generator) paramLocation(f reflect.StructField, in reflect.Type) (string, error) {
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
	var parameterLocations = []string{
		g.config.PathLocationTag,
		g.config.QueryLocationTag,
		g.config.HeaderLocationTag,
	}
	for i, n := range parameterLocations {
		has(n, f.Tag, i)
	}
	if c == 0 {
		// This will be considered to be part
		// of the request body.
		return "", nil
	}
	if c > 1 {
		return "", fmt.Errorf("field %s of %s has conflicting parameter location", f.Name, g.typeName(in))
	}
	return parameterLocations[p], nil
}

// newSchemaFromStructField returns a new Schema builded
// from the field's type and its tags.
func (g *Generator) newSchemaFromStructField(sf reflect.StructField, required bool, fname, pname string) *SchemaOrRef {
	sor := g.newSchemaFromType(sf.Type)
	if sor == nil {
		return nil
	}
	// Get the underlying schema, it may be a reference
	// to a component, and update its fields using the
	// informations in the struct field tags.
	schema := g.schemaFromComponents(sor)

	if schema == nil {
		return nil
	}
	// Default value.
	// See section 'Common Mistakes' at
	// https://swagger.io/docs/specification/describing-parameters/
	if d := sf.Tag.Get(g.config.DefaultTag); d != "" {
		if required {
			g.error("openapi/gen: field %s of type %s cannot be required and have a default value", fname, pname)
		} else {
			if v, err := stringToType(d, sf.Type); err != nil {
				g.error(
					"openapi/gen: default value %s of field %s in type %s cannot be converted to field's type: %s",
					d, fname, pname, err,
				)
			} else {
				schema.Default = v
			}
		}
	}
	// Enum.
	es := sf.Tag.Get(g.config.EnumTag)
	if es != "" {
		values := strings.Split(es, ",")
		for _, val := range values {
			if v, err := stringToType(val, sf.Type); err != nil {
				g.error(
					"openapi/gen: enum value %s of field %s in type %s cannot be converted to field's type: %s",
					val, fname, pname, err,
				)
			} else {
				schema.Enum = append(schema.Enum, v)
			}
		}
	}
	// Field description.
	if desc, ok := sf.Tag.Lookup(descriptionTag); ok {
		schema.Description = desc
	}
	// Deprecated.
	deprecated, err := strconv.ParseBool(sf.Tag.Get(deprecatedTag))
	if err == nil {
		// Consider invalid values as false.
		deprecated = false
	}
	schema.Deprecated = deprecated

	// Update schema fields related to the JSON Validation
	// spec based on the content of the validator tag.
	schema = g.updateSchemaValidation(schema, sf)

	// Allow overidding schema properties that were
	// auto inferred manually via tags.
	if t, ok := sf.Tag.Lookup(formatTag); ok {
		schema.Format = t
	}

	return sor
}

// newSchemaFromType creates a new OpenAPI schema from
// the given reflect type.
func (g *Generator) newSchemaFromType(t reflect.Type) *SchemaOrRef {
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
		g.error("openapi/gen: encountered unsupported type %s", t.Kind().String())
		return nil
	}
	if dt == TypeComplex {
		switch t.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return g.buildSchemaRecursive(t)
		case reflect.Struct:
			return g.newSchemaFromStruct(t)
		default:
			g.error("openapi/gen: encountered unknown type %s", t.Kind().String())
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
func (g *Generator) buildSchemaRecursive(t reflect.Type) *SchemaOrRef {
	schema := &Schema{}

	switch t.Kind() {
	case reflect.Ptr:
		return g.buildSchemaRecursive(t.Elem())
	case reflect.Struct:
		return g.newSchemaFromStruct(t)
	case reflect.Map:
		// Map type is considered as a type "object"
		// and should declare underlying items type
		// in additional properties field.
		schema.Type = "object"

		// JSON Schema allow only strings as
		// object key.
		if t.Key().Kind() != reflect.String {
			g.error("openapi/gen: encountered type Map with keys of unsupported type %s", t.Key().Kind().String())
			return &SchemaOrRef{Schema: schema}
		}
		schema.AdditionalProperties = g.buildSchemaRecursive(t.Elem())
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
		schema.Items = g.buildSchemaRecursive(t.Elem())
	default:
		dt := DataTypeFromGo(t)
		schema.Type, schema.Format = dt.Type(), dt.Format()
	}
	return &SchemaOrRef{Schema: schema}
}

// structSchema returns an OpenAPI schema that describe
// the Go struct represented by the type t.
func (g *Generator) newSchemaFromStruct(t reflect.Type) *SchemaOrRef {
	if t.Kind() != reflect.Struct {
		return nil
	}
	name := g.typeName(t)

	// If the type of the field has already been registered,
	// skip the schema generation to avoid a recursive loop.
	// We're not returning directly a reference from the components,
	// because there is no guarantee the generation is complete yet.
	if _, ok := g.schemaTypes[t]; ok {
		return &SchemaOrRef{Reference: &Reference{
			Ref: "#/components/schemas/" + g.typeName(t),
		}}
	}
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*SchemaOrRef),
	}
	// Register the type once before diving into
	// the recursive hole if it has a name. Anonymous
	// struct are all considered unique.
	if name != "" {
		g.schemaTypes[t] = struct{}{}
	}
	schema = g.flattenStructSchema(t, t, schema)

	sor := &SchemaOrRef{Schema: schema}

	// If the type has a name, register the schema with the
	// API specification components and return a relative reference.
	// Unnamed types, like anonymous structs, will always be inlined
	// in the specification.
	if name != "" {
		g.api.Components.Schemas[name] = sor

		return &SchemaOrRef{Reference: &Reference{
			Ref: "#/components/schemas/" + name,
		}}
	}
	// Return an inlined schema for types with no name.
	return sor
}

// flattenStructSchema recursively flatten the embedded
// fields of the struct type t to the given schema.
func (g *Generator) flattenStructSchema(t, parent reflect.Type, schema *Schema) *Schema {
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
		if f.Anonymous && ft.Kind() == reflect.Struct {
			// If the type of the embedded field is the same as
			// the topmost parent, skip it to avoid an infinite
			// recursive loop.
			if ft == parent {
				logrus.Warnf("openapi/gen: skipped recursive embeding of type %s", g.typeName(parent))
			} else {
				schema = g.flattenStructSchema(ft, parent, schema)
			}
			continue
		}
		fname := fieldNameFromTag(f, mediaTags[tonic.MediaType()])
		if fname == "" {
			// Field has no name,, skip it.
			return schema
		}
		var required bool
		// The required property of a field is not part of its
		// own schema but specified in the parent schema.
		if fname != "" && g.isStructFieldRequired(f) {
			required = true
			schema.Required = append(schema.Required, fname)
		}
		schema.Properties[fname] = g.newSchemaFromStructField(f, required, fname, g.typeName(t))
	}
	return schema
}

// isStructFieldRequired returns whether a struct field
// is required. The information is read from the field
// tag 'binding'.
func (g *Generator) isStructFieldRequired(sf reflect.StructField) bool {
	if t, ok := sf.Tag.Lookup(g.config.ValidatorTag); ok {
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
func (g *Generator) schemaFromComponents(schema *SchemaOrRef) *Schema {
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
				return g.api.Components.Schemas[parts[3]].Schema
			}
		}
	}
	return nil
}

// typeName returns the unique name of a type, which is
// the concatenation of the package name and the name
// of the given type, transformed to CamelCase without
// a dot separator between the two parts.
func (g *Generator) typeName(t reflect.Type) string {
	name := t.String() // package.name.
	sp := strings.Index(name, ".")

	pkg := name[:sp]
	// If the package is the main package, remove
	// the package part from the name.
	if pkg == "main" {
		pkg = ""
	}
	typ := name[sp+1:]

	return strings.Title(pkg) + strings.Title(typ)
}

// updateSchemaValidation fills the fields of the schema
// related to the JSON Schema Validation RFC based on the
// content of the validator tag.
// see https://godoc.org/gopkg.in/go-playground/validator.v8
func (g *Generator) updateSchemaValidation(schema *Schema, sf reflect.StructField) *Schema {
	ts := sf.Tag.Get(g.config.ValidatorTag)
	if ts == "" {
		return schema
	}
	ft := sf.Type
	if sf.Type.Kind() == reflect.Ptr {
		ft = sf.Type.Elem()
	}
	tags := strings.Split(ts, ",")

	for _, t := range tags {
		if t == "dive" || t == "keys" {
			break
		}
		// Tags can be joined together with an OR operator.
		parts := strings.Split(t, "|")

		for _, p := range parts {
			var k, v string
			// Split k/v pair using separator.
			sepIdx := strings.Index(p, "=")
			if sepIdx == -1 {
				k = p
			} else {
				k = p[:sepIdx]
				v = p[sepIdx+1:]
			}
			// Handle validators with value.
			switch k {
			case "len", "max", "min", "eq", "gt", "gte", "lt", "lte":
				n, err := strconv.Atoi(v)
				if err != nil {
					continue
				}
				switch k {
				case "len":
					setSchemaLen(schema, n, ft)
				case "max", "lte":
					setSchemaMax(schema, n, ft)
				case "min", "gte":
					setSchemaMin(schema, n, ft)
				case "lt":
					setSchemaMax(schema, n-1, ft)
				case "gt":
					setSchemaMin(schema, n+1, ft)
				case "eq":
					setSchemaEq(schema, n, ft)
				}
			}
		}
	}
	return schema
}

func (g *Generator) error(format string, a ...interface{}) {
	err := fmt.Errorf(format, a...)
	g.errors = append(g.errors, err)
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
	if name == "" {
		return sf.Name
	}
	if name == "-" {
		return ""
	}
	return name
}
