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
	"tradingViewWebhookBot/internal/controller"
	"tradingViewWebhookBot/internal/database"
	"tradingViewWebhookBot/internal/logger"
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
	// Initialize config
	err := initConfig()
	if err != nil {
		return nil, err
	}

	// Initialize logger
	logger := logger.InitLogger()
	zap.ReplaceGlobals(logger)

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Error("Warning: Error loading .env file: %v", zap.Error(err))
	}

	// Initialize database
	db, err := database.NewPostgresConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Initialize router
	router := initializeRouter(db)

	return &App{
		logger: logger,
		db:     db,
		router: router,
	}, nil
}

func initializeRouter(db *sqlx.DB) *chi.Mux {
	//repos := repository.NewRepositories(db)
	//
	//exchangeApi := bybit.NewBybitApi(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"))
	//
	//telegramClient := telegram.NewTelegramClient()
	//
	//orderManagerService := orders.NewOrderManagerService(
	//	repos.Transaction,
	//	exchangeApi,
	//	date.GetClock(),
	//	telegramClient,
	//	viper.GetInt64("default.leverage"))

	// Initialize controllers
	healthController := controller.NewHealthController()
	//coinController := controller.NewCoinController(repos.Coin, exchangeApi, telegramClient)
	//webhookController := controller.NewAlertWebhookController(repos.TradingStrategy, repos.Transaction, repos.Coin, exchangeApi, telegramClient, orderManagerService)

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Get("/health", healthController.HealthCheck)
	//setupRoutes(r, healthController, coinController, webhookController)

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
