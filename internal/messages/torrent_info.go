package messages

import (
	"fmt"
	"html"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

func BuildTorrentInfo(chatID int64, torrent qbt.TorrentInfo) tgbotapi.MessageConfig {
	progressBar := buildProgressBar(torrent.Progress)
	status := formatStatus(torrent.State)

	msgText := ""

	// Header: Escaped name to prevent HTML parsing failures
	msgText += fmt.Sprintf("📌 <b>%s</b>\n\n", html.EscapeString(torrent.Name))

	// Torrent Details Card
	msgText += fmt.Sprintf("<b>State:</b> %s \n", status)
	msgText += fmt.Sprintf("<b>Progress:</b> <code>%s </code>\n", progressBar)
	msgText += fmt.Sprintf("<b>Size:</b> <code>%s</code>\n", formatBytes(torrent.Size))
	msgText += fmt.Sprintf(
		"<b>Speed:</b> ⬇️ <code>%s/s</code> | ⬆️ <code>%s/s</code>\n",
		formatBytes(torrent.Dlspeed),
		formatBytes(torrent.Upspeed),
	)
	msgText += fmt.Sprintf(
		"<b>Seeds / Leechs:</b> <code>%d / %d</code>\n",
		torrent.NumSeeds,
		torrent.NumLeechs,
	)
	msgText += fmt.Sprintf(
		"<b>ETA:</b> <code>%s</code>\n\n",
		formatETA(torrent.Eta, torrent.Progress),
	)

	/* // Dynamic Action Buttons (Pause vs Resume depending on state)
	var toggleBtn tgbotapi.InlineKeyboardButton
	if strings.HasPrefix(torrent.State, "paused") || strings.HasPrefix(torrent.State, "stalled") {
		toggleBtn = tgbotapi.NewInlineKeyboardButtonData("▶️ Resume", "resume:"+torrent.Hash)
	} else {
		toggleBtn = tgbotapi.NewInlineKeyboardButtonData("⏸️ Pause", "pause:"+torrent.Hash)
	} */

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cancel:"+torrent.Hash),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "refresh:"+torrent.Hash),
		),
	)
	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	return msg

}
