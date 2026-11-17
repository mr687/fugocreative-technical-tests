package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mr687/fugocreative-technical-tests/internal/model"
)

type BriefRepository interface {
	Create(context.Context, model.Brief) (string, error)
}

type briefRepository struct {
	db *sql.DB
}

func NewBriefRepository(db *sql.DB) BriefRepository {
	return &briefRepository{db: db}
}

func (r *briefRepository) Create(ctx context.Context, brief model.Brief) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin create brief transaction: %w", err)
	}
	defer tx.Rollback()

	var id string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO briefs (contact_name, email, company_name, need_description, budget_range, target_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text`,
		brief.ContactName, brief.Email, brief.CompanyName, brief.NeedDescription, brief.BudgetRange, brief.TargetDate,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert brief: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO ai_jobs (brief_id) VALUES ($1)`, id); err != nil {
		return "", fmt.Errorf("insert initial AI job: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit create brief transaction: %w", err)
	}
	return id, nil
}
