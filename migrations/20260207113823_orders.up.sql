CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number TEXT NOT NULL,
    accrual DOUBLE PRECISION DEFAULT NULL,
    status order_status NOT NULL DEFAULT 'NEW',
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT orders_same_user_unique
        UNIQUE (user_id, order_number),

    CONSTRAINT orders_global_unique
        UNIQUE (order_number)
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);
