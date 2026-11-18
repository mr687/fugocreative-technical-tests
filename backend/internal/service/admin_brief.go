package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mr687/fugocreative-technical-tests/internal/model"
	"github.com/mr687/fugocreative-technical-tests/internal/repository"
)

var (
	ErrInvalidBriefFilter = errors.New("invalid brief filter")
	ErrBriefNotFound      = repository.ErrBriefNotFound
	ErrJobAlreadyActive   = repository.ErrJobAlreadyActive
	ErrAIContextNotReady  = repository.ErrAIContextNotReady
	ErrEmailDraftFailed   = errors.New("email draft generation failed")
)

type AdminBriefService interface {
	List(context.Context, string, string, int, int) (model.AdminBriefListResponse, error)
	Get(context.Context, string) (model.AdminBriefDetail, error)
	Reprocess(context.Context, string) (model.ReprocessBriefResponse, error)
	GenerateEmailDraft(context.Context, string) (model.GenerateEmailDraftResponse, error)
	UpdatePipeline(context.Context, string, string) (model.UpdatePipelineResponse, error)
}

type adminBriefService struct {
	repository repository.AdminBriefRepository
	wake       func()
	generator  EmailDraftGenerator
	timeout    time.Duration
}

type EmailDraftGenerator interface {
	GenerateEmailDraft(context.Context, model.EmailDraftContext) (model.EmailDraft, error)
}

func NewAdminBriefService(repo repository.AdminBriefRepository, wake func(), generator EmailDraftGenerator, timeout time.Duration) AdminBriefService {
	return &adminBriefService{repository: repo, wake: wake, generator: generator, timeout: timeout}
}

func (s *adminBriefService) List(ctx context.Context, priority, category string, page, pageSize int) (model.AdminBriefListResponse, error) {
	if priority != "" && priority != "High" && priority != "Medium" && priority != "Low" {
		return model.AdminBriefListResponse{}, ErrInvalidBriefFilter
	}
	if category != "" && category != "Branding" && category != "Website" && category != "Mobile App" && category != "AI Automation" && category != "Social Media" && category != "Lainnya" {
		return model.AdminBriefListResponse{}, ErrInvalidBriefFilter
	}
	if page < 1 || pageSize < 1 || pageSize > 100 || page > int(^uint(0)>>1)/pageSize {
		return model.AdminBriefListResponse{}, ErrInvalidBriefFilter
	}
	items, total, err := s.repository.List(ctx, priority, category, pageSize, (page-1)*pageSize)
	if err != nil {
		return model.AdminBriefListResponse{}, err
	}
	totalPages := total / int64(pageSize)
	if total%int64(pageSize) != 0 {
		totalPages++
	}
	return model.AdminBriefListResponse{Items: items, Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}, nil
}

func (s *adminBriefService) Get(ctx context.Context, id string) (model.AdminBriefDetail, error) {
	return s.repository.Get(ctx, id)
}

func (s *adminBriefService) Reprocess(ctx context.Context, id string) (model.ReprocessBriefResponse, error) {
	jobID, err := s.repository.Reprocess(ctx, id)
	if err != nil {
		return model.ReprocessBriefResponse{}, err
	}
	if s.wake != nil {
		s.wake()
	}
	return model.ReprocessBriefResponse{JobID: jobID, AIStatus: "pending", JobStatus: "queued"}, nil
}

func (s *adminBriefService) GenerateEmailDraft(ctx context.Context, id string) (model.GenerateEmailDraftResponse, error) {
	draftContext, err := s.repository.GetEmailDraftContext(ctx, id)
	if err != nil {
		return model.GenerateEmailDraftResponse{}, err
	}
	if strings.TrimSpace(draftContext.Summary) == "" || draftContext.MissingInfo == nil {
		return model.GenerateEmailDraftResponse{}, ErrAIContextNotReady
	}
	if s.generator == nil {
		return model.GenerateEmailDraftResponse{}, ErrEmailDraftFailed
	}
	generationCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	draft, err := s.generator.GenerateEmailDraft(generationCtx, draftContext)
	if err != nil || strings.TrimSpace(draft.Subject) == "" || strings.TrimSpace(draft.Body) == "" {
		return model.GenerateEmailDraftResponse{}, errors.Join(ErrEmailDraftFailed, err)
	}
	if err := s.repository.SaveEmailDraft(ctx, id, draft); err != nil {
		return model.GenerateEmailDraftResponse{}, err
	}
	return model.GenerateEmailDraftResponse{ID: id, AIEmailDraft: draft}, nil
}

func (s *adminBriefService) UpdatePipeline(ctx context.Context, id, status string) (model.UpdatePipelineResponse, error) {
	return s.repository.UpdatePipeline(ctx, id, status)
}
