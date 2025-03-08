package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	ID                 int64         `json:"id"`
	Username           string        `json:"username"`
	Email              string        `json:"email"`
	Password           string        `json:"-"`
	Pinged             bool          `json:"pinged"`               // user is pinged
	LastPingedAt       sql.NullTime  `json:"last_pinged_at"`       // last time user was pinged
	Verified           bool          `json:"verified"`             // email is verified
	UpdatedAt          time.Time     `json:"updated_at"`           // last time user was updated
	CreatedAt          time.Time     `json:"created_at"`           // user's account creation date
	PingedPartnerCount int64         `json:"pinged_partner_count"` // number of times user has pinged partner without response
	PartnerID          sql.NullInt64 `json:"partner_id"`           // user's partner's userID
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
		SELECT id, username, email, pinged, last_pinged_at, verified, pinged_partner_count, partner_id, updated_at, created_at
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
		&user.CreatedAt,
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

func (s *UserStore) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, updated_at = NOW()
		WHERE id = $3 AND updated_at = $4
    RETURNING updated_at
	`

	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.ID,
		user.UpdatedAt,
	).Scan(&user.UpdatedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *UserStore) Ping(ctx context.Context, user *User) error {
	if !user.PartnerID.Valid {
		return ErrPartnerNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var partnerExists bool
	query := `
    SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
  `
	err = tx.QueryRowContext(
		ctx,
		query,
		user.PartnerID.Int64,
	).Scan(
		&partnerExists,
	)
	if err != nil {
		return err
	}

	if !partnerExists {
		return ErrPartnerNotFound
	}

	query = `
    UPDATE users
    SET pinged_partner_count = pinged_partner_count + 1,
        updated_at = NOW()
    WHERE id = $1 AND updated_at = $2
    RETURNING updated_at
  `

	err = tx.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.UpdatedAt,
	).Scan(&user.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	var partnerUpdatedAt time.Time
	query = `
      SELECT updated_at
      FROM users
      WHERE id = $1
  `
	err = tx.QueryRowContext(ctx, query, user.PartnerID.Int64).Scan(&partnerUpdatedAt)
	if err != nil {
		return err
	}

	query = `
    UPDATE users
    SET pinged = true, last_pinged_at = NOW(), updated_at = NOW()
    WHERE id = $1 AND updated_at = $2
  `

	var newPartnerUpdatedAt time.Time
	err = tx.QueryRowContext(
		ctx,
		query,
		user.PartnerID,
		partnerUpdatedAt,
	).Scan(&newPartnerUpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
