CREATE TABLE pmm_instances
(
    id  uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    url VARCHAR NOT NULL,
    api_key_secret_id VARCHAR NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);
