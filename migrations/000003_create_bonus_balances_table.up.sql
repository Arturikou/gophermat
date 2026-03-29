-- migrations/000003_create_bonus_balances_table.up.sql
-- Create bonus_balances table

CREATE TABLE IF NOT EXISTS bonus_balances
(
    user_id         INTEGER PRIMARY KEY REFERENCES users (id),
    current_amount  NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    total_withdrawn NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT check_bonus_current_positive CHECK (current_amount >= 0),
    CONSTRAINT check_bonus_withdrawn_positive CHECK (total_withdrawn >= 0)
);