package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mr687/fugocreative-technical-tests/internal/model"
	"github.com/mr687/fugocreative-technical-tests/internal/service"
	requestutil "github.com/mr687/fugocreative-technical-tests/pkg/request"
	"github.com/mr687/fugocreative-technical-tests/pkg/response"
)

type BriefHandler interface {
	Start() chi.Router
	Path() string
}

type briefHandler struct {
	service service.BriefService
}

func NewBriefHandler(briefService service.BriefService) BriefHandler {
	return &briefHandler{service: briefService}
}

func (h *briefHandler) Path() string {
	return "/public/briefs"
}

func (h *briefHandler) Start() chi.Router {
	router := chi.NewRouter()
	router.Post("/", h.Create)
	return router
}

func (h *briefHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request model.CreateBriefRequest
	fieldErrors, err := requestutil.DecodeAndValidate(w, r, &request)
	if err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeValidationError, "Periksa kembali input.", nil)
		return
	}
	if len(fieldErrors) > 0 {
		response.Error(w, http.StatusBadRequest, response.CodeValidationError, "Periksa kembali input.", fieldErrors)
		return
	}
	request.ContactName = strings.TrimSpace(request.ContactName)
	request.Email = strings.TrimSpace(request.Email)
	request.CompanyName = strings.TrimSpace(request.CompanyName)
	request.NeedDescription = strings.TrimSpace(request.NeedDescription)
	request.BudgetRange = strings.TrimSpace(request.BudgetRange)

	created, err := h.service.Create(r.Context(), request)
	if err != nil {
		slog.Error("create brief failed", "error", err)
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Terjadi kesalahan pada server.", nil)
		return
	}
	response.Success(w, http.StatusAccepted, created)
}
