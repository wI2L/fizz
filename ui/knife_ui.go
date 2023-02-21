package ui

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed knife4go
var swaggerUiRes embed.FS

// AddUIHandler adds handler that serves html for Swagger UI
func AddUIHandler(ginEngine *gin.Engine, path string, openApiJsonPath string) {
	sub, err := fs.Sub(swaggerUiRes, "knife4go")
	if err != nil {
		panic(err)
	}

	ginEngine.StaticFS(path, http.FS(FsWrapper(sub, openApiJsonPath)))
}
