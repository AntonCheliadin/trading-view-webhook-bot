-- +migrate Up
CREATE TABLE IF NOT EXISTS coins (
    id SERIAL PRIMARY KEY,
    coin_name VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL UNIQUE
);

-- +migrate Up
INSERT INTO coins (coin_name, symbol) VALUES ('Bitcoin', 'BTCUSDT');
INSERT INTO coins (coin_name, symbol) VALUES ('Binance coin', 'BNBUSDT');
INSERT INTO coins (coin_name, symbol) VALUES ('Solana', 'SOLUSDT'); 