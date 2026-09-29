-- +goose Up
ALTER TABLE vendors
    ADD COLUMN pic_name TEXT,
    ADD COLUMN whatsapp TEXT,
    ADD COLUMN address TEXT;

-- +goose Down
ALTER TABLE vendors
    DROP COLUMN IF EXISTS address,
    DROP COLUMN IF EXISTS whatsapp,
    DROP COLUMN IF EXISTS pic_name;
