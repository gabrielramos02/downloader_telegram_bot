package messages

import (
	"fmt"
	"html"

	gopeed "github.com/gabrielramos02/gopeed-api-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func BuildDirectDownloadProgress(chatID int64, task gopeed.GopeedTask) tgbotapi.MessageConfig {
	status := formatGopeedStatus(task.Status)
	size := effectiveTaskSize(task)
	progress := computeGopeedProgress(task, size)
	progressBar := buildProgressBar(progress)

	htmlString := fmt.Sprintf(
		"<b>📌 %s</b>\n\n"+
			"<b>Status:</b> %s\n"+
			"<b>Progress:</b> %s\n"+
			"<b>Downloaded:</b> <code>%s / %s</code>\n"+
			"<b>Speed:</b> ⬇️ <code>%s/s</code>\n"+
			"<b>Protocol:</b> <code>%s</code>\n",
		html.EscapeString(task.Name),
		status,
		progressBar,
		formatBytes(task.Progress.Downloaded),
		formatBytes(size),
		formatBytes(task.Progress.Speed),
		html.EscapeString(task.Protocol),
	)

	if task.Meta.Req.URL != "" {
		htmlString += fmt.Sprintf("<b>URL:</b> <code>%s</code>\n", escapeHTML(task.Meta.Req.URL))
	}

	msg := tgbotapi.NewMessage(chatID, htmlString)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = buildDirectDownloadCancelMarkup(task.ID)
	return msg
}

func computeGopeedProgress(task gopeed.GopeedTask, size int64) float64 {
	if size <= 0 {
		if task.Status == gopeed.GopeedStatusDone {
			return 1.0
		}
		return 0
	}
	return float64(task.Progress.Downloaded) / float64(size)
}

func effectiveTaskSize(task gopeed.GopeedTask) int64 {
	if task.Size > 0 {
		return task.Size
	}
	if task.Meta.Res.Size > 0 {
		return task.Meta.Res.Size
	}
	var fileSize int64
	for _, f := range task.Meta.Res.Files {
		fileSize += f.Size
	}
	return fileSize
}

func buildDirectDownloadCancelMarkup(taskID string) tgbotapi.InlineKeyboardMarkup {
	data := fmt.Sprintf("dd:cancel:%s", taskID)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", data),
		),
	)
}
