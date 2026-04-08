-- migrations/000004_create_bonus_withdrawals_table.up.sql
-- Create bonus_withdrawals table

CREATE TABLE IF NOT EXISTS bonus_withdrawals
(
    id           INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id      INTEGER        NOT NULL REFERENCES users (id),
    order_number VARCHAR(255)   NOT NULL,
    amount       NUMERIC(15, 2) NOT NULL,
    processed_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT check_withdrawal_amount_positive CHECK (amount > 0)
);