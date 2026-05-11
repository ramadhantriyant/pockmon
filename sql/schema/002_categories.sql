-- +goose up
CREATE TABLE categories (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) CHECK (type IN ('expense', 'income')) NOT NULL,
    color VARCHAR(7), -- hex color code
    icon VARCHAR(50), -- icon identifier
    parent_category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    is_system BOOLEAN DEFAULT false, -- for pre-defined categories
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name, type)
);

CREATE INDEX idx_categories_user ON categories(user_id);
CREATE INDEX idx_categories_type ON categories(type);

-- +goose down
DROP INDEX IF EXISTS idx_categories_type;
DROP INDEX IF EXISTS idx_categories_user;
DROP TABLE IF EXISTS categories;