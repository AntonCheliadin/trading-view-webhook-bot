-- +migrate Up
UPDATE trading_strategies SET tag = 'ETH_SSL_WAE_LONG' where name = 'ETH 4H';