package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mr687/fugocreative-technical-tests/pkg/response"
)

type contextKey uint8

const usernameKey contextKey = iota

func JWTAuth(secret string) func(http.Handler) http.Handler {
	key := []byte(secret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				unauthorized(w)
				return
			}
			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}
				return key, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer("fugocreative"), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
			if err != nil || token == nil || !token.Valid || strings.TrimSpace(claims.Subject) == "" {
				unauthorized(w)
				return
			}
			ctx := context.WithValue(r.Context(), usernameKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok && username != ""
}

func unauthorized(w http.ResponseWriter) {
	response.Error(w, http.StatusUnauthorized, response.CodeUnauthorized, "Autentikasi diperlukan.", nil)
}
