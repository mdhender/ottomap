--  Copyright (c) 2024 Michael D Henderson. All rights reserved.


-- name: CreateUser :one
INSERT INTO users (handle, hashed_password, clan, magic_key, path)
VALUES (:handle, :hashed_password, :clan, :magic_key, :path)
RETURNING uid;

