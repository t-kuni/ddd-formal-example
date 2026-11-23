CREATE TABLE IF NOT EXISTS actors(
    id uuid PRIMARY KEY,
    name text NOT NULL,
    role text NOT NULL CHECK (role IN ('customer','warehouse','carrier','operator')),
    email text NOT NULL
);
CREATE TABLE IF NOT EXISTS customer_addresses(
    customer_id uuid PRIMARY KEY REFERENCES actors(id),
    company text NOT NULL,
    recipient text NOT NULL,
    postal text NOT NULL,
    address text NOT NULL,
    phone text NOT NULL
);
CREATE TABLE IF NOT EXISTS products(
    id uuid PRIMARY KEY,
    sku text NOT NULL UNIQUE,
    name text NOT NULL,
    category text NOT NULL,
    unit text NOT NULL,
    price bigint NOT NULL CHECK (price >= 0),
    image bytea NOT NULL
);
CREATE TABLE IF NOT EXISTS coupons(id uuid PRIMARY KEY,code text UNIQUE NOT NULL, percent integer NOT NULL CHECK(percent BETWEEN 0 AND 100));
CREATE TABLE IF NOT EXISTS inventory(
    product_id uuid PRIMARY KEY REFERENCES products(id),
    version bigint NOT NULL CHECK (version > 0),
    on_hand bigint NOT NULL CHECK (on_hand >= 0),
    reserved bigint NOT NULL CHECK (reserved BETWEEN 0 AND on_hand)
);
CREATE TABLE IF NOT EXISTS carts(
    customer_id uuid PRIMARY KEY REFERENCES actors(id),
    version bigint NOT NULL CHECK (version > 0),
    expires_at bigint NOT NULL CHECK (expires_at >= 0)
);
CREATE TABLE IF NOT EXISTS cart_items(
    customer_id uuid NOT NULL REFERENCES carts(customer_id),
    product_id uuid NOT NULL REFERENCES products(id),
    quantity bigint NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (customer_id, product_id)
);
CREATE TABLE IF NOT EXISTS orders(
    id uuid PRIMARY KEY,
    version bigint NOT NULL CHECK (version > 0),
    number text NOT NULL UNIQUE,
    customer_id uuid NOT NULL REFERENCES actors(id),
    status integer NOT NULL CHECK (status BETWEEN 0 AND 10),
    requested_date text NOT NULL,
    note text NOT NULL,
    coupon text NOT NULL,
    discount_percent bigint NOT NULL CHECK (discount_percent BETWEEN 0 AND 100),
    subtotal bigint NOT NULL CHECK (subtotal >= 0),
    discount bigint NOT NULL CHECK (discount BETWEEN 0 AND subtotal),
    net bigint NOT NULL CHECK (net = subtotal - discount),
    tax bigint NOT NULL CHECK (tax >= 0),
    total bigint NOT NULL CHECK (total = net + tax),
    points bigint NOT NULL CHECK (points >= 0),
    worker_id uuid REFERENCES actors(id),
    created_at timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS orders_by_customer ON orders(customer_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS orders_by_status ON orders(status, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS orders_by_date ON orders(created_at DESC, id DESC);
CREATE TABLE IF NOT EXISTS order_addresses(
    order_id uuid PRIMARY KEY REFERENCES orders(id),
    company text NOT NULL,
    recipient text NOT NULL,
    postal text NOT NULL,
    address text NOT NULL,
    phone text NOT NULL
);
CREATE TABLE IF NOT EXISTS order_items(
    order_id uuid NOT NULL REFERENCES orders(id),
    position integer NOT NULL CHECK (position > 0),
    product_id uuid NOT NULL REFERENCES products(id),
    sku text NOT NULL,
    name text NOT NULL,
    unit text NOT NULL,
    quantity bigint NOT NULL CHECK (quantity > 0),
    unit_price bigint NOT NULL CHECK (unit_price >= 0),
    PRIMARY KEY (order_id, position)
);
CREATE TABLE IF NOT EXISTS shipments(
    order_id uuid PRIMARY KEY REFERENCES orders(id),
    carrier_id uuid REFERENCES actors(id),
    tracking text NOT NULL
);
CREATE TABLE IF NOT EXISTS shipment_boxes(
    id uuid PRIMARY KEY,
    order_id uuid NOT NULL REFERENCES shipments(order_id),
    position integer NOT NULL CHECK (position > 0),
    label text NOT NULL,
    weight_grams bigint NOT NULL CHECK (weight_grams > 0),
    UNIQUE (order_id, position)
);
CREATE TABLE IF NOT EXISTS order_status_history(
    order_id uuid NOT NULL REFERENCES orders(id),
    sequence integer NOT NULL CHECK (sequence > 0),
    occurred_at timestamptz NOT NULL,
    actor_id uuid NOT NULL,
    status_from integer NOT NULL CHECK (status_from BETWEEN -1 AND 10),
    status_to integer NOT NULL CHECK (status_to BETWEEN 0 AND 10),
    note text NOT NULL,
    PRIMARY KEY (order_id, sequence)
);
CREATE TABLE IF NOT EXISTS loyalty_accounts(
    customer_id uuid PRIMARY KEY REFERENCES actors(id),
    version bigint NOT NULL CHECK (version > 0),
    points bigint NOT NULL CHECK (points >= 0)
);
CREATE TABLE IF NOT EXISTS order_point_awards(
    order_id uuid PRIMARY KEY REFERENCES orders(id)
);
CREATE TABLE IF NOT EXISTS commands(id uuid PRIMARY KEY,actor_id uuid NOT NULL,type text NOT NULL,payload jsonb NOT NULL,fingerprint text NOT NULL,status text NOT NULL DEFAULT 'queued',result jsonb,error text NOT NULL DEFAULT '',attempts integer NOT NULL DEFAULT 0,last_dispatched timestamptz,created_at timestamptz NOT NULL DEFAULT now(),finished_at timestamptz);
CREATE INDEX IF NOT EXISTS pending_commands ON commands(created_at) WHERE status='queued';
CREATE TABLE IF NOT EXISTS invoices(id uuid PRIMARY KEY,order_id uuid UNIQUE NOT NULL REFERENCES orders(id),pdf bytea NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),sent_at timestamptz,attempts integer NOT NULL DEFAULT 0,last_error text NOT NULL DEFAULT '',lease_until timestamptz);
