// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: sessions.sql

package sqlc

import (
	"context"
	"time"
)

const createSession = `-- name: CreateSession :exec
INSERT INTO sessions (sid, uid, expires_at)
VALUES (?, ?, ?)
`

type CreateSessionParams struct {
	Sid       string
	Uid       string
	ExpiresAt time.Time
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) error {
	_, err := q.db.ExecContext(ctx, createSession, arg.Sid, arg.Uid, arg.ExpiresAt)
	return err
}

const deleteExpiredSessions = `-- name: DeleteExpiredSessions :exec
DELETE
FROM sessions
WHERE CURRENT_TIMESTAMP < expires_at
`

func (q *Queries) DeleteExpiredSessions(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteExpiredSessions)
	return err
}

const deleteSession = `-- name: DeleteSession :exec
DELETE
FROM sessions
WHERE sid = ?
`

func (q *Queries) DeleteSession(ctx context.Context, sid string) error {
	_, err := q.db.ExecContext(ctx, deleteSession, sid)
	return err
}

const deleteUserSessions = `-- name: DeleteUserSessions :exec
DELETE
FROM sessions
WHERE uid = ?
`

func (q *Queries) DeleteUserSessions(ctx context.Context, uid string) error {
	_, err := q.db.ExecContext(ctx, deleteUserSessions, uid)
	return err
}

const getSession = `-- name: GetSession :one
SELECT uid, expires_at
FROM sessions
WHERE sid = ?
  AND CURRENT_TIMESTAMP < expires_at
`

type GetSessionRow struct {
	Uid       string
	ExpiresAt time.Time
}

func (q *Queries) GetSession(ctx context.Context, sid string) (GetSessionRow, error) {
	row := q.db.QueryRowContext(ctx, getSession, sid)
	var i GetSessionRow
	err := row.Scan(&i.Uid, &i.ExpiresAt)
	return i, err
}

const getSessions = `-- name: GetSessions :many
SELECT sid, uid, expires_at
FROM sessions
WHERE CURRENT_TIMESTAMP < expires_at
`

func (q *Queries) GetSessions(ctx context.Context) ([]Session, error) {
	rows, err := q.db.QueryContext(ctx, getSessions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Session
	for rows.Next() {
		var i Session
		if err := rows.Scan(&i.Sid, &i.Uid, &i.ExpiresAt); err != nil {
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
