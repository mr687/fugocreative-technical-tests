package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mr687/fugocreative-technical-tests/internal/model"
)

type AIJobRepository interface {
	ClaimNext(context.Context) (*model.AIJob, error)
	Complete(context.Context, model.AIJob, model.AIResult, error) error
	RecoverStale(context.Context, time.Duration) error
}

func NewAIJobRepository(db *sql.DB) AIJobRepository {
	return &briefRepository{db: db}
}

func (r *briefRepository) ClaimNext(ctx context.Context) (*model.AIJob, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin claim AI job: %w", err)
	}
	defer tx.Rollback()
	var job model.AIJob
	err = tx.QueryRowContext(ctx, `
		SELECT j.id::text, j.brief_id::text, b.contact_name, b.email, b.company_name,
		       b.need_description, b.budget_range, b.target_date
		FROM ai_jobs j JOIN briefs b ON b.id = j.brief_id
		WHERE j.status = 'queued'
		ORDER BY j.created_at
		FOR UPDATE OF j SKIP LOCKED
		LIMIT 1`).Scan(
		&job.ID, &job.BriefID, &job.Brief.ContactName, &job.Brief.Email, &job.Brief.CompanyName,
		&job.Brief.NeedDescription, &job.Brief.BudgetRange, &job.Brief.TargetDate,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select queued AI job: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE ai_jobs SET status = 'processing', started_at = now() WHERE id = $1`, job.ID); err != nil {
		return nil, fmt.Errorf("mark AI job processing: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE briefs SET ai_status = 'pending', ai_error = NULL WHERE id = $1`, job.BriefID); err != nil {
		return nil, fmt.Errorf("mark brief processing: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit claimed AI job: %w", err)
	}
	return &job, nil
}

func (r *briefRepository) Complete(ctx context.Context, job model.AIJob, result model.AIResult, processErr error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin complete AI job: %w", err)
	}
	defer tx.Rollback()
	if processErr != nil {
		message := processErr.Error()
		if _, err := tx.ExecContext(ctx, `UPDATE briefs SET ai_status = 'failed', ai_result = NULL, ai_error = $2, updated_at = now() WHERE id = $1`, job.BriefID, message); err != nil {
			return fmt.Errorf("store brief AI error: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE ai_jobs SET status = 'failed', error = $2, finished_at = now() WHERE id = $1`, job.ID, message); err != nil {
			return fmt.Errorf("store AI job error: %w", err)
		}
	} else {
		encoded, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("encode AI result: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE briefs SET ai_status = 'succeeded', ai_result = $2, ai_error = NULL, updated_at = now() WHERE id = $1`, job.BriefID, string(encoded)); err != nil {
			return fmt.Errorf("store brief AI result: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE ai_jobs SET status = 'succeeded', error = NULL, finished_at = now() WHERE id = $1`, job.ID); err != nil {
			return fmt.Errorf("mark AI job succeeded: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit completed AI job: %w", err)
	}
	return nil
}

func (r *briefRepository) RecoverStale(ctx context.Context, staleAfter time.Duration) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE ai_jobs SET status = 'queued', attempt = attempt + 1, started_at = NULL
		WHERE status = 'processing' AND started_at < now() - ($1 * interval '1 second')`, staleAfter.Seconds())
	if err != nil {
		return fmt.Errorf("requeue stale AI jobs: %w", err)
	}
	return nil
}
