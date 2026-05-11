-- +goose up
CREATE TABLE transfers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    to_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    from_transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    to_transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    to_amount NUMERIC(15, 2) NOT NULL CHECK (to_amount > 0),
    exchange_rate NUMERIC(10, 6) DEFAULT 1.0,
    transfer_date DATE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK (from_account_id != to_account_id)
);

CREATE INDEX idx_transfers_user ON transfers(user_id);
CREATE INDEX idx_transfers_from_account ON transfers(from_account_id);
CREATE INDEX idx_transfers_to_account ON transfers(to_account_id);

-- +goose down
DROP INDEX IF EXISTS idx_transfers_to_account;
DROP INDEX IF EXISTS idx_transfers_from_account;
DROP INDEX IF EXISTS idx_transfers_user;
DROP TABLE IF EXISTS transfers;
