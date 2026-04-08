-- +goose Up
-- Make user_id nullable in transactions table to support guest purchases
ALTER TABLE transactions ALTER COLUMN user_id DROP NOT NULL;

-- +goose Down
-- Revert: make user_id NOT NULL again (this will fail if there are guest transactions)
ALTER TABLE transactions ALTER COLUMN user_id SET NOT NULL;
