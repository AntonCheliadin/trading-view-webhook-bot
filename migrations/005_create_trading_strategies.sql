-- +migrate Up
INSERT INTO trading_strategies (name, description, tag, created_at, updated_at)
VALUES ('ATOM 4h', 'SSL WAE_LB ATR LONG+SHORT', 'ATOM_SSL_WAE_LONG_SHORT', NOW(), NOW());

-- +migrate Up
INSERT INTO trading_strategies (name, description, tag, created_at, updated_at)
VALUES ('AVAX 4h', 'SSL WAE_LB ATR LONG+SHORT', 'AVAX_SSL_WAE_LONG_SHORT', NOW(), NOW());

-- +migrate Up
INSERT INTO trading_strategies (name, description, tag, created_at, updated_at)
VALUES ('BTC 4h', 'Test Trading strategy (BTC simple pullback 4H)', 'BTC_SIMPLE_PULLBACK', NOW(), NOW());

-- +migrate Up
INSERT INTO trading_strategies (name, description, tag, created_at, updated_at)
VALUES ('BTC 1D', 'ChatGpt without SL and TP (BTC ChatGpt 1D)', 'BTC_CHAT_GPT', NOW(), NOW());

-- +migrate Up
INSERT INTO trading_strategies (name, description, tag, created_at, updated_at)
VALUES ('ETH 4H', 'SSL WAE_LB ATR LONG (ETH 4H):', 'ETH_SSL_WAE_LONG', NOW(), NOW());
