-- name: CreateUser :one
INSERT INTO users (email, password_hash, role, handle, must_change_password)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByHandle :one
SELECT * FROM users WHERE handle = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: SetPassword :exec
UPDATE users
SET password_hash = $2, must_change_password = $3, updated_at = NOW()
WHERE id = $1;

-- name: CountUsers :one
SELECT count(*)::bigint FROM users;
