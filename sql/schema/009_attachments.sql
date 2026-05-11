-- +goose up
CREATE TABLE attachments (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_type VARCHAR(50),
    file_size INTEGER, -- in bytes
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_attachments_transaction ON attachments(transaction_id);

-- +goose down
DROP INDEX IF EXISTS idx_attachments_transaction;
DROP TABLE IF EXISTS attachments;