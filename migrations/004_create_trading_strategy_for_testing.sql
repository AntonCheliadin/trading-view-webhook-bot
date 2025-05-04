-- +migrate Up
DELETE trading_strategies;

-- +migrate Up
INSERT INTO trading_strategies (name, description, tag, created_at, updated_at)
VALUES ('Test Trading strategy', 'for testing', 'TEST_STRATEGY', NOW(), NOW());