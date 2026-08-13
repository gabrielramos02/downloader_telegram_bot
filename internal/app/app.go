package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	gopeed "github.com/gabrielramos02/gopeed-api-go"
	"github.com/gabrielramos02/telegram-bot-go/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI
var gp *gopeed.GopeedClient
var db *database.Queries

type loggerClient struct {
	log   *slog.Logger
	close closeFunc
}

var l *loggerClient = &loggerClient{
	log:   slog.Default(),
	close: func() error { return nil },
}

func Run(cfg Config) error {

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

	// Initialize Gopeed client
	gp, err = gopeed.NewClient(cfg.GPURL, gopeed.WithAPIToken(cfg.GPToken))
	if err != nil {
		l.log.Error("error during Gopeed client initialization", slog.Any("error", err))
	}
	ctxInfo, cancelInfo := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelInfo()
	if info, err := gp.GetInfo(ctxInfo); err != nil {
		l.log.Error("error getting Gopeed info", slog.Any("error", err))
	} else {
		l.log.Info("Gopeed info", slog.String("version", info.Version))
	}

	// Initialize database connection
	dbIntance, err := database.OpenDB(cfg.DBURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		if err = dbIntance.Close(); err != nil {
			l.log.Error("failed to close database", slog.Any("error", err))
		}
	}()
	err = database.Migrate(dbIntance)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	db = database.New(dbIntance)

	// Initialize Telegram bot
	bot, err = tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		l.log.Error("error during bot initialization", slog.Any("error", err))
		return err
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
	return nil
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
		err = handleCommand(message, text)
	} else {
		err = handleUrl(message, text)
	}
	if err != nil {
		l.log.Error("Error handling message", slog.Any("error", err))
	}
}
func isCommand(text string) bool {
	return strings.HasPrefix(text, "/")
}
