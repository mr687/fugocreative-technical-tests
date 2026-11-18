package handler

import (
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mr687/fugocreative-technical-tests/internal/middleware"
	"github.com/mr687/fugocreative-technical-tests/internal/model"
	"github.com/mr687/fugocreative-technical-tests/internal/service"
	requestutil "github.com/mr687/fugocreative-technical-tests/pkg/request"
	"github.com/mr687/fugocreative-technical-tests/pkg/response"
)

type AdminBriefHandler interface {
	Start() chi.Router
	Path() string
}

type adminBriefHandler struct {
	service service.AdminBriefService
	secret  string
}

func NewAdminBriefHandler(adminBriefService service.AdminBriefService, jwtSecret string) AdminBriefHandler {
	return &adminBriefHandler{service: adminBriefService, secret: jwtSecret}
}

func (h *adminBriefHandler) Path() string { return "/admin/briefs" }

func (h *adminBriefHandler) Start() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(h.secret))
	router.Get("/", h.List)
	router.Get("/{id}", h.Get)
	router.Post("/{id}/reprocess", h.Reprocess)
	router.Patch("/{id}/pipeline", h.UpdatePipeline)
	return router
}

func (h *adminBriefHandler) List(w http.ResponseWriter, r *http.Request) {
	page, ok := positiveQueryInt(r, "page", 1)
	if !ok {
		invalidBriefListQuery(w)
		return
	}
	pageSize, ok := positiveQueryInt(r, "page_size", 20)
	if !ok || pageSize > 100 {
		invalidBriefListQuery(w)
		return
	}
	result, err := h.service.List(r.Context(), r.URL.Query().Get("priority"), r.URL.Query().Get("category"), page, pageSize)
	if errors.Is(err, service.ErrInvalidBriefFilter) {
		invalidBriefListQuery(w)
		return
	}
	if err != nil {
		slog.Error("list admin briefs failed", "error", err)
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Terjadi kesalahan pada server.", nil)
		return
	}
	response.Success(w, http.StatusOK, result)
}

func (h *adminBriefHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !validUUID(id) {
		adminBriefNotFound(w)
		return
	}
	result, err := h.service.Get(r.Context(), id)
	if errors.Is(err, service.ErrBriefNotFound) {
		adminBriefNotFound(w)
		return
	}
	if err != nil {
		slog.Error("get admin brief failed", "error", err)
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Terjadi kesalahan pada server.", nil)
		return
	}
	response.Success(w, http.StatusOK, result)
}

func (h *adminBriefHandler) Reprocess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !validUUID(id) {
		adminBriefNotFound(w)
		return
	}
	result, err := h.service.Reprocess(r.Context(), id)
	if errors.Is(err, service.ErrBriefNotFound) {
		adminBriefNotFound(w)
		return
	}
	if errors.Is(err, service.ErrJobAlreadyActive) {
		response.Error(w, http.StatusConflict, response.CodeJobAlreadyActive, "Brief sedang diproses.", nil)
		return
	}
	if err != nil {
		slog.Error("reprocess admin brief failed", "error", err)
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Terjadi kesalahan pada server.", nil)
		return
	}
	response.Success(w, http.StatusAccepted, result)
}

func (h *adminBriefHandler) UpdatePipeline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !validUUID(id) {
		adminBriefNotFound(w)
		return
	}
	var request model.UpdatePipelineRequest
	fieldErrors, err := requestutil.DecodeAndValidate(w, r, &request)
	if err != nil || len(fieldErrors) > 0 {
		response.Error(w, http.StatusBadRequest, response.CodeValidationError, "Periksa kembali input.", fieldErrors)
		return
	}
	request.PipelineStatus = strings.TrimSpace(request.PipelineStatus)
	result, err := h.service.UpdatePipeline(r.Context(), id, request.PipelineStatus)
	if errors.Is(err, service.ErrBriefNotFound) {
		adminBriefNotFound(w)
		return
	}
	if err != nil {
		slog.Error("update admin brief pipeline failed", "error", err)
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Terjadi kesalahan pada server.", nil)
		return
	}
	response.Success(w, http.StatusOK, result)
}

func adminBriefNotFound(w http.ResponseWriter) {
	response.Error(w, http.StatusNotFound, response.CodeNotFound, "Brief tidak ditemukan.", nil)
}

func validUUID(id string) bool {
	if len(id) != 36 || id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		return false
	}
	compact := strings.ReplaceAll(id, "-", "")
	decoded := make([]byte, 16)
	_, err := hex.Decode(decoded, []byte(compact))
	return err == nil
}
func positiveQueryInt(r *http.Request, key string, fallback int) (int, bool) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback, true
	}
	parsed, err := strconv.Atoi(value)
	return parsed, err == nil && parsed > 0
}

func invalidBriefListQuery(w http.ResponseWriter) {
	response.Error(w, http.StatusBadRequest, response.CodeInvalidFilter, "Filter atau parameter pagination tidak valid.", nil)
}
