package app

import (
	"context"
	"log/slog"
	"time"

	gopeed "github.com/gabrielramos02/gopeed-api-go"
	"github.com/gabrielramos02/telegram-bot-go/internal/database"
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
	_, err := db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        chatID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	_, err = bot.Send(msg)
	if err != nil {
		return err
	}
	l.log.Debug("Start command executed", slog.Int64("chatID", chatID))
	return nil
}

func getDirectDownloads(chatID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	user, err := db.GetUserByID(ctx, chatID)
	if err != nil {
		return err
	}
	userFiles, err := db.GetUserFiles(ctx, user.ID)
	if err != nil {
		return err
	}
	var tasks []gopeed.GopeedTask
	if len(userFiles) > 0 {
		var fileIDs []string
		for _, userFile := range userFiles {
			fileIDs = append(fileIDs, userFile.FileID)
		}
		tasks, err = gp.GetTasks(ctx, fileIDs, "")
		if err != nil {
			return err
		}
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
