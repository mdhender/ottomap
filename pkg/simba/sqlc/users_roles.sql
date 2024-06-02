-- name: CreateUserRole :exec
INSERT INTO users_roles (uid, rid, value)
VALUES (?, ?, ?)
ON CONFLICT (uid, rid) DO UPDATE SET value = ?;

-- name: ReadUserRole :one
SELECT value
FROM users_roles
WHERE uid = ?
  AND rid = ?;

-- name: ReadUserRoles :many
SELECT rid, value
FROM users_roles
WHERE uid = ?;

-- name: UpdateUserRole :exec
UPDATE users_roles
SET value = ?
WHERE uid = ?
  AND rid = ?;

-- name: DeleteUserRole :exec
DELETE
FROM users_roles
WHERE uid = ?
  AND rid = ?;

-- name: DeleteUserRoles :exec
DELETE
FROM users_roles
WHERE uid = ?;

