package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
)

type User struct {
	ID                 int64         `json:"id"`
	Username           string        `json:"username"`
	Email              string        `json:"email"`
	Password           password      `json:"-"`
	Pinged             bool          `json:"pinged"`               // user is pinged
	LastPingedAt       sql.NullTime  `json:"last_pinged_at"`       // last time user was pinged
	Verified           bool          `json:"verified"`             // email is verified
	UpdatedAt          time.Time     `json:"updated_at"`           // last time user was updated
	CreatedAt          time.Time     `json:"created_at"`           // user's account creation date
	PingedPartnerCount int64         `json:"pinged_partner_count"` // number of times user has pinged partner without response
	PartnerID          sql.NullInt64 `json:"partner_id"`           // user's partner's userID
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, password, email)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := tx.QueryRowContext(
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
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, username, email, pinged, last_pinged_at, verified, pinged_partner_count, partner_id, updated_at, created_at
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

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
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	// transaction wrapper
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// create the user
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		// create the user invite
		if err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		query := `
			INSERT INTO user_invitations (token, user_id, expires_at)
			VALUES ($1, NOW() + $2, $3)
		`

		_, err := tx.ExecContext(ctx, query, token, userID, exp)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

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
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

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
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

// Partner sets two users as partners.
func (s *UserStore) Partner(ctx context.Context, user *User, partner *User) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		query := `
		UPDATE users
		SET partner_id = $1, updated_at = NOW()
		WHERE id = $2 AND updated_at = $3
		RETURNING updated_at
	`

		err := tx.QueryRowContext(
			ctx,
			query,
			partner.ID,
			user.ID,
			user.UpdatedAt,
		).Scan(&user.UpdatedAt)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return ErrNotFound
			default:
				return err
			}
		}

		query = `
		UPDATE users
		SET partner_id = $1, updated_at = NOW()
		WHERE id = $2 AND updated_at = $3
		RETURNING updated_at
	`

		err = tx.QueryRowContext(
			ctx,
			query,
			user.ID,
			partner.ID,
			partner.UpdatedAt,
		).Scan(&partner.UpdatedAt)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
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
	})
}

// Unpartner removes the partner relationship between two users.
func (s *UserStore) Unpartner(ctx context.Context, user *User) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if !user.PartnerID.Valid {
			return ErrPartnerNotFound
		}

		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		var partnerUpdatedAt time.Time
		query := `
				SELECT updated_at
				FROM users
				WHERE id = $1
		`
		err := tx.QueryRowContext(
			ctx,
			query,
			user.PartnerID.Int64,
		).Scan(&partnerUpdatedAt)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return ErrNotFound
			default:
				return err
			}
		}

		query = `
			UPDATE users
			SET pinged = false, pinged_partner_count = 0, partner_id = NULL, updated_at = NOW()
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
			switch err {
			case sql.ErrNoRows:
				return ErrNotFound
			default:
				return err
			}
		}

		query = `
			UPDATE users
			SET pinged = false, pinged_partner_count = 0, partner_id = NULL, updated_at = NOW()
			WHERE id = $1 AND updated_at = $2
			RETURNING updated_at
		`

		var newPartnerUpdatedAt time.Time
		err = tx.QueryRowContext(
			ctx,
			query,
			user.PartnerID.Int64,
			partnerUpdatedAt,
		).Scan(&newPartnerUpdatedAt)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return ErrNotFound
			default:
				return err
			}
		}

		return nil
	})
}

// Pings a user's partner and updates the user's pinged_partner_count.
func (s *UserStore) Ping(ctx context.Context, user *User) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if !user.PartnerID.Valid {
			return ErrPartnerNotFound
		}

		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		var partnerUpdatedAt time.Time
		query := `
				SELECT updated_at
				FROM users
				WHERE id = $1
		`
		err := tx.QueryRowContext(
			ctx,
			query,
			user.PartnerID.Int64,
		).Scan(&partnerUpdatedAt)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return ErrNotFound
			default:
				return err
			}
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
			switch err {
			case sql.ErrNoRows:
				return ErrNotFound
			default:
				return err
			}
		}

		query = `
			UPDATE users
			SET pinged = true, last_pinged_at = NOW(), updated_at = NOW()
			WHERE id = $1 AND updated_at = $2
			RETURNING updated_at
		`

		var newPartnerUpdatedAt time.Time
		err = tx.QueryRowContext(
			ctx,
			query,
			user.PartnerID.Int64,
			partnerUpdatedAt,
		).Scan(&newPartnerUpdatedAt)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return ErrNotFound
			default:
				return err
			}
		}

		return nil
	})
}

// Partner answers a user's partner's ping and turns off the user's pinged status.
func (s *UserStore) Pong(ctx context.Context, user *User) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if !user.PartnerID.Valid {
			return ErrPartnerNotFound
		}

		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		query := `
			UPDATE users
			SET pinged = false, updated_at = NOW()
			WHERE id = $1 AND updated_at = $2
			RETURNING updated_at
		`

		err := tx.QueryRowContext(
			ctx,
			query,
			user.ID,
			user.UpdatedAt,
		).Scan(&user.UpdatedAt)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return ErrNotFound
			default:
				return err
			}
		}

		return nil
	})
}
