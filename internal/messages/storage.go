package messages

import (
	"fmt"

	"github.com/gabrielramos02/telegram-bot-go/internal/glances"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func BuildStorageInfo(chatID int64, filesystems []glances.FileSystem) tgbotapi.MessageConfig {
	rootFS := filterRootMount(filesystems)

	var messageText string

	if len(rootFS) == 0 {
		messageText = "No root filesystem found."
	} else {
		messageText = "<b>💾 Storage</b>\n\n"
		for i, fs := range rootFS {
			messageText += formatFileSystem(i+1, fs)
			if i < len(rootFS)-1 {
				messageText += "\n"
			}
		}
	}

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = buildStorageInfoMarkup()
	return msg
}

func buildStorageInfoMarkup() tgbotapi.InlineKeyboardMarkup {
	refreshBtn := tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "st:refresh:")
	closeBtn := tgbotapi.NewInlineKeyboardButtonData("❌ Close", "st:close:")
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(refreshBtn, closeBtn),
	)
}

func filterRootMount(filesystems []glances.FileSystem) []glances.FileSystem {
	var rootFS []glances.FileSystem
	for _, fs := range filesystems {
		if fs.MountPoint == "/rootfs" {
			rootFS = append(rootFS, fs)
		}
	}
	return rootFS
}

func formatFileSystem(index int, fs glances.FileSystem) string {
	return fmt.Sprintf(
		"<b>#%d %s</b>\n"+
			"<b>Mount:</b> <code>%s</code>\n"+
			"<b>Type:</b> <code>%s</code>\n"+
			"<b>Usage:</b> %s <code>%s / %s</code>\n",
		index,
		escapeHTML(fs.MountPoint),
		escapeHTML(fs.MountPoint),
		escapeHTML(fs.FSType),
		buildProgressBar(fs.Percent/100),
		formatBytes(fs.Used),
		formatBytes(fs.Size),
	)
}
