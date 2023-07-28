ALTER TABLE backup_storages ADD COLUMN id uuid DEFAULT uuid_generate_v4() PRIMARY KEY;
ALTER TABLE backup_storages DROP CONSTRAINT backup_storages_pkey;
ALTER TABLE backup_storages ADD PRIMARY KEY (id);
ALTER TABLE backup_storages ADD CONSTRAINT backup_storages_name_key UNIQUE (name);
