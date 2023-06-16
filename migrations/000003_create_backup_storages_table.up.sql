CREATE TABLE backup_storages
(
    id            uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    name          VARCHAR   NOT NULL,
    bucket_name   VARCHAR   NOT NULL,
    url           VARCHAR   NOT NULL,
    region        VARCHAR   NOT NULL,
    access_key_id VARCHAR   NOT NULL,
    secret_key_id VARCHAR   NOT NULL,

    created_at    TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP
);
