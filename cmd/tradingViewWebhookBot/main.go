package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"tradingViewWebhookBot/internal/api/bybit"
	"tradingViewWebhookBot/internal/controller"
	"tradingViewWebhookBot/internal/database"
	"tradingViewWebhookBot/internal/logger"
	"tradingViewWebhookBot/internal/repository"
	"tradingViewWebhookBot/internal/service/date"
	"tradingViewWebhookBot/internal/service/orders"
	"tradingViewWebhookBot/internal/telegram"
)

type App struct {
	logger *zap.Logger
	db     *sqlx.DB
	router *chi.Mux
}

func main() {
	app, err := initializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.cleanup()

	if err := app.run(); err != nil {
		app.logger.Fatal("Server failed to start", zap.Error(err))
	}
}

func initializeApp() (*App, error) {
	err := initConfig()
	if err != nil {
		return nil, err
	}

	logger := logger.InitLogger()
	zap.ReplaceGlobals(logger)

	logger.Info("tradingViewWebhookBot starting...")

	if _, err := os.Stat(".env"); err == nil {
		logger.Info(".env file found")
		if err := godotenv.Load(); err != nil {
			logger.Error(".env file found but loading failed: %v", zap.Error(err))
		}
	} else {
		logger.Warn(".env file not found")
	}

	db, err := database.NewPostgresConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	router := initializeRouter(db)

	return &App{
		logger: logger,
		db:     db,
		router: router,
	}, nil
}

func initializeRouter(db *sqlx.DB) *chi.Mux {
	repos := repository.NewRepositories(db)

	exchangeApi := bybit.NewBybitApi(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"))

	telegramClient := telegram.NewTelegramClient()

	orderManagerService := orders.NewOrderManagerService(
		repos.Transaction,
		exchangeApi,
		date.GetClock(),
		telegramClient,
		viper.GetInt64("default.leverage"))

	healthController := controller.NewHealthController()
	coinController := controller.NewCoinController(repos.Coin, exchangeApi, telegramClient)
	webhookController := controller.NewAlertWebhookController(repos.TradingStrategy, repos.Transaction, repos.Coin, exchangeApi, telegramClient, orderManagerService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	setupRoutes(r, healthController, coinController, webhookController)

	return r
}

func setupRoutes(r *chi.Mux, healthController *controller.HealthController, coinController *controller.CoinController,
	webhookController *controller.AlertWebhookController) {
	r.Get("/health", healthController.HealthCheck)

	// Coin routes
	r.Route("/coins", func(r chi.Router) {
		r.Get("/symbol/{symbol}", coinController.GetCoinBySymbol)
		r.Get("/id/{id}", coinController.GetCoinByID)
		r.Get("/price/{symbol}", coinController.GetCurrentPrice)
	})

	r.HandleFunc("/webhook/alert", webhookController.HandleAlert)

}

func (a *App) run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	a.logger.Info("Server starting", zap.String("port", port))
	return http.ListenAndServe(":"+port, a.router)
}

func (a *App) cleanup() {
	if err := a.logger.Sync(); err != nil {
		log.Printf("Failed to sync logger: %v", err)
	}
	if err := a.db.Close(); err != nil {
		log.Printf("Failed to close database connection: %v", err)
	}
}

func initConfig() error {
	viper.AddConfigPath("internal/configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
