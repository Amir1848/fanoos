CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    user_name VARCHAR(50) NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    full_name VARCHAR(101) NOT NULL,
    email VARCHAR(100) NULL,
    phone_number VARCHAR(100) NULL,
    updated_at TIMESTAMP
);