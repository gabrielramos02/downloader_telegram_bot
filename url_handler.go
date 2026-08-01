package main

import (
	"fmt"
	"net/url"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

func handleUrl(chatID int64, urlString string) (string, error) {
	hash := ""
	paused := true
	var urlList []string
	urlObject, err := url.Parse(urlString)
	if err != nil {
		return hash, fmt.Errorf("invalid URL: %v", err)
	}
	switch urlObject.Scheme {
	case "magnet":
		urlList = append(urlList, urlString)
		err := qb.DownloadLinks(urlList, qbt.DownloadOptions{Paused: &paused})
		if err != nil {
			return hash, err
		}
		hash, err = extractHashFromMagnet(urlString)
		if err != nil {
			return hash, err
		}
		_, err = bot.Send(tgbotapi.NewMessage(chatID, "URL added to download queue"))
		return hash, err

	default:
		return hash, fmt.Errorf("unsupported URL scheme: %s", urlObject.Scheme)
	}

}

func extractHashFromMagnet(magnetURL string) (string, error) {
	u, err := url.Parse(magnetURL)
	if err != nil {
		return "", err
	}

	xtParams := u.Query()["xt"]
	for _, xt := range xtParams {
		if hash, found := strings.CutPrefix(xt, "urn:btih:"); found {
			hash = strings.Split(hash, "&")[0]
			return strings.ToLower(hash), nil

		}
	}

	return "", fmt.Errorf("no BTIH hash found in magnet URL")
}
