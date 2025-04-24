CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    username VARCHAR(50),
    email VARCHAR(50) UNIQUE,
    password TEXT
);

CREATE INDEX email_idx ON users USING HASH (email);