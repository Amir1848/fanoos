CREATE TABLE device_links (
    id SERIAL PRIMARY KEY,
    source_device_id INT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    target_device_id INT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    link_type VARCHAR(50) NOT NULL DEFAULT 'toggle',
    UNIQUE (source_device_id, target_device_id)
);
