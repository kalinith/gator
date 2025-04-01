-- +goose Up
CREATE TABLE feeds(
	id UUID PRIMARY KEY,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	name varchar NOT NULL,
	url varchar UNIQUE NOT NULL,
	user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE feeds;