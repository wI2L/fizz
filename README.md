
<h1 align="center">Fizz</h1>
<p align="center"><img src="images/lemon.png" height="200px" width="auto" alt="Gin Fizz"></p><p align="center">Fizz is a wrapper for <strong>Gin</strong> based on <i>gadgeto/tonic</i>.</p>
<p align="center">It generates wrapping gin-compatible handlers that do all the repetitive work and wrap the call to your handlers. It can also generates an almost complete <strong>OpenAPI 3</strong> specification of your API.</p>
<p align="center"><br>
<a href="https://godoc.org/github.com/wI2L/fizz"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"></a> <a href="https://goreportcard.com/report/wI2L/fizz"><img src="https://goreportcard.com/badge/github.com/wI2L/fizz"></a> <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
<br>
</p>

---

## Routing

First of all, you need to create a Fizz instance. You can pass an existing Gin engine instance to `fizz.NewFromEngine()`, or use `fizz.New()` that will use a new default Gin engine.

```go
engine := gin.Default()
engine.Use(...) // register global middlewares

f := fizz.NewFromEngine(engine)
```

A Fizz instance abstracts the `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`, `HEAD`, and `Use` functions of a Gin engine that you use to declare routes and register middlewares.

The bare minimum to register a handler is:
```go
f.GET("/foo/bar", MyHandler, fizz.StatusCode(200))
```
Those functions takes variadics arguments using the option pattern to let you enrich the OpenAPI specification for each operation.

```go
// set the default status code used by the tonic render hook.
// Warning: this option is mandatory for tonic. It will panic
// if it's missing.
fizz.StatusCode(code int)

// set the default response description.
// a default text status created from the default status code
// will be used in case its missing.
fizz.StatusDescription(desc string)

// the sumamry of the operation.
fizz.Summary(summary string)

// the description of the operation.
fizz.Description(desc string)

// the ID of the operation.
// Must be a unique string used to identify the operation among
// all operations described in the API.
fizz.ID(id string)

// mark the operation as deprecated.
fizz.Deprecated(deprecated bool)

// add an additional response to the operation.
// model and header may be `nil`.
fizz.Response(code, desc string, model interface{}, headers []*ResponseHeader)

// add an additional header to the default response.
// model can be of any type, and may also be `nil`,
// in which case the string type will be used as default.
fizz.Header(name, desc string, model interface{})
```

To help you declare additional headers, predefined variables for Go primitives types that you can pass as the third argument of `fizz.Header()` are available.
```go
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
```

You can also creates subgroups of routes using the method `Group()`. Unlike Gin own function, the Fizz method takes two other arguments `name` and `description`. These parameters will be used to create an **OpenAPI** tag that will be applied to all the sub-routes registered to the group.

```go
grp := f.Group("/subpath", "MyGroup", "Group description", middlewares...)

// The Use() method can also be used on groups to
// register middlewares after their creation.
grp.Use(moreMiddlewares...)

grp.GET("/:id", MyGetHandler,
   fizz.StatusCode(200),
   fizz.Summary("Get a resource by its ID"),
   fizz.Response(400, "Bad request", nil, nil),
)

type MyFooBar struct {}

grp.POST("", MyPostHandler,
   fizz.StatusCode(201),
   fizz.StatusDescription("resource created"),
   fizz.Header("X-Custom-Header", "custom header", fizz.Integer),
   fizz.Header("X-Foo-Bar", "foobar", MyFooBar{})
)
```

Subgroups of subgroups can be created to an infinite depth, according yo your needs.

```go
foo := f.Group("/foo", "Foo", "Foo group")

// all routes registered on group bar will have
// a relative path starting with /foo/bar
bar := f.Group("/bar", "Bar", "Bar group")

// /foo/bar/{barID}
bar.GET("/:barID", MyBarHandler, fizz.StatusCode(200))
```

Finally, use the Fizz instance as the base handler of your HTTP server.
```go
srv := &http.Server{
   Addr:    ":4242",
   Handler: f,
}
srv.ListenAndServe()
```

## Tonic

The subpackage **tonic** handles path/query/header/body parameters binding in a single consolidated input object which allows you to remove all the boilerplate code that retrieves and tests the presence of various parameters.

### Handler signature

The handlers registered with Fizz are automatically wrapped with Tonic, and must follow a specific signature.
```go
func(*gin.Context, [input object ptr]) ([output object], error)
```
Input and output objects are both optional, as such, the minimal accepted signature is:
```go
func(*gin.Context) error
```

Output objects can be of any type, and will be marshalled to the desired media type.
Note that the input object MUST always be a pointer to a struct, or the tonic wrapping will panic at runtime.

If you want to register handlers that don't follow the Tonic signature, you can get the underlying Gin engine of a Fizz instance using the `Engine()` method. Be aware that the routes added this way won't be documented by the OpenAPI spec gnerator since it is tightly coupled with Tonic.

### Location tags

Tonic uses three struct tags to recognize the parameters it should bind from the input types of your objects:
- `path`: bind from the request path
- `query`: bind from the query string
- `header`: bind from the request header

The fields that doesn't use one those four tags will be considered as part of the request body.

The value of each struct tags represents the name of the field in each location, with options.
```go
type MyHandlerParams struct {
   ID  int64     `path:"id"`
   Foo string    `query:"foo"`
   Bar time.Time `header:"x-foo-bar"`
}
```

Tonic will automatically convert the value extracted from the location described by the tag to the appropriate type before binding.

**NOTE**: A path parameter is always required and will appear required in the spec regardless of the `validate` tag content.

### Additional tags

You can use additional tags. Some will be interpreted by Tonic, others will be exclusively used to enrich the **OpenAPI** specification.
- `default`: Tonic will bind this value if none was passed with the request. This should not be used if a field is also required. Read the [documentation](https://swagger.io/docs/specification/describing-parameters/) (section _Common Mistakes_) for more informations about this tag behaviour.
- `description`: describe the field in the spec.
- `deprecated`: indicates if the field is deprecated. accepted values are _1_, _t_, _T_, _TRUE_, _true_, _True_, _0_, _f_, _F_, _FALSE_. Invalid value are considered to be false.
- `enum`: a coma separated list of acceptable values for the parameter. Tonic will verify that the given value is one of those.
- `format`: override the format of the field in the spec. [documentation](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#dataTypeFormat).
- `validate`: field validation rules. [documentation](https://godoc.org/gopkg.in/go-playground/validator.v8).

### JSON/XML

The JSON/XML encoders usually omit a field that has the tag `"-"`. This behaviour is reproduced by the spec generator ; a field with this tag won't appear in the schema properties of the model.

In the following example, the field `Input` is used only for binding request body parameters and won't appear in the output encoding while `Output` will be marshaled but will not be used for parameters binding.
```go
type Model struct {
	Input  string `json:"-"`
	Output string `json:"output" binding:"-"`
}
```

### Request body

If you want to make a request body field mandatory, you can use the tag `validate:"required"`. The validator used by tonic will ensure that the field is present.
To be able to make a difference between a missing value and the zero value of a type, use a pointer.

To explicitly ignore a parameter from the request body, use the tag `binding:"-"`.

Not that the generator will ignore request body parameters for the operations with method `GET`, `DELETE` or `HEAD`.
   > GET, DELETE and HEAD are no longer allowed to have request body because it does not have defined semantics as per [RFC 7231](https://tools.ietf.org/html/rfc7231#section-4.3).
   _[Swagger Documentation](https://swagger.io/docs/specification/describing-request-body/)_

### Schema validation

The **OpenAPI** generator recognize some tags of the [go-playground/validator.v8](https://gopkg.in/go-playground/validator.v8) package and translate those to the [properties of the schema](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.1.md#properties) that are taken from the [JSON Schema definition](http://json-schema.org/latest/json-schema-validation.html#rfc.section.6).

The supported tags are: [len](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Length), [max](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Maximum), [min](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Mininum), [eq](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Equals), [gt](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Greater_Than), [gte](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Greater_Than_or_Equal), [lt](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Less_Than), [lte](https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Less_Than_or_Equal).

Based on the type of the field that carry the tag, the fields `maximum`, `minimum`, `minLength`, `maxLength`, `minIntems`, `maxItems`, `minProperties` and `maxProperties` of its **JSON Schema** will be filled accordingly.

## OpenAPI specification

You can serve the generated OpenAPI specification in either `JSON` or `YAML` format using the handler returned by the `fizz.OpenAPI()` method.

To enrich the specification, you can provide additional informations. Head to the [OpenAPI 3 spec](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#infoObject) for more informations about the API informations that you can specify, or take a look at the type `openapi.Info` in the file [_openapi/objects.go_](openapi/object.go#L25).

```go
infos := &openapi.Info{
   Title:       "Fruits Market",
   Description: `This is a sample Fruits market server.`,
   Version:     "1.0.0",
}

f.Engine().GET("/openapi.json", fizz.OpenAPI(infos, "json"))
f.Engine().GET("/openapi.yaml", fizz.OpenAPI(infos, "yaml"))
```
**NOTES**:
* The handler is not compliant with the Tonic signature, you **have to** use the underlying Gin engine of the Fizz instance to register it.
* The spec generator will never panic. However, it is strongly recommended to call `fizz.Errors()` to retrieve and handle the errors that may have occured during the generation before starting your API.

## Known limitations

- Since **OpenAPI** is based on **JSON Schema**, maps with keys that are not strings are not supported and will be ignored during the generation of the spec.
- The output types of your handlers are registered as components within the generated spec. By default, the name used for each component is composed of the package and type name concatenated using CamelCase, and does not contain the full import path. As such, please ensure that you don't use the same type name in two eponym package in your application. See the method `Generator.UseFullSchemaNames()` for more control over this behaviour.
- Recursive embedding of the same type is not supported, at any level of recursion. The generator will warn and skip the offending fields.
   ```go
   type A struct {
      Foo int
      *A   // ko, embedded and same type as parent
      A *A // ok, not embedded
      *B   // ok, different type
   }

   type B struct {
      Bar string
      *A // ko, type B is embedded in type A
      *C // ok, type C does not contains an embedded field of type A
   }

   type C struct {
      Baz bool
   }
   ```

## Examples

A simple runnable API is available in `examples/market`.
```shell
go build
./market
# Retrieve the specification marshaled in JSON.
curl -i http://localhost:4242/openapi.json
```

## Credits

Fizz is based on [gadgeto/tonic](https://github.com/loopfz/gadgeto/tree/master/tonic) and [gin-gonic/gin](https://github.com/gin-gonic/gin). :heart:

<p align="right"><img src="https://forthebadge.com/images/badges/built-with-swag.svg"></p>
