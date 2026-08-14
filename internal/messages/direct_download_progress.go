package messages

import (
	"fmt"
	"html"

	gopeed "github.com/gabrielramos02/gopeed-api-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func BuildDirectDownloadProgress(chatID int64, task gopeed.GopeedTask) tgbotapi.MessageConfig {
	return buildDirectDownloadProgress(chatID, task, true)
}

func BuildDirectDownloadProgressAuto(chatID int64, task gopeed.GopeedTask) tgbotapi.MessageConfig {
	return buildDirectDownloadProgress(chatID, task, false)
}

func buildDirectDownloadProgress(chatID int64, task gopeed.GopeedTask, showRefresh bool) tgbotapi.MessageConfig {
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
		htmlString += fmt.Sprintf("<b>URL:</b> %s\n", formatURLLink(task.Meta.Req.URL))
	}

	msg := tgbotapi.NewMessage(chatID, htmlString)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = buildDirectDownloadInfoMarkup(task, showRefresh)
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

func buildDirectDownloadInfoMarkup(task gopeed.GopeedTask, showRefresh bool) tgbotapi.InlineKeyboardMarkup {
	cancelBtn := tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", fmt.Sprintf("dd:cancel:%s", task.ID))
	switch task.Status {
	case gopeed.GopeedStatusPause:
		continueBtn := tgbotapi.NewInlineKeyboardButtonData(
			"▶️ Continuar",
			fmt.Sprintf("dd:continue:%s", task.ID),
		)
		return buildDirectDownloadActionMarkup(
			tgbotapi.NewInlineKeyboardRow(continueBtn, cancelBtn),
			task.ID,
			showRefresh,
		)
	case gopeed.GopeedStatusRunning:
		pauseBtn := tgbotapi.NewInlineKeyboardButtonData(
			"⏸️ Pausar",
			fmt.Sprintf("dd:pause:%s", task.ID),
		)
		return buildDirectDownloadActionMarkup(
			tgbotapi.NewInlineKeyboardRow(pauseBtn, cancelBtn),
			task.ID,
			showRefresh,
		)
	case gopeed.GopeedStatusDone:
		deleteBtn := tgbotapi.NewInlineKeyboardButtonData(
			"🗑️ Eliminar",
			fmt.Sprintf("dd:delete:%s", task.ID),
		)
		return buildDirectDownloadActionMarkup(
			tgbotapi.NewInlineKeyboardRow(deleteBtn),
			task.ID,
			showRefresh,
		)
	default:
		return buildDirectDownloadCancelMarkup(task.ID, showRefresh)
	}
}

func buildDirectDownloadActionMarkup(
	actionRow []tgbotapi.InlineKeyboardButton,
	taskID string,
	showRefresh bool,
) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{actionRow}
	if showRefresh {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(buildDirectDownloadRefreshButton(taskID)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildDirectDownloadRefreshButton(taskID string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(
		"🔄 Refresh",
		fmt.Sprintf("dd:refresh:%s", taskID),
	)
}

func buildDirectDownloadCancelMarkup(taskID string, showRefresh bool) tgbotapi.InlineKeyboardMarkup {
	data := fmt.Sprintf("dd:cancel:%s", taskID)
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", data),
		),
	}
	if showRefresh {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(buildDirectDownloadRefreshButton(taskID)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
