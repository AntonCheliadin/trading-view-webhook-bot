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
	"tradingViewWebhookBot/internal/api"
	"tradingViewWebhookBot/internal/api/bybit"
	"tradingViewWebhookBot/internal/controller"
	"tradingViewWebhookBot/internal/database"
	"tradingViewWebhookBot/internal/logger"
	"tradingViewWebhookBot/internal/repository"
	"tradingViewWebhookBot/internal/service/date"
	"tradingViewWebhookBot/internal/service/orders"
	"tradingViewWebhookBot/internal/telegram"
	"tradingViewWebhookBot/internal/worker"
)

type App struct {
	logger             *zap.Logger
	db                 *sqlx.DB
	router             *chi.Mux
	orderProfitChecker *worker.OrderProfitChecker
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

	// Initialize repositories, services, and other components
	repos := repository.NewRepositories(db)

	exchangeApi := bybit.NewBybitApi(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"))

	telegramClient := telegram.NewTelegramClient()

	orderManagerService := orders.NewOrderManagerService(
		repos.Transaction,
		exchangeApi,
		date.GetClock(),
		telegramClient,
		viper.GetInt64("default.leverage"))

	// Initialize the order profit checker worker
	orderProfitChecker := worker.NewOrderProfitChecker(
		repos.TradingStrategy,
		repos.Transaction,
		repos.Coin,
		exchangeApi,
		telegramClient,
		orderManagerService,
	)

	// Set the profit checker on the telegram client to handle "positions" command
	telegramClient.SetProfitChecker(orderProfitChecker)

	// Initialize controllers and router
	router := initializeRouter(repos, exchangeApi, telegramClient, orderManagerService)

	return &App{
		logger:             logger,
		db:                 db,
		router:             router,
		orderProfitChecker: orderProfitChecker,
	}, nil
}

func initializeRouter(
	repos *repository.Repository,
	exchangeApi api.ExchangeApi,
	telegramClient *telegram.TelegramClient,
	orderManagerService *orders.OrderManagerService,
) *chi.Mux {
	healthController := controller.NewHealthController()
	coinController := controller.NewCoinController(repos.Coin, exchangeApi, telegramClient)
	webhookController := controller.NewAlertWebhookController(repos.TradingStrategy, repos.Transaction, repos.Coin, exchangeApi, telegramClient, orderManagerService)
	debugController := controller.NewDebugController(repos.TradingStrategy, repos.Transaction, repos.Coin, exchangeApi, telegramClient, orderManagerService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	setupRoutes(r, healthController, coinController, webhookController, debugController)

	return r
}

func setupRoutes(r *chi.Mux, healthController *controller.HealthController, coinController *controller.CoinController,
	webhookController *controller.AlertWebhookController, debugController *controller.DebugController) {
	r.Get("/health", healthController.HealthCheck)

	// Coin routes
	r.Route("/coins", func(r chi.Router) {
		r.Get("/symbol/{symbol}", coinController.GetCoinBySymbol)
		r.Get("/id/{id}", coinController.GetCoinByID)
		r.Get("/price/{symbol}", coinController.GetCurrentPrice)
	})

	r.HandleFunc("/webhook/alert", webhookController.HandleAlert)

	// Debug routes
	r.Route("/debug", func(r chi.Router) {
		r.Get("/open/{symbol}/{strategyTag}/{futureType}", debugController.OpenOrderAllIn)
		r.Get("/close/{symbol}/{strategyTag}", debugController.CloseOrder)
		// Profit check is now handled by the Telegram "positions" command
	})
}

func (a *App) run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the order profit checker worker
	a.orderProfitChecker.Start()
	a.logger.Info("Order profit checker worker started")

	a.logger.Info("Server starting", zap.String("port", port))
	return http.ListenAndServe(":"+port, a.router)
}

func (a *App) cleanup() {
	// Stop the order profit checker worker
	a.orderProfitChecker.Stop()
	a.logger.Info("Order profit checker worker stopped")

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
