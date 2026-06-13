-- name: CreateMessage :one
INSERT INTO messages (room_id, user_id, content, type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMessagesByRoom :many
SELECT m.*, u.username, u.display_name FROM messages m
JOIN users u ON m.user_id = u.id
WHERE m.room_id = $1
ORDER BY m.created_at DESC
LIMIT $2 OFFSET $3;

-- name: EditMessage :one
UPDATE messages
SET content = $3, edited_at = NOW()
WHERE id = $1 AND room_id = $2
RETURNING *;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = $1 AND room_id = $2;
