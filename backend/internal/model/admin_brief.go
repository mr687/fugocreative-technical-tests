package model

import "time"

type AdminBriefListItem struct {
	ID              string    `json:"id"`
	CompanyName     string    `json:"company_name"`
	ServiceCategory string    `json:"service_category"`
	Priority        string    `json:"priority"`
	AIStatus        string    `json:"ai_status"`
	PipelineStatus  string    `json:"pipeline_status"`
	CreatedAt       time.Time `json:"created_at"`
}

type AdminBriefListResponse struct {
	Items      []AdminBriefListItem `json:"items"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	Total      int64                `json:"total"`
	TotalPages int64                `json:"total_pages"`
}

type AdminBriefDetail struct {
	ID              string      `json:"id"`
	ContactName     string      `json:"contact_name"`
	Email           string      `json:"email"`
	CompanyName     string      `json:"company_name"`
	NeedDescription string      `json:"need_description"`
	BudgetRange     string      `json:"budget_range"`
	TargetDate      string      `json:"target_date"`
	AIStatus        string      `json:"ai_status"`
	AIResult        *AIResult   `json:"ai_result"`
	PipelineStatus  string      `json:"pipeline_status"`
	LatestJob       *AdminAIJob `json:"latest_job"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type AdminAIJob struct {
	ID      string  `json:"id"`
	Status  string  `json:"status"`
	Attempt int     `json:"attempt"`
	Error   *string `json:"error"`
}

type ReprocessBriefResponse struct {
	JobID     string `json:"job_id"`
	AIStatus  string `json:"ai_status"`
	JobStatus string `json:"job_status"`
}

type UpdatePipelineRequest struct {
	PipelineStatus string `json:"pipeline_status" validate:"required,oneof=New Contacted Proposal Won Lost"`
}

type UpdatePipelineResponse struct {
	ID             string    `json:"id"`
	PipelineStatus string    `json:"pipeline_status"`
	UpdatedAt      time.Time `json:"updated_at"`
}
