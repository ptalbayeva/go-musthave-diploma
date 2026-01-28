CREATE TABLE transactions
(
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID REFERENCES users (id) ON DELETE CASCADE,
    order_id         VARCHAR(255) NOT NULL,
    points           INT          NOT NULL,
    transaction_type VARCHAR(50)  NOT NULL, -- ACCRUAL, WITHDRAWAL
    created_at       TIMESTAMP        DEFAULT NOW()
);
