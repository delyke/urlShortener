CREATE TABLE users (
                      id IDENTITY PRIMARY KEY,
                      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                      updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);