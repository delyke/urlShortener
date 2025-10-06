CREATE TABLE users (
                      id BIGSERIAL PRIMARY KEY,
                      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                      updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);