-- name: CountQueuedReports :one
SELECT COUNT(*)
FROM report_queue;

-- name: CountQueuedByChecksum :one
SELECT COUNT(*)
FROM report_queue_data
WHERE cksum = ?1;

-- name: CountQueuedInProgressReports :one
SELECT COUNT(*)
FROM report_queue
WHERE status != "completed";

-- name: CreateQueuedReport :exec
INSERT INTO report_queue(qid, cid, status)
VALUES (?1, ?2, ?3);

-- name: CreateQueuedReportData :exec
INSERT INTO report_queue_data(qid, name, cksum, lines)
VALUES (?1, ?2, ?3, ?4);

-- name: ReadQueuedReport :one
SELECT report_queue.cid,
       report_queue.status,
       report_queue.crdttm,
       report_queue.updttm,
       report_queue_data.name,
       report_queue_data.cksum
FROM report_queue
         LEFT OUTER JOIN report_queue_data ON report_queue.qid = report_queue_data.qid
WHERE report_queue.qid = ?1
  AND report_queue.cid = ?2;

-- name: ReadQueuedReports :many
SELECT qid, cid, status, crdttm, updttm
FROM report_queue
WHERE cid = ?1;