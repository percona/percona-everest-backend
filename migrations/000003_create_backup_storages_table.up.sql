CREATE TABLE backup_storages
(
    id            uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    type          VARCHAR   NOT NULL,
    name          VARCHAR   NOT NULL UNIQUE,
    bucket_name   VARCHAR   NOT NULL,
    url           VARCHAR,
    region        VARCHAR   NOT NULL,
    access_key_id VARCHAR   NOT NULL,
    secret_key_id VARCHAR   NOT NULL,

    created_at    TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP
);
