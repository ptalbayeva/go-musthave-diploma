CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    order_id VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL,
    points_accumulated INT DEFAULT 0,
    withdrawal_status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT NOW()
);
