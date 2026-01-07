-- name: CreateMessage :one
INSERT INTO messages (sender_id, receiver_id, content, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMessageByID :one
SELECT * FROM messages WHERE id = $1;

-- name: ListMessagesBetweenUsers :many
SELECT * FROM messages
WHERE (sender_id = $1 AND receiver_id = $2)
   OR (sender_id = $2 AND receiver_id = $1)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateMessageStatus :exec
UPDATE messages SET status = $2 WHERE id = $1;