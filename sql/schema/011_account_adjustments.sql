-- +goose Up
CREATE TABLE account_adjustments (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(15, 2) NOT NULL,
    previous_balance NUMERIC(15, 2) NOT NULL,
    new_balance NUMERIC(15, 2) NOT NULL,
    reason TEXT,
    adjustment_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for faster queries
CREATE INDEX idx_account_adjustments_account_id ON account_adjustments(account_id);
CREATE INDEX idx_account_adjustments_user_id ON account_adjustments(user_id);
CREATE INDEX idx_account_adjustments_adjustment_date ON account_adjustments(adjustment_date);
CREATE INDEX idx_account_adjustments_created_at ON account_adjustments(created_at);

-- +goose Down
DROP TABLE IF EXISTS account_adjustments;
