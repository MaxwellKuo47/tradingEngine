package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// system
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// user
	router.HandlerFunc(http.MethodPost, "/v1/users", app.userRegisterHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users/authentication", app.userLoginHandler)

	return app.recoverPanic(app.rateLimit(router))
}
