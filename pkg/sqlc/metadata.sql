-- name: ReadMetadataInputOutputPaths :one
SELECT input_path, output_path
FROM metadata;

-- name: ReadMetadataPublic :one
SELECT public_path
FROM metadata;

-- name: ReadMetadataTemplates :one
SELECT templates_path
FROM metadata;

-- name: UpdateMetadataInputOutputPaths :exec
UPDATE metadata
SET input_path = ?1,
    output_path = ?2;

-- name: UpdateMetadataPublic :exec
UPDATE metadata
SET public_path = ?;

-- name: UpdateMetadataTemplates :exec
UPDATE metadata
SET templates_path = ?;