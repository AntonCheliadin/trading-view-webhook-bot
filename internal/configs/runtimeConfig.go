package configs

import (
	"go.uber.org/zap"
	"time"
	telegramApi "tradingViewWebhookBot/internal/telegram"
)

var RuntimeConfig *config

func NewRuntimeConfig(logger *zap.Logger) *config {
	if RuntimeConfig != nil {
		panic("Unexpected try to create second instance")
	}

	RuntimeConfig = &config{
		TradingEnabled: true,
		telegramClient: telegramApi.NewTelegramClient(),
	}
	return RuntimeConfig
}

type config struct {
	/**
	Transactions switcher, enable/disable buy and sell transactions.
	*/
	TradingEnabled bool

	/**
	Limit spend money for the last 24 hours.
	 0 - without limit.
	*/
	LimitSpendDay int

	telegramClient *telegramApi.TelegramClient
}

func (c *config) DisableBuyingForHour() {
	c.TradingEnabled = false
	c.telegramClient.SendMessage("Trading has been disabled for an hour.")

	select {
	case <-time.After(time.Hour):
		c.TradingEnabled = true
		c.telegramClient.SendMessage("Trading has been enabled.")
	}
}

func (c *config) HasLimitSpendDay() bool {
	return c.LimitSpendDay > 0
}
