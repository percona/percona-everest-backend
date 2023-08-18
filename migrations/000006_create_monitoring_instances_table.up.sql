CREATE TABLE monitoring_instances
(
    name VARCHAR NOT NULL PRIMARY KEY,
    type VARCHAR NOT NULL,
    url VARCHAR NOT NULL,
    api_key_secret_id VARCHAR NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);

INSERT INTO monitoring_instances
    SELECT id, 'pmm', url, api_key_secret_id, created_at, updated_at FROM pmm_instances;

DROP TABLE pmm_instances;
