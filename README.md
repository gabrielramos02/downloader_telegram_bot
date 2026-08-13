# Telegram Download Bot

![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
[![Powered by gopeed-api-go](https://img.shields.io/badge/Powered%20by-gopeed--api--go-blue)](https://github.com/gabrielramos02/gopeed-api-go)

A Telegram bot that lets you control your downloads from anywhere. Manage **Gopeed** direct downloads through simple chat messages and inline buttons.

## Motivation

I run my downloads on a home server but I want to control them from my phone without opening the Web UI of each download manager. This bot centralizes Gopeed into a single Telegram chat, with live progress updates and one-tap actions. It was also a good excuse to extract the Gopeed client into its own reusable Go library: [gopeed-api-go](https://github.com/gabrielramos02/gopeed-api-go).

## Features

- Start direct downloads by sending an HTTP/HTTPS URL or a magnet link.
- List active direct downloads with formatted messages.
- View detailed progress cards for each download.
- Pause, continue, cancel, refresh, or get more info through inline keyboards.
- Live progress updates every 5 seconds while a download is active.
- Structured logging with `slog` and optional file rotation.

## Quick Start

```bash
# Clone the repository
git clone https://github.com/gabrielramos02/telegram-bot-go.git
cd telegram-bot-go

# Download dependencies
go mod download

# Create and edit your .env file
cp .env .env.local
# Edit .env.local with your credentials

# Run the bot
go run ./cmd/bot
```

> Note: if you don't have an `.env` file yet, create one from the example in the [Configuration](#configuration) section.

## Powered by [gopeed-api-go](https://github.com/gabrielramos02/gopeed-api-go)

This bot uses **`gopeed-api-go`**, a standalone Go client for the [Gopeed](https://gopeed.com/) download manager REST API.

`gopeed-api-go` was originally extracted from this bot and packaged as a reusable library so that other Go projects can interact with Gopeed without writing their own HTTP client from scratch.

If you only need a Go client for Gopeed, you can use it directly:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    gopeed "github.com/gabrielramos02/gopeed-api-go"
)

func main() {
    client, err := gopeed.NewClient(
        "http://localhost:9999",
        gopeed.WithAPIToken("your-token"),
        gopeed.WithTimeout(30*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    taskID, err := client.CreateTaskFromURL(ctx, "https://example.com/file.iso", gopeed.GopeedOptions{
        Extra: &gopeed.GopeedExtraOptions{Connections: 32},
    })
    if err != nil {
        log.Fatal(err)
    }

    task, err := client.GetTask(ctx, taskID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Task: %s | Progress: %d/%d\n", task.Name, task.Progress.Downloaded, task.Size)
}
```

## Tech Stack

- [Go](https://go.dev/) 1.26+
- [go-telegram-bot-api/v5](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [gopeed-api-go](https://github.com/gabrielramos02/gopeed-api-go)
- `log/slog` for structured logging
- `gopkg.in/natefinch/lumberjack.v2` for log rotation

## Prerequisites

- A running [Gopeed](https://gopeed.com/) instance with API access enabled and a token configured.
- A Telegram bot token from [@BotFather](https://t.me/BotFather).
- Go 1.26 or later installed on your machine.

## Configuration

Create a `.env` file in the project root with the following variables:

| Variable      | Description                                  | Required |
|---------------|----------------------------------------------|----------|
| `BOT_TOKEN`   | Telegram bot token from BotFather            | Yes      |
| `ENV`         | Environment name, e.g., `development`        | Yes      |
| `GP_URL`      | Gopeed API base URL                          | Yes      |
| `GP_TOKEN`    | Gopeed API token                             | Yes      |
| `LOG_FILE`    | Path to rotating log file (optional)         | No       |

Example:

```env
BOT_TOKEN=your-telegram-bot-token
ENV=development
GP_URL=http://127.0.0.1:9999
GP_TOKEN=your-gopeed-token
LOG_FILE=bot.log
```

## Usage

Send commands or links to your bot:

- `/start` — Welcome message.
- `/get_direct_downloads` — List active Gopeed direct downloads.
- Send an **HTTP/HTTPS URL** or a **magnet link** to start a direct download.

When a download starts, the bot sends a progress card with an inline keyboard. The available actions depend on the download status:

- **Running:** pause or cancel the download.
- **Paused:** continue or cancel the download.
- **Done:** delete the finished download.
- **Info** — open a detailed view for a download from the list.

### Example messages

```text
/get_direct_downloads
https://example.com/file.iso
magnet:?xt=urn:btih:...
```

## Development & Testing

The project includes a `Makefile` with common tasks:

```bash
# Run all checks (format, staticcheck, gosec, tests with coverage)
make check

# Only run tests
make test

# Only run tests with coverage
make test-cover

# Update the gopeed-api-go dependency and tidy modules
make update-deps

# Update deps and run all checks (before committing)
make pre-commit

# Build the binary
go build -o telegram-bot ./cmd/bot
```

## Contributing

Contributions are welcome. If you find a bug or want to add a feature:

1. Fork the repository.
2. Create a new branch for your change.
3. Make your changes and add tests when possible.
4. Run `make check` to make sure everything passes.
5. Open a pull request with a clear description.

## License

This project is licensed under the [MIT License](LICENSE).

## Author

Made by [Gabriel Ramos](https://github.com/gabrielramos02).
