-- +goose Up
CREATE TABLE feed_follows (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    user_id UUID NOT NULL,
    feed_id UUID NOT NULL,

    CONSTRAINT uq_feed_follows
        UNIQUE (user_id, feed_id),

    CONSTRAINT fk_feed_follows_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,

    CONSTRAINT fk_feed_follows_feed
        FOREIGN KEY (feed_id)
        REFERENCES feeds (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

-- +goose Down
DROP TABLE feed_follows;
