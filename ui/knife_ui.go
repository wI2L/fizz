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

	// for `v3/api-docs/swagger-config`, as springdoc
	configPath := path + "v3/api-docs/swagger-config"
	ginEngine.GET(configPath, func(c *gin.Context) {

		c.JSON(200, &SwaggerConfig{ConfigUrl: configPath, DisplayRequestDuration: true, OperationsSorter: "method", Urls: &[]SwaggerUrl{
			SwaggerUrl{
				Url:  openApiJsonPath,
				Name: "default",
			},
		}})
	})

	ginEngine.StaticFS(path, http.FS(FsWrapper(sub, openApiJsonPath)))
}
