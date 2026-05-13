-- +goose Up
ALTER TABLE transactions ADD COLUMN transfer_id UUID REFERENCES transfers(id) ON DELETE SET NULL;
CREATE INDEX idx_transactions_transfer ON transactions(transfer_id) WHERE transfer_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_transfer;
ALTER TABLE transactions DROP COLUMN IF EXISTS transfer_id;
