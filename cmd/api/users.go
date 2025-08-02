package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ab0utbla-k/rvt-hello-app/internal/data"
	"github.com/ab0utbla-k/rvt-hello-app/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) saveUserHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	username := params.ByName("username")

	var input struct {
		DateOfBirth string `json:"dateOfBirth"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	dateOfBirth, err := time.Parse("2006-01-02", input.DateOfBirth)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid date format, use YYYY-MM-DD"))
		return
	}

	user := &data.User{
		Username:    username,
		DateOfBirth: dateOfBirth,
	}

	v := validator.New()
	data.ValidateUser(v, user)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) getBirthdayMessageHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	username := params.ByName("username")

	user, err := app.models.Users.Get(username)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	message := user.GetBirthdayMessage()

	env := envelope{"message": message}
	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
