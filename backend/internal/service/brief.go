package service

import (
	"context"
	"strings"
	"time"

	"github.com/mr687/fugocreative-technical-tests/internal/model"
	"github.com/mr687/fugocreative-technical-tests/internal/repository"
)

type BriefService interface {
	Create(context.Context, model.CreateBriefRequest) (model.CreateBriefResponse, error)
}

type briefService struct {
	repository repository.BriefRepository
	notify     func()
}

func NewBriefService(briefRepository repository.BriefRepository, notify ...func()) BriefService {
	service := &briefService{repository: briefRepository}
	if len(notify) > 0 {
		service.notify = notify[0]
	}
	return service
}

func (s *briefService) Create(ctx context.Context, request model.CreateBriefRequest) (model.CreateBriefResponse, error) {
	targetDate, err := time.Parse("2006-01-02", request.TargetDate)
	if err != nil {
		return model.CreateBriefResponse{}, err
	}
	brief := model.Brief{
		ContactName:     strings.TrimSpace(request.ContactName),
		Email:           strings.TrimSpace(request.Email),
		CompanyName:     strings.TrimSpace(request.CompanyName),
		NeedDescription: strings.TrimSpace(request.NeedDescription),
		BudgetRange:     strings.TrimSpace(request.BudgetRange),
		TargetDate:      targetDate,
	}
	id, err := s.repository.Create(ctx, brief)
	if err != nil {
		return model.CreateBriefResponse{}, err
	}
	if s.notify != nil {
		s.notify()
	}
	return model.CreateBriefResponse{ID: id, AIStatus: "pending", JobStatus: "queued", Message: "Brief berhasil dikirim"}, nil
}
