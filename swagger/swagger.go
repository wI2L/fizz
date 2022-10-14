package swagger

import (
	"embed"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
)

//go:embed swagger-ui
var swaggerUiRes embed.FS

// AddUIHandler adds handler that serves html for Swagger UI
func AddUIHandler(ginEngine *gin.Engine, path string, openApiSpecPath string) error {
	sub, err := fs.Sub(swaggerUiRes, "swagger-ui")
	if err != nil {
		return err
	}

	ginEngine.StaticFS(path, http.FS(newFsWrapper(sub, openApiSpecPath)))
	return nil
}
