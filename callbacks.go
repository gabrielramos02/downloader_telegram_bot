package main

import (
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	if hash, exist := strings.CutPrefix(query.Data, "cancel:"); exist {
		err := handleTorrentCancel(query.Message, hash)
		if err != nil {
			l.log.Error("Error canceling torrent", slog.Any("error", err))
		}
	}
}

func handleTorrentCancel(message *tgbotapi.Message, hash string) error {
	err := qb.Delete([]string{hash}, true)
	if err != nil {
		return err
	}
	cancelGoroutines[message.Chat.ID]()
	_, err = bot.Send(tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, "Torrent canceled successfully."))
	if err != nil {
		l.log.Error("Error sending message", slog.Any("error", err))
		return err
	}
	return nil
}
