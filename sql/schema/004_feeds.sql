-- +goose Up
ALTER TABLE feeds
ADD COLUMN last_fetched_at timestamp;

ALTER TABLE feeds
DROP COLUMN user_id;

-- +goose Down
ALTER TABLE feeds
DROP COLUMN last_fetched_at;

ALTER TABLE feeds
ADD COLUMN user_id UUID;

UPDATE feeds set user_id = (select user_id from feed_follows where feed_id = feeds.id limit 1);

DELETE FROM feeds WHERE user_id IS NULL;

ALTER TABLE feeds
ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE feeds
ADD CONSTRAINT fk_feeds_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
