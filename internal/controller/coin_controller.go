package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"tradingViewWebhookBot/internal/api"
	"tradingViewWebhookBot/internal/repository"
	"tradingViewWebhookBot/internal/telegram"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type CoinController struct {
	repo           repository.Coin
	exchangeApi    api.ExchangeApi
	telegramClient *telegram.TelegramClient
	logger         *zap.Logger
}

func NewCoinController(
	repo repository.Coin,
	exchangeApi api.ExchangeApi,
	telegramClient *telegram.TelegramClient,
) *CoinController {
	return &CoinController{
		repo:           repo,
		exchangeApi:    exchangeApi,
		telegramClient: telegramClient,
		logger:         zap.L(),
	}
}

func (c *CoinController) GetCoinBySymbol(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	if symbol == "" {
		http.Error(w, "symbol is required", http.StatusBadRequest)
		return
	}

	coin, err := c.repo.FindBySymbol(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(coin)
}

func (c *CoinController) GetCoinByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	coin, err := c.repo.FindById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(coin)
}

func (c *CoinController) GetCurrentPrice(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	if symbol == "" {
		http.Error(w, "symbol is required", http.StatusBadRequest)
		return
	}

	coin, err := c.repo.FindBySymbol(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	price, err := c.exchangeApi.GetCurrentCoinPrice(coin)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get price: %v", err), http.StatusInternalServerError)
		return
	}

	c.telegramClient.SendMessage(fmt.Sprintf("Current %s price: $%.2f", coin.Symbol, price))

	response := struct {
		Symbol string  `json:"symbol"`
		Price  float64 `json:"price"`
	}{
		Symbol: coin.Symbol,
		Price:  price,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
