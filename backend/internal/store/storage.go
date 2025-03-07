package store

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

var (
	ErrNotFound = errors.New("resource not found")
)

type Storage struct {
	Users interface {
		Create(context.Context, *User) error
		GetByID(context.Context, int64) (*User, error)
		// Update(context.Context, *User) error
		Delete(context.Context, int64) error
		Ping(context.Context) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users: &UserStore{db},
	}
}
