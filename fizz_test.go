package fizz

import (
	"encoding/json"
	"fmt"
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
	"github.com/gofrs/uuid"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

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

// customTime shows the date & time without timezone information
type customTime time.Time

func (c customTime) String() string {
	return time.Time(c).Format("2006-01-02T15:04:05")
}

func (c customTime) MarshalJSON() ([]byte, error) {
	// add quotes for JSON representation
	ts := fmt.Sprintf("\"%s\"", c.String())
	return []byte(ts), nil
}

func (c customTime) MarshalYAML() (interface{}, error) {
	return c.String(), nil
}

func (c customTime) ParseExample(v string) (interface{}, error) {
	t1, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil, err
	}
	return customTime(t1), nil
}

type T struct {
	X string     `json:"x" yaml:"x" description:"This is X"`
	Y int        `json:"y" yaml:"y" description:"This is Y"`
	Z customTime `json:"z" yaml:"z" example:"2022-02-07T18:00:00+09:00" description:"This is Z"`
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

	t1, err := time.Parse(time.RFC3339, "2022-02-07T18:00:00+09:00")
	assert.Nil(t, err)

	fizz.GET("/foo/:a", nil, tonic.Handler(func(c *gin.Context, params *In) (*T, error) {
		assert.Equal(t, 0, params.A)
		assert.Equal(t, "foobar", params.B)
		assert.Equal(t, "foobaz", params.C)

		return &T{X: "foo", Y: 1, Z: customTime(t1)}, nil
	}, 200))

	// Create a router group to test that tonic handlers works with router groups.
	grp := fizz.Group("/test", "Test Group", "Test Group")

	grp.GET("/bar/:a", nil, tonic.Handler(func(c *gin.Context, params *In) (*T, error) {
		assert.Equal(t, 42, params.A)
		assert.Equal(t, "group-foobar", params.B)
		assert.Equal(t, "group-foobaz", params.C)

		return &T{X: "group-foo", Y: 2, Z: customTime(t1)}, nil
	}, 200))

	srv := httptest.NewServer(fizz)
	defer srv.Close()

	c := srv.Client()
	c.Timeout = 1 * time.Second

	requests := []struct {
		url    string
		method string
		header http.Header

		expectStatus int
		expectBody   string
	}{
		{
			url:    "/foo/0?b=foobar",
			method: http.MethodGet,
			header: http.Header{
				"X-Test-C": []string{"foobaz"},
			},
			expectStatus: 200,
			expectBody:   `{"x":"foo","y":1,"z":"2022-02-07T18:00:00"}`,
		},
		{
			url:    "/test/bar/42?b=group-foobar",
			method: http.MethodGet,
			header: http.Header{
				"X-Test-C": []string{"group-foobaz"},
			},
			expectStatus: 200,
			expectBody:   `{"x":"group-foo","y":2,"z":"2022-02-07T18:00:00"}`,
		},
		{
			url:    "/bar/42?b=group-foobar",
			method: http.MethodGet,
			header: http.Header{
				"X-Test-C": []string{"group-foobaz"},
			},
			expectStatus: 404,
		},
	}

	for _, req := range requests {
		url, err := url.Parse(srv.URL + req.url)
		if err != nil {
			t.Error(err)
			break
		}

		resp, err := c.Do(&http.Request{
			URL:    url,
			Method: req.method,
			Header: req.header,
		})
		if err != nil {
			t.Error(err)
			break
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
			break
		}

		if req.expectStatus < 300 {
			assert.Equal(t, req.expectStatus, resp.StatusCode)
			assert.Equal(t, req.expectBody, string(body))
		} else {
			assert.Equal(t, req.expectStatus, resp.StatusCode)
		}
	}
}

type testInputModel struct {
	PathParam1 string `path:"a"`
	PathParam2 int    `path:"b"`
	QueryParam string `query:"q"`
}

type testInputModel1 struct {
	PathParam1 string `path:"a"`
}

type testInputModel2 struct {
	C        string      `path:"c"`
	Message  string      `json:"message" description:"A short message"`
	AnyValue interface{} `json:"value" description:"A nullable value of arbitrary type"`
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
				{
					Name:        "X-Rate-Limit",
					Description: "Rate limit",
					Model:       Integer,
				},
			}, nil),
			Response("404", "", String, nil, "not-found-example"),
			ResponseWithExamples("400", "", String, nil, map[string]interface{}{
				"one": "message1",
				"two": "message2",
			}),
			XCodeSample(&openapi.XCodeSample{
				Lang:   "Shell",
				Label:  "v4.4",
				Source: "curl http://0.0.0.0:8080",
			}),
			// Explicit override for SecurityRequirement (allow-all)
			WithoutSecurity(),
			XInternal(),
		},
		tonic.Handler(func(c *gin.Context, in *testInputModel1) (*T, error) {
			return &T{}, nil
		}, 200),
	)

	fizz.GET("/test/:a/:b", []OperationOption{
		ID("GetTest2"),
		InputModel(&testInputModel{}),
		WithOptionalSecurity(),
		Security(&openapi.SecurityRequirement{"oauth2": []string{"write:pets", "read:pets"}}),
	}, tonic.Handler(func(c *gin.Context) error {
		return nil
	}, 200))
	infos := &openapi.Info{
		Title:       "Test Server",
		Description: `This is a test server.`,
		Version:     "1.0.0",
	}

	fizz.POST("/test/:c",
		[]OperationOption{
			ID("PostTest"),
			StatusDescription("201"),
			StatusDescription("Created"),
		},
		tonic.Handler(func(c *gin.Context, in *testInputModel2) error {
			return nil
		}, 201),
	)

	servers := []*openapi.Server{
		{
			URL:         "https://foo.bar/{basePath}",
			Description: "Such Server, Very Wow",
			Variables: map[string]*openapi.ServerVariable{
				"basePath": {
					Default:     "v2",
					Description: "version of the API",
					Enum:        []string{"v1", "v2", "beta"},
				},
			},
		},
	}
	fizz.Generator().SetServers(servers)

	security := []*openapi.SecurityRequirement{
		{"api_key": []string{}},
		{"oauth2": []string{"write:pets", "read:pets"}},
	}
	fizz.Generator().SetSecurityRequirement(security)

	fizz.Generator().API().Components.SecuritySchemes = map[string]*openapi.SecuritySchemeOrRef{
		"api_key": {
			SecurityScheme: &openapi.SecurityScheme{
				Type: "apiKey",
				Name: "api_key",
				In:   "header",
			},
		},
		"oauth2": {
			SecurityScheme: &openapi.SecurityScheme{
				Type: "oauth2",
				Flows: &openapi.OAuthFlows{
					Implicit: &openapi.OAuthFlow{
						AuthorizationURL: "https://example.com/api/oauth/dialog",
						Scopes: map[string]string{
							"write:pets": "modify pets in your account",
							"read:pets":  "read your pets",
						},
					},
				},
			},
		},
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

func TestOperationContext(t *testing.T) {
	fizz := New()

	const (
		id   = "OperationContext"
		desc = "Test for OpenAPI operation instance in Gin context"
	)
	tonicHandler := tonic.Handler(func(c *gin.Context) error {
		op, err := OperationFromContext(c)
		if err == nil && op.ID == id && op.Description == desc {
			c.Status(http.StatusOK)
			return nil
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return nil
	}, http.StatusOK)

	fizz.GET("/op",
		[]OperationOption{
			ID(id),
			Description(desc),
		}, tonicHandler,
	)
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/op", nil)
	if err != nil {
		t.Fatal(err)
	}
	fizz.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status: got %v want %v",
			status, http.StatusOK,
		)
	}

	fizz.POST("/noop", nil, func(c *gin.Context) {
		_, err := OperationFromContext(c)
		if err != nil {
			c.Status(http.StatusOK)
			return
		}
		c.Status(http.StatusInternalServerError)
	})
	req, err = http.NewRequest("POST", "/noop", nil)
	if err != nil {
		t.Fatal(err)
	}
	fizz.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status: got %v want %v",
			status, http.StatusOK,
		)
	}
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
	return reflect.DeepEqual(j2, j1), nil
}
