CREATE TABLE urls (
                      id BIGSERIAL PRIMARY KEY,
                      original_url VARCHAR(256) NOT NULL,
                      short_url VARCHAR(16) NOT NULL UNIQUE
);

CREATE INDEX idx_short_url ON urls(short_url);