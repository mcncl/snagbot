# SnagBot

A Slack bot that automatically converts dollar amounts in messages to a fun comparison.

## Overview

SnagBot monitors messages in Slack channels where it is added. When users mention dollar values (e.g., "This would cost $35"), the bot automatically replies in a thread with a fun conversion message. By default, the conversion is based on the cost of a Bunnings snag ($3.50 each), resulting in messages like "That's 10 Bunnings snags!"

The bot supports custom configuration via Slack commands, allowing users to set custom items and prices on a per-channel basis.

## Features

- Automatically detects and processes dollar amounts in Slack messages
- Converts dollar amounts to fun equivalents (e.g., "That's 10 Bunnings snags!")
- Supports custom items and prices per channel
- Handles multiple dollar amounts in a single message
- Provides slash commands for configuration management

## Available Commands

- `/snagbot` or `/snagbot status` - Show current configuration
- `/snagbot item "coffee" price 5.00` - Set custom item and price
- `/snagbot reset` - Reset to default configuration
- `/snagbot help` - Show help information

## Setup Instructions

### Prerequisites

- Go 1.21 or higher
- A Slack workspace with permission to add custom apps
- Slack API credentials (Bot Token and Signing Secret)

### Environment Variables

Create a `.env` file in the project root with the following variables (based on `.env.example`):

```
SLACK_BOT_TOKEN=xoxb-your-bot-token-here
SLACK_SIGNING_SECRET=your-signing-secret-here
PORT=8080
DEFAULT_ITEM_NAME="Bunnings snags"
DEFAULT_ITEM_PRICE=3.50
```

### Build and Run

1. Clone the repository:
   ```bash
   git clone https://github.com/mcncl/snagbot.git
   cd snagbot
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   make build
   ```

4. Run the bot:
   ```bash
   make run
   ```

Alternatively, you can use Docker:

```bash
docker build -t snagbot .
docker run -p 8080:8080 --env-file .env snagbot
```

### Setting up in Slack

1. Create a new Slack App in the [Slack API Console](https://api.slack.com/apps)
2. Under "OAuth & Permissions", add the following scopes:
   - `channels:history`
   - `chat:write`
   - `commands`
3. Create a slash command `/snagbot` with the Request URL pointing to your server: `https://your-server.com/api/commands`
4. Under "Event Subscriptions", enable events and add the following:
   - Subscribe to bot events: `message.channels`
   - Set the Request URL to: `https://your-server.com/api/events`
5. Install the app to your workspace
6. Add the bot to desired channels

## Development

### Project Structure

```
snagbot/
├── cmd/
│   └── server/            # Application entry point
├── internal/
│   ├── api/               # HTTP API handlers
│   ├── app/               # Application setup
│   ├── calculator/        # Dollar value extraction and conversion
│   ├── command/           # Slash command handling
│   ├── config/            # Configuration management
│   ├── errors/            # Error handling utilities
│   ├── logging/           # Logging utilities
│   ├── service/           # Business logic services
│   └── slack/             # Slack API integration
├── pkg/
│   └── models/            # Shared data models
└── test/
    └── integration/       # Integration tests
```

### Available Make Commands

- `make build` - Build the application
- `make run` - Run the application
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make lint` - Run linter

## Testing

Run all tests with:

```bash
make test
```

Or run specific tests:

```bash
go test ./internal/calculator/...
go test ./test/integration/...
```

## Deployment

### Heroku

```bash
heroku create
heroku config:set SLACK_BOT_TOKEN=xoxb-your-token
heroku config:set SLACK_SIGNING_SECRET=your-secret
git push heroku main
```

### Docker / Kubernetes

A Dockerfile is provided for containerized deployments. For Kubernetes, configure your deployment to include the necessary environment variables.

## Architecture

For detailed information about the application architecture, error handling strategy, and future improvement plans, see [ARCHITECTURE.md](ARCHITECTURE.md).

## License

MIT
