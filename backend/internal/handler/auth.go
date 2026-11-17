package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mr687/fugocreative-technical-tests/internal/model"
	"github.com/mr687/fugocreative-technical-tests/internal/service"
	requestutil "github.com/mr687/fugocreative-technical-tests/pkg/request"
	"github.com/mr687/fugocreative-technical-tests/pkg/response"
)

type AuthHandler interface {
	Start() chi.Router
	Path() string
}

type authHandler struct {
	service service.AuthService
}

func NewAuthHandler(authService service.AuthService) AuthHandler {
	return &authHandler{service: authService}
}

func (h *authHandler) Path() string {
	return "/auth"
}

func (h *authHandler) Start() chi.Router {
	router := chi.NewRouter()

	router.Post("/login", h.Login)

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
