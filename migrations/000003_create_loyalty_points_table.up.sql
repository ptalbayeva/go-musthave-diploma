CREATE TABLE loyalty_points (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance FLOAT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
