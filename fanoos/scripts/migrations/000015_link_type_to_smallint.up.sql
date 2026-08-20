ALTER TABLE device_links
    ALTER COLUMN link_type DROP DEFAULT,
    ALTER COLUMN link_type TYPE SMALLINT USING 1,
    ALTER COLUMN link_type SET NOT NULL,
    ALTER COLUMN link_type SET DEFAULT 1;

ALTER TABLE device_links
    ADD CONSTRAINT device_links_link_type_check CHECK (link_type BETWEEN 1 AND 4);
