-- name: InsertPasswordToken :one
INSERT INTO password_tokens (user_id, token_hash, purpose, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUnusedTokenByHash :one
SELECT * FROM password_tokens
WHERE token_hash = $1
  AND used_at IS NULL
  AND expires_at > NOW();

-- name: MarkTokenUsed :exec
UPDATE password_tokens SET used_at = NOW() WHERE id = $1;

-- name: InvalidateUnusedTokens :exec
UPDATE password_tokens
SET used_at = NOW()
WHERE user_id = $1
  AND purpose = $2
  AND used_at IS NULL;
