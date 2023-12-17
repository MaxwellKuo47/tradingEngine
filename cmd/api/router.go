package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResp)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResp)

	// system
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// user
	router.HandlerFunc(http.MethodPost, "/v1/users", app.userRegisterHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users/authentication", app.userLoginHandler)

	// order
	router.HandlerFunc(http.MethodPost, "/v1/orders", app.requireAuthenticatedUser(app.orderCreate))

	// for adjust fake stock value
	router.HandlerFunc(http.MethodPost, "/v1/stockValueChangeHandler", app.adjustStockPrice)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
