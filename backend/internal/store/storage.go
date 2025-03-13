package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrPartnerNotFound   = errors.New("partner not found")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Users interface {
		Create(context.Context, *User) error
		GetByID(context.Context, int64) (*User, error)
		Update(context.Context, *User) error
		Delete(context.Context, int64) error
		Partner(context.Context, *User, *User) error
		Ping(context.Context, *User) error
		Pong(context.Context, *User) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users: &UserStore{db},
	}
}
