package swagger

import (
	"github.com/stretchr/testify/assert"
	"github.com/wI2L/fizz"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddUIHandler(t *testing.T) {
	const testPath = "/test/swagger"
	f := fizz.New()
	err := AddUIHandler(f.Engine(), testPath, "openapi.json")

	assert.NoError(t, err)

	srv := httptest.NewServer(f)
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + testPath)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := new(strings.Builder)
	_, err = io.Copy(body, resp.Body)
	assert.Contains(t, body.String(), "openapi.json")
	assert.Contains(t, body.String(), "<title>Swagger UI</title>")
}
