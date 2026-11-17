-- +goose Up
CREATE TABLE briefs (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    contact_name text NOT NULL CHECK (length(trim(contact_name)) > 0),
    email text NOT NULL CHECK (length(trim(email)) > 0),
    company_name text NOT NULL CHECK (length(trim(company_name)) > 0),
    need_description text NOT NULL CHECK (length(trim(need_description)) > 0),
    budget_range text NOT NULL CHECK (length(trim(budget_range)) > 0),
    target_date date NOT NULL CHECK (target_date >= CURRENT_DATE),
    ai_status text NOT NULL DEFAULT 'pending'
        CHECK (ai_status IN ('pending', 'succeeded', 'failed')),
    ai_result jsonb,
    ai_error text,
    pipeline_status text NOT NULL DEFAULT 'New'
        CHECK (pipeline_status IN ('New', 'Contacted', 'Proposal', 'Won', 'Lost')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((ai_status = 'succeeded') = (ai_result IS NOT NULL))
);
CREATE INDEX briefs_created_at_idx ON briefs (created_at DESC);

CREATE TABLE ai_jobs (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    brief_id uuid NOT NULL REFERENCES briefs(id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'queued'
        CHECK (status IN ('queued', 'processing', 'succeeded', 'failed')),
    attempt integer NOT NULL DEFAULT 1 CHECK (attempt > 0),
    error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    finished_at timestamptz
);
CREATE INDEX ai_jobs_queued_idx ON ai_jobs (created_at) WHERE status = 'queued';
CREATE UNIQUE INDEX ai_jobs_one_active_per_brief_idx ON ai_jobs (brief_id)
    WHERE status IN ('queued', 'processing');

-- +goose Down
DROP TABLE ai_jobs;
DROP TABLE briefs;
