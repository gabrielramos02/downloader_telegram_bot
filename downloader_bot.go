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
var requiredEnvVars = []string{"BOT_TOKEN", "ENV", "QB_URL", "QB_USERNAME", "QB_PASSWORD"}

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
	cfg, err := loadConfig(map[string]string{
		"BOT_TOKEN":   os.Getenv("BOT_TOKEN"),
		"ENV":         os.Getenv("ENV"),
		"QB_URL":      os.Getenv("QB_URL"),
		"QB_USERNAME": os.Getenv("QB_USERNAME"),
		"QB_PASSWORD": os.Getenv("QB_PASSWORD"),
	})
	if err != nil {
		log.Panic(err.Error())
	}

	logger, closeLogger, err := initializeLogger()
	logger = logger.With(
		slog.String("env", cfg.Env),
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

	qb = qbt.NewClient(cfg.QBURL)
	err = qb.Login(cfg.QBUsername, cfg.QBPassword)
	if err != nil {
		l.log.Error("error during login", slog.Any("error", err))
	}

	bot, err = tgbotapi.NewBotAPI(cfg.BotToken)
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
func validateEnvVars(vars map[string]string) error {
	for _, v := range requiredEnvVars {
		if vars[v] == "" {
			return fmt.Errorf("required env variable %s not set", v)
		}
	}
	return nil
}

type config struct {
	BotToken   string
	Env        string
	QBURL      string
	QBUsername string
	QBPassword string
}

func loadConfig(vars map[string]string) (config, error) {
	if err := validateEnvVars(vars); err != nil {
		return config{}, err
	}
	return config{
		BotToken:   vars["BOT_TOKEN"],
		Env:        vars["ENV"],
		QBURL:      vars["QB_URL"],
		QBUsername: vars["QB_USERNAME"],
		QBPassword: vars["QB_PASSWORD"],
	}, nil
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
	if isCommand(text) {
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
func isCommand(text string) bool {
	return strings.HasPrefix(text, "/")
}
