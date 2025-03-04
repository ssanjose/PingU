package store

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Storage struct {
	Users interface {
		Create(context.Context, *User) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users: &UserStore{db},
	}
}
