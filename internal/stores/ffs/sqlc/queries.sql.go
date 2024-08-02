// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package sqlc

import (
	"context"
)

const createUser = `-- name: CreateUser :one


INSERT INTO users (handle, hashed_password, clan, magic_key, path)
VALUES (?1, ?2, ?3, ?4, ?5)
RETURNING uid
`

type CreateUserParams struct {
	Handle         string
	HashedPassword string
	Clan           string
	MagicKey       string
	Path           string
}

// Copyright (c) 2024 Michael D Henderson. All rights reserved.
func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.Handle,
		arg.HashedPassword,
		arg.Clan,
		arg.MagicKey,
		arg.Path,
	)
	var uid int64
	err := row.Scan(&uid)
	return uid, err
}
