-- noinspection SqlNoDataSourceInspectionForFiles
-- +migrate Up

CREATE TABLE orders (
    order_uid TEXT PRIMARY KEY,
    track_number TEXT NOT NULL,
    entry TEXT,
    locale TEXT NOT NULL, 
    internal_signature TEXT, 
    customer_id TEXT NOT NULL, 
    delivery_service TEXT NOT NULL, 
    shardkey TEXT,
    sm_id INTEGER, 
    date_created TIMESTAMP NOT NULL, 
    oof_shard TEXT
);

CREATE TABLE delivery (
    delivery_id SERIAL PRIMARY KEY, 
    order_uid TEXT NOT NULL,
    name TEXT NOT NULL, 
    phone TEXT NOT NULL, 
    zip TEXT, 
    city TEXT NOT NULL, 
    address TEXT NOT NULL, 
    region TEXT, 
    email TEXT,
    CONSTRAINT fk_order_uid FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
);

CREATE TABLE payment (
    payment_id SERIAL PRIMARY KEY, 
    order_uid TEXT NOT NULL,
    transaction TEXT NOT NULL, 
    request_id TEXT,
    currency TEXT NOT NULL,
    provider TEXT NOT NULL,
    amount DECIMAL NOT NULL,
    payment_dt TIMESTAMP NOT NULL,
    bank TEXT,
    delivery_cost DECIMAL,
    goods_total DECIMAL NOT NULL,
    custom_fee DECIMAL NOT NULL,
    CONSTRAINT fk_order_uid FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
);

CREATE TABLE items (
    item_id SERIAL PRIMARY KEY,
    order_uid TEXT NOT NULL,
    track_number TEXT NOT NULL,
    price DECIMAL NOT NULL,
    rid TEXT,
    name TEXT NOT NULL,
    sale INTEGER,
    size TEXT,
    total_price DECIMAL NOT NULL,
    nm_id INTEGER NOT NULL,
    brand TEXT NOT NULL,
    status INTEGER NOT NULL,
    CONSTRAINT fk_order_uid FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
);

CREATE INDEX idx_track_number ON items(track_number);
CREATE INDEX idx_payment_transaction ON payment(transaction);

-- +migrate Down

DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS delivery;
DROP TABLE IF EXISTS payment;
DROP TABLE IF EXISTS items;
