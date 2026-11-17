package repository

import (
	"context"
	"crypto/subtle"
)

type AdminRepository interface {
	VerifyCredentials(ctx context.Context, username, password string) bool
}

type adminRepository struct {
	username string
	password string
}

func NewAdminRepository(username, password string) AdminRepository {
	return &adminRepository{username: username, password: password}
}

func (r *adminRepository) VerifyCredentials(ctx context.Context, username, password string) bool {
	if ctx.Err() != nil {
		return false
	}
	usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(r.username))
	passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(r.password))
	return usernameMatch&passwordMatch == 1
}
