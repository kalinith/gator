-- name: CreateFeed :one
INSERT INTO feeds (id,created_at,updated_at,name,url,user_id)
VALUES (
$1,
$2,
$3,
$4,
$5,
$6)
RETURNING *;

-- name: SelectFeeds :many
SELECT f.name AS FeedName,f.url AS URL,u.name AS UserName
  FROM feeds f INNER JOIN users u
  ON f.user_id = u.id;

-- name: SelectFeedURL :one
SELECT ID, name, url
  FROM feeds
  WHERE url = $1;

