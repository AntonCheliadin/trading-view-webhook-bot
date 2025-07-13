package telegram

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// ProfitChecker defines the interface for checking order profits
type ProfitChecker interface {
	CheckOrderProfits()
}

type TelegramClient struct {
	logger        *zap.Logger
	apiKey        string
	chatID        string
	apiBaseURL    string
	enabled       bool
	bot           *tgbotapi.BotAPI
	profitChecker ProfitChecker
}

func (t *TelegramClient) SetProfitChecker(profitChecker ProfitChecker) {
	t.profitChecker = profitChecker
}

func NewTelegramClient() *TelegramClient {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_API_KEY"))
	if err != nil {
		zap.L().Error("Failed to initialize Telegram bot", zap.Error(err))
		return &TelegramClient{
			logger:     zap.L(),
			apiKey:     os.Getenv("TELEGRAM_BOT_API_KEY"),
			chatID:     os.Getenv("TELEGRAM_BOT_CHAT_ID"),
			apiBaseURL: os.Getenv("TELEGRAM_API_BASE_URL"),
			enabled:    os.Getenv("TELEGRAM_ENABLED") == "true",
		}
	}

	bot.Debug = true // Set to false in production
	zap.L().Info("Authorized on account", zap.String("username", bot.Self.UserName))

	telegramClient := &TelegramClient{
		logger:     zap.L(),
		apiKey:     os.Getenv("TELEGRAM_BOT_API_KEY"),
		chatID:     os.Getenv("TELEGRAM_BOT_CHAT_ID"),
		apiBaseURL: os.Getenv("TELEGRAM_API_BASE_URL"),
		enabled:    os.Getenv("TELEGRAM_ENABLED") == "true",
		bot:        bot,
	}

	telegramClient.StartMessageHandler()

	return telegramClient
}

// StartMessageHandler starts listening for messages
func (t *TelegramClient) StartMessageHandler() {
	if !t.enabled || t.bot == nil {
		t.logger.Info("Telegram bot is disabled or not initialized")
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 1000 * 10

	updates := t.bot.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			// Log message metadata
			t.logger.Info("New Message Received",
				zap.Int("message_id", update.Message.MessageID),
				zap.Int64("user_id", update.Message.From.ID),
				zap.String("username", update.Message.From.UserName),
				zap.Int64("chat_id", update.Message.Chat.ID),
				zap.String("chat_type", update.Message.Chat.Type),
				zap.Time("time", update.Message.Time()),
				zap.String("text", update.Message.Text),
			)

			if update.Message.ForwardFrom != nil {
				t.logger.Info("Forwarded Message Info",
					zap.String("from_username", update.Message.ForwardFrom.UserName),
					zap.Int64("from_user_id", update.Message.ForwardFrom.ID),
				)
			}

			if strings.ToLower(update.Message.Text) == "positions" {
				t.profitChecker.CheckOrderProfits()
			} else {
				// For other messages, send a confirmation reply
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					fmt.Sprintf("Received your message: %s", update.Message.Text))
				msg.ReplyToMessageID = update.Message.MessageID

				if _, err := t.bot.Send(msg); err != nil {
					t.logger.Error("Error sending reply",
						zap.Error(err),
						zap.Int64("chat_id", update.Message.Chat.ID),
					)
				}
			}
		}
	}()
}

func (t *TelegramClient) SendMessage(text string) {
	if !t.enabled {
		t.logger.Debug("Telegram message (disabled)", zap.String("text", text))
		return
	}

	if t.apiKey == "" || t.chatID == "" {
		t.logger.Error("telegram configuration missing: API_KEY or CHAT_ID not set")
		return
	}

	apiURL := t.apiBaseURL + t.apiKey + "/sendMessage"

	response, err := http.PostForm(
		apiURL,
		url.Values{
			"chat_id":    {t.chatID},
			"text":       {text},
			"parse_mode": {"HTML"},
		})

	if err != nil {
		t.logger.Error("failed to send telegram message", zap.Error(err))
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.logger.Error("failed to read telegram response", zap.Error(err))
		return
	}

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("telegram API error: status=%d, body=%s", response.StatusCode, string(body))
		t.logger.Error("telegram API error", zap.Error(err))
		return
	}

	t.logger.Debug("Telegram message sent successfully",
		zap.String("text", text),
		zap.Int("status", response.StatusCode))
}
