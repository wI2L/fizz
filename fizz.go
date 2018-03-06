package fizz

import (
	"fmt"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/wI2L/fizz/openapi"
	"github.com/wI2L/fizz/tonic"
)

// Fizz is an abstraction of a Gin engine that wraps the
// routes handlers with Tonic and generates an OpenAPI
// 3.0 specification from it.
type Fizz struct {
	spec   *openapi.Spec
	engine *gin.Engine
	*RouterGroup
}

// RouterGroup is an abstraction of a Gin router group.
type RouterGroup struct {
	group       *gin.RouterGroup
	spec        *openapi.Spec
	Name        string
	Description string
}

// New creates a new Fizz wrapper for
// a default Gin engine.
func New() *Fizz {
	return NewFromEngine(gin.New())
}

// NewFromEngine creates a new Fizz wrapper
// from an existing Gin engine.
func NewFromEngine(e *gin.Engine) *Fizz {
	spec := openapi.NewSpec()

	return &Fizz{
		engine: e,
		spec:   spec,
		RouterGroup: &RouterGroup{
			group: &e.RouterGroup,
			spec:  spec,
		},
	}
}

// Router returns the Gin underlying engine.
func (f *Fizz) Router() *gin.Engine {
	return f.engine
}

// Group creates a new group of routes.
func (g *RouterGroup) Group(path, name, description string, handlers ...gin.HandlerFunc) *RouterGroup {
	// Create the tag in the specification
	// for this groups.
	g.spec.AddTag(name, description)

	return &RouterGroup{
		spec:        g.spec,
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
func (g *RouterGroup) GET(path string, h interface{}, infos ...func(*tonic.OperationInfo)) *RouterGroup {
	return g.Handle(path, "GET", h, infos...)
}

// POST is a shortcut to register a new handler with the POST method.
func (g *RouterGroup) POST(path string, h interface{}, infos ...func(*tonic.OperationInfo)) *RouterGroup {
	return g.Handle(path, "POST", h, infos...)
}

// PUT is a shortcut to register a new handler with the PUT method.
func (g *RouterGroup) PUT(path string, h interface{}, infos ...func(*tonic.OperationInfo)) *RouterGroup {
	return g.Handle(path, "PUT", h, infos...)
}

// PATCH is a shortcut to register a new handler with the PATCH method.
func (g *RouterGroup) PATCH(path string, h interface{}, infos ...func(*tonic.OperationInfo)) *RouterGroup {
	return g.Handle(path, "PATCH", h, infos...)
}

// DELETE is a shortcut to register a new handler with the DELETE method.
func (g *RouterGroup) DELETE(path string, h interface{}, infos ...func(*tonic.OperationInfo)) *RouterGroup {
	return g.Handle(path, "DELETE", h, infos...)
}

// OPTIONS is a shortcut to register a new handler with the OPTIONS method.
func (g *RouterGroup) OPTIONS(path string, h interface{}, infos ...func(*tonic.OperationInfo)) *RouterGroup {
	return g.Handle(path, "OPTIONS", h, infos...)
}

// HEAD is a shortcut to register a new handler with the HEAD method.
func (g *RouterGroup) HEAD(path string, h interface{}, infos ...func(*tonic.OperationInfo)) *RouterGroup {
	return g.Handle(path, "HEAD", h, infos...)
}

// Handle registers a new request handler that is wrapped
// with Tonic and documented in the OpenAPI specification.
func (g *RouterGroup) Handle(path, method string, h interface{}, infos ...func(*tonic.OperationInfo)) *RouterGroup {
	oi := &tonic.OperationInfo{}
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
	hfunc, hinfo := tonic.Handler(h, oi.StatusCode)
	g.group.Handle(method, path, hfunc)

	// Add operation to the OpenAPI spec.
	err := g.spec.AddOperation(
		joinPaths(g.group.BasePath(), path),
		method, hinfo, oi, g.Name,
	)
	if err != nil {
		panic(fmt.Sprintf(
			"error while creating OpenAPI spec on operation %s %s: %s",
			method, path, err,
		))
	}
	return g
}

// OpenAPI returns a Gin HandlerFunc that serves
// the marshalled OpenAPI specification of the API.
func (f *Fizz) OpenAPI(info *openapi.Info, ct string) gin.HandlerFunc {
	f.spec.SetInfo(info)

	if ct == "" {
		ct = "json"
	}
	switch ct {
	case "json":
		return func(c *gin.Context) {
			b, err := f.spec.JSON()
			if err != nil {
				c.Error(err)
			}
			c.Data(200, "application/json", b)
		}
	case "yaml":
		return func(c *gin.Context) {
			b, err := f.spec.YAML()
			if err != nil {
				c.Error(err)
			}
			c.Data(200, "application/yaml", b)
		}
	}
	return nil
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
