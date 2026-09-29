-- +goose Up
ALTER TABLE vendors ADD COLUMN bio TEXT;

-- +goose Down
ALTER TABLE vendors DROP COLUMN IF EXISTS bio;
