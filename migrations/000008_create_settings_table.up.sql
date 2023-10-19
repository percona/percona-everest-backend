CREATE TABLE settings
(
    id     SERIAL PRIMARY KEY,
    key    TEXT UNIQUE NOT NULL,
    value  TEXT
);
