package main

import (
	"log/slog"

	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

const (
	startCommand       = "start"
	getTorrentsCommand = "get_torrents"
)

func handleCommand(chatID int64, command string) error {
	action, ok := commandAction(command)
	if !ok {
		return nil
	}
	switch action {
	case startCommand:
		return sendStart(chatID)
	case getTorrentsCommand:
		return getTorrents(chatID)
	default:
		return nil

	}

}

func commandAction(command string) (string, bool) {
	switch command {
	case "/start":
		return startCommand, true
	case "/get_torrents":
		return getTorrentsCommand, true
	default:
		return "", false
	}
}

func sendStart(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Hello to my new bot")
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)
	l.log.Debug("Start command executed", slog.Int64("chatID", chatID))
	return err
}

func getTorrents(chatID int64) error {
	torrentList, err := qb.Torrents(qbt.TorrentsOptions{})
	if err != nil {
		l.log.Error("Error getting torrents", slog.Any("error", err))

	}
	var msg tgbotapi.MessageConfig
	if len(torrentList) == 0 {
		msg = tgbotapi.NewMessage(chatID, "No torrents found.")
	} else {
		msg = messages.BuildTorrentList(chatID, torrentList)

	}
	_, err = bot.Send(msg)
	l.log.Debug(
		"Get torrents command executed",
		slog.Int64("chatID", chatID),
		slog.Int("torrentCount", len(torrentList)),
	)
	return err
}
