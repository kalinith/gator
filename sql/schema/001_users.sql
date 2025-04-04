-- +goose Up
CREATE TABLE users(
	id UUID PRIMARY KEY,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	name varchar(250) NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE users;
