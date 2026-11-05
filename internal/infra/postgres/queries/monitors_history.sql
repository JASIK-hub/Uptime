-- name: AddMonitorCheck :exec
INSERT INTO monitors_history (
    monitor_id,
    duration_ms,
    started_at,
    success,
    status_code,
    error_message
)
VALUES ($1, $2, $3, $4, $5, $6);