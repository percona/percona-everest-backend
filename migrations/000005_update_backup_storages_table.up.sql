ALTER TABLE backup_storages DROP CONSTRAINT backup_storages_pkey;
ALTER TABLE backup_storages ADD PRIMARY KEY (name);
ALTER TABLE backup_storages DROP COLUMN id;
ALTER TABLE backup_storages ADD COLUMN description TEXT;

