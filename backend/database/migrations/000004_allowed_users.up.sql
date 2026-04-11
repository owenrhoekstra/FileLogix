CREATE TABLE approved_users (
                       id BYTEA PRIMARY KEY,
                       email TEXT UNIQUE NOT NULL,
                       allowed BOOLEAN NOT NULL,
                       permission TEXT NOT NULL
);