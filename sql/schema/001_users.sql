-- +goose up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    firebase_uid VARCHAR(255) UNIQUE NOT NULL,
    currency_code CHAR(3) NOT NULL DEFAULT 'IDR',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);

-- +goose down
DROP INDEX IF EXISTS idx_users_firebase_uid;
DROP TABLE IF EXISTS users;