-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
	INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id) VALUES ($1, $2, $3, $4, $5)
	RETURNING *
)
SELECT
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds ON feeds.ID = inserted_feed_follow.feed_id
INNER JOIN users ON users.ID = inserted_feed_follow.user_id;

-- name: GetFeedFollowsForUser :many
SELECT feeds.* FROM feeds
INNER JOIN feed_follows ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1;