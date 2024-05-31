CREATE TABLE avatars
(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id      TEXT NOT NULL ,
    bot_name TEXT NOT NULL,
    url TEXT NOT NULL,
    deleted TIMESTAMP DEFAULT NULL
)