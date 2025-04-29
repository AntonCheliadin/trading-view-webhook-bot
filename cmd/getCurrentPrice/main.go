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
	// Initialize logger
	logger := logger.InitLogger()
	defer logger.Sync()

	// Load environment variables
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			logger.Error("Warning: Error loading .env file: %v", zap.Error(err))
		}
	}

	// Initialize database connection
	db, err := database.NewPostgresConnection()
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Create repositories and services
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
