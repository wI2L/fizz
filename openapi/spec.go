package openapi

import "encoding/json"

// OpenAPI represents the root document object of
// an OpenAPI document.
type OpenAPI struct {
	OpenAPI    string                 `json:"openapi" yaml:"openapi"`
	Info       *Info                  `json:"info" yaml:"info"`
	Servers    []*Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths      Paths                  `json:"paths" yaml:"paths"`
	Components *Components            `json:"components,omitempty" yaml:"components,omitempty"`
	Tags       []*Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	Security   []*SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	XTagGroups []*XTagGroup           `json:"x-tagGroups,omitempty" yaml:"x-tagGroups,omitempty"`
}

// Components holds a set of reusable objects for different
// aspects of the specification.
type Components struct {
	Schemas         map[string]*SchemaOrRef         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses       map[string]*ResponseOrRef       `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      map[string]*ParameterOrRef      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples        map[string]*ExampleOrRef        `json:"examples,omitempty" yaml:"examples,omitempty"`
	Headers         map[string]*HeaderOrRef         `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes map[string]*SecuritySchemeOrRef `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

// Info represents the metadata of an API.
type Info struct {
	Title          string   `json:"title" yaml:"title"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string   `json:"version" yaml:"version"`
	XLogo          *XLogo   `json:"x-logo,omitempty" yaml:"x-logo,omitempty"`
}

// Contact represents the the contact informations
// exposed for an API.
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// License represents the license informations
// exposed for an API.
type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Server represents a server.
type Server struct {
	URL         string                     `json:"url" yaml:"url"`
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]*ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// ServerVariable represents a server variable for server
// URL template substitution.
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// Paths represents the relative paths to the individual
// endpoints and their operations.
type Paths map[string]*PathItem

// PathItem describes the operations available on a single
// API path.
type PathItem struct {
	Ref         string            `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string            `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	GET         *Operation        `json:"get,omitempty" yaml:"get,omitempty"`
	PUT         *Operation        `json:"put,omitempty" yaml:"put,omitempty"`
	POST        *Operation        `json:"post,omitempty" yaml:"post,omitempty"`
	DELETE      *Operation        `json:"delete,omitempty" yaml:"delete,omitempty"`
	OPTIONS     *Operation        `json:"options,omitempty" yaml:"options,omitempty"`
	HEAD        *Operation        `json:"head,omitempty" yaml:"head,omitempty"`
	PATCH       *Operation        `json:"patch,omitempty" yaml:"patch,omitempty"`
	TRACE       *Operation        `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers     []*Server         `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters  []*ParameterOrRef `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Reference is a simple object to allow referencing
// other components in the specification, internally and
// externally.
type Reference struct {
	Ref string `json:"$ref" yaml:"$ref"`
}

// Parameter describes a single operation parameter.
type Parameter struct {
	Name            string       `json:"name" yaml:"name"`
	In              string       `json:"in" yaml:"in"`
	Description     string       `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool         `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool         `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool         `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Schema          *SchemaOrRef `json:"schema,omitempty" yaml:"schema,omitempty"`
	Style           string       `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool         `json:"explode,omitempty" yaml:"explode,omitempty"`
}

// ParameterOrRef represents a Parameter that can be inlined
// or referenced in the API description.
type ParameterOrRef struct {
	*Parameter
	*Reference
}

// MarshalYAML implements yaml.Marshaler for ParameterOrRef.
func (por *ParameterOrRef) MarshalYAML() (interface{}, error) {
	if por.Parameter != nil {
		return por.Parameter, nil
	}
	return por.Reference, nil
}

// RequestBody represents a request body.
type RequestBody struct {
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]*MediaType `json:"content" yaml:"content"`
	Required    bool                  `json:"required,omitempty" yaml:"required,omitempty"`
}

// SchemaOrRef represents a Schema that can be inlined
// or referenced in the API description.
type SchemaOrRef struct {
	*Schema
	*Reference
}

// MarshalYAML implements yaml.Marshaler for SchemaOrRef.
func (sor *SchemaOrRef) MarshalYAML() (interface{}, error) {
	if sor.Schema != nil {
		return sor.Schema, nil
	}
	return sor.Reference, nil
}

// Schema represents the definition of input and output data
// types of the API.
type Schema struct {
	// The following properties are taken from the JSON Schema
	// definition but their definitions were adjusted to the
	// OpenAPI Specification.
	Type                 string                  `json:"type,omitempty" yaml:"type,omitempty"`
	AllOf                *SchemaOrRef            `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                *SchemaOrRef            `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                *SchemaOrRef            `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Items                *SchemaOrRef            `json:"items,omitempty" yaml:"items,omitempty"`
	Properties           map[string]*SchemaOrRef `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties *SchemaOrRef            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Description          string                  `json:"description,omitempty" yaml:"description,omitempty"`
	Format               string                  `json:"format,omitempty" yaml:"format,omitempty"`
	Default              interface{}             `json:"default,omitempty" yaml:"default,omitempty"`
	Example              interface{}             `json:"example,omitempty" yaml:"example,omitempty"`

	// The following properties are taken directly from the
	// JSON Schema definition and follow the same specifications
	Title            string        `json:"title,omitempty" yaml:"title,omitempty"`
	MultipleOf       int           `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum          int           `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum          int           `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength        int           `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength        int           `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern          string        `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MaxItems         int           `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems         int           `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems      bool          `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxProperties    int           `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties    int           `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Required         []string      `json:"required,omitempty" yaml:"required,omitempty"`
	Enum             []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
	Nullable         bool          `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Deprecated       bool          `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

// Operation describes an API operation on a path.
type Operation struct {
	Tags         []string               `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary      string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	ID           string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   []*ParameterOrRef      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  *RequestBody           `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses    Responses              `json:"responses,omitempty" yaml:"responses,omitempty"`
	Deprecated   bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Servers      []*Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Security     []*SecurityRequirement `json:"security" yaml:"security"`
	XCodeSamples []*XCodeSample         `json:"x-codeSamples,omitempty" yaml:"x-codeSamples,omitempty"`
	XInternal    bool                   `json:"x-internal,omitempty" yaml:"x-internal,omitempty"`
}

// A workaround for missing omitnil functionality.
// Explicitely omit the Security field from marshaling when it is nil, but not when empty.
type operationNilOmitted struct {
	Tags         []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary      string            `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string            `json:"description,omitempty" yaml:"description,omitempty"`
	ID           string            `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   []*ParameterOrRef `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  *RequestBody      `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses    Responses         `json:"responses,omitempty" yaml:"responses,omitempty"`
	Deprecated   bool              `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Servers      []*Server         `json:"servers,omitempty" yaml:"servers,omitempty"`
	XCodeSamples []*XCodeSample    `json:"x-codeSamples,omitempty" yaml:"x-codeSamples,omitempty"`
	XInternal    bool              `json:"x-internal,omitempty" yaml:"x-internal,omitempty"`
}

// MarshalYAML implements yaml.Marshaler for Operation.
// Needed to marshall empty but non-null SecurityRequirements.
func (o *Operation) MarshalYAML() (interface{}, error) {
	if o.Security == nil {
		return omitOperationNilFields(o), nil
	}
	return o, nil
}

// MarshalJSON excludes empty but non-null SecurityRequirements.
func (o *Operation) MarshalJSON() ([]byte, error) {
	if o.Security == nil {
		return json.Marshal(omitOperationNilFields(o))
	}
	return json.Marshal(*o)
}

func omitOperationNilFields(o *Operation) *operationNilOmitted {
	return &operationNilOmitted{
		Tags:         o.Tags,
		Summary:      o.Summary,
		Description:  o.Description,
		ID:           o.ID,
		Parameters:   o.Parameters,
		RequestBody:  o.RequestBody,
		Responses:    o.Responses,
		Deprecated:   o.Deprecated,
		Servers:      o.Servers,
		XCodeSamples: o.XCodeSamples,
		XInternal:    o.XInternal,
	}
}

// Responses represents a container for the expected responses
// of an opration. It maps a HTTP response code to the expected
// response.
type Responses map[string]*ResponseOrRef

// ResponseOrRef represents a Response that can be inlined
// or referenced in the API description.
type ResponseOrRef struct {
	*Response
	*Reference
}

// MarshalYAML implements yaml.Marshaler for ResponseOrRef.
func (ror *ResponseOrRef) MarshalYAML() (interface{}, error) {
	if ror.Response != nil {
		return ror.Response, nil
	}
	return ror.Reference, nil
}

// Response describes a single response from an API.
type Response struct {
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Headers     map[string]*HeaderOrRef    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]*MediaTypeOrRef `json:"content,omitempty" yaml:"content,omitempty"`
}

// HeaderOrRef represents a Header that can be inlined
// or referenced in the API description.
type HeaderOrRef struct {
	*Header
	*Reference
}

// MarshalYAML implements yaml.Marshaler for HeaderOrRef.
func (hor *HeaderOrRef) MarshalYAML() (interface{}, error) {
	if hor.Header != nil {
		return hor.Header, nil
	}
	return hor.Reference, nil
}

// Header represents an HTTP header.
type Header struct {
	Description     string       `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool         `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool         `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool         `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Schema          *SchemaOrRef `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// MediaTypeOrRef represents a MediaType that can be inlined
// or referenced in the API description.
type MediaTypeOrRef struct {
	*MediaType
	*Reference
}

// MarshalYAML implements yaml.Marshaler for MediaTypeOrRef.
func (mtor *MediaTypeOrRef) MarshalYAML() (interface{}, error) {
	if mtor.MediaType != nil {
		return mtor.MediaType, nil
	}
	return mtor.Reference, nil
}

// MediaType represents the type of a media.
type MediaType struct {
	Schema   *SchemaOrRef             `json:"schema" yaml:"schema"`
	Example  interface{}              `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]*ExampleOrRef `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]*Encoding     `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// ExampleOrRef represents an Example that can be inlined
// or referenced in the API description.
type ExampleOrRef struct {
	*Example
	*Reference
}

// MarshalYAML implements yaml.Marshaler for ExampleOrRef.
func (eor *ExampleOrRef) MarshalYAML() (interface{}, error) {
	if eor.Example != nil {
		return eor.Example, nil
	}
	return eor.Reference, nil
}

// Example represents the example of a media type.
type Example struct {
	Summary       string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"`
	Value         interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Encoding represents a single encoding definition
// applied to a single schema property.
type Encoding struct {
	ContentType   string                  `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]*HeaderOrRef `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string                  `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       bool                    `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool                    `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

// Tag represents the metadata of a single tag.
type Tag struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// SecuritySchemeOrRef represents a SecurityScheme that can be inlined
// or referenced in the API description.
type SecuritySchemeOrRef struct {
	*SecurityScheme
	*Reference
}

// MarshalYAML implements yaml.Marshaler for SecuritySchemeOrRef.
func (sor *SecuritySchemeOrRef) MarshalYAML() (interface{}, error) {
	if sor.SecurityScheme != nil {
		return sor.SecurityScheme, nil
	}
	return sor.Reference, nil
}

// SecurityScheme represents a security scheme that can be used by an operation.
type SecurityScheme struct {
	Type             string      `json:"type,omitempty" yaml:"type,omitempty"`
	Scheme           string      `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Description      string      `json:"description,omitempty" yaml:"description,omitempty"`
	In               string      `json:"in,omitempty" yaml:"in,omitempty"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty" yaml:"flows,omitempty"`
}

// OAuthFlows represents all the supported OAuth flows.
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow represents an OAuth security scheme.
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

// MarshalYAML implements yaml.Marshaler for OAuthFlow.
func (f OAuthFlow) MarshalYAML() ([]byte, error) {
	type flow OAuthFlow
	if f.Scopes == nil {
		// The field is REQUIRED and MAY be empty according to the spec.
		f.Scopes = map[string]string{}
	}
	return json.Marshal(flow(f))
}

// SecurityRequirement represents the security object in the API specification.
type SecurityRequirement map[string][]string

// XLogo represents the information about the x-logo extension.
// See: https://github.com/Redocly/redoc/blob/master/docs/redoc-vendor-extensions.md#x-logo
type XLogo struct {
	URL             string `json:"url,omitempty" yaml:"url,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty" yaml:"backgroundColor,omitempty"`
	AltText         string `json:"altText,omitempty" yaml:"altText,omitempty"`
	Href            string `json:"href,omitempty" yaml:"href,omitempty"`
}

// XTagGroup represents the information about the x-tagGroups extension.
// See: https://github.com/Redocly/redoc/blob/master/docs/redoc-vendor-extensions.md#x-taggroups
type XTagGroup struct {
	Name string   `json:"name,omitempty" yaml:"name,omitempty"`
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// XCodeSample represents the information about the x-codeSample extension.
// See: https://github.com/Redocly/redoc/blob/master/docs/redoc-vendor-extensions.md#x-codesamples
type XCodeSample struct {
	Lang   string `json:"lang,omitempty" yaml:"lang,omitempty"`
	Label  string `json:"label,omitempty" yaml:"label,omitempty"`
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
}
