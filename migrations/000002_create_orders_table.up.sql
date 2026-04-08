-- migrations/000002_create_orders_table.up.sql
-- Create orders table

CREATE TABLE IF NOT EXISTS orders
(
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     INTEGER        NOT NULL REFERENCES users (id),
    number      VARCHAR(255)   NOT NULL UNIQUE,
    status      VARCHAR(30)    NOT NULL DEFAULT 'NEW',
    accrual     NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    uploaded_at TIMESTAMPTZ    NOT NULL DEFAULT Now(),

    CONSTRAINT check_order_status CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED'))
);