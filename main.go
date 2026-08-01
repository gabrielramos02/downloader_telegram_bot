package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

var bot *tgbotapi.BotAPI
var qb *qbt.Client

const (
)

func main() {
	var err error

	err = godotenv.Load(".env")
	if err != nil {
		log.Panic(err.Error())
	}

	BOT_TOKEN := os.Getenv("BOT_TOKEN")
	if BOT_TOKEN == "" {
		log.Panic("BOT_TOKEN env variable not set")
	}

	QB_URL := os.Getenv("QB_URL")
	if QB_URL == "" {
		log.Panic("QB_URL env variable not set")
	}

	QB_USERNAME := os.Getenv("QB_USERNAME")
	if QB_USERNAME == "" {
		log.Panic("QB_USERNAME env variable not set")
	}

	QB_PASSWORD := os.Getenv("QB_PASSWORD")
	if QB_PASSWORD == "" {
		log.Panic("QB_USERNAME env variable not set")
	}

	qb = qbt.NewClient(QB_URL)
	err = qb.Login(QB_USERNAME, QB_PASSWORD)
	if err != nil {
		log.Fatalf("Error during login: %s", err.Error())
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
