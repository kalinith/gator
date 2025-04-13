-- name: CreateFeed :one
INSERT INTO feeds (id,created_at,updated_at,name,url,last_fetched_at)
VALUES (
$1,
$2,
$3,
$4,
$5,
NULL)
RETURNING *;

-- name: SelectFeeds :many
SELECT f.name AS FeedName,f.url AS URL,u.name AS UserName
  FROM feeds f INNER JOIN feed_follows
  ON f.ID = feed_follows.feed_id
  INNER JOIN users u
  ON feed_follows.user_id = u.id;

-- name: SelectFeedURL :one
SELECT ID, name, url
  FROM feeds
  WHERE url = $1;

-- name: MarkFeedFetched :exec
 UPDATE feeds
  SET last_fetched_at = $1,
  updated_at = $1
  WHERE ID = $2;

-- name: GetNextFeedToFetch :one
 SELECT *
  FROM feeds
  Order by last_fetched_at Desc NULLS FIRST, name ASC
  LIMIT 1;

-- name: DeleteFeed :exec
DELETE
  FROM feeds
  WHERE ID = $1;

-- name: DeleteFeeds :exec
DELETE
  FROM feeds;