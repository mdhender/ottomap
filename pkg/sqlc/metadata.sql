-- name: ReadMetadataPublic :one
SELECT public_path
FROM metadata;

-- name: ReadMetadataTemplates :one
SELECT templates_path
FROM metadata;

-- name: UpdateMetadataPublic :exec
UPDATE metadata
SET public_path = ?;

-- name: UpdateMetadataTemplates :exec
UPDATE metadata
SET templates_path = ?;