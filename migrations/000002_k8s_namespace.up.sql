ALTER TABLE kubernetes_clusters
    ADD COLUMN namespace VARCHAR NOT NULL CHECK (namespace <> '');
