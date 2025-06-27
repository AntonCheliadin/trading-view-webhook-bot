package controller

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"tradingViewWebhookBot/internal/api"
	"tradingViewWebhookBot/internal/constants"
	"tradingViewWebhookBot/internal/constants/futureType"
	"tradingViewWebhookBot/internal/repository"
	"tradingViewWebhookBot/internal/service/orders"
	"tradingViewWebhookBot/internal/telegram"
)

type DebugController struct {
	strategyRepo        repository.TradingStrategy
	transactionRepo     repository.Transaction
	coinRepository      repository.Coin
	exchangeApi         api.ExchangeApi
	telegramClient      *telegram.TelegramClient
	orderManagerService *orders.OrderManagerService
}

func NewDebugController(
	strategyRepo repository.TradingStrategy,
	transactionRepo repository.Transaction,
	coinRepo repository.Coin,
	exchangeApi api.ExchangeApi,
	telegramClient *telegram.TelegramClient,
	orderManagerService *orders.OrderManagerService,
) *DebugController {
	return &DebugController{
		strategyRepo:        strategyRepo,
		transactionRepo:     transactionRepo,
		coinRepository:      coinRepo,
		exchangeApi:         exchangeApi,
		telegramClient:      telegramClient,
		orderManagerService: orderManagerService,
	}
}

// OpenOrderAllIn handles the debug endpoint for opening an order with all available funds
func (c *DebugController) OpenOrderAllIn(w http.ResponseWriter, r *http.Request) {
	// Get parameters from URL
	symbol := chi.URLParam(r, "symbol")
	if symbol == "" {
		http.Error(w, "symbol is required", http.StatusBadRequest)
		return
	}

	strategyTag := chi.URLParam(r, "strategyTag")
	if strategyTag == "" {
		http.Error(w, "strategyTag is required", http.StatusBadRequest)
		return
	}

	futureTypeStr := chi.URLParam(r, "futureType")
	if futureTypeStr == "" {
		http.Error(w, "futureType is required (0 for LONG, 1 for SHORT)", http.StatusBadRequest)
		return
	}

	// Parse futureType
	var ft futureType.FuturesType
	if futureTypeStr == "0" {
		ft = futureType.LONG
	} else if futureTypeStr == "1" {
		ft = futureType.SHORT
	} else {
		http.Error(w, "futureType must be 0 (LONG) or 1 (SHORT)", http.StatusBadRequest)
		return
	}

	// Find coin
	coin, err := c.coinRepository.FindBySymbol(symbol)
	if err != nil || coin == nil {
		zap.S().Error("Failed to find coin by symbol. Error: ", err)
		http.Error(w, fmt.Sprintf("Failed to find coin by symbol %s. Error: %s", symbol, err), http.StatusBadRequest)
		return
	}

	// Find strategy
	strategy, err := c.strategyRepo.FindByTag(strategyTag)
	if err != nil || strategy == nil {
		zap.S().Error("Failed to find strategy by tag. Error: ", err)
		http.Error(w, fmt.Sprintf("Failed to find strategy by tag %s. Error: %s", strategyTag, err), http.StatusBadRequest)
		return
	}

	// Execute the order
	c.orderManagerService.OpenOrderAllIn(strategy, coin, ft)

	// Return success response
	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Order opened for %s with strategy %s and futureType %s", symbol, strategyTag, futureTypeStr),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CloseOrder handles the debug endpoint for closing an order
func (c *DebugController) CloseOrder(w http.ResponseWriter, r *http.Request) {
	// Get parameters from URL
	symbol := chi.URLParam(r, "symbol")
	if symbol == "" {
		http.Error(w, "symbol is required", http.StatusBadRequest)
		return
	}

	strategyTag := chi.URLParam(r, "strategyTag")
	if strategyTag == "" {
		http.Error(w, "strategyTag is required", http.StatusBadRequest)
		return
	}

	// Find coin
	coin, err := c.coinRepository.FindBySymbol(symbol)
	if err != nil || coin == nil {
		zap.S().Error("Failed to find coin by symbol. Error: ", err)
		http.Error(w, fmt.Sprintf("Failed to find coin by symbol %s. Error: %s", symbol, err), http.StatusBadRequest)
		return
	}

	// Find strategy
	strategy, err := c.strategyRepo.FindByTag(strategyTag)
	if err != nil || strategy == nil {
		zap.S().Error("Failed to find strategy by tag. Error: ", err)
		http.Error(w, fmt.Sprintf("Failed to find strategy by tag %s. Error: %s", strategyTag, err), http.StatusBadRequest)
		return
	}

	// Find open transaction
	openedTransaction, err := c.transactionRepo.FindOpenedTransactionByCoin(strategy.Id, coin.Id)
	if err != nil {
		zap.S().Error("Error during FindOpenedTransactionByCoin: ", err)
		http.Error(w, fmt.Sprintf("Error during FindOpenedTransactionByCoin: %s", err), http.StatusInternalServerError)
		return
	}

	if openedTransaction == nil {
		http.Error(w, fmt.Sprintf("No open transaction found for coin %s and strategy %s", symbol, strategyTag), http.StatusNotFound)
		return
	}

	// Get current price
	currentPrice, err := c.exchangeApi.GetCurrentCoinPrice(coin)
	if err != nil {
		zap.S().Error("Error getting current price: ", err)
		http.Error(w, fmt.Sprintf("Error getting current price: %s", err), http.StatusInternalServerError)
		return
	}

	// Close the order
	closedTransaction := c.orderManagerService.CloseOrder(
		strategy,
		openedTransaction,
		coin,
		currentPrice,
		constants.FUTURES,
	)

	if closedTransaction == nil {
		http.Error(w, "Failed to close order", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Order closed for %s with strategy %s", symbol, strategyTag),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
