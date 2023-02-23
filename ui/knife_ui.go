package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

//go:embed knife4go/webjars
//go:embed knife4go/img
//go:embed knife4go/oauth
var statics embed.FS

//go:embed knife4go/doc.html
var docHtml []byte

// AddUIHandler adds handler that serves html for Swagger UI
func AddUIHandler(ginEngine *gin.Engine, path string, openApiJsonPath string) {

	// for `v3/api-docs/swagger-config`, as springdoc
	configPath, _ := url.JoinPath(path, "v3/api-docs/swagger-config")
	ginEngine.GET(configPath, func(c *gin.Context) {
		c.JSON(200, &SwaggerConfig{ConfigUrl: configPath, DisplayRequestDuration: true, OperationsSorter: "method", Urls: &[]SwaggerUrl{
			SwaggerUrl{
				Url:  openApiJsonPath,
				Name: "default",
			},
		}})
	})

	ginEngine.GET(path+"/index.html", func(c *gin.Context) {
		c.Writer.WriteHeader(200)
		c.Writer.Write(docHtml)
		c.Writer.Header().Add("Accept", "text/html")
		c.Writer.Flush()
	})

	// webjars
	subWebjars, err := fs.Sub(statics, "knife4go/webjars")
	if err != nil {
		panic(err)
	}

	urlSubWebJars, _ := url.JoinPath(path, "webjars")
	ginEngine.StaticFS(urlSubWebJars, http.FS(subWebjars))

	// img
	subImg, err := fs.Sub(statics, "knife4go/img")
	if err != nil {
		panic(err)
	}

	urlSubImg, _ := url.JoinPath(path, "img")
	ginEngine.StaticFS(urlSubImg, http.FS(subImg))

	// oauth
	subOauth, err := fs.Sub(statics, "knife4go/oauth")
	if err != nil {
		panic(err)
	}

	urlSubOauth, _ := url.JoinPath(path, "oauth")
	ginEngine.StaticFS(urlSubOauth, http.FS(subOauth))

}
