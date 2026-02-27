-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, display_name, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
