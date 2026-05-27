package authservice

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"nex_play_auth/github.com/config"
	domainerrors "nex_play_auth/github.com/internal/domain/errors"
	"nex_play_auth/github.com/internal/domain/otp"
	"nex_play_auth/github.com/internal/domain/user"
	generateotp "nex_play_auth/github.com/pkg/generate_otp"
	"nex_play_auth/github.com/pkg/hash"
	"nex_play_auth/github.com/pkg/jwt"
	"nex_play_auth/github.com/pkg/mailer"
	passwordstrengthchecker "nex_play_auth/github.com/pkg/password_strength_checker"
	"time"
)

// Auth Service
type Service struct {
	userRepo user.UserRepository
	otpRepo  otp.OTPRepository
	jwt      *jwt.Manager
	mailer   *mailer.Mailer
	cfg      *config.Config
}

// New service
func NewService(userRepo user.UserRepository,
	otpRepo otp.OTPRepository,
	jwtManager *jwt.Manager,
	m *mailer.Mailer,
	cfg *config.Config,
) *Service {

	return &Service{
		userRepo: userRepo,
		otpRepo:  otpRepo,
		jwt:      jwtManager,
		mailer:   m,
		cfg:      cfg,
	}
}

// Helpers
func (s *Service) sentOTP(ctx context.Context, u *user.User, otpType otp.OTPType, purpose string) error {

	if err := s.otpRepo.InvalidatePrevious(ctx, u.ID, otpType); err != nil {

		return err
	}

	code, err := generateotp.New(6)

	if err != nil {

		return err
	}

	otpRecord := &otp.OTP{
		UserID:   u.ID,
		Code:     code,
		Type:     otpType,
		ExpireAt: time.Now().UTC().Add(s.cfg.OTPExpiry),
	}

	if err := s.otpRepo.Create(ctx, otpRecord); err != nil {

		return err
	}

	switch purpose {
	case "signup":
		purpose = "SignUp"
	case "reset_password":
		purpose = "reset password"
	}

	go func() {

		if err := s.mailer.SendOTP(u.Email, code, purpose); err != nil {

			slog.Error("failed to send OTP email",
				"user_id", u.ID,
				"purpose", purpose,
				"error", err,
			)
		}
	}()

	return nil
}

// Helper
func (s *Service) consumeOTP(ctx context.Context, userID int64, otpType otp.OTPType, code string) error {

	otpRecord, err := s.otpRepo.GetLatest(ctx, userID, otpType)

	if err != nil {

		return domainerrors.ErrInvalidOTP
	}

	//print logs
	slog.Info("OTP debug",
		"stored_code", otpRecord.Code,
		"entered_code", code,
		"used", otpRecord.Used,
		"expires_at", otpRecord.ExpireAt,
		"now", time.Now().UTC(),
		"expired", time.Now().UTC().After(otpRecord.ExpireAt),
	)

	if otpRecord.Used {

		return domainerrors.ErrOTPAlreadyUsed
	}

	if time.Now().UTC().After(otpRecord.ExpireAt) {

		return domainerrors.ErrOTPExpired
	}

	// Constant time comparison to prevent timing attacks.
	if subtle.ConstantTimeCompare([]byte(otpRecord.Code), []byte(code)) != 1 {

		return domainerrors.ErrInvalidOTP
	}

	return s.otpRepo.MarkUsed(ctx, userID)

}

// SignUp
func (s *Service) SignUp(ctx context.Context, email, userName, password string) error {

	if err := passwordstrengthchecker.Check(password); err != nil {

		return err
	}

	passwordHash, err := hash.Password(password)

	if err != nil {

		return err
	}

	newUser := &user.User{
		Email:        email,
		UserName:     userName,
		PasswordHash: passwordHash,
	}

	//save user in db (verified == 0)
	if err := s.userRepo.Create(ctx, newUser); err != nil {

		return err
	}

	//send OTP
	if err := s.sentOTP(ctx, newUser, otp.OTPTypeSignUp, "SignUp"); err != nil {

		return err
	}

	slog.Info("user registered", "user_id", newUser.ID, "email", email)

	return nil
}

// Verify validates an OTP for the given type and returns a token pair on success (signUp only)
func (s *Service) Verify(ctx context.Context, email, code string) (*jwt.TokenPair, error) {

	u, err := s.userRepo.GetByEmail(ctx, email)

	if err != nil {

		return nil, domainerrors.ErrInvalidCredentials
	}

	if err := s.consumeOTP(ctx, u.ID, otp.OTPTypeSignUp, code); err != nil {

		return nil, err
	}

	if !u.IsVerified {

		if err := s.userRepo.MarkVerified(ctx, u.ID); err != nil {

			slog.Warn("could not mark user verified", "user_id", u.ID, "error", err)
		}
	}

	token, err := s.jwt.GenerateTokenPair(u.ID, email)

	if err != nil {

		return nil, err
	}

	return token, nil
}

// SignIn
func (s *Service) SignIn(ctx context.Context, email, password string) (*jwt.TokenPair, error) {

	u, err := s.userRepo.GetByEmail(ctx, email)

	if err != nil {

		return nil, domainerrors.ErrInvalidCredentials
	}

	if !hash.CheckPassword(u.PasswordHash, password) {

		return nil, domainerrors.ErrInvalidCredentials
	}

	if !u.IsVerified {

		return nil, domainerrors.ErrAccountNotVerified
	}

	return s.jwt.GenerateTokenPair(u.ID, u.Email)
}

// resent OTP
func (s *Service) ResentOTP(ctx context.Context, email, purpose string) error {

	u, err := s.userRepo.GetByEmail(ctx, email)

	//Always returns nil to avoid leaking whether the email exists
	if err != nil {

		return nil
	}

	var otpType otp.OTPType
	switch purpose {
	case "signup":
		otpType = otp.OTPTypeSignUp
	case "reset_password":
		otpType = otp.OTPTypeResetPassword
	}

	return s.sentOTP(ctx, u, otpType, purpose)
}

// forgot send OTP to the given Email
func (s *Service) ForgotPassword(ctx context.Context, email string) error {

	u, err := s.userRepo.GetByEmail(ctx, email)

	//Always returns nil to avoid leaking whether the email exists
	if err != nil {

		return nil
	}

	return s.sentOTP(ctx, u, otp.OTPTypeResetPassword, "password reset")
}

// Reset Password verify OTP, Update password and returns new JWT Token Par
func (s *Service) ResetPassword(ctx context.Context, email, code, newPassword string) (*jwt.TokenPair, error) {

	if err := passwordstrengthchecker.Check(newPassword); err != nil {

		return nil, err
	}

	u, err := s.userRepo.GetByEmail(ctx, email)

	if err != nil {
		// Don't reveal whether the email exists.
		return nil, domainerrors.ErrInvalidOTP
	}

	if err := s.consumeOTP(ctx, u.ID, otp.OTPTypeResetPassword, code); err != nil {

		return nil, err
	}

	passwordHash, err := hash.Password(newPassword)

	if err != nil {

		return nil, err
	}

	if err := s.userRepo.UpdatePassword(ctx, u.ID, passwordHash); err != nil {

		return nil, err
	}

	slog.Info("password reset", "user_id", u.ID)

	token, err := s.jwt.GenerateTokenPair(u.ID, email)

	if err != nil {

		return nil, err
	}

	return token, nil
}

// Refresh Token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {

	claims, err := s.jwt.VerifyRefresh(refreshToken)

	if err != nil {

		return nil, domainerrors.ErrInvalidToken
	}

	u, err := s.userRepo.GetByID(ctx, claims.UserID)

	if err != nil {

		return nil, domainerrors.ErrInvalidToken
	}

	return s.jwt.GenerateTokenPair(u.ID, u.Email)
}
