# Telegram Download Bot

![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
[![Powered by gopeed-api-go](https://img.shields.io/badge/Powered%20by-gopeed--api--go-blue)](https://github.com/gabrielramos02/gopeed-api-go)

A Telegram bot that lets you control your download managers from anywhere. Manage **qBittorrent** torrents and **Gopeed** direct downloads through simple chat messages and inline buttons.

## Features

- Add torrents by sending a magnet link.
- Start direct downloads by sending an HTTP/HTTPS URL.
- List active torrents and direct downloads with formatted messages.
- View detailed progress cards for each download.
- Cancel, refresh, or get more info through inline keyboards.
- Live progress updates every 5 seconds while a download is active.
- Structured logging with `slog` and optional file rotation.

## Powered by [gopeed-api-go](https://github.com/gabrielramos02/gopeed-api-go)

This bot uses **`gopeed-api-go`**, a standalone Go client for the [Gopeed](https://gopeed.com/) download manager REST API.

`gopeed-api-go` was originally extracted from this bot and packaged as a reusable library so that other Go projects can interact with Gopeed without writing their own HTTP client from scratch.

If you only need a Go client for Gopeed, you can use it directly:

```go
package main

import (
    "context"
    "fmt"
    "time"

    gopeed "github.com/gabrielramos02/gopeed-api-go"
)

func main() {
    client, err := gopeed.NewClient("http://localhost:9999", "your-token")
    if err != nil {
        panic(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    taskID, err := client.CreateTaskFromURL(ctx, "https://example.com/file.iso", gopeed.GopeedOptions{
        Extra: &gopeed.GopeedExtraOptions{Connections: 32},
    })
    if err != nil {
        panic(err)
    }

    task, err := client.GetTask(ctx, taskID)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Task: %s | Progress: %d/%d\n", task.Name, task.Progress.Downloaded, task.Size)
}
```

## Tech Stack

- [Go](https://go.dev/) 1.26+
- [go-telegram-bot-api/v5](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [superturkey650/go-qbittorrent](https://github.com/superturkey650/go-qbittorrent)
- [gopeed-api-go](https://github.com/gabrielramos02/gopeed-api-go)
- `log/slog` for structured logging
- `gopkg.in/natefinch/lumberjack.v2` for log rotation

## Prerequisites

- A running qBittorrent instance with the Web UI enabled.
- A running [Gopeed](https://gopeed.com/) instance with API access enabled and a token configured.
- A Telegram bot token from [@BotFather](https://t.me/BotFather).
- Go 1.26 or later installed on your machine.

## Configuration

Create a `.env` file in the project root with the following variables:

| Variable      | Description                                  | Required |
|---------------|----------------------------------------------|----------|
| `BOT_TOKEN`   | Telegram bot token from BotFather            | Yes      |
| `QB_URL`      | qBittorrent Web UI URL                       | Yes      |
| `QB_USERNAME` | qBittorrent Web UI username                  | Yes      |
| `QB_PASSWORD` | qBittorrent Web UI password                  | Yes      |
| `GP_URL`      | Gopeed API base URL                          | Yes      |
| `GP_TOKEN`    | Gopeed API token                             | Yes      |
| `ENV`         | Environment name, e.g., `development`        | Yes      |
| `LOG_FILE`    | Path to rotating log file (optional)         | No       |

Example:

```env
BOT_TOKEN=your-telegram-bot-token
QB_URL=http://127.0.0.1:8081
QB_USERNAME=your-qbittorrent-username
QB_PASSWORD=your-qbittorrent-password
GP_URL=http://127.0.0.1:9999
GP_TOKEN=your-gopeed-token
ENV=development
LOG_FILE=bot.log
```

## Installation & Run

```bash
# Clone the repository
git clone https://github.com/gabrielramos02/telegram-bot-go.git
cd telegram-bot-go

# Download dependencies
go mod download

# Create your .env file
cp .env.example .env
# Edit .env with your credentials

# Run the bot
go run .
```

## Usage

Send commands or links to your bot:

- `/start` — Welcome message.
- `/get_torrents` — List active torrents.
- `/get_direct_downloads` — List active Gopeed direct downloads.
- Send a **magnet link** to add a new torrent.
- Send an **HTTP/HTTPS URL** to start a direct download.

When a download starts, the bot sends a progress card with an inline keyboard so you can:

- **Refresh** the info.
- **Cancel** or **Delete** a download.
- **Info** to open a detailed view.

## Development & Testing

```bash
# Run all tests
go test ./...

# Build the binary
go build -o telegram-bot-go .
```

## License

This project is licensed under the [MIT License](LICENSE).

## Author

Made by [Gabriel Ramos](https://github.com/gabrielramos02).
