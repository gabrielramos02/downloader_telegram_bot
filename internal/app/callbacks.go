package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	gopeed "github.com/gabrielramos02/gopeed-api-go"
	"github.com/gabrielramos02/telegram-bot-go/internal/glances"
	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)
const RefreshList string = "refresh_list"

type CallbackAction string

const (
	ActionCancel      CallbackAction = "cancel"
	ActionInfo        CallbackAction = "info"
	ActionRefresh     CallbackAction = "refresh"
	ActionRefreshList CallbackAction = "refresh_list"
	ActionPause       CallbackAction = "pause"
	ActionContinue    CallbackAction = "continue"
	ActionDelete      CallbackAction = "delete"
	ActionClose       CallbackAction = "close"
)

type CallbackScope string

const (
	ScopeDD      CallbackScope = "dd"
	ScopeStorage CallbackScope = "st"
)

var handlers = map[CallbackScope]map[CallbackAction]func(query *tgbotapi.CallbackQuery, id string) error{
	ScopeDD: {
		ActionCancel:      handleDirectDownloadCancel,
		ActionInfo:        handleDirectDownloadInfo,
		ActionRefresh:     handleDirectDownloadRefresh,
		ActionPause:       handleDirectDownloadPause,
		ActionContinue:    handleDirectDownloadContinue,
		ActionDelete:      handleDirectDownloadDelete,
		ActionRefreshList: handleDirectDownloadRefreshList,
	},
	ScopeStorage: {
		ActionRefresh: handleStorageRefresh,
		ActionClose:   handleStorageClose,
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
	_, err = bot.Request(
		tgbotapi.NewCallbackWithAlert(query.ID, "Direct download canceled successfully."),
	)
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
	goroutineKey := fmt.Sprintf("%d-%s", query.Message.Chat.ID, id)
	if cancel, exists := cancelGoroutines[goroutineKey]; exists && cancel != nil {
		cancel()
	}
	_, err = bot.Request(
		tgbotapi.NewCallbackWithAlert(query.ID, "Direct download deleted successfully."),
	)
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
	if id == RefreshList {
		return handleDirectDownloadRefreshList(query, id)
	}
	task, err := gp.GetTask(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "task not found") {
			msg := tgbotapi.NewEditMessageText(
				query.Message.Chat.ID,
				query.Message.MessageID,
				"Direct download deleted successfully.",
			)
			_, err = bot.Request(msg)
			if err != nil {
				return err
			}
			return nil
		}
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

func handleStorageRefresh(query *tgbotapi.CallbackQuery, id string) error {
	_, err := bot.Request(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID))
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Loading storage info...")
	loadingMsg, err := bot.Send(msg)
	if err != nil {
		return err
	}

	gl := glances.NewClient()
	fs, err := gl.GetFS(context.Background())
	if err != nil {
		return err
	}
	_, err = bot.Request(tgbotapi.NewDeleteMessage(loadingMsg.Chat.ID, loadingMsg.MessageID))
	if err != nil {
		return err
	}
	msg = messages.BuildStorageInfo(query.Message.Chat.ID, fs)
	msg.ReplyToMessageID = query.Message.ReplyToMessage.MessageID
	_, err = bot.Send(msg)
	return err
}

func handleStorageClose(query *tgbotapi.CallbackQuery, id string) error {
	_, err := bot.Request(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID))
	_, err = bot.Request(
		tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.ReplyToMessage.MessageID),
	)
	return err
}

func handleDirectDownloadRefreshList(query *tgbotapi.CallbackQuery, id string) error {
	chatID := query.Message.Chat.ID
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
	var newMsg tgbotapi.EditMessageTextConfig
	if markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup); ok {
		newMsg = tgbotapi.NewEditMessageTextAndMarkup(
			query.Message.Chat.ID,
			query.Message.MessageID,
			msg.Text,
			markup,
		)
	}
	_, err = bot.Send(newMsg)
	return err

}
