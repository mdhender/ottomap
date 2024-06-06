// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queues.sql

package sqlc

import (
	"context"
	"database/sql"
	"time"
)

const countQueuedByChecksum = `-- name: CountQueuedByChecksum :one
SELECT COUNT(*)
FROM report_queue_data
WHERE cksum = ?1
`

func (q *Queries) CountQueuedByChecksum(ctx context.Context, cksum string) (int64, error) {
	row := q.db.QueryRowContext(ctx, countQueuedByChecksum, cksum)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countQueuedInProgressReports = `-- name: CountQueuedInProgressReports :one
SELECT COUNT(*)
FROM report_queue
WHERE status != "completed"
`

func (q *Queries) CountQueuedInProgressReports(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countQueuedInProgressReports)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countQueuedReports = `-- name: CountQueuedReports :one
SELECT COUNT(*)
FROM report_queue
`

func (q *Queries) CountQueuedReports(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countQueuedReports)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createQueuedReport = `-- name: CreateQueuedReport :exec
INSERT INTO report_queue(qid, cid, status)
VALUES (?1, ?2, ?3)
`

type CreateQueuedReportParams struct {
	Qid    string
	Cid    string
	Status string
}

func (q *Queries) CreateQueuedReport(ctx context.Context, arg CreateQueuedReportParams) error {
	_, err := q.db.ExecContext(ctx, createQueuedReport, arg.Qid, arg.Cid, arg.Status)
	return err
}

const createQueuedReportData = `-- name: CreateQueuedReportData :exec
INSERT INTO report_queue_data(qid, name, cksum, lines)
VALUES (?1, ?2, ?3, ?4)
`

type CreateQueuedReportDataParams struct {
	Qid   string
	Name  string
	Cksum string
	Lines string
}

func (q *Queries) CreateQueuedReportData(ctx context.Context, arg CreateQueuedReportDataParams) error {
	_, err := q.db.ExecContext(ctx, createQueuedReportData,
		arg.Qid,
		arg.Name,
		arg.Cksum,
		arg.Lines,
	)
	return err
}

const readQueuedReport = `-- name: ReadQueuedReport :one
SELECT report_queue.cid,
       report_queue.status,
       report_queue.crdttm,
       report_queue.updttm,
       report_queue_data.name,
       report_queue_data.cksum
FROM report_queue
         LEFT OUTER JOIN report_queue_data ON report_queue.qid = report_queue_data.qid
WHERE report_queue.qid = ?1
  AND report_queue.cid = ?2
`

type ReadQueuedReportParams struct {
	Qid string
	Cid string
}

type ReadQueuedReportRow struct {
	Cid    string
	Status string
	Crdttm time.Time
	Updttm time.Time
	Name   sql.NullString
	Cksum  sql.NullString
}

func (q *Queries) ReadQueuedReport(ctx context.Context, arg ReadQueuedReportParams) (ReadQueuedReportRow, error) {
	row := q.db.QueryRowContext(ctx, readQueuedReport, arg.Qid, arg.Cid)
	var i ReadQueuedReportRow
	err := row.Scan(
		&i.Cid,
		&i.Status,
		&i.Crdttm,
		&i.Updttm,
		&i.Name,
		&i.Cksum,
	)
	return i, err
}

const readQueuedReports = `-- name: ReadQueuedReports :many
SELECT qid, cid, status, crdttm, updttm
FROM report_queue
WHERE cid = ?1
`

func (q *Queries) ReadQueuedReports(ctx context.Context, cid string) ([]ReportQueue, error) {
	rows, err := q.db.QueryContext(ctx, readQueuedReports, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReportQueue
	for rows.Next() {
		var i ReportQueue
		if err := rows.Scan(
			&i.Qid,
			&i.Cid,
			&i.Status,
			&i.Crdttm,
			&i.Updttm,
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
