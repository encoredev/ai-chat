CREATE TABLE IF NOT EXISTS bot(
    id uuid PRIMARY KEY,
    name TEXT NOT NULL,
    prompt TEXT NOT NULL,
    profile TEXT NOT NULL,
    avatar bytea,
    provider TEXT NOT NULL,
    deleted TIMESTAMP DEFAULT NULL

);