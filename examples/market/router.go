package main

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
)

// NewRouter returns a new router for the
// Pet Store.
func NewRouter() (*fizz.Fizz, error) {
	engine := gin.New()
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
	fizz.Engine().GET("/openapi.json", fizz.OpenAPI(infos, "json"))

	// Setup routes.
	routes(fizz.Group("/market", "market", "Your daily dose of freshness"))

	if len(fizz.Errors()) != 0 {
		return nil, fmt.Errorf("fizz errors: %v", fizz.Errors())
	}
	return fizz, nil
}

func routes(grp *fizz.RouterGroup) {
	// Add a new fruit to the market.
	grp.POST("", CreateFruit,
		fizz.StatusCode(200),
		fizz.Summary("Add a fruit to the market"),
		fizz.Response("400", "Bad request", nil, nil),
	)
	// Remove a fruit from the market,
	// probably because it rotted.
	grp.DELETE("/:name", DeleteFruit,
		fizz.StatusCode(204),
		fizz.Summary("Remove a fruit from the market"),
		fizz.Response("400", "Fruit not found", nil, nil),
	)
	grp.GET("", ListFruits,
		fizz.StatusCode(200),
		fizz.Summary("List the fruits of the market"),
		fizz.Response("400", "Bad request", nil, nil),
		fizz.Header("X-Market-Listing-Size", "Listing size", fizz.Long),
	)
}
