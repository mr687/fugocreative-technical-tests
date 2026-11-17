package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mr687/fugocreative-technical-tests/internal/model"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type CredentialRepository interface {
	VerifyCredentials(context.Context, string, string) bool
}

type AuthService interface {
	Login(context.Context, model.LoginRequest) (model.LoginResponse, error)
}

type authService struct {
	credentials CredentialRepository
	secret      []byte
	ttl         time.Duration
}

func NewAuthService(credentials CredentialRepository, secret string, ttl time.Duration) AuthService {
	return &authService{credentials: credentials, secret: []byte(secret), ttl: ttl}
}

func (s *authService) Login(ctx context.Context, request model.LoginRequest) (model.LoginResponse, error) {
	if !s.credentials.VerifyCredentials(ctx, request.Username, request.Password) {
		return model.LoginResponse{}, ErrInvalidCredentials
	}
	now := time.Now()
	expiresAt := now.Add(s.ttl)
	claims := jwt.RegisteredClaims{
		Issuer:    "fugocreative",
		Subject:   request.Username,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return model.LoginResponse{}, fmt.Errorf("sign access token: %w", err)
	}
	return model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.ttl.Seconds()),
	}, nil
}
