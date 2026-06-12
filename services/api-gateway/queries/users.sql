-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (username, email, password, display_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET display_name = COALESCE($2, display_name),
    updated_at = NOW()
WHERE id = $1
RETURNING *;
