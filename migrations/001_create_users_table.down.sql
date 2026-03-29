-- migrations/001_create_users_table.down.sql
-- Rollback users database

DROP TABLE IF EXISTS users CASCADE;