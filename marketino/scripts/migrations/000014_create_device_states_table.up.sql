CREATE TABLE device_states (
    device_guid UUID PRIMARY KEY REFERENCES devices(guid) ON DELETE CASCADE,
    is_on BOOLEAN NOT NULL DEFAULT false,
    online BOOLEAN NOT NULL DEFAULT false,
    ts BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
