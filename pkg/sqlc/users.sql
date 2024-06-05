-- name: CreateUser :exec
INSERT INTO users (uid, username, email, hashed_password)
VALUES (?, ?, ?, ?);

-- name: DeleteUser :exec
DELETE
FROM users
WHERE uid = ?;

-- name: ReadUser :one
SELECT username, email
FROM users
WHERE uid = ?;

-- name: ReadUserAuthData :one
SELECT uid, hashed_password
FROM users
WHERE username = ?;

-- name: ReadUserByEmail :one
SELECT username
FROM users
WHERE email = ?;


-- name: ReadUsers :many
SELECT users.uid, username, email
FROM users
WHERE users.uid = ?
ORDER BY username;

