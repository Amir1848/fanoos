ALTER TABLE device_links DROP CONSTRAINT device_links_link_type_check;

ALTER TABLE device_links
    ALTER COLUMN link_type DROP DEFAULT,
    ALTER COLUMN link_type TYPE VARCHAR(50) USING 'toggle'::varchar,
    ALTER COLUMN link_type SET DEFAULT 'toggle';
