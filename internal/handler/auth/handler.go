package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	domainerrors "nex_play_auth/github.com/internal/domain/errors"
	authservice "nex_play_auth/github.com/internal/service/auth_service"
	emailformatechecker "nex_play_auth/github.com/pkg/email_formate_checker"
	"nex_play_auth/github.com/pkg/response"
	usernamechecker "nex_play_auth/github.com/pkg/username_checker"
	"strings"
)

type Handler struct {
	authSVC *authservice.Service
}

func NewHander(authSVC *authservice.Service) *Handler {

	return &Handler{authSVC: authSVC}
}

// Register wires all auth routes onto the mux
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/signup", h.signUp)
	mux.HandleFunc("POST /api/v1/auth/signin", h.signIn)
	mux.HandleFunc("POST /api/v1/auth/verify", h.verify)
	mux.HandleFunc("POST /api/v1/auth/otp/resend", h.resentOTP)
	mux.HandleFunc("POST /api/v1/auth/password/forgot", h.forgotPassword)
	mux.HandleFunc("POST /api/v1/auth/password/reset", h.resetPassword)
	mux.HandleFunc("POST /api/v1/auth/token/refresh", h.refreshToken)
}

// Halper
func decode(r *http.Request, v any) error {

	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)

	dec.DisallowUnknownFields()

	return dec.Decode(v)
}

//Handlers

// signUp
func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) {

	var req SignUpRequest

	if err := decode(r, &req); err != nil {

		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	req.UserName = strings.TrimSpace(req.UserName)

	switch {

	case req.Email == "" || req.UserName == "" || req.Password == "":

		response.Error(w, http.StatusBadRequest, "email, username and password are required")
		return

	case !emailformatechecker.Check(req.Email):

		response.Error(w, http.StatusBadRequest, "invalid email address")
		return

	case !usernamechecker.Check(req.UserName):

		response.Error(w, http.StatusBadRequest, "username must be 3–40 characters")
		return
	}

	if err := h.authSVC.SignUp(r.Context(), req.Email, req.UserName, req.Password); err != nil {

		switch {

		case errors.Is(err, domainerrors.ErrUserAlreadyExists):
			response.Error(w, http.StatusConflict, "email or username is already taken")

		case errors.Is(err, domainerrors.ErrWeakPassword):
			response.Error(w, http.StatusBadRequest, "password must be at least 8 characters")

		default:
			response.Error(w, http.StatusInternalServerError, "could not create account")
		}

		return
	}

	response.Message(w, http.StatusCreated, "account created")
}

// verify OTP signUp
func (h *Handler) verify(w http.ResponseWriter, r *http.Request) {

	var req VerifyOTPRequest

	if err := decode(r, &req); err != nil {

		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	req.Code = strings.TrimSpace(req.Code)

	if req.Email == "" || req.Code == "" {

		response.Error(w, http.StatusBadRequest, "email and code are required")
		return
	}

	tokens, err := h.authSVC.Verify(r.Context(), req.Email, req.Code)

	if err != nil {

		switch {

		case errors.Is(err, domainerrors.ErrInvalidOTP):
			response.Error(w, http.StatusUnauthorized, "invalid code")

		case errors.Is(err, domainerrors.ErrOTPExpired):
			response.Error(w, http.StatusUnauthorized, "code has expired — please sign in again")

		case errors.Is(err, domainerrors.ErrOTPAlreadyUsed):
			response.Error(w, http.StatusUnauthorized, "code was already used")

		default:
			slog.Error("verify error", "error", err)
			response.Error(w, http.StatusInternalServerError, "could not verify code")
		}

		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

// SignIn
func (h *Handler) signIn(w http.ResponseWriter, r *http.Request) {

	var req SignInRequest

	if err := decode(r, &req); err != nil {

		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Email == "" || req.Password == "" {

		response.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	tokens, err := h.authSVC.SignIn(r.Context(), req.Email, req.Password)

	if err != nil {

		switch {

		case errors.Is(err, domainerrors.ErrInvalidCredentials):
			// Same message for wrong email and wrong password avoids enumeration
			response.Error(w, http.StatusUnauthorized, "invalid email or password")

		default:
			slog.Error("could not sign in", err)
			response.Error(w, http.StatusInternalServerError, "could not sign in")
		}
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) resentOTP(w http.ResponseWriter, r *http.Request) {

	var req ResentotpRequest

	if err := decode(r, &req); err != nil {

		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Service silently no error if the email doesn't exist
	_ = h.authSVC.ResentOTP(r.Context(), req.Email, req.Purpose)

	msg := fmt.Sprintf("a %s code has been sent successfully", req.Purpose)

	response.Message(w, http.StatusOK, msg)
}

// Forgot password
func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {

	var req ForgotPasswordRequest

	if err := decode(r, &req); err != nil {

		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Service silently no error if the email doesn't exist
	_ = h.authSVC.ForgotPassword(r.Context(), req.Email)

	response.Message(w, http.StatusOK, "if that email is registered, a reset code has been sent")
}

// Reset Password set new password (Forgot password -> resetPassword)
func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {

	var req ResetPasswordRequest

	if err := decode(r, &req); err != nil {

		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	req.Code = strings.TrimSpace(req.Code)

	if req.Email == "" || req.Code == "" || req.NewPassword == "" {

		response.Error(w, http.StatusBadRequest, "email, code and new_password are required")
		return
	}

	tokens, err := h.authSVC.ResetPassword(r.Context(), req.Email, req.Code, req.NewPassword)

	if err != nil {

		switch {

		case errors.Is(err, domainerrors.ErrInvalidOTP):
			response.Error(w, http.StatusUnauthorized, "invalid code")

		case errors.Is(err, domainerrors.ErrOTPExpired):
			response.Error(w, http.StatusUnauthorized, "code has expired — request a new one")

		case errors.Is(err, domainerrors.ErrOTPAlreadyUsed):
			response.Error(w, http.StatusUnauthorized, "code was already used")

		case errors.Is(err, domainerrors.ErrWeakPassword):
			response.Error(w, http.StatusBadRequest, "password must be at least 8 characters")

		default:
			response.Error(w, http.StatusInternalServerError, "could not reset password")
		}

		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

// Refesh Access JWT token
func (h *Handler) refreshToken(w http.ResponseWriter, r *http.Request) {

	var req RefreshTokenRequest

	if err := decode(r, &req); err != nil {

		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {

		response.Error(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	tokens, err := h.authSVC.RefreshToken(r.Context(), req.RefreshToken)

	if err != nil {

		response.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}
