package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	gopeed "github.com/gabrielramos02/gopeed-api-go"
	"github.com/gabrielramos02/telegram-bot-go/internal/database"
	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var cancelGoroutines = make(map[string]context.CancelFunc)

var urlHandlers = map[string]func(replyToID int, chatID int64, URL string) error{
	"magnet": handleHttpURL,
	"http":   handleHttpURL,
	"https":  handleHttpURL,
}

func handleUrl(message *tgbotapi.Message, urlString string) error {
	ctx := context.Background()
	_, err := db.GetUserByID(ctx, message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			"You are not registered. Please use /start to register.",
		)
		_, sendErr := bot.Send(msg)
		if sendErr != nil {
			return fmt.Errorf("error sending message: %w", sendErr)
		}
		return fmt.Errorf("error getting user from database: %w", err)
	}
	chatID := message.Chat.ID
	scheme, err := parseURL(urlString)
	if err != nil {
		return err
	}
	handler, ok := urlHandlers[scheme]
	if !ok {
		_, err = bot.Send(tgbotapi.NewMessage(chatID, "Unsupported URL scheme"))
		if err != nil {
			return fmt.Errorf("error sending message: %w", err)
		}
		return fmt.Errorf("no handler for URL scheme: %s", scheme)
	}
	err = handler(message.MessageID, chatID, urlString)
	if err != nil {
		return fmt.Errorf("error handling URL: %w", err)
	}
	return nil
}
func parseURL(urlString string) (scheme string, err error) {
	urlObject, err := url.Parse(urlString)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if urlObject.Scheme == "" {
		return "", fmt.Errorf("unsupported URL scheme")
	}
	return urlObject.Scheme, nil
}

func handleHttpURL(replyToID int, chatID int64, URL string) error {
	opts := gopeed.GopeedOptions{
		Path:  "/downloads" + "/" + fmt.Sprintf("%d", chatID),
		Extra: &gopeed.GopeedExtraOptions{Connections: 32},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	msg := tgbotapi.NewMessage(chatID, "⏳ Creating direct download task...")
	msgSended, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}
	ddId, err := gp.CreateTaskFromURL(ctx, URL, opts)
	if err != nil {
		return fmt.Errorf("failed to create direct download task: %w", err)
	}
	l.log.Debug("Direct download task created with ID:", slog.String("ddId", ddId))
	_, err = db.CreateFile(context.Background(), database.CreateFileParams{
		ID:        ddId,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to create file in database: %w", err)
	}
	err = db.LinkFileToUser(
		context.Background(),
		database.LinkFileToUserParams{FileID: ddId, UserID: chatID},
	)
	if err != nil {
		return fmt.Errorf("failed to link file to user in database: %w", err)
	}
	ddInfo, err := gp.GetTask(ctx, ddId)
	if err != nil {
		l.log.Error("Error getting direct download task info", slog.Any("error", err))
		return err
	}
	_, err = bot.Request(tgbotapi.NewDeleteMessage(chatID, msgSended.MessageID))
	if err != nil {
		return fmt.Errorf("error deleting message: %w", err)
	}
	msg = messages.BuildDirectDownloadProgress(chatID, ddInfo)
	msg.ReplyToMessageID = replyToID
	msgSended, err = bot.Send(msg)
	sendDirectDownloadInfo(chatID, ddInfo, msgSended)
	return err
}

func sendDirectDownloadInfo(chatID int64, ddInfo gopeed.GopeedTask, msgSended tgbotapi.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	mutex := sync.Mutex{}
	goroutineKey := fmt.Sprintf("%d-%s", chatID, ddInfo.ID)
	cancelGoroutines[goroutineKey] = cancel
	go func() {
		var err error
		defer func() {
			mutex.Lock()
			delete(cancelGoroutines, ddInfo.ID)
			mutex.Unlock()
			_, err = bot.Send(
				tgbotapi.NewEditMessageText(
					chatID,
					msgSended.MessageID,
					"Direct download canceled successfully.",
				))
			if err != nil {
				l.log.Error("Error sending message", slog.Any("error", err))
			}
			l.log.Debug(
				"End of goroutine for MessageID",
				slog.Int("messageid", msgSended.MessageID),
			)
		}()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for ddInfo.Status != gopeed.GopeedStatusDone && ddInfo.Status != gopeed.GopeedStatusError {
			select {
			case <-ctx.Done():
				l.log.Debug(
					"Cancelled goroutine for MessageID",
					slog.Int("messageid", msgSended.MessageID),
				)
				return
			case <-ticker.C:
				contextInfo, cancelInfo := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancelInfo()
				ddInfo, err = gp.GetTask(contextInfo, ddInfo.ID)
				if err != nil {
					l.log.Error("Error getting direct download task info", slog.Any("error", err))
					return
				}
				msg := messages.BuildDirectDownloadProgress(msgSended.Chat.ID, ddInfo)
				newMsg := tgbotapi.NewEditMessageText(msg.ChatID, msgSended.MessageID, msg.Text)
				newMsg.ParseMode = tgbotapi.ModeHTML
				if markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup); ok {
					newMsg.ReplyMarkup = &markup
				}
				if msgSended.Text != newMsg.Text {
					_, err = bot.Send(newMsg)
					if err != nil {
						l.log.Error("Error sending message", slog.Any("error", err))
					}
				}

			}
		}
		l.log.Debug(
			"Direct download task completed for chatID",
			slog.Int64("chatID", chatID),
			slog.String("ddName", ddInfo.Meta.Res.Name),
			slog.String("ddStatus", string(ddInfo.Status)),
		)
		if ddInfo.Status == gopeed.GopeedStatusError {
			l.log.Error(
				"Direct download task ended with error",
				slog.Int64("chatID", chatID))
		}
		msgText := fmt.Sprintf("✅ <b>Download Complete!</b> Your file: %s is ready.", ddInfo.Name)
		finalMsg := tgbotapi.NewMessage(chatID, msgText)
		finalMsg.ParseMode = tgbotapi.ModeHTML
		_, err = bot.Send(finalMsg)
		if err != nil {
			l.log.Error("Error sending final message", slog.Any("error", err))
		}
		_, err = bot.Request(tgbotapi.NewDeleteMessage(chatID, msgSended.ReplyToMessage.MessageID))
		if err != nil {
			l.log.Error("Error deleting message", slog.Any("error", err))
		}
		_, err = bot.Request(tgbotapi.NewDeleteMessage(chatID, msgSended.MessageID))
		if err != nil {
			l.log.Error("Error deleting message", slog.Any("error", err))
		}
	}()
}
