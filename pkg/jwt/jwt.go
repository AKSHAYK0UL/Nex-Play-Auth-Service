package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// token type
type JWTTokenType string

const (
	accessToken  JWTTokenType = "access"
	refreshToken JWTTokenType = "refresh"
)

// payload
type Claims struct {
	UserID    int64        `json:"user_id"`
	Email     string       `json:"email"`
	TokenType JWTTokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// send to the user after login
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // in seconds
}

// manager
type Manager struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// create manager
func NewManager(secret string, accessExpiry, refreshExpiry time.Duration) *Manager {

	return &Manager{
		secret:        []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// sign creates a signed JWT with the given param
// tokenType distinguishes access tokens from refresh tokens so they
// cannot be used interchangeably
func (m *Manager) sign(userID int64, email string, expiry time.Duration, tokenType JWTTokenType) (string, error) {

	now := time.Now()

	claims := &Claims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Genrate Token Pair
func (m *Manager) GenerateTokenPair(UserID int64, email string) (*TokenPair, error) {

	access, err := m.sign(UserID, email, m.accessExpiry, accessToken)

	if err != nil {
		return nil, err
	}

	refresh, err := m.sign(UserID, email, m.refreshExpiry, refreshToken)

	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(m.accessExpiry.Seconds()),
	}, nil
}

// Verify parses and validates the token signature and expiry
func (m *Manager) Verify(tokenStr string) (*Claims, error) {

	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {

			return nil, errors.New("invalid signing method")
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(*Claims)

	if !ok || !t.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// VerifyAccess validates the token and ensures it is an access token
func (m *Manager) VerifyAccess(tokenStr string) (*Claims, error) {

	claims, err := m.Verify(tokenStr)

	if err != nil {
		return nil, err
	}

	if claims.TokenType != accessToken {
		return nil, errors.New("not an access token")
	}

	return claims, nil
}

// VerifyRefresh validates the token and ensures it is a refresh token
func (m *Manager) VerifyRefresh(tokenStr string) (*Claims, error) {

	claims, err := m.Verify(tokenStr)

	if err != nil {
		return nil, err
	}

	if claims.TokenType != refreshToken {
		return nil, errors.New("not a refresh token")
	}

	return claims, nil
}
