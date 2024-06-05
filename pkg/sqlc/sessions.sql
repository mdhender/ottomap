-- name: CreateSession :exec
INSERT INTO sessions (sid, uid, expires_at)
VALUES (?, ?, ?);

-- name: ReadSession :one
SELECT uid, expires_at
FROM sessions
WHERE sid = ?
  AND CURRENT_TIMESTAMP < expires_at;

-- name: ReadSessions :many
SELECT sid, uid, expires_at
FROM sessions
WHERE CURRENT_TIMESTAMP < expires_at;

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