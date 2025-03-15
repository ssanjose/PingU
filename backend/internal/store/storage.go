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
		Create(context.Context, *sql.Tx, *User) error
		CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error
		GetByID(context.Context, int64) (*User, error)
		Update(context.Context, *User) error
		Delete(context.Context, int64) error
		Partner(context.Context, *User, *User) error
		Unpartner(context.Context, *User) error
		Ping(context.Context, *User) error
		Pong(context.Context, *User) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users: &UserStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
