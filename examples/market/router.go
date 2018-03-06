package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
	"github.com/wI2L/fizz/tonic"
)

// NewRouter returns a new router for the
// Pet Store.
func NewRouter() *gin.Engine {
	engine := gin.Default()
	engine.Use(cors.Default())

	fizz := fizz.NewFromEngine(engine)

	// Initialize the informations of
	// the API that will be served with
	// the specification.
	infos := &openapi.Info{
		Title:       "Fruits Market",
		Description: `This is a sample Fruits market server.`,
		Version:     "1.0.0",
	}
	// Get the underlying fizz instance router and
	// create a new route that serve the spec without
	// any tonic wrapping of the handler.
	fizz.Router().GET("/openapi.json", fizz.OpenAPI(infos, "json"))

	// Setup routes.
	routes(fizz.Group("/market", "market", "Your daily dose of freshness"))

	return fizz.Router()
}

func routes(grp *fizz.RouterGroup) {
	// Add a new fruit to the market.
	grp.POST("", CreateFruit,
		tonic.StatusCode(200),
		tonic.Summary("Add a fruit to the market"),
		tonic.Response("400", "Bad request", nil, nil),
	)
	// Remove a fruit from the market,
	// probably because it rotted.
	grp.DELETE("/:name", DeleteFruit,
		tonic.StatusCode(204),
		tonic.Summary("Remove a fruit from the market"),
		tonic.Response("400", "Fruit not found", nil, nil),
	)
	grp.GET("", ListFruits,
		tonic.StatusCode(200),
		tonic.Summary("List the fruits of the market"),
		tonic.Response("400", "Bad request", nil, nil),
	)
}
