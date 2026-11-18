package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mr687/fugocreative-technical-tests/internal/model"
	"github.com/mr687/fugocreative-technical-tests/internal/service"
)

const adminTestSecret = "01234567890123456789012345678901"
const adminTestBriefID = "01900000-0000-7000-8000-000000000000"

type adminBriefServiceStub struct {
	list      func(context.Context, string, string, int, int) (model.AdminBriefListResponse, error)
	get       func(context.Context, string) (model.AdminBriefDetail, error)
	reprocess func(context.Context, string) (model.ReprocessBriefResponse, error)
	pipeline  func(context.Context, string, string) (model.UpdatePipelineResponse, error)
}

func (s adminBriefServiceStub) List(ctx context.Context, priority, category string, page, pageSize int) (model.AdminBriefListResponse, error) {
	return s.list(ctx, priority, category, page, pageSize)
}
func (s adminBriefServiceStub) Get(ctx context.Context, id string) (model.AdminBriefDetail, error) {
	return s.get(ctx, id)
}
func (s adminBriefServiceStub) Reprocess(ctx context.Context, id string) (model.ReprocessBriefResponse, error) {
	return s.reprocess(ctx, id)
}
func (s adminBriefServiceStub) UpdatePipeline(ctx context.Context, id, status string) (model.UpdatePipelineResponse, error) {
	return s.pipeline(ctx, id, status)
}

func TestAdminBriefRoutesRequireJWT(t *testing.T) {
	stub := adminBriefServiceStub{
		list: func(context.Context, string, string, int, int) (model.AdminBriefListResponse, error) {
			return model.AdminBriefListResponse{}, nil
		},
		get: func(context.Context, string) (model.AdminBriefDetail, error) { return model.AdminBriefDetail{}, nil },
		reprocess: func(context.Context, string) (model.ReprocessBriefResponse, error) {
			return model.ReprocessBriefResponse{}, nil
		},
		pipeline: func(context.Context, string, string) (model.UpdatePipelineResponse, error) {
			return model.UpdatePipelineResponse{}, nil
		},
	}
	router := chi.NewRouter()
	handler := NewAdminBriefHandler(stub, adminTestSecret)
	router.Mount("/api"+handler.Path(), handler.Start())

	for _, test := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/admin/briefs", ""},
		{http.MethodGet, "/api/admin/briefs/" + adminTestBriefID, ""},
		{http.MethodPost, "/api/admin/briefs/" + adminTestBriefID + "/reprocess", `{}`},
		{http.MethodPatch, "/api/admin/briefs/" + adminTestBriefID + "/pipeline", `{"pipeline_status":"Won"}`},
	} {
		request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", test.method, test.path, response.Code)
		}
	}
}

func TestAdminBriefHandlers(t *testing.T) {
	var gotPriority, gotCategory, gotPipeline string
	var gotPage, gotPageSize int
	stub := adminBriefServiceStub{
		list: func(_ context.Context, priority, category string, page, pageSize int) (model.AdminBriefListResponse, error) {
			gotPriority, gotCategory, gotPage, gotPageSize = priority, category, page, pageSize
			if priority == "bad" {
				return model.AdminBriefListResponse{}, service.ErrInvalidBriefFilter
			}
			return model.AdminBriefListResponse{
				Items: []model.AdminBriefListItem{{ID: adminTestBriefID, CompanyName: "Studio"}},
				Page:  page, PageSize: pageSize, Total: 21, TotalPages: 3,
			}, nil
		},
		get: func(_ context.Context, id string) (model.AdminBriefDetail, error) {
			return model.AdminBriefDetail{ID: id, TargetDate: "2026-12-20"}, nil
		},
		reprocess: func(_ context.Context, id string) (model.ReprocessBriefResponse, error) {
			return model.ReprocessBriefResponse{JobID: "01900000-0000-7000-8000-000000000001", AIStatus: "pending", JobStatus: "queued"}, nil
		},
		pipeline: func(_ context.Context, id, status string) (model.UpdatePipelineResponse, error) {
			gotPipeline = status
			return model.UpdatePipelineResponse{ID: id, PipelineStatus: status}, nil
		},
	}
	router := chi.NewRouter()
	handler := NewAdminBriefHandler(stub, adminTestSecret)
	router.Mount("/api"+handler.Path(), handler.Start())
	token := adminTestToken(t)

	t.Run("lists with combined filters and pagination", func(t *testing.T) {
		response := adminRequest(router, token, http.MethodGet, "/api/admin/briefs?priority=High&category=Website&page=2&page_size=10", "")
		if response.Code != http.StatusOK || gotPriority != "High" || gotCategory != "Website" || gotPage != 2 || gotPageSize != 10 {
			t.Fatalf("status=%d filters=(%q,%q) page=(%d,%d) body=%s", response.Code, gotPriority, gotCategory, gotPage, gotPageSize, response.Body.String())
		}
		var body model.AdminBriefListResponse
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil || len(body.Items) != 1 ||
			body.Page != 2 || body.PageSize != 10 || body.Total != 21 || body.TotalPages != 3 {
			t.Fatalf("unexpected list response: %+v, err=%v", body, err)
		}
	})
	t.Run("defaults pagination", func(t *testing.T) {
		response := adminRequest(router, token, http.MethodGet, "/api/admin/briefs", "")
		if response.Code != http.StatusOK || gotPage != 1 || gotPageSize != 20 {
			t.Fatalf("status=%d page=(%d,%d) body=%s", response.Code, gotPage, gotPageSize, response.Body.String())
		}
	})

	t.Run("rejects invalid filter", func(t *testing.T) {
		response := adminRequest(router, token, http.MethodGet, "/api/admin/briefs?priority=bad", "")
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "INVALID_FILTER") {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})
	t.Run("rejects invalid pagination parameters", func(t *testing.T) {
		for _, query := range []string{"page=0", "page=invalid", "page_size=101"} {
			response := adminRequest(router, token, http.MethodGet, "/api/admin/briefs?"+query, "")
			if response.Code != http.StatusBadRequest {
				t.Errorf("query %q status=%d body=%s, want 400", query, response.Code, response.Body.String())
			}
		}
	})

	t.Run("validates and updates pipeline", func(t *testing.T) {
		invalid := adminRequest(router, token, http.MethodPatch, "/api/admin/briefs/"+adminTestBriefID+"/pipeline", `{"pipeline_status":"Unknown"}`)
		if invalid.Code != http.StatusBadRequest {
			t.Fatalf("invalid status = %d body=%s", invalid.Code, invalid.Body.String())
		}
		response := adminRequest(router, token, http.MethodPatch, "/api/admin/briefs/"+adminTestBriefID+"/pipeline", `{"pipeline_status":"Contacted"}`)
		if response.Code != http.StatusOK || gotPipeline != "Contacted" {
			t.Fatalf("status=%d pipeline=%q body=%s", response.Code, gotPipeline, response.Body.String())
		}
	})

	t.Run("reprocesses and returns accepted", func(t *testing.T) {
		response := adminRequest(router, token, http.MethodPost, "/api/admin/briefs/"+adminTestBriefID+"/reprocess", `{}`)
		if response.Code != http.StatusAccepted || !strings.Contains(response.Body.String(), `"job_status":"queued"`) {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	})
}

func adminTestToken(t *testing.T) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "fugocreative", Subject: "admin", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte(adminTestSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func adminRequest(router http.Handler, token, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
