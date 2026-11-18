package model

import "time"

type CreateBriefRequest struct {
	ContactName     string `json:"contact_name" validate:"required,notblank,max=200"`
	Email           string `json:"email" validate:"required,email,max=320"`
	CompanyName     string `json:"company_name" validate:"required,notblank,max=200"`
	NeedDescription string `json:"need_description" validate:"required,notblank,max=5000"`
	BudgetRange     string `json:"budget_range" validate:"required,notblank,max=100"`
	TargetDate      string `json:"target_date" validate:"required,datetime=2006-01-02,date_not_past"`
}

type CreateBriefResponse struct {
	ID        string `json:"id"`
	AIStatus  string `json:"ai_status"`
	JobStatus string `json:"job_status"`
	Message   string `json:"message"`
}

type Brief struct {
	ContactName     string    `json:"contact_name"`
	Email           string    `json:"email"`
	CompanyName     string    `json:"company_name"`
	NeedDescription string    `json:"need_description"`
	BudgetRange     string    `json:"budget_range"`
	TargetDate      time.Time `json:"target_date"`
}

type AIResult struct {
	Summary         string   `json:"summary"`
	ServiceCategory string   `json:"service_category"`
	Priority        string   `json:"priority"`
	PriorityReason  string   `json:"priority_reason"`
	MissingInfo     []string `json:"missing_info"`
}

type AIJob struct {
	ID      string
	BriefID string
	Brief   Brief
}
