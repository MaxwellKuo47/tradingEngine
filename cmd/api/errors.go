package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.errorLogger.Error(
		"logError",
		slog.String("msg", err.Error()),
		slog.String("request_method", r.Method),
		slog.String("request_url", r.URL.String()),
	)
}

func (app *application) errResp(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrResp(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errResp(w, r, http.StatusInternalServerError, message)
}

func (app *application) notFoundResp(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errResp(w, r, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResp(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errResp(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) rateLimitExceededResp(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errResp(w, r, http.StatusTooManyRequests, message)
}

func (app *application) badReqResp(w http.ResponseWriter, r *http.Request, err error) {
	app.errResp(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResp(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errResp(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) editConflictResp(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errResp(w, r, http.StatusConflict, message)
}

func (app *application) invalidCredentialsResp(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errResp(w, r, http.StatusUnauthorized, message)
}

func (app *application) invalidAuthTokenResp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	app.errResp(w, r, http.StatusUnauthorized, message)
}

func (app *application) authRequiredResp(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errResp(w, r, http.StatusUnauthorized, message)
}
