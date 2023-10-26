ALTER TABLE hash_link ADD COLUMN OWNER_UUID UUID NOT NULL default '';
CREATE INDEX hash_link_owner_uuid ON hash_link (OWNER_UUID);