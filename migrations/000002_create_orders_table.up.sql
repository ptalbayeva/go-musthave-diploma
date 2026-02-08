CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    order_id VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL,
    points_accumulated FLOAT DEFAULT 0,
    withdrawal_status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_user_id_order_id ON orders (user_id, order_id);
CREATE INDEX idx_user_id_order_id_status ON orders (user_id, order_id, status);
CREATE INDEX idx_user_id_withdrawal_status ON orders (user_id, withdrawal_status);
