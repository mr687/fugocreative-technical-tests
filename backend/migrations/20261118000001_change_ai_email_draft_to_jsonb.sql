-- +goose Up
ALTER TABLE briefs
    ALTER COLUMN ai_email_draft TYPE jsonb
    USING CASE
        WHEN ai_email_draft IS NULL THEN NULL
        ELSE jsonb_build_object('subject', '', 'body', ai_email_draft)
    END;

-- +goose Down
ALTER TABLE briefs
    ALTER COLUMN ai_email_draft TYPE text
    USING CASE
        WHEN ai_email_draft IS NULL THEN NULL
        ELSE concat_ws(E'\n\n', ai_email_draft ->> 'subject', ai_email_draft ->> 'body')
    END;
