package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	gopeed "github.com/gabrielramos02/gopeed-api-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

var bot *tgbotapi.BotAPI
var qb *qbt.Client
var gp *gopeed.GopeedClient

var requiredEnvVars = []string{
	"BOT_TOKEN",
	"ENV",
	"QB_URL",
	"QB_USERNAME",
	"QB_PASSWORD",
	"GP_URL",
	"GP_TOKEN",
}

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
		"GP_URL":      os.Getenv("GP_URL"),
		"GP_TOKEN":    os.Getenv("GP_TOKEN"),
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

	// Initialize qBittorrent client
	qb = qbt.NewClient(cfg.QBURL)
	err = qb.Login(cfg.QBUsername, cfg.QBPassword)
	if err != nil {
		l.log.Error("error during login", slog.Any("error", err))
	}
	if qbVersion, err := qb.WebAPIVersion(); err != nil {
		l.log.Error("error getting qBittorrent version", slog.Any("error", err))
	} else {
		l.log.Info("qBittorrent version", slog.String("version", qbVersion))
	}
	// Initialize Gopeed client
	gp, err = gopeed.NewClient(cfg.GPURL, cfg.GPToken)
	if err != nil {
		l.log.Error("error during Gopeed client initialization", slog.Any("error", err))
	}
	if info, err := gp.GetInfo(""); err != nil {
		l.log.Error("error getting Gopeed info", slog.Any("error", err))
	} else {
		l.log.Info("Gopeed info", slog.String("version", info.Version))
	}

	bot, err = tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		l.log.Error("error during bot initialization", slog.Any("error", err))
		return
	}

	//bot.Debug = true

	l.log.Info("Authorized on account", slog.String("account", bot.Self.UserName))
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	go receiveUpdates(ctx, updates)
	stopCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-stopCtx.Done()
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
	GPURL      string
	GPToken    string
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
		GPURL:      vars["GP_URL"],
		GPToken:    vars["GP_TOKEN"],
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
