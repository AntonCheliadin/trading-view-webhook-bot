package controller

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"io"
	"net/http"
	"tradingViewWebhookBot/internal/api"
	"tradingViewWebhookBot/internal/constants"
	"tradingViewWebhookBot/internal/dto/tradingview"
	"tradingViewWebhookBot/internal/repository"
	"tradingViewWebhookBot/internal/service/orders"
	"tradingViewWebhookBot/internal/telegram"
)

type AlertWebhookController struct {
	strategyRepo        repository.TradingStrategy
	transactionRepo     repository.Transaction
	coinRepository      repository.Coin
	exchangeApi         api.ExchangeApi
	telegramClient      *telegram.TelegramClient
	orderManagerService *orders.OrderManagerService
}

func NewAlertWebhookController(
	strategyRepo repository.TradingStrategy,
	transactionRepo repository.Transaction,
	coinRepo repository.Coin,
	exchangeApi api.ExchangeApi,
	telegramClient *telegram.TelegramClient,
	orderManagerService *orders.OrderManagerService,
) *AlertWebhookController {
	return &AlertWebhookController{
		strategyRepo:        strategyRepo,
		transactionRepo:     transactionRepo,
		coinRepository:      coinRepo,
		exchangeApi:         exchangeApi,
		telegramClient:      telegramClient,
		orderManagerService: orderManagerService,
	}
}

func (c *AlertWebhookController) HandleAlert(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		zap.L().Error("Error reading request body", zap.Error(err))
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	c.telegramClient.SendMessage(fmt.Sprintf("Alert triggered: %s", body))

	var alertRequest tradingview.AlertRequestDto
	if err := json.Unmarshal(body, &alertRequest); err != nil {
		zap.L().Error("Error parsing request body", zap.Error(err))
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(alertRequest); err != nil {
		c.telegramClient.SendMessage(fmt.Sprintf("AlertRequest is not valid: %s", alertRequest.String()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	strategy, err := c.strategyRepo.FindByTag(alertRequest.Tag)
	if err != nil || strategy == nil {
		c.telegramClient.SendMessage(fmt.Sprintf("Trading strategy not found: %s", alertRequest.Tag))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	coin, err := c.coinRepository.FindBySymbol(alertRequest.Ticker)
	if err != nil || coin == nil {
		c.telegramClient.SendMessage(fmt.Sprintf("Coin not found: %s", alertRequest.Ticker))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	openedTransaction, err := c.transactionRepo.FindOpenedTransactionByCoin(strategy.Id, coin.Id)
	if err != nil {
		c.telegramClient.SendMessage(fmt.Sprintf("Error during FindOpenedTransactionByCoin: %s", coin.Id))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if openedTransaction == nil {
		c.orderManagerService.OpenOrderAllIn(
			strategy,
			coin,
			alertRequest.GetFuturesType(),
		)
	} else {
		c.orderManagerService.CloseOrder(
			strategy,
			openedTransaction,
			coin,
			alertRequest.GetPriceFloat(),
			constants.FUTURES,
		)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alert processed successfully"))
}
