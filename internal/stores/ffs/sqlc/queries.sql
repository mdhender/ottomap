--  Copyright (c) 2024 Michael D Henderson. All rights reserved.


-- name: CreateUser :one
INSERT INTO users (handle, hashed_password, clan, magic_key, path)
VALUES (:handle, :hashed_password, :clan, :magic_key, :path)
RETURNING id;

-- name: CreateTurnReport :one
INSERT INTO reports (uid, turn, clan, path)
VALUES (:uid, :turn, :clan, :path)
RETURNING id;

-- name: CreateTurnMap :one
INSERT INTO maps (uid, turn, clan, path)
VALUES (:uid, :turn, :clan, :path)
RETURNING id;

-- name: CreateUnit :exec
INSERT INTO units (rid, turn, name, starting_hex, ending_hex)
VALUES (:rid, :turn, :name, :starting_hex, :ending_hex);
