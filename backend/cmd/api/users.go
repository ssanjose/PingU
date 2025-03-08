package main

import (
	"errors"
	"net/http"

	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ssanjose/PingU/internal/store"
	"golang.org/x/net/context"
)

type userKey string

const userCtx userKey = "user"

type CreateUserPayload struct {
	Username string `json:"username" validate:"required,max=35"`
	Password string `json:"password" validate:"required,min=6,max=72"`
	Email    string `json:"email" validate:"required,email"`
}

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserPayload
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
		Password: payload.Password,
	}

	ctx := r.Context()

	if err := app.store.Users.Create(ctx, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getPostFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type UpdateUserPayload struct {
	Username *string `json:"username" validate:"omitempty,max=35"`
	Email    *string `json:"email" validate:"omitempty,email"`
}

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getPostFromCtx(r)
	var payload UpdateUserPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Username != nil {
		user.Username = *payload.Username
	}

	if payload.Email != nil {
		user.Email = *payload.Email
	}

	if err := app.store.Users.Update(r.Context(), user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.Users.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) pingUserPartnerHandler(w http.ResponseWriter, r *http.Request) {
	user := getPostFromCtx(r)

	if err := app.store.Users.Ping(r.Context(), user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, "pong")
}

func getPostFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "userID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
