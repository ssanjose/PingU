package store

import (
	"context"
	"database/sql"

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
