package main

import (
	"net/http"

	"github.com/ssanjose/PingU/internal/store"
)

type CreatePostPayload struct {
	Username string `json:"username"`
	Password string `json:"-"`
	Email    string `json:"email"`
}

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: payload.Password,
	}

	ctx := r.Context()

	if err := app.store.Users.Create(ctx, user); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := writeJSON(w, http.StatusCreated, user); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (app *application) pingUserPartnerHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, "pong")
}
