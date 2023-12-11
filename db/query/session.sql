-- name: CreateSession :one
INSERT INTO sessions (
  id,
  username, 
  refresh_token, 
  user_agent,
  client_ip,
  is_blocked,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1 LIMIT 1;

-- name: GetSessionForUpdate :one
SELECT * FROM sessions
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: UpdateBlockSession :one
UPDATE sessions 
SET is_blocked = $2, expires_at = $3
WHERE id = $1
RETURNING *;