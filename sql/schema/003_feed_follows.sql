-- +goose Up
CREATE TABLE feed_follows (
	id UUID PRIMARY KEY,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL,
	feed_id UUID REFERENCES feeds(id) ON DELETE CASCADE NOT NULL,
	UNIQUE (feed_id, user_id)
);

-- +goose Down
DROP TABLE feed_follows;
