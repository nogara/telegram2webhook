package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	TelegramToken string
	WebhookURL    string
	Debug         bool
}

type WebhookPayload struct {
	Message Message `json:"message"`
}

type Message struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
	From From   `json:"from"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type From struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

func loadConfig() (*Config, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	webhookURL := os.Getenv("N8N_WEBHOOK_URL")
	if webhookURL == "" {
		return nil, fmt.Errorf("N8N_WEBHOOK_URL environment variable is not set")
	}

	debug := os.Getenv("DEBUG") == "true"

	return &Config{
		TelegramToken: token,
		WebhookURL:    webhookURL,
		Debug:         debug,
	}, nil
}

func sendToWebhook(webhookURL string, payload WebhookPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status code: %d", resp.StatusCode)
	}

	log.Printf("Successfully sent message to webhook. Status: %d", resp.StatusCode)
	return nil
}

func main() {
	log.Println("Starting Telegram to Webhook bridge...")

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	bot.Debug = config.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)
	log.Printf("Webhook URL: %s", config.WebhookURL)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Println("Listening for messages...")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("Received message from %s: %s", update.Message.From.UserName, update.Message.Text)

		payload := WebhookPayload{
			Message: Message{
				Text: update.Message.Text,
				Chat: Chat{
					ID: update.Message.Chat.ID,
				},
				From: From{
					Username:  update.Message.From.UserName,
					FirstName: update.Message.From.FirstName,
				},
			},
		}

		if err := sendToWebhook(config.WebhookURL, payload); err != nil {
			log.Printf("Error sending to webhook: %v", err)
			continue
		}

		log.Printf("Message forwarded successfully")
	}
}
