package worker

import (
	"fmt"
	"go.uber.org/zap"
	"time"
	"tradingViewWebhookBot/internal/api"
	"tradingViewWebhookBot/internal/repository"
	"tradingViewWebhookBot/internal/service/orders"
	"tradingViewWebhookBot/internal/telegram"
)

// OrderProfitChecker is a worker that periodically checks the profit of opened orders
type OrderProfitChecker struct {
	strategyRepo        repository.TradingStrategy
	transactionRepo     repository.Transaction
	coinRepository      repository.Coin
	exchangeApi         api.ExchangeApi
	telegramClient      *telegram.TelegramClient
	orderManagerService *orders.OrderManagerService
	checkInterval       time.Duration
	stopChan            chan struct{}
}

// NewOrderProfitChecker creates a new OrderProfitChecker
func NewOrderProfitChecker(
	strategyRepo repository.TradingStrategy,
	transactionRepo repository.Transaction,
	coinRepo repository.Coin,
	exchangeApi api.ExchangeApi,
	telegramClient *telegram.TelegramClient,
	orderManagerService *orders.OrderManagerService,
) *OrderProfitChecker {
	return &OrderProfitChecker{
		strategyRepo:        strategyRepo,
		transactionRepo:     transactionRepo,
		coinRepository:      coinRepo,
		exchangeApi:         exchangeApi,
		telegramClient:      telegramClient,
		orderManagerService: orderManagerService,
		checkInterval:       4 * time.Hour, // Check every 4 hours
		stopChan:            make(chan struct{}),
	}
}

// Start begins the periodic checking of order profits
func (c *OrderProfitChecker) Start() {
	go func() {
		// Run immediately on start
		c.CheckOrderProfits()

		ticker := time.NewTicker(c.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.CheckOrderProfits()
			case <-c.stopChan:
				return
			}
		}
	}()
}

// Stop halts the periodic checking
func (c *OrderProfitChecker) Stop() {
	close(c.stopChan)
}

// CheckOrderProfits checks the profit of all opened orders and sends the information to Telegram
func (c *OrderProfitChecker) CheckOrderProfits() {
	zap.S().Info("Starting to check profits for all opened orders")

	// Get all trading strategies
	strategies, err := c.strategyRepo.List()
	if err != nil {
		zap.S().Error("Failed to get trading strategies:", err)
		return
	}

	var totalOpenedOrders int
	var message string

	// For each strategy, get all opened transactions
	for _, strategy := range strategies {
		openedTransactions, err := c.transactionRepo.FindAllOpenedTransactions(strategy)
		if err != nil {
			zap.S().Error("Failed to get opened transactions for strategy: ", strategy.Tag, err)
			continue
		}

		if len(openedTransactions) == 0 {
			continue
		}

		totalOpenedOrders += len(openedTransactions)

		// Add strategy information to the message
		message += fmt.Sprintf("\n<b>Strategy: %s</b>\n", strategy.Tag)

		// For each opened transaction, calculate profit
		for _, transaction := range openedTransactions {
			coin, err := c.coinRepository.FindById(transaction.CoinId)
			if err != nil {
				zap.S().Error("Failed to get coin for transaction:", transaction.Id, err)
				continue
			}

			// Calculate profit with leverage (percentage and dollars)
			profitPercent, profitDollars, err := c.orderManagerService.CalculateCurrentProfitWithLeverage(coin, transaction)
			if err != nil {
				zap.S().Error("Failed to calculate profit for transaction:", transaction.Id, err)
				continue
			}

			// Get current price
			currentPrice, err := c.exchangeApi.GetCurrentCoinPrice(coin)
			if err != nil {
				zap.S().Error("Failed to get current price for coin:", coin.Symbol, err)
				continue
			}

			// Add transaction information to the message
			message += fmt.Sprintf("Coin: %s, Entry: %.2f, Current: %.2f, Cost: %.2f Profit: %.2f%% ($%.2f)\n",
				coin.Symbol,
				transaction.Price,
				currentPrice,
				transaction.TotalCost,
				profitPercent,
				profitDollars)
		}
	}

	// Send message to Telegram if there are opened orders
	if totalOpenedOrders > 0 {
		fullMessage := fmt.Sprintf("<b>Order Profit Report</b>\nTotal opened orders: %d\n%s", totalOpenedOrders, message)
		c.telegramClient.SendMessage(fullMessage)
		zap.S().Info("Sent profit report to Telegram")
	} else {
		zap.S().Info("No opened orders found")
	}
}
