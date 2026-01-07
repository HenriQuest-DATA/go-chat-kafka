-- name: CreateFriendship :one
INSERT INTO friendships (user_id, friend_id, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetFriendship :one
SELECT * FROM friendships
WHERE (user_id = $1 AND friend_id = $2)
   OR (user_id = $2 AND friend_id = $1);

-- name: UpdateFriendshipStatus :exec
UPDATE friendships SET status = $2 WHERE id = $1;

-- name: ListUserFriends :many
SELECT u.* FROM users u
INNER JOIN friendships f ON u.id = f.friend_id
WHERE f.user_id = $1 AND f.status = 'accepted'
UNION
SELECT u.* FROM users u
INNER JOIN friendships f ON u.id = f.user_id
WHERE f.friend_id = $1 AND f.status = 'accepted';