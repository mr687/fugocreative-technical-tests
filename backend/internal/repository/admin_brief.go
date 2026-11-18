package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mr687/fugocreative-technical-tests/internal/model"
)

var (
	ErrBriefNotFound    = errors.New("brief not found")
	ErrJobAlreadyActive = errors.New("AI job already active")
)

type AdminBriefRepository interface {
	List(context.Context, string, string, int, int) ([]model.AdminBriefListItem, int64, error)
	Get(context.Context, string) (model.AdminBriefDetail, error)
	Reprocess(context.Context, string) (string, error)
	UpdatePipeline(context.Context, string, string) (model.UpdatePipelineResponse, error)
}

type adminBriefRepository struct {
	db *sql.DB
}

func NewAdminBriefRepository(db *sql.DB) AdminBriefRepository {
	return &adminBriefRepository{db: db}
}

func (r *adminBriefRepository) List(ctx context.Context, priority, category string, limit, offset int) ([]model.AdminBriefListItem, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT count(*) FROM briefs
		WHERE ($1 = '' OR ai_result ->> 'priority' = $1)
		  AND ($2 = '' OR ai_result ->> 'service_category' = $2)`, priority, category).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin briefs: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, company_name, COALESCE(ai_result ->> 'service_category', 'Belum dikategorikan'),
		       COALESCE(ai_result ->> 'priority', ''), ai_status, pipeline_status, created_at
		FROM briefs
		WHERE ($1 = '' OR ai_result ->> 'priority' = $1)
		  AND ($2 = '' OR ai_result ->> 'service_category' = $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4`, priority, category, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin briefs: %w", err)
	}
	defer rows.Close()

	items := make([]model.AdminBriefListItem, 0)
	for rows.Next() {
		var item model.AdminBriefListItem
		if err := rows.Scan(&item.ID, &item.CompanyName, &item.ServiceCategory, &item.Priority,
			&item.AIStatus, &item.PipelineStatus, &item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan admin brief: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate admin briefs: %w", err)
	}
	return items, total, nil
}

func (r *adminBriefRepository) Get(ctx context.Context, id string) (model.AdminBriefDetail, error) {
	var brief model.AdminBriefDetail
	var aiResult []byte
	var jobID, jobStatus, jobError sql.NullString
	var attempt sql.NullInt64
	err := r.db.QueryRowContext(ctx, `
		SELECT b.id::text, b.contact_name, b.email, b.company_name, b.need_description,
		       b.budget_range, b.target_date::text, b.ai_status, b.ai_result,
		       b.pipeline_status, b.created_at, b.updated_at,
		       j.id::text, j.status, j.attempt, j.error
		FROM briefs b
		LEFT JOIN LATERAL (
			SELECT id, status, attempt, error
			FROM ai_jobs WHERE brief_id = b.id
			ORDER BY created_at DESC LIMIT 1
		) j ON true
		WHERE b.id = $1`, id).Scan(
		&brief.ID, &brief.ContactName, &brief.Email, &brief.CompanyName,
		&brief.NeedDescription, &brief.BudgetRange, &brief.TargetDate, &brief.AIStatus,
		&aiResult, &brief.PipelineStatus, &brief.CreatedAt, &brief.UpdatedAt,
		&jobID, &jobStatus, &attempt, &jobError,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.AdminBriefDetail{}, ErrBriefNotFound
	}
	if err != nil {
		return model.AdminBriefDetail{}, fmt.Errorf("get admin brief: %w", err)
	}
	if len(aiResult) > 0 {
		var result model.AIResult
		if err := json.Unmarshal(aiResult, &result); err != nil {
			return model.AdminBriefDetail{}, fmt.Errorf("decode brief AI result: %w", err)
		}
		brief.AIResult = &result
	}
	if jobID.Valid {
		job := &model.AdminAIJob{ID: jobID.String, Status: jobStatus.String, Attempt: int(attempt.Int64)}
		if jobError.Valid {
			job.Error = &jobError.String
		}
		brief.LatestJob = job
	}
	return brief, nil
}

func (r *adminBriefRepository) Reprocess(ctx context.Context, briefID string) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin brief reprocess transaction: %w", err)
	}
	defer tx.Rollback()

	var id string
	if err := tx.QueryRowContext(ctx, `SELECT id::text FROM briefs WHERE id = $1 FOR UPDATE`, briefID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrBriefNotFound
		}
		return "", fmt.Errorf("lock brief for reprocess: %w", err)
	}
	var active bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM ai_jobs WHERE brief_id = $1 AND status IN ('queued', 'processing'))`, briefID).Scan(&active); err != nil {
		return "", fmt.Errorf("check active brief job: %w", err)
	}
	if active {
		return "", ErrJobAlreadyActive
	}

	var jobID string
	if err := tx.QueryRowContext(ctx, `INSERT INTO ai_jobs (brief_id) VALUES ($1) RETURNING id::text`, briefID).Scan(&jobID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrJobAlreadyActive
		}
		return "", fmt.Errorf("create brief reprocess job: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE briefs SET ai_status = 'pending', ai_result = NULL, ai_error = NULL, updated_at = now() WHERE id = $1`, briefID); err != nil {
		return "", fmt.Errorf("reset brief AI result: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brief reprocess transaction: %w", err)
	}
	return jobID, nil
}

func (r *adminBriefRepository) UpdatePipeline(ctx context.Context, id, status string) (model.UpdatePipelineResponse, error) {
	var result model.UpdatePipelineResponse
	err := r.db.QueryRowContext(ctx, `
		UPDATE briefs SET pipeline_status = $2, updated_at = now()
		WHERE id = $1
		RETURNING id::text, pipeline_status, updated_at`, id, status).Scan(
		&result.ID, &result.PipelineStatus, &result.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.UpdatePipelineResponse{}, ErrBriefNotFound
	}
	if err != nil {
		return model.UpdatePipelineResponse{}, fmt.Errorf("update brief pipeline: %w", err)
	}
	return result, nil
}
