package messages

import (
	"fmt"
	"html"
	"strings"

	"github.com/gabrielramos02/telegram-bot-go/internal/gopeed"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func BuildDirectDownloads(chatID int64, taskList []gopeed.GopeedTask) tgbotapi.MessageConfig {
	var messageText string

	if len(taskList) == 0 {
		messageText = "No direct downloads found."
	} else {
		messageText = "<b>⬇️ Direct Downloads</b>\n\n"
		for i, task := range taskList {
			messageText += formatDirectDownloadTask(i+1, task)
			if i < len(taskList)-1 {
				messageText += "\n"
			}
		}
	}

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = tgbotapi.ModeHTML
	if len(taskList) > 0 {
		msg.ReplyMarkup = buildDirectDownloadKeyboard(taskList)
	}
	return msg
}

func formatDirectDownloadTask(index int, task gopeed.GopeedTask) string {
	status := formatGopeedStatus(task.Status)
	size := effectiveTaskSize(task)
	progress := computeGopeedProgress(task, size)
	progressBar := buildProgressBar(progress)

	var urlText string
	url := strings.TrimSpace(task.Meta.Req.URL)
	if url != "" {
		urlText = fmt.Sprintf("<b>URL:</b> <code>%s</code>\n", escapeHTML(url))
	}

	return fmt.Sprintf(
		"<b>#%d %s</b>\n%s"+
			"<b>Status:</b> %s\n"+
			"<b>Progress:</b> %s <code>%s / %s</code>\n"+
			"<b>Speed:</b> ⬇️ <code>%s/s</code>\n",
		index,
		html.EscapeString(task.Name),
		urlText,
		status,
		progressBar,
		formatBytes(task.Progress.Downloaded),
		formatBytes(size),
		formatBytes(task.Progress.Speed),
	)
}

func formatGopeedStatus(status gopeed.GopeedStatus) string {
	switch status {
	case gopeed.GopeedStatusReady:
		return "🟢 Ready"
	case gopeed.GopeedStatusRunning:
		return "⏬ Running"
	case gopeed.GopeedStatusPause:
		return "⏸️ Paused"
	case gopeed.GopeedStatusWait:
		return "⏳ Waiting"
	case gopeed.GopeedStatusError:
		return "❌ Error"
	case gopeed.GopeedStatusDone:
		return "✅ Done"
	default:
		return string(status)
	}
}

func buildDirectDownloadKeyboard(taskList []gopeed.GopeedTask) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, task := range taskList {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("🗑 Eliminar %s", truncateFilename(task.Name, 20)),
			fmt.Sprintf("dd:cancel:%s", task.ID),
		)
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
