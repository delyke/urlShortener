CREATE TABLE users (
                      id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
                      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                      updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);