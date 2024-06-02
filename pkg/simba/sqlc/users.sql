-- name: CreateUser :exec
INSERT INTO users (uid, username, email, hashed_password, clan)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteUser :exec
DELETE
FROM users
WHERE uid = ?;

-- name: ReadUser :one
SELECT username, email, clan
FROM users
WHERE uid = ?;

-- name: ReadUserAuthenticationData :one
SELECT uid, hashed_password
FROM users
WHERE username = ?;

-- name: ReadUsers :many
SELECT uid, username, email, clan
FROM users
ORDER BY username;

