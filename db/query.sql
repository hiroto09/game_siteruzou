-- name: InsertPingLog :exec
INSERT INTO ping_log (timestamp, status, loss_rate)
VALUES (?, ?, ?);

-- name: GetLatestLogs :many
SELECT * FROM ping_log
ORDER BY timestamp DESC
LIMIT 10;
