CREATE TABLE resources (
    id SERIAL PRIMARY KEY,
    resource_key VARCHAR(100) NOT NULL UNIQUE
);

INSERT INTO resources (resource_key) VALUES
('users.create'),
('users.delete'),
('users.view'),
('roles.create'),
('roles.delete'),
('roles.view');
