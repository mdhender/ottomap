-- name: InsertInput :one
INSERT INTO input (path, name, cksum)
VALUES (?1, ?2, ?3)
RETURNING id;

-- name: InsertInputLine :exec
INSERT INTO input_lines (iid, sect_no, line_no, line)
VALUES (?1, ?2, ?3, ?4);

-- name: ReadAllInputMetadata :many
SELECT id, path, name, cksum, status, crdttm
FROM input;

-- name: ReadInputLines :many
SELECT input_lines.line_no
     , input_lines.line
FROM input,
     input_lines
WHERE input.id = ?1
  AND input.status = 'parsing'
  AND input_lines.sect_no = ?2
  AND input_lines.iid = input.id
ORDER BY input_lines.line_no;

-- name: ReadInputSections :many
SELECT DISTINCT input.id, input_lines.sect_no
FROM input,
     input_lines
WHERE input.id = ?1
  AND input.status = 'parsing'
  AND input_lines.iid = input.id
ORDER BY input.id, input_lines.sect_no;

-- name: ReadInputMetadata :one
SELECT path, name, cksum, crdttm
FROM input
WHERE id = ?1;

-- name: ReadInputMetadataByChecksum :many
SELECT id, path, name, crdttm
FROM input
WHERE cksum = ?1;

-- name: ReadPendingInputMetadata :many
SELECT id, path, name, cksum, crdttm
FROM input
WHERE status = 'pending'
ORDER BY crdttm DESC;

-- name: UpdateInputStatus :exec
UPDATE input
SET status =?3,
    updttm = CURRENT_TIMESTAMP
WHERE id = ?1
  AND status = ?2;

-- name: CreateLogMessage :exec
INSERT INTO log_messages (arg_1, arg_2, arg_3, message)
VALUES (?1, ?2, ?3, ?4);