-- +goose Up
ALTER TABLE artists
    ADD COLUMN province TEXT,
    ADD COLUMN province_code TEXT,
    ADD COLUMN city TEXT,
    ADD COLUMN city_code TEXT,
    ADD COLUMN district TEXT,
    ADD COLUMN district_code TEXT;

ALTER TABLE vendors
    ADD COLUMN province TEXT,
    ADD COLUMN province_code TEXT,
    ADD COLUMN city_code TEXT,
    ADD COLUMN district TEXT,
    ADD COLUMN district_code TEXT;

-- +goose Down
ALTER TABLE artists
    DROP COLUMN IF EXISTS district_code,
    DROP COLUMN IF EXISTS district,
    DROP COLUMN IF EXISTS city_code,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS province_code,
    DROP COLUMN IF EXISTS province;

ALTER TABLE vendors
    DROP COLUMN IF EXISTS district_code,
    DROP COLUMN IF EXISTS district,
    DROP COLUMN IF EXISTS city_code,
    DROP COLUMN IF EXISTS province_code,
    DROP COLUMN IF EXISTS province;
