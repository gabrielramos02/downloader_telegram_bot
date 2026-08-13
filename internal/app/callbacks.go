package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackAction string

const (
	ActionCancel   CallbackAction = "cancel"
	ActionInfo     CallbackAction = "info"
	ActionRefresh  CallbackAction = "refresh"
	ActionPause    CallbackAction = "pause"
	ActionContinue CallbackAction = "continue"
	ActionDelete   CallbackAction = "delete"
)

type CallbackScope string

const (
	ScopeDD CallbackScope = "dd"
)

var handlers = map[CallbackScope]map[CallbackAction]func(query *tgbotapi.CallbackQuery, id string) error{
	ScopeDD: {
		ActionCancel:   handleDirectDownloadCancel,
		ActionInfo:     handleDirectDownloadInfo,
		ActionRefresh:  handleDirectDownloadRefresh,
		ActionPause:    handleDirectDownloadPause,
		ActionContinue: handleDirectDownloadContinue,
		ActionDelete:   handleDirectDownloadDelete,
	},
}

func handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	scope, action, id, ok := parseCallbackData(query.Data)
	if !ok {
		return
	}
	handler, exist := handlers[scope][action]
	l.log.Debug(
		"Handling callback query",
		slog.String("scope", string(scope)),
		slog.String("action", string(action)),
		slog.String("id", id),
		slog.Bool("handler_exists", exist),
	)
	if !exist {
		return
	}
	err := handler(query, id)
	if err != nil {
		l.log.Error("Error handling data", slog.Any("error", err), slog.String("data", query.Data))
	}
}

func parseCallbackData(
	data string,
) (scope CallbackScope, action CallbackAction, id string, ok bool) {
	parts := strings.SplitN(data, ":", 3)
	if len(parts) != 3 {
		return
	}
	return CallbackScope(parts[0]), CallbackAction(parts[1]), parts[2], true
}

func handleDirectDownloadCancel(query *tgbotapi.CallbackQuery, id string) error {
	l.log.Debug("Handling direct download cancel", slog.String("id", id))
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer httpCancel()
	err := gp.DeleteTask(httpCtx, id)
	if err != nil {
		return err
	}
	goroutineKey := fmt.Sprintf("%d-%s", query.Message.Chat.ID, id)
	if cancel, exists := cancelGoroutines[goroutineKey]; exists && cancel != nil {
		cancel()
	}
	err = db.DeleteFile(context.Background(), id)
	if err != nil {
		return err
	}
	// err = db.UnlinkFileFromUser(context.Background(), database.UnlinkFileFromUserParams{
	// 	FileID: id,
	// 	UserID: query.Message.Chat.ID,
	// })
	// if err != nil {
	// 	return err
	// }
	_, err = bot.Request(
		tgbotapi.NewCallbackWithAlert(query.ID, "Direct download canceled successfully."),
	)
	if err != nil {
		return err
	}
	_, err = bot.Send(
		tgbotapi.NewEditMessageText(
			query.Message.Chat.ID,
			query.Message.MessageID,
			"Direct download canceled successfully.",
		))
	return err
}

func handleDirectDownloadDelete(query *tgbotapi.CallbackQuery, id string) error {
	l.log.Debug("Handling direct download delete", slog.String("id", id))
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer httpCancel()
	err := gp.DeleteTask(httpCtx, id)
	if err != nil {
		return err
	}
	err = db.DeleteFile(context.Background(), id)
	if err != nil {
		return err
	}
	// err = db.UnlinkFileFromUser(context.Background(), database.UnlinkFileFromUserParams{
	// 	FileID: id,
	// 	UserID: query.Message.Chat.ID,
	// })
	// if err != nil {
	// 	return err
	// }
	goroutineKey := fmt.Sprintf("%d-%s", query.Message.Chat.ID, id)
	if cancel, exists := cancelGoroutines[goroutineKey]; exists && cancel != nil {
		cancel()
	}
	_, err = bot.Request(
		tgbotapi.NewCallbackWithAlert(query.ID, "Direct download deleted successfully."),
	)
	if err != nil {
		return err
	}
	_, err = bot.Send(
		tgbotapi.NewEditMessageText(
			query.Message.Chat.ID,
			query.Message.MessageID,
			"Direct download deleted successfully.",
		))
	return err
}

func handleDirectDownloadInfo(query *tgbotapi.CallbackQuery, id string) error {
	l.log.Debug("Handling direct download info", slog.String("id", id))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	task, err := gp.GetTask(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "task not found") {
			msg := tgbotapi.NewCallbackWithAlert(query.ID, "Direct download task not found.")
			_, err = bot.Request(msg)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	msg := messages.BuildDirectDownloadProgress(query.Message.Chat.ID, task)
	_, err = bot.Send(msg)
	return err
}

func handleDirectDownloadRefresh(query *tgbotapi.CallbackQuery, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	task, err := gp.GetTask(ctx, id)
	if err != nil {
		return err
	}
	msg := messages.BuildDirectDownloadProgress(query.Message.Chat.ID, task)
	var newMsg tgbotapi.EditMessageTextConfig

	if replyMarkup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup); ok {
		newMsg = tgbotapi.NewEditMessageTextAndMarkup(
			query.Message.Chat.ID,
			query.Message.MessageID,
			msg.Text,
			replyMarkup,
		)
		newMsg.ParseMode = tgbotapi.ModeHTML
	}
	_, err = bot.Send(newMsg)
	return err
}

func handleDirectDownloadPause(query *tgbotapi.CallbackQuery, id string) error {
	ctxReq, cancelReq := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelReq()
	err := gp.PauseTask(ctxReq, id)
	if err != nil {
		return err
	}
	goroutineKey := fmt.Sprintf("%d-%s", query.Message.Chat.ID, id)
	if cancel, exists := cancelGoroutines[goroutineKey]; exists && cancel != nil {
		cancel()
	}
	task, err := gp.GetTask(ctxReq, id)
	if err != nil {
		return err
	}
	msg := messages.BuildDirectDownloadProgress(query.Message.Chat.ID, task)
	var newMsg tgbotapi.EditMessageTextConfig
	if replyMarkup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup); ok {
		newMsg = tgbotapi.NewEditMessageTextAndMarkup(
			query.Message.Chat.ID,
			query.Message.MessageID,
			msg.Text,
			replyMarkup,
		)
		newMsg.ParseMode = tgbotapi.ModeHTML
	}
	_, err = bot.Send(newMsg)
	return err
}

func handleDirectDownloadContinue(query *tgbotapi.CallbackQuery, id string) error {
	ctxReq, cancelReq := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelReq()
	err := gp.ContinueTask(ctxReq, id)
	if err != nil {
		return err
	}
	task, err := gp.GetTask(ctxReq, id)
	if err != nil {
		return err
	}
	msg := messages.BuildDirectDownloadProgress(query.Message.Chat.ID, task)
	var newMsg tgbotapi.EditMessageTextConfig
	if replyMarkup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup); ok {
		newMsg = tgbotapi.NewEditMessageTextAndMarkup(
			query.Message.Chat.ID,
			query.Message.MessageID,
			msg.Text,
			replyMarkup,
		)
		newMsg.ParseMode = tgbotapi.ModeHTML
	}
	msgSended, err := bot.Send(newMsg)
	sendDirectDownloadInfo(query.Message.Chat.ID, task, msgSended)
	return err
}
