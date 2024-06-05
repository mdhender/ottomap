-- name: CreateUserRole :exec
INSERT INTO users_roles (uid, rlid, value)
VALUES (?1, ?2, ?3)
ON CONFLICT (uid, rlid) DO UPDATE SET value = ?3;

-- name: ReadUserRole :one
SELECT value
FROM users_roles
WHERE uid = ?
  AND rlid = ?;

-- name: ReadUserRoles :many
SELECT rlid, value
FROM users_roles
WHERE uid = ?;

-- name: UpdateUserRole :exec
UPDATE users_roles
SET value = ?
WHERE uid = ?
  AND rlid = ?;

-- name: DeleteUserRole :exec
DELETE
FROM users_roles
WHERE uid = ?
  AND rlid = ?;

-- name: DeleteUserRoles :exec
DELETE
FROM users_roles
WHERE uid = ?;

