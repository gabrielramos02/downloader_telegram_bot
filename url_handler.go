package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

var cancelGoroutines = make(map[int64]context.CancelFunc)

func handleUrl(chatID int64, urlString string) error {
	var urlList []string
	urlObject, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	urlList = append(urlList, urlString)
	switch urlObject.Scheme {
	case "magnet":
		err = handleMagnetURL(chatID, urlList)
		if err != nil {
			return fmt.Errorf("failed to handle magnet URL: %v", err)
		}
		return err

	default:
		return fmt.Errorf("unsupported URL scheme: %s", urlObject.Scheme)
	}

}

func handleMagnetURL(chatID int64, urlList []string) error {
	var err error
	hash, err := addTorrent(urlList)
	if err != nil {
		return fmt.Errorf("failed to add torrent: %v", err)
	}
	torrent, err := getTorrentInfo(hash)
	if err != nil {
		return fmt.Errorf("failed to get torrent info: %v", err)
	}
	msg := messages.BuildTorrentInfo(chatID, torrent)

	msgSended, err := bot.Send(msg)
	sendTorrentInfo(chatID, hash, msgSended)
	return err
}

func sendTorrentInfo(chatID int64, hash string, msgSended tgbotapi.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	cancelGoroutines[chatID] = cancel
	go func() {
		var torrent qbt.TorrentInfo
		var err error
		defer func() {
			delete(cancelGoroutines, chatID)
			log.Printf("Goroutine for chatID %d has finished.", chatID)
		}()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for torrent.State != "stalledUP" && torrent.State != "stalledDL" && torrent.State != "error" && torrent.Progress < 1.0 {
			select {
			case <-ctx.Done():
				log.Printf("Goroutine for chatID %d has been canceled.", chatID)
				return
			case <-ticker.C:
				torrent, err = getTorrentInfo(hash)
				if err != nil {
					log.Printf("failed to get torrent info: %v", err)
					continue
				}
				msg := messages.BuildTorrentInfo(chatID, torrent)
				newMsg := tgbotapi.NewEditMessageText(chatID, msgSended.MessageID, msg.Text)
				newMsg.ParseMode = tgbotapi.ModeHTML
				if markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup); ok {
					newMsg.ReplyMarkup = &markup
				}

				_, err = bot.Send(newMsg)
			}
		}
		msgText := fmt.Sprintf("✅ <b>Download Complete!</b> Your file: %s is ready.", torrent.Name)
		finalMsg := tgbotapi.NewMessage(chatID, msgText)
		finalMsg.ParseMode = tgbotapi.ModeHTML
		bot.Send(finalMsg)
		_, err = bot.Send(tgbotapi.NewDeleteMessage(chatID, msgSended.MessageID))
	}()
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

func addTorrent(urlList []string) (string, error) {
	var hash string
	paused := true
	err := qb.DownloadLinks(urlList, qbt.DownloadOptions{Paused: &paused})
	if err != nil {
		return hash, err
	}
	hash, err = extractHashFromMagnet(urlList[0])
	if err != nil {
		return hash, err
	}
	return hash, nil

}

func getTorrentInfo(hash string) (qbt.TorrentInfo, error) {
	torrentInfoList, err := qb.Torrents(qbt.TorrentsOptions{Hashes: []string{hash}})
	if err != nil {
		return torrentInfoList[0], err
	}
	return torrentInfoList[0], nil

}
