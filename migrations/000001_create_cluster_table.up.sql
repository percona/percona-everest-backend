CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

SELECT uuid_generate_v4();

CREATE TABLE kubernetes_clusters
(
    id         uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    name       VARCHAR UNIQUE NOT NULL,
    created_at TIMESTAMP      NOT NULL,
    updated_at TIMESTAMP
);

CREATE TABLE secrets
(
    id         VARCHAR UNIQUE not null,
    value      TEXT           NOT NULL,
    created_at TIMESTAMP      NOT NULL,
    updated_at TIMESTAMP
);
