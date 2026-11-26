

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    sku TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS stock_movements (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    type TEXT NOT NULL CHECK (type IN ('in', 'out')),
    qty INTEGER NOT NULL CHECK (qty > 0),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sales_orders (
    id SERIAL PRIMARY KEY,
    customer TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'CREATED' CHECK (status IN ('CREATED', 'CANCELLED')),
    total INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    cancelled_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS sales_order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES sales_orders(id),
    sku TEXT NOT NULL REFERENCES products(sku),
    qty INTEGER NOT NULL CHECK (qty > 0),
    unit_price INTEGER NOT NULL
);

INSERT INTO products (sku, name, price, stock) VALUES
('LAPTOP-001', 'Lenovo IdeaPad Slim 3', 7500000, 10),
('LAPTOP-002', 'Asus Vivobook Go 14', 6200000, 15),
('HP-GAMING', 'HP Victus 15', 12500000, 5),
('MOUSE-01', 'Logitech G102 Lightsync', 250000, 50),
('KEYBOARD-01', 'Fantech MK853', 350000, 30),
('MONITOR-01', 'LG UltraGear 24"', 2800000, 8),
('HEADSET-01', 'Razer Kraken X', 650000, 20),
('WEBCAM-01', 'Logitech C920 HD Pro', 1100000, 12),
('SSD-512', 'Kingston NV2 512GB', 650000, 40),
('RAM-16', 'Corsair Vengeance 16GB DDR4', 850000, 25),
('CHAIR-01', 'Secretlab Titan Evo', 5500000, 3),
('DESK-01', 'Meja Standing Desk Electric', 3200000, 7);
