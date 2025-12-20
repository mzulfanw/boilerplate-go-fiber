package auth

type LoginRequest struct {
	Email    string `json:"email" validate:"required,notblank"`
	Password string `json:"password" validate:"required,notblank"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,notblank"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,notblank"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required,notblank"`
	Password string `json:"password" validate:"required,notblank,min=8"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}
