-- +goose Up
ALTER TABLE users
    ADD COLUMN email_verified_at TIMESTAMPTZ;

UPDATE users
SET email_verified_at = NOW()
WHERE password_hash IS NOT NULL
  AND email_verified_at IS NULL;

CREATE TABLE email_otps (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_canonical  CITEXT NOT NULL,
    purpose          TEXT NOT NULL CHECK (purpose IN ('signup', 'claim', 'reset', 'verify')),
    code_hash        TEXT NOT NULL,
    expires_at       TIMESTAMPTZ NOT NULL,
    consumed_at      TIMESTAMPTZ,
    attempt_count    INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_otps_lookup
    ON email_otps (email_canonical, purpose, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_email_otps_lookup;
DROP TABLE IF EXISTS email_otps;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
