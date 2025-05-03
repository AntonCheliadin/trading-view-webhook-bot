package main

import (
	"os"
	"tradingViewWebhookBot/internal/api"
	"tradingViewWebhookBot/internal/api/bybit"
	"tradingViewWebhookBot/internal/database"
	"tradingViewWebhookBot/internal/domain"
	"tradingViewWebhookBot/internal/logger"
	"tradingViewWebhookBot/internal/repository"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger := logger.InitLogger()
	defer logger.Sync()

	logger.Info("getCurrentPrice starting...")

	if err := godotenv.Load(); err != nil {
		logger.Fatal("Error loading .env file", zap.Error(err))
	}

	db, err := database.NewPostgresConnection()
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	coinRepo := repository.NewCoinRepository(db)
	exchangeApi := bybit.NewBybitApi(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"))
	//exchangeApi := binance.NewBinanceApi()

	coin, err := coinRepo.FindBySymbol("BTCUSDT")
	if err != nil {
		logger.Fatal("Failed to find coin", zap.Error(err))
	}

	getCurrentPrice(err, exchangeApi, coin, logger)
}

func getCurrentPrice(err error, exchangeApi api.ExchangeApi, coin *domain.Coin, logger *zap.Logger) {
	price, err := exchangeApi.GetCurrentCoinPrice(coin)
	if err != nil {
		logger.Fatal("Failed to get current price",
			zap.String("symbol", coin.Symbol),
			zap.Error(err),
		)
	}

	logger.Info("Current price",
		zap.String("symbol", coin.Symbol),
		zap.Float64("price", price),
	)
}
