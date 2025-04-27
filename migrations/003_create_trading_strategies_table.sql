-- +migrate Up
CREATE TABLE IF NOT EXISTS trading_strategies
(
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    tag TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMP    NOT NULL,
    updated_at  TIMESTAMP    NOT NULL
);

-- +migrate Up
INSERT INTO trading_strategies (name, description, tag, created_at, updated_at)
VALUES ('BTC Simple Pullback Strategy', NULL, 'BTC_PULLBACK', NOW(), NOW());

-- +migrate Up
INSERT INTO trading_strategies (name, description, tag, created_at, updated_at)
VALUES ('ChatGpt Trading strategy', 'BTC 1D ChatGpt without SL and TP', 'ChatGpt', NOW(), NOW());