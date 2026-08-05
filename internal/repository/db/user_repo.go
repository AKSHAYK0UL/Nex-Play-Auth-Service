package db

import (
	"context"
	"database/sql"
	"errors"
	domainerrors "nex_play_auth/github.com/internal/domain/errors"
	"nex_play_auth/github.com/internal/domain/user"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// isEmailUniqueErr checks for a PostgreSQL unique constraint violation (23505)
// on the email column specifically, so duplicate usernames are allowed.
func isEmailUniqueErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key"
}

// Create inserts a new user and populates the ID, CreatedAt, UpdatedAt fields
func (r *UserRepo) Create(ctx context.Context, u *user.User) error {

	const q = `
		INSERT INTO users (email, username, password_hash, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, FALSE, $4, $5)
		RETURNING id
	`

	now := time.Now().UTC()

	err := r.db.QueryRowContext(ctx, q, u.Email, u.UserName, u.PasswordHash, now, now).Scan(&u.ID)

	if err != nil {
		if isEmailUniqueErr(err) {
			return domainerrors.ErrUserAlreadyExists
		}
		return err
	}

	u.CreatedAt = now
	u.UpdatedAt = now

	return nil
}

// GetByID fetches a user by their ID
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*user.User, error) {

	const q = `
		SELECT id, email, username, password_hash, is_verified, password_changed_at, created_at, updated_at
		FROM users WHERE id = $1
	`

	u := new(user.User)

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.UserName, &u.PasswordHash,
		&u.IsVerified, &u.PasswordChangedAt, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, err
	}

	return u, nil
}

// GetByEmail fetches a user by their email
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {

	const q = `
		SELECT id, email, username, password_hash, is_verified, password_changed_at, created_at, updated_at
		FROM users WHERE email = $1
	`

	u := new(user.User)

	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.UserName, &u.PasswordHash,
		&u.IsVerified, &u.PasswordChangedAt, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, err
	}

	return u, nil
}

// UpdatePassword updates the password hash and timestamps for a user
func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {

	const q = `
		UPDATE users SET password_hash = $1, password_changed_at = $2, updated_at = $3 WHERE id = $4
	`

	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, q, passwordHash, now, now, id)

	return err
}

// GetPasswordChangedAt returns the time the user last changed their password
func (r *UserRepo) GetPasswordChangedAt(ctx context.Context, id int64) (time.Time, error) {

	var t time.Time

	err := r.db.QueryRowContext(ctx,
		`SELECT password_changed_at FROM users WHERE id = $1`, id,
	).Scan(&t)

	return t, err
}

// MarkVerified sets is_verified to TRUE for the given user ID
func (r *UserRepo) MarkVerified(ctx context.Context, id int64) error {

	const q = `
		UPDATE users SET is_verified = TRUE, updated_at = $1 WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, q, time.Now().UTC(), id)

	return err
}
