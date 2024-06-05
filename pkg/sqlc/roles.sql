-- name: CreateRole :exec
INSERT INTO roles(rlid)
VALUES (?);

-- name: ReadAllRoles :many
SELECT rlid
FROM roles;
