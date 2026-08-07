package messages

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

func BuildTorrentList(chatID int64, torrentList []qbt.TorrentInfo) tgbotapi.MessageConfig {
	var messageText string
	messageText = "<pre>\nFILE                      SIZE     STATUS        PROGRESS       SPEED      ETA\n-------------------------------------------------------------------------------\n"
	for _, torrent := range torrentList {
		messageText += formatRow(
			torrent.Name,
			torrent.Size,
			torrent.State,
			torrent.Progress,
			torrent.Dlspeed,
			torrent.Eta,
		)
	}
	messageText += "</pre>"
	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = buildInlineKeyboard(torrentList)
	return msg

}

func formatRow(
	fileName string,
	sizeBytes int64,
	rawStatus string,
	progress float64,
	speedBytes int64,
	etaSeconds int64,
) string {
	cleanName := truncateFilename(fileName, 23)
	cleanSize := formatBytes(sizeBytes)
	cleanStatusStr := formatStatus(rawStatus)
	cleanSpeed := formatBytes(speedBytes) + "/s"
	cleanETA := formatETA(etaSeconds, progress)

	progressBar := buildProgressBar(progress)

	return fmt.Sprintf("%-22s %-8s %-13s %-12s   %-8s %s\n",
		cleanName,
		cleanSize,
		cleanStatusStr,
		progressBar,
		cleanSpeed,
		cleanETA,
	)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatStatus(raw string) string {
	switch raw {
	case "downloading":
		return "⏬ Downloading"
	case "stalledDL":
		return "⏳ Stalled DL"
	case "stalledUP":
		return "⏸️ Stalled UP"
	case "stoppedDL", "pausedDL":
		return "⏹️ Stopped"
	case "completed", "uploading":
		return "✅ Completed"
	default:
		return raw
	}
}

func formatETA(etaSeconds int64, progress float64) string {
	if progress >= 1.0 {
		return "Done"
	}
	if etaSeconds >= 8640000 || etaSeconds <= 0 {
		return "--"
	}

	d := time.Duration(etaSeconds) * time.Second
	if d >= 24*time.Hour {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	if d >= time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
}

func truncateFilename(name string, maxLen int) string {
	runes := []rune(name)
	if len(runes) <= maxLen {
		return name
	}
	return string(runes[:maxLen-3]) + "..."
}

func buildInlineKeyboard(torrentList []qbt.TorrentInfo) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton
	for i, t := range torrentList {
		btnTxt := fmt.Sprintf("🔍 Ver #%d (%s)", i+1, truncateFilename(t.Name, 15))
		btn := tgbotapi.NewInlineKeyboardButtonData(btnTxt, fmt.Sprintf("torrent:info:%s", t.Hash))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
