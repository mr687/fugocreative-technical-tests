package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mr687/fugocreative-technical-tests/internal/model"
)

type briefServiceFunc func(context.Context, model.CreateBriefRequest) (model.CreateBriefResponse, error)

func (f briefServiceFunc) Create(ctx context.Context, request model.CreateBriefRequest) (model.CreateBriefResponse, error) {
	return f(ctx, request)
}

func TestPublicBriefCreate(t *testing.T) {
	serviceCalled := false
	router := chi.NewRouter()
	handler := NewBriefHandler(briefServiceFunc(func(_ context.Context, request model.CreateBriefRequest) (model.CreateBriefResponse, error) {
		serviceCalled = true
		if request.Email != "ayu@example.com" {
			t.Errorf("email = %q, want trimmed email", request.Email)
		}
		return model.CreateBriefResponse{ID: "01900000-0000-7000-8000-000000000000", AIStatus: "pending", JobStatus: "queued", Message: "Brief berhasil dikirim"}, nil
	}))
	router.Mount(handler.Path(), handler.Start())

	t.Run("rejects past target date before persistence", func(t *testing.T) {
		serviceCalled = false
		response := postBrief(router, `{"contact_name":"Ayu","email":"ayu@example.com","company_name":"Studio","need_description":"Website","budget_range":"Rp10–25 juta","target_date":"2000-01-01"}`)
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "target_date") {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		if serviceCalled {
			t.Fatal("service called for invalid brief")
		}
	})
	t.Run("creates brief and returns accepted response", func(t *testing.T) {
		serviceCalled = false
		response := postBrief(router, `{"contact_name":" Ayu ","email":"ayu@example.com","company_name":"Studio","need_description":"Website","budget_range":"Rp10–25 juta","target_date":"2099-12-20"}`)
		if response.Code != http.StatusAccepted || !serviceCalled {
			t.Fatalf("status=%d serviceCalled=%t body=%s", response.Code, serviceCalled, response.Body.String())
		}
		var body model.CreateBriefResponse
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.AIStatus != "pending" || body.JobStatus != "queued" {
			t.Fatalf("unexpected response: %+v", body)
		}
	})
}

func postBrief(handler http.Handler, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/public/briefs", strings.NewReader(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
