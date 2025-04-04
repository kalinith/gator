-- name: CreateFeedFollow :one
WITH inserted_feed_follows AS (
INSERT INTO feed_follows (id,created_at,updated_at,user_id,feed_id)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
) RETURNING *)
SELECT ff.id, ff.created_at, ff.updated_at, ff.user_id, ff.feed_id, u.name AS UserName, f.name as FeedName
FROM inserted_feed_follows AS ff
INNER JOIN users AS u ON ff.user_id = u.id
INNER JOIN feeds AS f ON ff.feed_id = f.id;

-- name: GetFeedFollowsForUser :many
SELECT u.name AS Username, f.name AS FeedName
FROM users AS u
INNER JOIN feed_follows AS ff ON ff.user_id = u.id
INNER JOIN feeds AS f ON ff.feed_id = f.id
WHERE u.name = $1
ORDER BY f.name;
