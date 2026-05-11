-- +goose up
CREATE TABLE recurring_transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    type VARCHAR(20) CHECK (type IN ('expense', 'income')) NOT NULL,
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    currency_code CHAR(3) DEFAULT 'IDR',
    description TEXT NOT NULL,
    frequency VARCHAR(20) CHECK (frequency IN ('daily', 'weekly', 'biweekly', 'monthly', 'quarterly', 'yearly')) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    next_due_date DATE NOT NULL,
    last_processed_date DATE,
    is_active BOOLEAN DEFAULT true,
    auto_create BOOLEAN DEFAULT false, -- automatically create transactions
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_recurring_user ON recurring_transactions(user_id);
CREATE INDEX idx_recurring_next_due ON recurring_transactions(next_due_date) WHERE is_active = true;

-- +goose down
DROP INDEX IF EXISTS idx_recurring_next_due;
DROP INDEX IF EXISTS idx_recurring_user;
DROP TABLE IF EXISTS recurring_transactions;