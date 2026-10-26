-- name: CreatePost :one
INSERT INTO posts (title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *
;

-- name: CheckPostExists :one
SELECT EXISTS(
    SELECT 1
    FROM posts
    WHERE url = $1
);

-- name: UpdatePost :one
UPDATE posts
SET
    title = $1,
    url = $2,
    description = $3,
    published_at = $4,
    feed_id = $5
WHERE
    url = sqlc.arg(where_url)::text
RETURNING *
;

-- name: GetPost :one
SELECT
    *
FROM posts
WHERE
    url = $1
;

-- name: GetUserPosts :many
SELECT
    posts.*,
    feeds.name AS feed_name,
    feeds.url AS feed_url,
    users.name AS user_name
FROM posts
INNER JOIN feeds ON posts.feed_id = feeds.id
INNER JOIN users ON feeds.user_id = users.id
WHERE
    feeds.user_id = $1
ORDER BY
    posts.created_at DESC
LIMIT $2
;