-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    description TEXT,
    published_at TIMESTAMPTZ,
    feed_id UUID NOT NULL,

    CONSTRAINT fk_posts_feed_id
        FOREIGN KEY (feed_id)
        REFERENCES feeds (id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
