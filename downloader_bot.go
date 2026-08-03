package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

var bot *tgbotapi.BotAPI
var qb *qbt.Client

type loggerClient struct {
	log   *slog.Logger
	close closeFunc
}

var l *loggerClient = &loggerClient{
	log:   slog.Default(),
	close: func() error { return nil },
}

const ()

func main() {
	var err error

	err = godotenv.Load(".env")
	if err != nil {
		log.Panicf("failed to load .env file: %q", err)
	}

	BOT_TOKEN := os.Getenv("BOT_TOKEN")
	if BOT_TOKEN == "" {
		log.Panic("BOT_TOKEN env variable not set")
	}
	env := os.Getenv("ENV")
	if env == "" {
		log.Panic("ENV env variable not set")
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
	logger, closeLogger, err := initializeLogger()
	logger = logger.With(
		slog.String("env", env),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
	}
	l = &loggerClient{
		logger,
		closeLogger,
	}
	defer func() {
		if err = l.close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close logger: %v\n", err)
		}
	}()

	qb = qbt.NewClient(QB_URL)
	err = qb.Login(QB_USERNAME, QB_PASSWORD)
	if err != nil {
		l.log.Error("error during login", slog.Any("error", err))
	}

	bot, err = tgbotapi.NewBotAPI(BOT_TOKEN)
	if err != nil {
		l.log.Error("error during bot initialization", slog.Any("error", err))
		return
	}

	bot.Debug = true

	l.log.Info("Authorized on account", slog.String("account", bot.Self.UserName))
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	go receiveUpdates(ctx, updates)
	_, err = bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		l.log.Error("Error reading from stdin", slog.Any("error", err))
	}
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
	} else {
		err := handleUrl(message.Chat.ID, text)
		if err != nil {
			_, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error: "+err.Error()))
			if err != nil {
				l.log.Error("Error sending error message", slog.Any("error", err))
			}
		}

	}
	if err != nil {
		l.log.Error("Error handling message", slog.Any("error", err))
	}
}
