-- name: AddMonitor :one
INSERT INTO monitors (url, timeout_seconds, interval_seconds)
VALUES ($1, $2, $3)
RETURNING id;

-- name: ListMonitors :many
SELECT id, url, timeout_seconds, interval_seconds
FROM monitors
ORDER BY id;