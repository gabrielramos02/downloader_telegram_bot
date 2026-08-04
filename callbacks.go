package main

import (
	"log/slog"
	"strings"

	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	if hash, exist := strings.CutPrefix(query.Data, "cancel:"); exist {
		err := handleTorrentCancel(query.Message, hash)
		if err != nil {
			l.log.Error("Error canceling torrent", slog.Any("error", err))
		}
	} else if hash, exist := strings.CutPrefix(query.Data, "info:"); exist {
		err := handleTorrentInfo(query.From.ID, hash)
		if err != nil {
			l.log.Error("Error getting torrent info", slog.Any("error", err))
		}
	} else if hash, exist := strings.CutPrefix(query.Data, "refresh:"); exist {
		err := handleRefreshTorrentInfo(query, hash)
		if err != nil {
			return
		}
	}
}

func handleTorrentCancel(message *tgbotapi.Message, hash string) error {
	err := qb.Delete([]string{hash}, true)
	if err != nil {
		return err
	}
	if cancel, exists := cancelGoroutines[message.Chat.ID]; exists && cancel != nil {
		cancel()
	}
	_, err = bot.Send(tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, "Torrent canceled successfully."))
	if err != nil {
		l.log.Error("Error sending message", slog.Any("error", err))
		return err
	}
	return nil
}

func handleTorrentInfo(chatID int64, hash string) error {
	torrent, err := getTorrentInfo(hash)
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

func handleRefreshTorrentInfo(query *tgbotapi.CallbackQuery, hash string) error {
	torrent, err := getTorrentInfo(hash)
	if err != nil {
		return err
	}
	msg := messages.BuildTorrentInfo(query.Message.Chat.ID, torrent)
	var newMsg tgbotapi.EditMessageTextConfig
	if replyMarkup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup); ok {
		newMsg = tgbotapi.NewEditMessageTextAndMarkup(query.Message.Chat.ID, query.Message.MessageID, msg.Text, replyMarkup)
		newMsg.ParseMode = tgbotapi.ModeHTML
	}
	_, err = bot.Send(newMsg)
	if err != nil {
		return err
	}
	return nil

}
