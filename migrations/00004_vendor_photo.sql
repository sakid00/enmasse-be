-- +goose Up
ALTER TABLE vendors ADD COLUMN photo_url TEXT;

-- +goose Down
ALTER TABLE vendors DROP COLUMN IF EXISTS photo_url;
