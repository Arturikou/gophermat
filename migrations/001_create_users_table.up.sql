-- migrations/001_create_users_table.up.sql
-- Create users table

CREATE TABLE IF NOT EXISTS users
(
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);