CREATE TABLE role_resources (
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    resource_id INT NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, resource_id)
);
