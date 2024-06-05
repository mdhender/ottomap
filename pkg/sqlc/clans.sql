-- name: CreateClan :exec
INSERT INTO clans (cid, uid)
VALUES (?1, ?2);

-- name: ReadAllClanReports :many
SELECT rid, tid, cid, crdttm
FROM reports
WHERE cid = ?1;

-- name: ReadUserClan :one
SELECT cid
FROM clans
WHERE uid = ?1;