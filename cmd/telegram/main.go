package main

import (
	"flag"
	"fmt"
	"os"

	"tradingViewWebhookBot/internal/telegram"

	"go.uber.org/zap"
)

// go run cmd/telegram/main.go -message="Your message here"
func main() {
	// Parse command line arguments
	message := flag.String("message", "", "Message to send via Telegram")
	flag.Parse()

	if *message == "" {
		defaultMessage := "Test message"
		message = &defaultMessage
	}

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	client := telegram.NewTelegramClient()

	client.SendMessage(*message)

	logger.Info("Message sent successfully",
		zap.String("message", *message),
	)
}
