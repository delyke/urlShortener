ALTER TABLE urls
ADD COLUMN user_id bigint REFERENCES users(id) ON DELETE cascade;