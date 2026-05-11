-- +goose up
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) CHECK (type IN ('cash', 'bank', 'credit_card', 'debit_card', 'investment', 'loan')) NOT NULL,
    currency_code CHAR(3) DEFAULT 'IDR',
    initial_balance NUMERIC(15, 2) DEFAULT 0,
    current_balance NUMERIC(15, 2) DEFAULT 0,
    include_in_total BOOLEAN DEFAULT true,
    color VARCHAR(7),
    icon VARCHAR(50),
    notes TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_accounts_user ON accounts(user_id);
CREATE INDEX idx_accounts_active ON accounts(user_id) WHERE is_active = true;

-- +goose down
DROP INDEX IF EXISTS idx_accounts_active;
DROP INDEX IF EXISTS idx_accounts_user;
DROP TABLE IF EXISTS accounts;
