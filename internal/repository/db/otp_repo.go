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

//OTP interface

// create
func (r *OTPRepo) Create(ctx context.Context, otp *otp.OTP) error {

	const q = `
		INSERT INTO otps (user_id, code, type, expires_at, used, created_at)
		VALUES (?, ?, ?, ?, 0, ?)
	`

	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, q, otp.UserID, otp.Code, otp.Type, otp.ExpireAt, now)

	return err
}

// GET Latest OTP for the Given Type
func (r *OTPRepo) GetLatest(ctx context.Context, userID int64, otpType *otp.OTPType) (*otp.OTP, error) {

	const q = `
		SELECT id, user_id, code, type, expires_at, used, created_at
    	FROM   otps
    	WHERE  user_id = ? AND type = ? AND used = 0
    	ORDER  BY created_at DESC
    	LIMIT  1
	`

	otp := new(otp.OTP)

	err := r.db.QueryRowContext(ctx, q, userID, otpType).Scan(
		&otp.ID, &otp.UserID, &otp.Code, &otp.Type,
		&otp.ExpireAt, &otp.Used, &otp.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrInvalidOTP
		}
		return nil, err
	}
	return otp, nil
}

// mark used otp used = 1
func (r *OTPRepo) MarkUsed(ctx context.Context, id int64) error {

	const q = `UPDATE otps SET used = 1 WHERE id = ?`

	_, err := r.db.ExecContext(ctx, q, id)

	return err
}

// marks all unused OTPs of the given type for a user as used = 1
func (r *OTPRepo) InvalidatePrevious(ctx context.Context, userID int64, otpType otp.OTPType) error {

	const q = `UPDATE otps SET used = 1 WHERE user_id = ? AND type = ? AND used = 0`

	_, err := r.db.ExecContext(ctx, q, userID, otpType)

	return err
}
