-- name: LinkFileToUser :exec
INSERT INTO files_users (file_id, user_id)
VALUES (?, ?);

-- name: GetUserFiles :many
SELECT file_id, user_id
FROM files_users
WHERE user_id = ?;

-- name: GetFileUsers :many
SELECT file_id, user_id
FROM files_users
WHERE file_id = ?;

-- name: UnlinkFileFromUser :exec
DELETE FROM files_users
WHERE file_id = ? AND user_id = ?;

-- name: UnlinkAllUserFiles :exec
DELETE FROM files_users
WHERE user_id = ?;

-- name: UnlinkAllFileUsers :exec
DELETE FROM files_users
WHERE file_id = ?;

-- name: CountUserFiles :one
SELECT COUNT(*) FROM files_users WHERE user_id = ?;
