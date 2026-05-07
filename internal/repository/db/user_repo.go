package db

import (
	"context"
	"database/sql"
	"errors"
	domainerrors "nex_play_auth/github.com/internal/domain/errors"
	"nex_play_auth/github.com/internal/domain/user"
	"strings"
	"time"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func isUniqueErr(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

//User interface

// Create User
func (r *UserRepo) Create(ctx context.Context, u *user.User) error {

	const q = `	
			INSERT INTO users (email, username, password_hash, is_verified, created_at, updated_at)
			VALUES (?, ?, ?, 0, ?, ?)
	`
	now := time.Now().UTC()

	res, err := r.db.ExecContext(ctx, q, u.Email, u.UserName, u.PasswordHash, now, now)

	if err != nil {
		if isUniqueErr(err) {
			return domainerrors.ErrUserAlreadyExists
		}
		return err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return err
	}

	u.ID = id
	u.CreatedAt = now
	u.UpdatedAt = now

	return nil
}

// get user by ID
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*user.User, error) {

	const q = `
		SELECT id, email, username, password_hash, is_verified, password_changed_at, created_at, updated_at
		FROM users WHERE id = ?
	`
	u := new(user.User)

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.UserName, &u.PasswordHash, &u.IsVerified, &u.PasswordChangedAt, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, err
	}

	return u, nil
}

// get user by email
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {

	const q = `
		SELECT id, email, username, password_hash, is_verified, password_changed_at, created_at, updated_at
		FROM users WHERE email = ?
	`
	u := new(user.User)

	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.UserName, &u.PasswordHash, &u.IsVerified, &u.PasswordChangedAt, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, err
	}

	return u, nil
}

// Updated Password
func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {

	const q = `
		UPDATE users SET password_hash = ?, password_changed_at = ?, updated_at = ? WHERE id = ?
	`
	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, q, passwordHash, now, now, id)

	return err
}

// Get Password Changed at
func (r *UserRepo) GetPasswordChangedAt(ctx context.Context, id int64) (time.Time, error) {

	var t time.Time

	err := r.db.QueryRowContext(ctx,
		`SELECT password_changed_at FROM users WHERE id = ?`, id,
	).Scan(&t)

	return t, err
}

// MarkVerified
func (r *UserRepo) MarkVerified(ctx context.Context, id int64) error {

	const q = `
		UPDATE users SET is_verified = 1, updated_at = ? WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, q, time.Now().UTC(), id)

	return err
}
