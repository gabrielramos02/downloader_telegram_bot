-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: ListUsersWithSubscription :many
SELECT * FROM users WHERE subscription_id IS NOT NULL;

-- name: CreateUser :exec
INSERT OR IGNORE INTO users (id, created_at, updated_at, subscription_id)
VALUES (?, ?, ?, ?);

-- name: UpdateUser :exec
UPDATE users
SET updated_at = ?,
    subscription_id = ?
WHERE id = ?;

-- name: UpdateUserSubscription :exec
UPDATE users
SET subscription_id = ?,
    updated_at = ?
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;
