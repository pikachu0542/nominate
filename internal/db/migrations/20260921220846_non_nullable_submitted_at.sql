-- +goose Up
ALTER TABLE nomination ALTER COLUMN submitted_at SET NOT NULL;

-- +goose Down
ALTER TABLE nomination ALTER COLUMN submitted_at DROP NOT NULL;