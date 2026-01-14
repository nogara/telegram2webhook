package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadConfigPairsSuccess(t *testing.T) {
	t.Setenv("TELEGRAM_WEBHOOK_PAIRS", `[{"token":"token-one","webhook":"https://example.com/one"},{"token":"token-two","webhook":"https://example.com/two"}]`)
	t.Setenv("DEBUG", "true")

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("expected config, got error: %v", err)
	}

	if !config.Debug {
		t.Fatalf("expected debug true")
	}

	if len(config.Bots) != 2 {
		t.Fatalf("expected 2 bots, got %d", len(config.Bots))
	}

	if config.Bots[0].TelegramToken != "token-one" || config.Bots[0].WebhookURL != "https://example.com/one" {
		t.Fatalf("unexpected first bot config: %+v", config.Bots[0])
	}
}

func TestLoadConfigPairsInvalidJSON(t *testing.T) {
	t.Setenv("TELEGRAM_WEBHOOK_PAIRS", `[{"token":"token-one","webhook":}`)

	_, err := loadConfig()
	if err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestLoadConfigPairsMissingFields(t *testing.T) {
	t.Setenv("TELEGRAM_WEBHOOK_PAIRS", `[{"token":"","webhook":"https://example.com"}]`)

	_, err := loadConfig()
	if err == nil {
		t.Fatalf("expected error for missing token")
	}
}

func TestLoadConfigFallback(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "legacy-token")
	t.Setenv("N8N_WEBHOOK_URL", "https://example.com/legacy")

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("expected config, got error: %v", err)
	}

	if len(config.Bots) != 1 {
		t.Fatalf("expected 1 bot, got %d", len(config.Bots))
	}

	if config.Bots[0].TelegramToken != "legacy-token" || config.Bots[0].WebhookURL != "https://example.com/legacy" {
		t.Fatalf("unexpected legacy config: %+v", config.Bots[0])
	}
}

func TestLoadConfigMissingLegacy(t *testing.T) {
	t.Setenv("TELEGRAM_WEBHOOK_PAIRS", "")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("N8N_WEBHOOK_URL", "")

	_, err := loadConfig()
	if err == nil {
		t.Fatalf("expected error when env vars missing")
	}
}

func TestSendToWebhookSuccess(t *testing.T) {
	var received WebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		decoder := json.NewDecoder(request.Body)
		if err := decoder.Decode(&received); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	payload := WebhookPayload{Message: Message{Text: "hello"}}

	if err := sendToWebhook(server.URL, payload); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if received.Message.Text != "hello" {
		t.Fatalf("expected payload to be received")
	}
}

func TestSendToWebhookNonSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	payload := WebhookPayload{Message: Message{Text: "hello"}}

	if err := sendToWebhook(server.URL, payload); err == nil {
		t.Fatalf("expected error for non-2xx response")
	}
}

func TestSendToWebhookInvalidURL(t *testing.T) {
	payload := WebhookPayload{Message: Message{Text: "hello"}}

	if err := sendToWebhook("http://invalid host", payload); err == nil {
		t.Fatalf("expected error for invalid URL")
	}
}
