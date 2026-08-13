package main

import (
	"log/slog"
	"os"

	"github.com/gabrielramos02/telegram-bot-go/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		slog.Error("failed to load .env file", slog.String("error", err.Error()))
		return
	}
	cfg, err := app.LoadConfig(map[string]string{
		"BOT_TOKEN": os.Getenv("BOT_TOKEN"),
		"ENV":       os.Getenv("ENV"),
		"GP_URL":    os.Getenv("GP_URL"),
		"GP_TOKEN":  os.Getenv("GP_TOKEN"),
		"DB_URL":    os.Getenv("DB_URL"),
	})
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		return
	}
	err = app.Run(cfg)
	if err != nil {
		slog.Error("failed to run app", slog.String("error", err.Error()))
		return
	}
}
