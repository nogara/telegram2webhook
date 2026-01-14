# Telegram to Webhook Bridge

A lightweight Golang application that bridges Telegram bot messages to n8n webhooks. This service listens for incoming Telegram messages and forwards them to a configurable webhook endpoint.

## Features

- Connects to Telegram Bot API using long polling
- Forwards messages to n8n webhook in real-time
- Comprehensive error handling and logging
- Configurable via environment variables
- Docker support for easy deployment
- Minimal resource footprint

## Prerequisites

- Go 1.21 or higher (for local development)
- Docker (for containerized deployment)
- Telegram Bot Token (obtain from [@BotFather](https://t.me/botfather))
- n8n webhook URL

## Configuration

The application is configured using environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `TELEGRAM_WEBHOOK_PAIRS` | JSON array of `{ "token": "...", "webhook": "..." }` objects | No |
| `TELEGRAM_BOT_TOKEN` | Your Telegram bot token from BotFather | Yes (if no pairs) |
| `N8N_WEBHOOK_URL` | The n8n webhook endpoint URL | Yes (if no pairs) |
| `DEBUG` | Enable debug logging (true/false) | No (default: false) |

When `TELEGRAM_WEBHOOK_PAIRS` is set, it takes precedence and lets you run multiple bots with different webhooks.

Example:

```bash
export TELEGRAM_WEBHOOK_PAIRS='[{"token":"bot_token_1","webhook":"https://your-n8n.com/webhook/one"},{"token":"bot_token_2","webhook":"https://your-n8n.com/webhook/two"}]'
```

## Local Development

### 1. Clone or navigate to the project directory

```bash
cd telegram2webhook
```

### 2. Create a `.env` file

```bash
cp .env.example .env
```

Edit `.env` and add your actual tokens and URLs.

### 3. Install dependencies

```bash
go mod download
```

### 4. Run the application

```bash
# Source environment variables
export $(cat .env | xargs)

# Run the app
go run main.go
```

## Docker Deployment

### Build the Docker image

```bash
docker build -t telegram2webhook .
```

### Run with Docker

```bash
docker run -d \
  --name telegram2webhook \
  -e TELEGRAM_WEBHOOK_PAIRS='[{"token":"bot_token_1","webhook":"https://your-n8n.com/webhook/one"},{"token":"bot_token_2","webhook":"https://your-n8n.com/webhook/two"}]' \
  -e DEBUG=false \
  --restart unless-stopped \
  telegram2webhook
```

### Run with Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'

services:
  telegram2webhook:
    build: .
    container_name: telegram2webhook
    environment:
      - TELEGRAM_WEBHOOK_PAIRS=${TELEGRAM_WEBHOOK_PAIRS}
      - DEBUG=${DEBUG:-false}
    restart: unless-stopped
```

Then run:

```bash
docker-compose up -d
```

## Webhook Payload Structure

The application sends the following JSON payload to your webhook:

```json
{
  "message": {
    "text": "Hello, world!",
    "chat": {
      "id": 123456789
    },
    "from": {
      "username": "johndoe",
      "first_name": "John"
    }
  }
}
```

## Setting Up Your Telegram Bot

1. Open Telegram and search for [@BotFather](https://t.me/botfather)
2. Send `/newbot` and follow the instructions
3. Copy the bot token provided
4. Add the token to your `.env` file or Docker environment

## Setting Up n8n Webhook

1. Create a new workflow in n8n
2. Add a "Webhook" trigger node
3. Set the HTTP Method to POST
4. Copy the webhook URL
5. Add the URL to your `.env` file or Docker environment

## Logs and Monitoring

The application logs all incoming messages and webhook deliveries:

```
2026/01/13 10:30:00 Starting Telegram to Webhook bridge...
2026/01/13 10:30:01 Authorized on account YourBotName
2026/01/13 10:30:01 Webhook URL: https://your-n8n.com/webhook/xxx
2026/01/13 10:30:01 Listening for messages...
2026/01/13 10:30:15 Received message from johndoe: Hello!
2026/01/13 10:30:15 Successfully sent message to webhook. Status: 200
2026/01/13 10:30:15 Message forwarded successfully
```

## Error Handling

The application includes comprehensive error handling:

- Missing environment variables are caught at startup
- Failed webhook deliveries are logged but don't stop the service
- Network errors are logged with detailed messages
- Non-2xx HTTP responses from webhooks are logged as errors

## Troubleshooting

### Bot not receiving messages

- Ensure your bot token is correct
- Make sure you've started a conversation with your bot
- Check if the bot has the necessary permissions

### Webhook not receiving data

- Verify the webhook URL is accessible
- Check n8n workflow is active
- Review application logs for error messages
- Test webhook manually with curl

### Docker container issues

```bash
# View logs
docker logs telegram2webhook

# Restart container
docker restart telegram2webhook
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
