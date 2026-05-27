package auth

type SignUpRequest struct {
	Email    string `json:"email"`
	UserName string `json:"name"`
	Password string `json:"password"`
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	Code  string `json:"otp"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResentotpRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"otp"`
	NewPassword string `json:"new_password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
