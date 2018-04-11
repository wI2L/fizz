package fizz

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/y0ssar1an/q"
	yaml "gopkg.in/yaml.v2"

	"github.com/wI2L/fizz/openapi"
)

func TestMain(m *testing.M) {
	// Don't print Gin debug in logs.
	gin.SetMode(gin.ReleaseMode)

	os.Exit(m.Run())
}

// TestInstance tests that a new Fizz
// instance can be created from scratch or
// from an existing Gin engine.
func TestInstance(t *testing.T) {
	fizz := New()
	assert.NotNil(t, fizz)

	engine := gin.Default()
	fizz = NewFromEngine(engine)
	assert.NotNil(t, fizz)

	assert.Equal(t, engine, fizz.Engine())
	assert.EqualValues(t, fizz.RouterGroup.group, &engine.RouterGroup)
	assert.NotNil(t, fizz.Generator())
	assert.Len(t, fizz.Errors(), 0)
}

// TestGroup tests that a router group can be created.
func TestGroup(t *testing.T) {
	engine := gin.New()
	fizz := NewFromEngine(engine)

	grp := fizz.Group("/test", "Test", "Test routes")
	assert.NotNil(t, grp)

	assert.Equal(t, grp.Name, "Test")
	assert.Equal(t, grp.Description, "Test routes")
	assert.Equal(t, grp.gen, fizz.gen)
	assert.NotNil(t, grp.group)
}

// TestHandler tests that handlers can be
// registered on the Fizz instance.
func TestHandler(t *testing.T) {
	fizz := New()

	rid := uuid.Must(uuid.NewV4())
	fizz.Use(func(c *gin.Context) {
		c.Header("X-Request-Id", rid.String())
	})

	wg := sync.WaitGroup{}
	h := func(c *gin.Context) {
		wg.Done()
	}
	fizz.GET("/", nil, h)
	fizz.POST("/", nil, h)
	fizz.PUT("/", nil, h)
	fizz.PATCH("/", nil, h)
	fizz.DELETE("/", nil, h)
	fizz.HEAD("/", nil, h)
	fizz.OPTIONS("/", nil, h)
	fizz.TRACE("/", nil, h)

	wg.Add(8)

	srv := httptest.NewServer(fizz)
	defer srv.Close()

	c := srv.Client()
	c.Timeout = 1 * time.Second

	for _, method := range []string{
		"GET",
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
		"HEAD",
		"OPTIONS",
		"TRACE",
	} {
		req, err := http.NewRequest(method, srv.URL, nil)
		if err != nil {
			t.Error(err)
		}
		resp, err := c.Do(req)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, rid.String(), resp.Header.Get("X-Request-Id"))
	}
	wg.Wait()
}

type T struct {
	X string `json:"x" description:"This is X"`
	Y int    `json:"y" description:"This is Y"`
}
type In struct {
	A int    `path:"a" description:"This is A"`
	B string `query:"b" description:"This is B"`
	C string `header:"X-Test-C" description:"This is C"`
}

// TestTonicHandler tests that a tonic-wrapped
// handler can be registered on a Fizz instance.
func TestTonicHandler(t *testing.T) {
	fizz := New()

	fizz.GET("/:a", nil, tonic.Handler(func(c *gin.Context, params *In) (*T, error) {
		assert.Equal(t, 0, params.A)
		assert.Equal(t, "foobar", params.B)
		assert.Equal(t, "foobaz", params.C)

		return &T{X: "foo", Y: 1}, nil
	}, 200))

	srv := httptest.NewServer(fizz)
	defer srv.Close()

	c := srv.Client()
	c.Timeout = 1 * time.Second

	url, err := url.Parse(srv.URL)
	if err != nil {
		t.Error(err)
	}
	url.Path = "/0"
	url.RawQuery = "b=foobar"

	resp, err := c.Do(&http.Request{
		URL:    url,
		Method: http.MethodGet,
		Header: http.Header{
			"X-Test-C": []string{"foobaz"},
		},
	})
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `{"x":"foo","y":1}`, string(body))
}

// TestSpecHandler tests that the OpenAPI handler
// return the spec properly marshaled in JSON.
func TestSpecHandler(t *testing.T) {
	fizz := New()

	fizz.GET("/test/:a",
		[]OperationOption{
			ID("GetTest"),
			Summary("Test"),
			Description("Test route"),
			StatusDescription("200"),
			StatusDescription("OK"),
			Deprecated(true),
			// Override summary and description
			// with printf-like options.
			Summaryf("Test-%s", "A"),
			Descriptionf("Test %s", "routes"),
			// Headers.
			Header("X-Request-Id", "Unique request ID", String),
			// Additional responses.
			Response("429", "", String, []*openapi.ResponseHeader{
				&openapi.ResponseHeader{
					Name:        "X-Rate-Limit",
					Description: "Rate limit",
					Model:       Integer,
				},
			}),
		},
		tonic.Handler(func(c *gin.Context) error {
			return nil
		}, 200),
	)
	infos := &openapi.Info{
		Title:       "Test Server",
		Description: `This is a test server.`,
		Version:     "1.0.0",
	}
	fizz.GET("/openapi.json", nil, fizz.OpenAPI(infos, "")) // default is JSON
	fizz.GET("/openapi.yaml", nil, fizz.OpenAPI(infos, "yaml"))

	srv := httptest.NewServer(fizz)
	defer srv.Close()

	c := srv.Client()
	c.Timeout = 1 * time.Second

	respJSON, err := c.Get(srv.URL + "/openapi.json")
	if err != nil {
		t.Error(err)
	}
	defer respJSON.Body.Close()

	assert.Equal(t, 200, respJSON.StatusCode)

	specJSON, err := ioutil.ReadAll(respJSON.Body)
	if err != nil {
		t.Error(err)
	}
	// see testdata/spec.json.
	expectedJSON, err := ioutil.ReadFile("testdata/spec.json")
	if err != nil {
		t.Error(err)
	}
	m, err := diffJSON(specJSON, expectedJSON)
	if err != nil {
		t.Error(err)
	}
	if !m {
		t.Error("invalid JSON spec")
	}

	respYAML, err := c.Get(srv.URL + "/openapi.yaml")
	if err != nil {
		t.Error(err)
	}
	defer respYAML.Body.Close()

	assert.Equal(t, 200, respYAML.StatusCode)

	specYAML, err := ioutil.ReadAll(respYAML.Body)
	if err != nil {
		t.Error(err)
	}
	// see testdata/spec.yaml.
	expectedYAML, err := ioutil.ReadFile("testdata/spec.yaml")
	if err != nil {
		t.Error(err)
	}
	m, err = diffYAML(specYAML, expectedYAML)
	if err != nil {
		t.Error(err)
	}
	if !m {
		t.Error("invalid YAML spec")
	}
}

// TestInvalidContentTypeOpenAPIHandler tests that the
// OpenAPI handler will panic if the given content type
// is invalid.
func TestInvalidContentTypeOpenAPIHandler(t *testing.T) {
	fizz := New()

	assert.Panics(t, func() {
		fizz.GET("/openapi.xml", nil, fizz.OpenAPI(nil, "xml"))
	})
}

// TestMultipleTonicHandler tests that adding more than
// one tonic-wrapped handler to a Fizz operation panics.
func TestMultipleTonicHandler(t *testing.T) {
	fizz := New()

	assert.Panics(t, func() {
		fizz.GET("/:a", nil,
			tonic.Handler(func(c *gin.Context) error { return nil }, 200),
			tonic.Handler(func(c *gin.Context) error { return nil }, 200),
		)
	})
}

// TestErrorGen tests that the generator panics if
// if fails to add an operation to the specification.
func TestErrorGen(t *testing.T) {
	type In struct {
		A string `path:"a" query:"b"`
	}
	fizz := New()

	assert.Panics(t, func() {
		fizz.GET("/a", nil, tonic.Handler(func(c *gin.Context, param *In) error { return nil }, 200))
	})
}

func TestJoinPaths(t *testing.T) {
	jp := joinPaths

	assert.Equal(t, "", jp("", ""))
	assert.Equal(t, "/", jp("", "/"))
	assert.Equal(t, "/a", jp("/a", ""))
	assert.Equal(t, "/a/", jp("/a/", ""))
	assert.Equal(t, "/a/", jp("/a/", "/"))
	assert.Equal(t, "/a/", jp("/a", "/"))
	assert.Equal(t, "/a/b", jp("/a", "/b"))
	assert.Equal(t, "/a/b", jp("/a/", "/b"))
	assert.Equal(t, "/a/b/", jp("/a/", "/b/"))
	assert.Equal(t, "/a/b/", jp("/a/", "/b//"))
}

func TestLastChar(t *testing.T) {
	assert.Equal(t, uint8('a'), lastChar("hola"))
	assert.Equal(t, uint8('s'), lastChar("adios"))

	assert.Panics(t, func() { lastChar("") })
}

func diffJSON(a, b []byte) (bool, error) {
	var j1, j2 interface{}
	if err := json.Unmarshal(a, &j1); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j1), nil
}

func diffYAML(a, b []byte) (bool, error) {
	var j1, j2 interface{}
	if err := yaml.Unmarshal(a, &j1); err != nil {
		return false, err
	}
	if err := yaml.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	q.Q(j1, j2)
	return reflect.DeepEqual(j2, j1), nil
}
