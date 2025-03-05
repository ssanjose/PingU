package store

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

type User struct {
	ID                 int64  `json:"id"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	Password           string `json:"-"`
	Pinged             bool   `json:"pinged"`               // user is pinged
	LastPingedAt       string `json:"last_pinged_at"`       // last time user was pinged
	Verified           bool   `json:"verified"`             // email is verified
	UpdatedAt          string `json:"updated_at"`           // last time user was updated
	CreatedAt          string `json:"created_at"`           // user's account creation date
	PingedPartnerCount int64  `json:"pinged_partner_count"` // number of times user has pinged partner without response
	PartnerID          int64  `json:"partner_id"`           // user's partner's userID
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (username, password, email)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, username, email, pinged, last_pinged_at, verified, pinged_partner_count, created_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Pinged,
		&user.LastPingedAt,
		&user.Verified,
		&user.PingedPartnerCount,
		&user.PartnerID,
		&user.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *UserStore) Ping(ctx context.Context) error {
	return nil
}
