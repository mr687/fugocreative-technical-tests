package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mr687/fugocreative-technical-tests/internal/model"
	"github.com/mr687/fugocreative-technical-tests/internal/repository"
	"github.com/mr687/fugocreative-technical-tests/internal/service"
)

func TestAdminLogin(t *testing.T) {
	const secret = "01234567890123456789012345678901"
	router := chi.NewRouter()
	authService := service.NewAuthService(repository.NewAdminRepository("admin", "correct-password"), secret, time.Hour)
	NewAuthHandler(authService).RegisterRoutes(router)

	t.Run("issues bearer token for valid credentials", func(t *testing.T) {
		response := postLogin(router, `{"username":"admin","password":"correct-password"}`)
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
		}
		var body model.LoginResponse
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.TokenType != "Bearer" || body.ExpiresIn != 3600 {
			t.Fatalf("unexpected login response: %+v", body)
		}
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(body.AccessToken, claims, func(token *jwt.Token) (any, error) { return []byte(secret), nil })
		if err != nil || !token.Valid || claims.Subject != "admin" {
			t.Fatalf("issued token invalid: valid=%t claims=%+v err=%v", token != nil && token.Valid, claims, err)
		}
	})

	t.Run("rejects invalid credentials generically", func(t *testing.T) {
		response := postLogin(router, `{"username":"admin","password":"wrong"}`)
		if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), "INVALID_CREDENTIALS") || strings.Contains(response.Body.String(), "wrong") {
			t.Fatalf("unexpected invalid-credentials response: status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("returns Indonesian field errors", func(t *testing.T) {
		response := postLogin(router, `{"username":"","password":""}`)
		if response.Code != http.StatusBadRequest || !strings.Contains(strings.ToLower(response.Body.String()), "wajib") {
			t.Fatalf("unexpected validation response: status=%d body=%s", response.Code, response.Body.String())
		}
	})
}

func postLogin(handler http.Handler, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
