package messages

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

func BuildTorrentProgress(chatID int64, t qbt.TorrentInfo) tgbotapi.MessageConfig {

	downloaded := t.Progress * float64(t.TotalSize)
	progressBar := buildProgressBar(t.Progress)

	htmlString := fmt.Sprintf(
		"<b>📌 %s</b>\n\n"+
			"<b>Progreso:</b> %s\n"+
			"<b>Descargado:</b> %s / %s\n"+
			"<b>Velocidad:</b> ⬇️ %s/s | ⬆️ %s/s\n"+
			"<b>Tiempo restante:</b> %s\n"+
			"<b>Semillas / Pares:</b> 🌱 %d | 👤 %d \n"+
			"<b>Ruta:</b> <code>%s</code>",
		escapeHTML(t.Name),
		progressBar,
		formatBytes(int64(downloaded)), formatBytes(int64(t.TotalSize)),
		formatBytes(int64(t.Dlspeed)), formatBytes(int64(t.Upspeed)),
		formatETA(int64(t.Eta), t.Progress),
		t.NumSeeds, t.NumLeechs,
		escapeHTML(t.SavePath),
	)
	msg := tgbotapi.NewMessage(chatID, htmlString)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = buildCancelMarkup(t.Hash)
	return msg
}
func escapeHTML(text string) string {
	r := strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	return r.Replace(text)
}
func buildProgressBar(progress float64) string {
	filled := int(progress * 10)
	if filled > 10 {
		filled = 10
	} else if filled < 0 {
		filled = 0
	}
	progressBar := fmt.Sprintf("<code>[%s%s]</code> %.1f%%",
		strings.Repeat("█", filled),
		strings.Repeat("░", 10-filled),
		progress*100,
	)
	return progressBar

}

func buildCancelMarkup(hash string) tgbotapi.InlineKeyboardMarkup {
	data := fmt.Sprintf("torrent:cancel:%s", hash)
	cancelMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", data),
		),
	)
	return cancelMarkup

}
