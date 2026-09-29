-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                CITEXT NOT NULL UNIQUE,
    password_hash        TEXT,
    role                 TEXT NOT NULL CHECK (role IN ('artist', 'vendor')),
    handle               CITEXT NOT NULL UNIQUE,
    must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE artists (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    first_name              TEXT NOT NULL DEFAULT '',
    last_name               TEXT NOT NULL DEFAULT '',
    artist_name             TEXT NOT NULL DEFAULT '',
    photo_url               TEXT,
    address                 TEXT,
    website_url             TEXT,
    instagram_handle        TEXT,
    twitter_handle          TEXT,
    whatsapp                TEXT,
    imported_artist_id      UUID UNIQUE,
    imported_portal_user_id UUID,
    bio                     TEXT,
    nationality             TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vendors (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    city       TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE password_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    purpose    TEXT NOT NULL CHECK (purpose IN ('claim', 'reset')),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_tokens_user ON password_tokens (user_id);
CREATE INDEX idx_password_tokens_hash ON password_tokens (token_hash);

-- +goose Down
DROP TABLE IF EXISTS password_tokens;
DROP TABLE IF EXISTS vendors;
DROP TABLE IF EXISTS artists;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS citext;
