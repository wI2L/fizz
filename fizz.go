package fizz

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/wI2L/fizz/openapi"
)

// Primitive type helpers.
var (
	Integer  int32
	Long     int64
	Float    float32
	Double   float64
	String   string
	Byte     []byte
	Binary   []byte
	Boolean  bool
	DateTime time.Time
)

// Fizz is an abstraction of a Gin engine that wraps the
// routes handlers with Tonic and generates an OpenAPI
// 3.0 specification from it.
type Fizz struct {
	gen    *openapi.Generator
	engine *gin.Engine
	*RouterGroup
}

// RouterGroup is an abstraction of a Gin router group.
type RouterGroup struct {
	group       *gin.RouterGroup
	gen         *openapi.Generator
	Name        string
	Description string
}

// New creates a new Fizz wrapper for
// a default Gin engine.
func New() *Fizz {
	return NewFromEngine(gin.Default())
}

// NewFromEngine creates a new Fizz wrapper
// from an existing Gin engine.
func NewFromEngine(e *gin.Engine) *Fizz {
	// Create a new spec with the config
	// based on tonic internals.
	gen, _ := openapi.NewGenerator(
		&openapi.SpecGenConfig{
			ValidatorTag:      tonic.ValidationTag,
			PathLocationTag:   tonic.PathTag,
			QueryLocationTag:  tonic.QueryTag,
			HeaderLocationTag: tonic.HeaderTag,
			EnumTag:           tonic.EnumTag,
			DefaultTag:        tonic.DefaultTag,
		},
	)
	return &Fizz{
		engine: e,
		gen:    gen,
		RouterGroup: &RouterGroup{
			group: &e.RouterGroup,
			gen:   gen,
		},
	}
}

// ServeHTTP implements http.HandlerFunc for Fizz.
func (f *Fizz) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.engine.ServeHTTP(w, r)
}

// Engine returns the underlying Gin engine.
func (f *Fizz) Engine() *gin.Engine {
	return f.engine
}

// Generator returns the underlying OpenAPI generator.
func (f *Fizz) Generator() *openapi.Generator {
	return f.gen
}

// Errors returns the errors that may have occurred
// during the spec generation.
func (f *Fizz) Errors() []error {
	return f.gen.Errors()
}

// Group creates a new group of routes.
func (g *RouterGroup) Group(path, name, description string, handlers ...gin.HandlerFunc) *RouterGroup {
	// Create the tag in the specification
	// for this groups.
	g.gen.AddTag(name, description)

	return &RouterGroup{
		gen:         g.gen,
		group:       g.group.Group(path, handlers...),
		Name:        name,
		Description: description,
	}
}

// Use adds middleware to the group.
func (g *RouterGroup) Use(handlers ...gin.HandlerFunc) {
	g.group.Use(handlers...)
}

// GET is a shortcut to register a new handler with the GET method.
func (g *RouterGroup) GET(path string, h interface{}, infos ...func(*openapi.OperationInfo)) *RouterGroup {
	return g.Handle(path, "GET", h, infos...)
}

// POST is a shortcut to register a new handler with the POST method.
func (g *RouterGroup) POST(path string, h interface{}, infos ...func(*openapi.OperationInfo)) *RouterGroup {
	return g.Handle(path, "POST", h, infos...)
}

// PUT is a shortcut to register a new handler with the PUT method.
func (g *RouterGroup) PUT(path string, h interface{}, infos ...func(*openapi.OperationInfo)) *RouterGroup {
	return g.Handle(path, "PUT", h, infos...)
}

// PATCH is a shortcut to register a new handler with the PATCH method.
func (g *RouterGroup) PATCH(path string, h interface{}, infos ...func(*openapi.OperationInfo)) *RouterGroup {
	return g.Handle(path, "PATCH", h, infos...)
}

// DELETE is a shortcut to register a new handler with the DELETE method.
func (g *RouterGroup) DELETE(path string, h interface{}, infos ...func(*openapi.OperationInfo)) *RouterGroup {
	return g.Handle(path, "DELETE", h, infos...)
}

// OPTIONS is a shortcut to register a new handler with the OPTIONS method.
func (g *RouterGroup) OPTIONS(path string, h interface{}, infos ...func(*openapi.OperationInfo)) *RouterGroup {
	return g.Handle(path, "OPTIONS", h, infos...)
}

// HEAD is a shortcut to register a new handler with the HEAD method.
func (g *RouterGroup) HEAD(path string, h interface{}, infos ...func(*openapi.OperationInfo)) *RouterGroup {
	return g.Handle(path, "HEAD", h, infos...)
}

// Handle registers a new request handler that is wrapped
// with Tonic and documented in the OpenAPI specification.
func (g *RouterGroup) Handle(path, method string, h interface{}, infos ...func(*openapi.OperationInfo)) *RouterGroup {
	oi := &openapi.OperationInfo{}
	for _, info := range infos {
		info(oi)
	}
	if oi.StatusCode == 0 {
		panic(fmt.Sprintf(
			"error while adding %s operation for path %s, missing default status code",
			method, path,
		))
	}
	// Generate tonic-wrapped handler and register
	// it with the underlying Gin RouterGroup.
	hfunc := tonic.Handler(h, oi.StatusCode)

	// Retrieve handler informations processed
	// by tonic.
	route, err := tonic.GetRouteByHandler(hfunc)
	if err != nil {
		panic(fmt.Sprintf(
			"error while creating OpenAPI spec on operation %s %s: %s",
			method, path, err,
		))
	}
	// Register the handler with Gin underlying group.
	g.group.Handle(method, path, hfunc)

	// Consolidate path.
	path = joinPaths(g.group.BasePath(), path)

	// Set an operation ID if none is provided.
	if oi.ID == "" {
		oi.ID = route.HandlerName()
	}
	// Add operation to the OpenAPI spec.
	if err := g.gen.AddOperation(path, method, g.Name, route.InputType(), route.OutputType(), oi); err != nil {
		panic(fmt.Sprintf(
			"error while generating OpenAPI spec on operation %s %s: %s",
			method, path, err,
		))
	}
	return g
}

// OpenAPI returns a Gin HandlerFunc that serves
// the marshalled OpenAPI specification of the API.
func (f *Fizz) OpenAPI(info *openapi.Info, ct string) gin.HandlerFunc {
	f.gen.SetInfo(info)

	if ct == "" {
		ct = "json"
	}
	switch ct {
	case "json":
		return func(c *gin.Context) {
			b, err := f.gen.JSON()
			if err != nil {
				c.Error(err)
			}
			c.Data(200, "application/json", b)
		}
	case "yaml":
		return func(c *gin.Context) {
			b, err := f.gen.YAML()
			if err != nil {
				c.Error(err)
			}
			c.Data(200, "application/yaml", b)
		}
	}
	return nil
}

// StatusCode sets the default status code of the operation.
func StatusCode(code int) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.StatusCode = code
	}
}

// StatusDescription sets the default status description of the operation.
func StatusDescription(desc string) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.StatusDescription = desc
	}
}

// Summary adds a summary to an operation.
func Summary(summary string) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Summary = summary
	}
}

// Description adds a description to an operation.
func Description(desc string) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Description = desc
	}
}

// ID overrides the operation ID.
func ID(id string) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.ID = id
	}
}

// Deprecated marks the operation as deprecated.
func Deprecated(deprecated bool) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Deprecated = deprecated
	}
}

// Response adds an additional response to the operation.
func Response(code, desc string, model interface{}, headers []*openapi.ResponseHeader) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Responses = append(o.Responses, &openapi.OperationReponse{
			Code:        code,
			Description: desc,
			Model:       model,
			Headers:     headers,
		})
	}
}

// Header adds a header to the operation.
func Header(name, desc string, model interface{}) func(*openapi.OperationInfo) {
	return func(o *openapi.OperationInfo) {
		o.Headers = append(o.Headers, &openapi.ResponseHeader{
			Name:        name,
			Description: desc,
			Model:       model,
		})
	}
}

func joinPaths(abs, rel string) string {
	if rel == "" {
		return abs
	}
	final := path.Join(abs, rel)
	as := lastChar(rel) == '/' && lastChar(final) != '/'
	if as {
		return final + "/"
	}
	return final
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("empty string")
	}
	return str[len(str)-1]
}
