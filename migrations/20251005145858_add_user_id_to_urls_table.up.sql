ALTER TABLE urls
ADD COLUMN user_id bigint;

/*ALTER TABLE urls
ADD CONSTRAINT fk_urls_user
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;*/