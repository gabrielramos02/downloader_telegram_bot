package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var bot *tgbotapi.BotAPI

func main() {
	var err error
	err = godotenv.Load(".env")
	if err != nil {
		log.Panic(err.Error())
	}
	BOT_TOKEN := os.Getenv("BOT_TOKEN")
	fmt.Printf("%v", BOT_TOKEN)
	if BOT_TOKEN == "" {
		log.Panic("BOT_TOKEN env variable not set")
	}
	bot, err = tgbotapi.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	go receiveUpdates(ctx, updates)
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			handleUpdate(update)
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handleMessage(update.Message)
	case update.CallbackQuery != nil:
		handleCallbackQuery(update.CallbackQuery)
	}

}

func handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}
	var err error
	if strings.HasPrefix(text, "/") {
		err = handleCommand(message.Chat.ID, text)
	}

	if err != nil {
		log.Printf("An error ocurred: %s", err)

	}
}

func handleCallbackQuery(query *tgbotapi.CallbackQuery) {

}

func handleCommand(chatID int64, command string) error {
	var err error
	switch command {
	case "/start":
		err = sendStart(chatID)
	}
	return err
}

func sendStart(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Hello to my new bot")
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)

	return err
}
