package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/maxwellkuo47/tradingEngine/internal/data"
	"github.com/maxwellkuo47/tradingEngine/internal/validator"
)

func (app *application) userRegisterHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badReqResp(w, r, err)
		return
	}

	user := &data.User{
		Name:  input.Name,
		Email: input.Email,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}
	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResp(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResp(w, r, v.Errors)
		default:
			app.serverErrResp(w, r, err)
		}
		return
	}

	err = app.models.UserWallet.New(user.ID)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
	}
}

func (app *application) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badReqResp(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResp(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResp(w, r)
		default:
			app.serverErrResp(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialsResp(w, r)
		return
	}

	token, err := app.models.Token.New(user.ID, 24*time.Hour)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}

}
