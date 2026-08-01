package main

import (
	"log"

	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

func handleCommand(chatID int64, command string) error {
	var err error
	switch command {
	case "/start":
		err = sendStart(chatID)
	case "/get_torrents":
		err = getTorrents(chatID)
	}

	return err
}

func sendStart(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Hello to my new bot")
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)

	return err
}

func getTorrents(chatID int64) error {
	torrentList, err := qb.Torrents(qbt.TorrentsOptions{})
	if err != nil {
		log.Printf("%v", err.Error())
	}
	for _, torrent := range torrentList {
		log.Printf("%v", torrent.Name)
	}
	msg := messages.BuildTorrentList(chatID, torrentList)
	_, err = bot.Send(msg)
	return err
}


