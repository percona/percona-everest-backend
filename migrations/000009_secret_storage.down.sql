CREATE TABLE plain_text_secrets
(
    id         VARCHAR UNIQUE not null,
    value      TEXT           NOT NULL,
    created_at TIMESTAMP      NOT NULL,
    updated_at TIMESTAMP
);
