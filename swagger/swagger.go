package swagger

import (
	"embed"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
)

//go:embed swagger-ui
var swaggerUiRes embed.FS

// AddOpenApiUIHandler adds handler that serves html for Swagger UI
func AddOpenApiUIHandler(ginEngine *gin.Engine, path string, openApiJsonPath string) {
	sub, err := fs.Sub(swaggerUiRes, "swagger-ui")
	if err != nil {
		panic(err)
	}

	ginEngine.StaticFS(path, http.FS(FsWrapper(sub, openApiJsonPath)))
}
