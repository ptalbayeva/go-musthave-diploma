CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id VARCHAR(255) NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    points FLOAT NOT NULL,
    type VARCHAR(50) NOT NULL,
    processed_at TIMESTAMP DEFAULT NOW()
);
