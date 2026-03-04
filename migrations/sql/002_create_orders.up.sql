-- Создание таблицы заказов
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'created',
    total_amount BIGINT NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    payment_status VARCHAR(50) DEFAULT 'pending',
    shipping_address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);
CREATE INDEX idx_orders_created_at ON orders(created_at);

-- Комментарий
COMMENT ON TABLE orders IS 'Заказы пользователей';
COMMENT ON COLUMN orders.status IS 'created, confirmed, paid, shipped, completed, cancelled';
COMMENT ON COLUMN orders.total_amount IS 'Сумма заказа в копейках';