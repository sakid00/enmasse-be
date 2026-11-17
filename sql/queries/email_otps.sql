-- name: CreateEmailOTP :one
INSERT INTO email_otps (
    email_canonical,
    purpose,
    code_hash,
    expires_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetLatestEmailOTP :one
SELECT * FROM email_otps
WHERE email_canonical = $1 AND purpose = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: IncrementEmailOTPAttempts :one
UPDATE email_otps
SET attempt_count = attempt_count + 1
WHERE id = $1
RETURNING *;

-- name: ConsumeEmailOTP :exec
UPDATE email_otps
SET consumed_at = NOW()
WHERE id = $1 AND consumed_at IS NULL;
