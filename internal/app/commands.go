package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	startCommand              = "start"
	getDirectDownloadsCommand = "get_direct_downloads"
)

func handleCommand(chatID int64, command string) error {
	action, ok := commandAction(command)
	if !ok {
		return nil
	}
	switch action {
	case startCommand:
		return sendStart(chatID)
	case getDirectDownloadsCommand:
		return getDirectDownloads(chatID)
	default:
		return nil

	}

}

func commandAction(command string) (string, bool) {
	switch command {
	case "/start":
		return startCommand, true
	case "/get_direct_downloads":
		return getDirectDownloadsCommand, true
	default:
		return "", false
	}
}

func sendStart(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Hello to my new bot")
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)
	if err != nil {
		return err
	}
	l.log.Debug("Start command executed", slog.Int64("chatID", chatID))
	return nil
}

func getDirectDownloads(chatID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tasks, err := gp.GetTasks(ctx)
	if err != nil {
		return err
	}
	var msg tgbotapi.MessageConfig
	if len(tasks) == 0 {
		msg = tgbotapi.NewMessage(chatID, "No direct download tasks found.")
	} else {
		msg = messages.BuildDirectDownloads(chatID, tasks)
	}
	_, err = bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
