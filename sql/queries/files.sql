-- name: GetFileByID :one
SELECT * FROM files WHERE id = ?;

-- name: ListFiles :many
SELECT * FROM files ORDER BY created_at DESC;

-- name: ListFilesByUser :many
SELECT f.*
FROM files f
JOIN files_users fu ON fu.file_id = f.id
WHERE fu.user_id = ?
ORDER BY f.created_at DESC;

-- name: CreateFile :one
INSERT INTO files (id, created_at, updated_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: DeleteFile :exec
DELETE FROM files WHERE id = ?;

-- name: DeleteFilesByUser :exec
DELETE FROM files
WHERE id IN (
    SELECT file_id FROM files_users WHERE user_id = ?
);
