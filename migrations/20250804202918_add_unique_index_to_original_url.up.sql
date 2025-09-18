-- up.sql
ALTER TABLE urls ADD CONSTRAINT unique_original_url UNIQUE (original_url);