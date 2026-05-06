package user

import (
	"context"
	"time"
)

type User struct {
	ID                int64     `json:"id"`
	Email             string    `json:"email"`
	UserName          string    `json:"username"`
	PasswordHash      string    `json:"-"`
	IsVerified        bool      `json:"is_verified"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// user Repo interface
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	GetPasswordChangedAt(ctx context.Context, id int64) (time.Time, error)
	MarkVerified(ctx context.Context, id int64) error
}
