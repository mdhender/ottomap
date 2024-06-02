-- name: GetSessions :many
SELECT sid, uid, expires_at
FROM sessions
WHERE CURRENT_TIMESTAMP < expires_at;

-- name: GetSession :one
SELECT uid, expires_at
FROM sessions
WHERE sid = ?
  AND CURRENT_TIMESTAMP < expires_at;

-- name: CreateSession :exec
INSERT INTO sessions (sid, uid, expires_at)
VALUES (?, ?, ?);

-- name: DeleteSession :exec
DELETE
FROM sessions
WHERE sid = ?;

-- name: DeleteExpiredSessions :exec
DELETE
FROM sessions
WHERE CURRENT_TIMESTAMP < expires_at;

-- name: DeleteUserSessions :exec
DELETE
FROM sessions
WHERE uid = ?;