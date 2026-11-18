package model

type LoginRequest struct {
	Username string `json:"username" validate:"required,notblank"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type AuthMeResponse struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username"`
}
