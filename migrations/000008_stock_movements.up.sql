CREATE TABLE IF NOT EXISTS stock_movements (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    operation VARCHAR(32) NOT NULL CHECK (operation IN ('set', 'increment', 'decrement', 'checkout_decrement')),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    before_quantity INTEGER NOT NULL CHECK (before_quantity >= 0),
    after_quantity INTEGER NOT NULL CHECK (after_quantity >= 0),
    reason VARCHAR(255) NULL,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (
        (operation = 'set' AND after_quantity = quantity)
        OR (operation = 'increment' AND after_quantity = before_quantity + quantity)
        OR (operation IN ('decrement', 'checkout_decrement') AND before_quantity >= quantity AND after_quantity = before_quantity - quantity)
    )
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_product_id ON stock_movements(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON stock_movements(created_at);
CREATE INDEX IF NOT EXISTS idx_stock_movements_operation ON stock_movements(operation);
