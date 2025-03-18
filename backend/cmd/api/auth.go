package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
	"github.com/ssanjose/PingU/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=35"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

// registerUserHandler godoc
//
// @Summary Register a new user
// @Description Register a new user
// @Tags authentication
// @Accept json
// @Produce json
// @Param payload body RegisterUserPayload true "User credentials"
// @Success 201 {object} User "User registered"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/users/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	// hash the user password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	// store the user
	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// send mail

	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	app.jsonResponse(w, http.StatusOK, nil)
}
