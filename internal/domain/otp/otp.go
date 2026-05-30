package otp

import (
	"context"
	"time"
)

type OTPType string

const (
	OTPTypeSignUp        OTPType = "signup"
	OTPTypeResetPassword OTPType = "reset_password"
)

// OTP struct
type OTP struct {
	ID        int64
	UserID    int64
	Code      string
	Type      OTPType
	ExpireAt  time.Time
	Used      bool
	CreatedAt time.Time
}

// OTP Repo interface
type OTPRepository interface {
	Create(ctx context.Context, otp *OTP) error                                  // create OTP and save in the OTP table on db
	GetLatest(ctx context.Context, userID int64, otpType OTPType) (*OTP, error)  // GET Latest OTP for the Given Type
	MarkUsed(ctx context.Context, id int64) error                                // mark the OTP as used in the db
	InvalidatePrevious(ctx context.Context, userID int64, otpType OTPType) error // InvalidatePrevious marks all previous unused OTPs for a user+type as used.
	DeleteExpiredOrUsed(ctx context.Context) error                               // Delete OTPs that are used or expired
}
