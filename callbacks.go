package main

import (
	"log/slog"
	"strings"

	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackAction string

const (
	ActionCancel  CallbackAction = "cancel"
	ActionInfo    CallbackAction = "info"
	ActionRefresh CallbackAction = "refresh"
)

type CallbackScope string

const (
	ScopeTorrent CallbackScope = "torrent"
	ScopeDD      CallbackScope = "dd"
)

var handlers = map[CallbackScope]map[CallbackAction]func(query *tgbotapi.CallbackQuery, id string) error{
	ScopeTorrent: {
		ActionCancel:  handleTorrentCancel,
		ActionInfo:    handleTorrentInfo,
		ActionRefresh: handleRefreshTorrentInfo,
	},
	ScopeDD: {
		ActionCancel:  handleDirectDownloadCancel,
		ActionInfo:    handleDirectDownloadInfo,
		ActionRefresh: handleDirectDownloadRefresh,
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
		l.log.Error("Error handling data")
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
func handleTorrentCancel(query *tgbotapi.CallbackQuery, id string) error {
	message := query.Message
	err := qb.Delete([]string{id}, true)
	if err != nil {
		return err
	}
	if cancel, exists := cancelGoroutines[message.MessageID]; exists && cancel != nil {
		cancel()
	}
	_, err = bot.Send(
		tgbotapi.NewEditMessageText(
			message.Chat.ID,
			message.MessageID,
			"Torrent canceled successfully.",
		),
	)
	if err != nil {
		l.log.Error("Error sending message", slog.Any("error", err))
		return err
	}
	return nil
}

func handleTorrentInfo(query *tgbotapi.CallbackQuery, id string) error {
	chatID := query.Message.Chat.ID
	torrent, err := getTorrentInfo(id)
	if err != nil {
		return err
	}
	msg := messages.BuildTorrentInfo(chatID, torrent)
	_, err = bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func handleRefreshTorrentInfo(query *tgbotapi.CallbackQuery, id string) error {
	torrent, err := getTorrentInfo(id)
	if err != nil {
		return err
	}
	msg := messages.BuildTorrentInfo(query.Message.Chat.ID, torrent)
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
	if err != nil {
		return err
	}
	return nil

}
func handleDirectDownloadCancel(query *tgbotapi.CallbackQuery, id string) error {
	l.log.Debug("Handling direct download cancel", slog.String("id", id))
	err := gp.DeleteTask(id)
	if err != nil {
		return err
	}
	if cancel, exists := cancelGoroutines[query.Message.MessageID]; exists && cancel != nil {
		cancel()
	}
	_, err = bot.Send(
		tgbotapi.NewEditMessageText(
			query.Message.Chat.ID,
			query.Message.MessageID,
			"Direct download canceled successfully.",
		))
	return err
}

func handleDirectDownloadInfo(query *tgbotapi.CallbackQuery, id string) error {
	l.log.Debug("Handling direct download info", slog.String("id", id))
	task, err := gp.GetTask(id)
	if err != nil {
		return err
	}
	msg := messages.BuildDirectDownloadProgress(query.Message.Chat.ID, task)
	_, err = bot.Send(msg)
	return nil
}

func handleDirectDownloadRefresh(query *tgbotapi.CallbackQuery, id string) error {
	task, err := gp.GetTask(id)
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
