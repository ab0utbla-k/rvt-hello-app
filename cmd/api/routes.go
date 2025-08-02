package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	router.HandlerFunc(http.MethodGet, "/hello/:username", app.getBirthdayMessageHandler)
	router.HandlerFunc(http.MethodPut, "/hello/:username", app.saveUserHandler)

	return app.metrics(app.recoverPanic(router))
}
