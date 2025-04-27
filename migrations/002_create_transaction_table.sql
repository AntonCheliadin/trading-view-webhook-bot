-- +migrate Up
CREATE TABLE IF NOT EXISTS transaction_table (
    id SERIAL PRIMARY KEY,
    coin_id BIGINT NOT NULL REFERENCES coins(id),
    transaction_type SMALLINT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    stop_loss_price DOUBLE PRECISION,
    take_profit_price DOUBLE PRECISION,
    total_cost DOUBLE PRECISION NOT NULL,
    commission DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    client_order_id VARCHAR(100),
    api_error TEXT,
    related_transaction_id BIGINT REFERENCES transaction_table(id),
    profit BIGINT,
    percent_profit DOUBLE PRECISION,
    trading_strategy_id BIGINT NOT NULL,
    futures_type SMALLINT NOT NULL,
    fake BOOLEAN NOT NULL DEFAULT false,
    trading_key VARCHAR(100) NOT NULL DEFAULT ''
);

-- +migrate Up
CREATE INDEX idx_transaction_trading_strategy ON transaction_table(trading_strategy_id);
CREATE INDEX idx_transaction_coin_id_created_at ON transaction_table(coin_id, created_at);
CREATE INDEX idx_transaction_created_at ON transaction_table(created_at);