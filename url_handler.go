package main

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

func handleUrl(chatID int64, url string) error {
	var err error
	var urlList []string
	if strings.HasPrefix(url, "magnet:") {
		urlList = append(urlList, url)
		err := qb.DownloadLinks(urlList, qbt.DownloadOptions{})
		if err != nil {
			return err
		}
	}
	_ ,err = bot.Send(tgbotapi.NewMessage(chatID, "URL added to download queue"))
	return err
}
