-- +goose Up
ALTER TABLE photos ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';

-- +goose Down
ALTER TABLE photos DROP COLUMN status;
