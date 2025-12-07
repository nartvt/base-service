-- name: GetUser :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT * FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (username, email, phone_number, first_name, last_name, hash_password) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetUserByUserName :one
SELECT * FROM users WHERE username = $1 AND deleted_at IS NULL;

-- name: GetUserByUsernameOrEmail :one
SELECT * FROM users WHERE (username = $1 OR email = $1) AND deleted_at IS NULL;

-- name: ValidateUserPasswordByUserName :one
-- DEPRECATED: This query has a SQL injection vulnerability. Use GetUserByUsernameOrEmail instead.
SELECT * FROM users WHERE (username = $1 OR email = $1) AND hash_password = $2 AND deleted_at IS NULL;
