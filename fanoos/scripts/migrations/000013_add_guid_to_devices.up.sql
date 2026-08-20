ALTER TABLE devices ADD COLUMN guid UUID NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE devices ADD CONSTRAINT devices_guid_unique UNIQUE (guid);
