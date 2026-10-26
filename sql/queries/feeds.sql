-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetFeed :one
SELECT * FROM feeds
WHERE url = $1 LIMIT 1;

-- name: CheckFeedExists :one
SELECT EXISTS(
    SELECT 1
    FROM feeds
    WHERE name = $1
);

-- name: ListFeeds :many
SELECT
    feeds.*,
    users.name AS user_name
FROM feeds
INNER JOIN users ON feeds.user_id = users.id
ORDER BY feeds.name;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET
    last_fetched_at = $1
WHERE
    id = $2
;

-- name: GetNextFeedToFetch :one
SELECT
    *
FROM feeds
ORDER BY
    last_fetched_at ASC NULLS FIRST
LIMIT 1
;