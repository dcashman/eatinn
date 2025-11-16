package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// Initialize a new httprouter router instance.
	router := httprouter.New()

	// Convert the notFoundResponse() helper to a http.Handler using the
	// http.HandlerFunc() adapter, and then set it as the custom error handler for 404
	// Not Found responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Likewise, convert the methodNotAllowedResponse() helper to a http.Handler and set
	// it as the custom error handler for 405 Method Not Allowed responses.
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/recipes", app.createRecipeHandler)
	router.HandlerFunc(http.MethodGet, "/v1/recipes/:id", app.showRecipeHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/recipes/:id", app.updateRecipeHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/recipes/:id", app.deleteRecipeHandler)

	// Return the httprouter instance.
	return app.recoverPanic(router)
}
