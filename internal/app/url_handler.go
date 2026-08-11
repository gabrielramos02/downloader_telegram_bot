package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	gopeed "github.com/gabrielramos02/gopeed-api-go"
	"github.com/gabrielramos02/telegram-bot-go/internal/messages"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

var cancelGoroutines = make(map[int]context.CancelFunc)

func handleUrl(chatID int64, urlString string) error {
	URL, scheme, err := classifyURL(urlString)
	if err != nil {
		return err
	}
	switch scheme {
	case "magnet":
		return handleMagnetURL(chatID, URL)
	case "http":
		return handleHttpURL(chatID, URL)
	default:
		return nil
	}
}
func classifyURL(urlString string) (string, string, error) {
	urlObject, err := url.Parse(urlString)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %v", err)
	}
	switch urlObject.Scheme {
	case "magnet":
		return urlString, "magnet", nil
	case "http", "https":
		return urlString, "http", nil
	default:
		return "", "", fmt.Errorf("unsupported URL scheme: %s", urlObject.Scheme)

	}
}

func handleMagnetURL(chatID int64, URL string) error {
	var err error
	hash, err := addTorrent(URL)
	if err != nil {
		return fmt.Errorf("failed to add torrent: %v", err)
	}
	torrent, err := getTorrentInfo(hash)
	if err != nil {
		return fmt.Errorf("failed to get torrent info: %v", err)
	}
	msg := messages.BuildTorrentProgress(chatID, torrent)

	msgSended, err := bot.Send(msg)
	sendTorrentInfo(chatID, hash, msgSended)
	return err
}

func sendTorrentInfo(chatID int64, hash string, msgSended tgbotapi.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	mutex := sync.Mutex{}
	cancelGoroutines[msgSended.MessageID] = cancel
	go func() {
		var torrent qbt.TorrentInfo
		var err error
		defer func() {
			mutex.Lock()
			delete(cancelGoroutines, msgSended.MessageID)
			mutex.Unlock()
			l.log.Debug(
				"End of goroutine for MessageID",
				slog.Int("messageid", msgSended.MessageID),
			)
		}()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for isTorrentInProgress(torrent) {
			select {
			case <-ctx.Done():
				l.log.Debug(
					"End of goroutine for MessageID",
					slog.Int("messageid", msgSended.MessageID),
				)
				return
			case <-ticker.C:
				torrent, err = getTorrentInfo(hash)
				if err != nil {
					l.log.Error("Error getting torrent info", slog.Any("error", err))
					continue
				}
				msg := messages.BuildTorrentProgress(chatID, torrent)
				newMsg := tgbotapi.NewEditMessageText(chatID, msgSended.MessageID, msg.Text)
				newMsg.ParseMode = tgbotapi.ModeHTML
				if markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup); ok {
					newMsg.ReplyMarkup = &markup
				}

				_, err = bot.Send(newMsg)
				if err != nil {
					l.log.Error("Error sending message", slog.Any("error", err))
				}

			}
		}
		l.log.Debug(
			"Torrent download completed or stopped for chatID",
			slog.Int64("chatID", chatID),
			slog.String("torrentName", torrent.Name),
			slog.String("torrentState", torrent.State),
			slog.Float64("torrentProgress", torrent.Progress),
		)
		msgText := fmt.Sprintf("✅ <b>Download Complete!</b> Your file: %s is ready.", torrent.Name)
		finalMsg := tgbotapi.NewMessage(chatID, msgText)
		finalMsg.ParseMode = tgbotapi.ModeHTML
		_, err = bot.Send(finalMsg)
		if err != nil {
			l.log.Error("Error sending final message", slog.Any("error", err))
		}
		_, err = bot.Send(tgbotapi.NewDeleteMessage(chatID, msgSended.MessageID))
		if err != nil {
			l.log.Error("Error deleting message", slog.Any("error", err))
		}
	}()
}
func isTorrentInProgress(torrent qbt.TorrentInfo) bool {
	return torrent.State != "stalledUP" && torrent.State != "error" && torrent.Progress < 1.0
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

func addTorrent(url string) (string, error) {
	var hash string
	paused := true
	URLs := []string{url}
	err := qb.DownloadLinks(URLs, qbt.DownloadOptions{Paused: &paused})
	if err != nil {
		return hash, err
	}
	hash, err = extractHashFromMagnet(URLs[0])
	if err != nil {
		return hash, err
	}
	return hash, nil

}

func getTorrentInfo(hash string) (qbt.TorrentInfo, error) {
	torrentInfoList, err := qb.Torrents(qbt.TorrentsOptions{Hashes: []string{hash}})
	if len(torrentInfoList) == 0 {
		return qbt.TorrentInfo{}, fmt.Errorf("no torrent info found for hash: %s", hash)
	}
	if err != nil {
		return torrentInfoList[0], err
	}
	return torrentInfoList[0], nil

}

func handleHttpURL(chatID int64, URL string) error {
	opts := gopeed.GopeedOptions{
		Extra: &gopeed.GopeedExtraOptions{Connections: 32},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ddId, err := gp.CreateTaskFromURL(ctx, URL, opts)
	if err != nil {
		return fmt.Errorf("failed to create direct download task: %v", err)
	}
	l.log.Debug("Direct download task created with ID:", slog.String("ddId", ddId))
	ddInfo, err := gp.GetTask(ctx, ddId)
	if err != nil {
		l.log.Error("Error getting direct download task info", slog.Any("error", err))
		return err
	}
	msg := messages.BuildDirectDownloadProgress(chatID, ddInfo)
	msgSended, err := bot.Send(msg)
	sendDirectDownloadInfo(chatID, ddInfo, msgSended)
	return err
}
func sendDirectDownloadInfo(chatID int64, ddInfo gopeed.GopeedTask, msgSended tgbotapi.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	mutex := sync.Mutex{}
	cancelGoroutines[msgSended.MessageID] = cancel
	go func() {
		var err error
		defer func() {
			mutex.Lock()
			delete(cancelGoroutines, msgSended.MessageID)
			mutex.Unlock()
			l.log.Debug(
				"End of goroutine for MessageID",
				slog.Int("messageid", msgSended.MessageID),
			)
		}()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for ddInfo.Status != gopeed.GopeedStatusDone && ddInfo.Status != gopeed.GopeedStatusError {
			select {
			case <-ctx.Done():
				l.log.Debug(
					"End of goroutine for MessageID",
					slog.Int("messageid", msgSended.MessageID),
				)
				return
			case <-ticker.C:
				contextInfo, cancelInfo := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancelInfo()
				ddInfo, err = gp.GetTask(contextInfo, ddInfo.ID)
				if err != nil {
					l.log.Error("Error getting direct download task info", slog.Any("error", err))
					return
				}
				msg := messages.BuildDirectDownloadProgress(chatID, ddInfo)
				newMsg := tgbotapi.NewEditMessageText(chatID, msgSended.MessageID, msg.Text)
				newMsg.ParseMode = tgbotapi.ModeHTML
				if markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup); ok {
					newMsg.ReplyMarkup = &markup
				}
				_, err = bot.Send(newMsg)
				if err != nil {
					l.log.Error("Error sending message", slog.Any("error", err))
				}

			}
		}
		l.log.Debug(
			"Direct download completed or stopped for chatID",
			slog.Int64("chatID", chatID),
			slog.String("ddName", ddInfo.Meta.Res.Name),
			slog.String("ddStatus", string(ddInfo.Status)),
		)
		if ddInfo.Status == gopeed.GopeedStatusError {
			l.log.Error(
				"Direct download task ended with error",
				slog.Int64("chatID", chatID))
		}
		msgText := fmt.Sprintf("✅ <b>Download Complete!</b> Your file: %s is ready.", ddInfo.Name)
		finalMsg := tgbotapi.NewMessage(chatID, msgText)
		finalMsg.ParseMode = tgbotapi.ModeHTML
		_, err = bot.Send(finalMsg)
		if err != nil {
			l.log.Error("Error sending final message", slog.Any("error", err))
		}
		_, err = bot.Send(tgbotapi.NewDeleteMessage(chatID, msgSended.MessageID))
		if err != nil {
			l.log.Error("Error deleting message", slog.Any("error", err))
		}
	}()
}
