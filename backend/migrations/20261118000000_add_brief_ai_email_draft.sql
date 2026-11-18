-- +goose Up
ALTER TABLE briefs ADD COLUMN ai_email_draft text;

-- +goose Down
ALTER TABLE briefs DROP COLUMN ai_email_draft;
