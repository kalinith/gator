-- +goose Up
CREATE TABLE posts(
	id UUID PRIMARY KEY,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	title varchar(250) NOT NULL,
	url varchar UNIQUE NOT NULL,
	description varchar,
	published_at timestamp NOT NULL,
	feed_id UUID REFERENCES feeds(id) ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE posts;
