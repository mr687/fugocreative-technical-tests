package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mr687/fugocreative-technical-tests/internal/middleware"
	"github.com/mr687/fugocreative-technical-tests/internal/model"
	"github.com/mr687/fugocreative-technical-tests/internal/service"
	requestutil "github.com/mr687/fugocreative-technical-tests/pkg/request"
	"github.com/mr687/fugocreative-technical-tests/pkg/response"
)

type AuthHandler interface {
	Start() chi.Router
	Path() string
	Me(http.ResponseWriter, *http.Request)
}

type authHandler struct {
	service service.AuthService
	secret  string
}

func NewAuthHandler(authService service.AuthService, jwtSecret string) AuthHandler {
	return &authHandler{service: authService, secret: jwtSecret}
}

func (h *authHandler) Path() string {
	return "/auth"
}

func (h *authHandler) Start() chi.Router {
	router := chi.NewRouter()

	router.Post("/login", h.Login)
	router.With(middleware.JWTAuth(h.secret)).Get("/me", h.Me)

	return router
}

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request model.LoginRequest
	fieldErrors, err := requestutil.DecodeAndValidate(w, r, &request)
	if err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeValidationError, "Periksa kembali input.", nil)
		return
	}
	if len(fieldErrors) > 0 {
		response.Error(w, http.StatusBadRequest, response.CodeValidationError, "Periksa kembali input.", fieldErrors)
		return
	}
	request.Username = strings.TrimSpace(request.Username)
	loginResponse, err := h.service.Login(r.Context(), request)
	if errors.Is(err, service.ErrInvalidCredentials) {
		response.Error(w, http.StatusUnauthorized, response.CodeInvalidCredentials, "Username atau password salah.", nil)
		return
	}
	if err != nil {
		slog.Error("admin login failed", "error", err)
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Terjadi kesalahan pada server.", nil)
		return
	}
	response.Success(w, http.StatusOK, loginResponse)
}

func (h *authHandler) Me(w http.ResponseWriter, r *http.Request) {
	username, ok := middleware.UsernameFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, response.CodeUnauthorized, "Autentikasi diperlukan.", nil)
		return
	}
	response.Success(w, http.StatusOK, model.AuthMeResponse{Authenticated: true, Username: username})
}
