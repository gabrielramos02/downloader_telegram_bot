package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gopeed "github.com/gabrielramos02/gopeed-api-go"
	"github.com/gabrielramos02/telegram-bot-go/internal/database"
	"github.com/gabrielramos02/telegram-bot-go/internal/glances"
	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var commandHandlers = map[string]func(message *tgbotapi.Message) error{
	"/start":                sendStart,
	"/get_direct_downloads": getDirectDownloads,
	"/get_storage_info":     getStorageInfo,
}

func handleCommand(message *tgbotapi.Message, command string) error {
	if handler, exists := commandHandlers[command]; exists {
		return handler(message)
	}
	return fmt.Errorf("no handler for command: %s", command)
}

func sendStart(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	msg := tgbotapi.NewMessage(chatID, "Hello to my new bot")
	msg.ParseMode = tgbotapi.ModeHTML
	err := db.CreateUser(context.Background(), database.CreateUserParams{
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

func getDirectDownloads(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
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

func getStorageInfo(message *tgbotapi.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	chatID := message.Chat.ID
	gl := glances.NewClient(glances.WithTimeout(10 * time.Second))
	fs, err := gl.GetFS(ctx)
	if err != nil {
		return err
	}
	msg := messages.BuildStorageInfo(chatID, fs)
	msg.ReplyToMessageID = message.MessageID
	_, err = bot.Send(msg)
	return err
}
