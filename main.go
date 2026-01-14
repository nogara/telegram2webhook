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

type BotConfig struct {
	TelegramToken string `json:"token"`
	WebhookURL    string `json:"webhook"`
}

type Config struct {
	Bots  []BotConfig
	Debug bool
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
	debug := os.Getenv("DEBUG") == "true"

	pairsJSON := os.Getenv("TELEGRAM_WEBHOOK_PAIRS")
	if pairsJSON != "" {
		var bots []BotConfig
		if err := json.Unmarshal([]byte(pairsJSON), &bots); err != nil {
			return nil, fmt.Errorf("failed to parse TELEGRAM_WEBHOOK_PAIRS: %w", err)
		}
		if len(bots) == 0 {
			return nil, fmt.Errorf("TELEGRAM_WEBHOOK_PAIRS must include at least one bot")
		}
		for index, bot := range bots {
			if bot.TelegramToken == "" {
				return nil, fmt.Errorf("TELEGRAM_WEBHOOK_PAIRS[%d] token is required", index)
			}
			if bot.WebhookURL == "" {
				return nil, fmt.Errorf("TELEGRAM_WEBHOOK_PAIRS[%d] webhook is required", index)
			}
		}
		return &Config{Bots: bots, Debug: debug}, nil
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	webhookURL := os.Getenv("N8N_WEBHOOK_URL")
	if webhookURL == "" {
		return nil, fmt.Errorf("N8N_WEBHOOK_URL environment variable is not set")
	}

	return &Config{
		Bots: []BotConfig{
			{
				TelegramToken: token,
				WebhookURL:    webhookURL,
			},
		},
		Debug: debug,
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

	for index, botConfig := range config.Bots {
		botIndex := index + 1
		go func(cfg BotConfig, label int) {
			bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
			if err != nil {
				log.Fatalf("Failed to create bot %d: %v", label, err)
			}

			bot.Debug = config.Debug

			log.Printf("Bot %d authorized on account %s", label, bot.Self.UserName)
			log.Printf("Bot %d webhook URL: %s", label, cfg.WebhookURL)

			u := tgbotapi.NewUpdate(0)
			u.Timeout = 60

			updates := bot.GetUpdatesChan(u)

			log.Printf("Bot %d listening for messages...", label)

			for update := range updates {
				if update.Message == nil {
					continue
				}

				log.Printf("Bot %d received message from %s: %s", label, update.Message.From.UserName, update.Message.Text)

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

				if err := sendToWebhook(cfg.WebhookURL, payload); err != nil {
					log.Printf("Bot %d error sending to webhook: %v", label, err)
					continue
				}

				log.Printf("Bot %d message forwarded successfully", label)
			}
		}(botConfig, botIndex)
	}

	select {}
}
