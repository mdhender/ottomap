// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: input.sql

package sqlc

import (
	"context"
	"time"
)

const createLogMessage = `-- name: CreateLogMessage :exec
INSERT INTO log_messages (arg_1, arg_2, arg_3, message)
VALUES (?1, ?2, ?3, ?4)
`

type CreateLogMessageParams struct {
	Arg1    string
	Arg2    string
	Arg3    string
	Message string
}

func (q *Queries) CreateLogMessage(ctx context.Context, arg CreateLogMessageParams) error {
	_, err := q.db.ExecContext(ctx, createLogMessage,
		arg.Arg1,
		arg.Arg2,
		arg.Arg3,
		arg.Message,
	)
	return err
}

const insertInput = `-- name: InsertInput :one
INSERT INTO input (path, name, cksum)
VALUES (?1, ?2, ?3)
RETURNING id
`

type InsertInputParams struct {
	Path  string
	Name  string
	Cksum string
}

func (q *Queries) InsertInput(ctx context.Context, arg InsertInputParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, insertInput, arg.Path, arg.Name, arg.Cksum)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const insertInputLine = `-- name: InsertInputLine :exec
INSERT INTO input_lines (iid, sect_no, line_no, line)
VALUES (?1, ?2, ?3, ?4)
`

type InsertInputLineParams struct {
	Iid    int64
	SectNo int64
	LineNo int64
	Line   string
}

func (q *Queries) InsertInputLine(ctx context.Context, arg InsertInputLineParams) error {
	_, err := q.db.ExecContext(ctx, insertInputLine,
		arg.Iid,
		arg.SectNo,
		arg.LineNo,
		arg.Line,
	)
	return err
}

const readAllInputMetadata = `-- name: ReadAllInputMetadata :many
SELECT id, path, name, cksum, status, crdttm
FROM input
`

type ReadAllInputMetadataRow struct {
	ID     int64
	Path   string
	Name   string
	Cksum  string
	Status string
	Crdttm time.Time
}

func (q *Queries) ReadAllInputMetadata(ctx context.Context) ([]ReadAllInputMetadataRow, error) {
	rows, err := q.db.QueryContext(ctx, readAllInputMetadata)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadAllInputMetadataRow
	for rows.Next() {
		var i ReadAllInputMetadataRow
		if err := rows.Scan(
			&i.ID,
			&i.Path,
			&i.Name,
			&i.Cksum,
			&i.Status,
			&i.Crdttm,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readInputLines = `-- name: ReadInputLines :many
SELECT input_lines.line_no
     , input_lines.line
FROM input,
     input_lines
WHERE input.id = ?1
  AND input.status = 'parsing'
  AND input_lines.sect_no = ?2
  AND input_lines.iid = input.id
ORDER BY input_lines.line_no
`

type ReadInputLinesParams struct {
	ID     int64
	SectNo int64
}

type ReadInputLinesRow struct {
	LineNo int64
	Line   string
}

func (q *Queries) ReadInputLines(ctx context.Context, arg ReadInputLinesParams) ([]ReadInputLinesRow, error) {
	rows, err := q.db.QueryContext(ctx, readInputLines, arg.ID, arg.SectNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadInputLinesRow
	for rows.Next() {
		var i ReadInputLinesRow
		if err := rows.Scan(&i.LineNo, &i.Line); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readInputMetadata = `-- name: ReadInputMetadata :one
SELECT path, name, cksum, crdttm
FROM input
WHERE id = ?1
`

type ReadInputMetadataRow struct {
	Path   string
	Name   string
	Cksum  string
	Crdttm time.Time
}

func (q *Queries) ReadInputMetadata(ctx context.Context, id int64) (ReadInputMetadataRow, error) {
	row := q.db.QueryRowContext(ctx, readInputMetadata, id)
	var i ReadInputMetadataRow
	err := row.Scan(
		&i.Path,
		&i.Name,
		&i.Cksum,
		&i.Crdttm,
	)
	return i, err
}

const readInputMetadataByChecksum = `-- name: ReadInputMetadataByChecksum :many
SELECT id, path, name, crdttm
FROM input
WHERE cksum = ?1
`

type ReadInputMetadataByChecksumRow struct {
	ID     int64
	Path   string
	Name   string
	Crdttm time.Time
}

func (q *Queries) ReadInputMetadataByChecksum(ctx context.Context, cksum string) ([]ReadInputMetadataByChecksumRow, error) {
	rows, err := q.db.QueryContext(ctx, readInputMetadataByChecksum, cksum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadInputMetadataByChecksumRow
	for rows.Next() {
		var i ReadInputMetadataByChecksumRow
		if err := rows.Scan(
			&i.ID,
			&i.Path,
			&i.Name,
			&i.Crdttm,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readInputSections = `-- name: ReadInputSections :many
SELECT DISTINCT input.id, input_lines.sect_no
FROM input,
     input_lines
WHERE input.id = ?1
  AND input.status = 'parsing'
  AND input_lines.iid = input.id
ORDER BY input.id, input_lines.sect_no
`

type ReadInputSectionsRow struct {
	ID     int64
	SectNo int64
}

func (q *Queries) ReadInputSections(ctx context.Context, id int64) ([]ReadInputSectionsRow, error) {
	rows, err := q.db.QueryContext(ctx, readInputSections, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadInputSectionsRow
	for rows.Next() {
		var i ReadInputSectionsRow
		if err := rows.Scan(&i.ID, &i.SectNo); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readPendingInputMetadata = `-- name: ReadPendingInputMetadata :many
SELECT id, path, name, cksum, crdttm
FROM input
WHERE status = 'pending'
ORDER BY crdttm DESC
`

type ReadPendingInputMetadataRow struct {
	ID     int64
	Path   string
	Name   string
	Cksum  string
	Crdttm time.Time
}

func (q *Queries) ReadPendingInputMetadata(ctx context.Context) ([]ReadPendingInputMetadataRow, error) {
	rows, err := q.db.QueryContext(ctx, readPendingInputMetadata)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadPendingInputMetadataRow
	for rows.Next() {
		var i ReadPendingInputMetadataRow
		if err := rows.Scan(
			&i.ID,
			&i.Path,
			&i.Name,
			&i.Cksum,
			&i.Crdttm,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateInputStatus = `-- name: UpdateInputStatus :exec
UPDATE input
SET status =?3,
    updttm = CURRENT_TIMESTAMP
WHERE id = ?1
  AND status = ?2
`

type UpdateInputStatusParams struct {
	ID       int64
	Status   string
	Status_2 string
}

func (q *Queries) UpdateInputStatus(ctx context.Context, arg UpdateInputStatusParams) error {
	_, err := q.db.ExecContext(ctx, updateInputStatus, arg.ID, arg.Status, arg.Status_2)
	return err
}
