-- +goose up
CREATE TABLE goals (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    target_amount NUMERIC(15, 2) NOT NULL CHECK (target_amount > 0),
    current_amount NUMERIC(15, 2) DEFAULT 0,
    currency_code CHAR(3) DEFAULT 'IDR',
    target_date DATE,
    goal_type VARCHAR(50) CHECK (goal_type IN ('savings', 'debt_payoff', 'investment', 'purchase', 'other')) NOT NULL,
    is_completed BOOLEAN DEFAULT false,
    completed_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_goals_user ON goals(user_id);
CREATE INDEX idx_goals_active ON goals(user_id) WHERE is_completed = false;

-- +goose down
DROP INDEX IF EXISTS idx_goals_active;
DROP INDEX IF EXISTS idx_goals_user;
DROP TABLE IF EXISTS goals;