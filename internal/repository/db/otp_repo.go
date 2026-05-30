package db

import (
	"context"
	"database/sql"
	"errors"
	domainerrors "nex_play_auth/github.com/internal/domain/errors"
	"nex_play_auth/github.com/internal/domain/otp"
	"time"
)

type OTPRepo struct {
	db *sql.DB
}

func NewOTPRepo(db *sql.DB) *OTPRepo {
	return &OTPRepo{db: db}
}

// Create inserts a new OTP record
func (r *OTPRepo) Create(ctx context.Context, otp *otp.OTP) error {

	const q = `
		INSERT INTO otps (user_id, code, type, expires_at, used, created_at)
		VALUES ($1, $2, $3, $4, FALSE, $5)
	`

	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, q, otp.UserID, otp.Code, otp.Type, otp.ExpireAt, now)

	return err
}

// GetLatest returns the most recent unused OTP for the given user and type
func (r *OTPRepo) GetLatest(ctx context.Context, userID int64, otpType otp.OTPType) (*otp.OTP, error) {

	const q = `
		SELECT id, user_id, code, type, expires_at, used, created_at
		FROM   otps
		WHERE  user_id = $1 AND type = $2 AND used = FALSE
		ORDER  BY created_at DESC
		LIMIT  1
	`

	o := new(otp.OTP)

	err := r.db.QueryRowContext(ctx, q, userID, otpType).Scan(
		&o.ID, &o.UserID, &o.Code, &o.Type,
		&o.ExpireAt, &o.Used, &o.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrInvalidOTP
		}
		return nil, err
	}

	return o, nil
}

// MarkUsed marks a single OTP as used by ID
func (r *OTPRepo) MarkUsed(ctx context.Context, id int64) error {

	const q = `UPDATE otps SET used = TRUE WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id)

	return err
}

// InvalidatePrevious marks all unused OTPs of the given type for a user as used
func (r *OTPRepo) InvalidatePrevious(ctx context.Context, userID int64, otpType otp.OTPType) error {

	const q = `UPDATE otps SET used = TRUE WHERE user_id = $1 AND type = $2 AND used = FALSE`

	_, err := r.db.ExecContext(ctx, q, userID, otpType)

	return err
}

// DeleteExpiredOrUsed deletes OTPs that are marked as used or have passed their expiration time
func (r *OTPRepo) DeleteExpiredOrUsed(ctx context.Context) error {

	const q = `DELETE FROM otps WHERE used = TRUE OR expires_at < $1`

	_, err := r.db.ExecContext(ctx, q, time.Now().UTC())

	return err
}