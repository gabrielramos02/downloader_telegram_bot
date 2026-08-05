package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"slices"

	"github.com/gabrielramos02/telegram-bot-go/internal/loggerErr"
	pkgerr "github.com/pkg/errors"

	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
)

type closeFunc func() error

func initializeLogger() (*slog.Logger, closeFunc, error) {
	logFileEnv := os.Getenv("LOG_FILE")
	var handlers []slog.Handler
	var closeFunction func() error
	if logFileEnv != "" {
		logger := &lumberjack.Logger{
			Filename:   logFileEnv,
			MaxSize:    1,
			MaxAge:     28,
			MaxBackups: 10,
			LocalTime:  false,
			Compress:   true,
		}
		closeFunction = func() error {
			if err := logger.Close(); err != nil {
				return fmt.Errorf("error flushing file: %v", err)
			}
			return nil
		}
		handlers = append(handlers, slog.NewJSONHandler(logger, &slog.HandlerOptions{
			ReplaceAttr: replaceAttr,
		}))
	} else {
		closeFunction = func() error { return nil }
	}
	handlers = append(handlers, tint.NewTextHandler(os.Stderr, &tint.Options{
		Level:       slog.LevelDebug,
		ReplaceAttr: replaceAttr,
	}))
	logger := slog.New(slog.NewMultiHandler(handlers...))
	return logger, closeFunction, nil
}

type multiError interface {
	error
	Unwrap() []error
}
type stackTracer interface {
	error
	StackTrace() pkgerr.StackTrace
}

func errorAttrs(err error) []slog.Attr {
	attrs := []slog.Attr{
		{Key: "message", Value: slog.StringValue(err.Error())},
	}
	attrs = append(attrs, loggerErr.Attrs(err)...)
	if stackErr, ok := errors.AsType[stackTracer](err); ok {
		attrs = append(attrs, slog.Attr{
			Key:   "stack_trace",
			Value: slog.StringValue(fmt.Sprintf("%+v", stackErr.StackTrace())),
		})
	}
	return attrs
}

var sensitiveKeys = []string{"password", "user", "key", "apikey", "secret", "pin", "creditcardno"}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if slices.Contains(sensitiveKeys, a.Key) {
		return slog.String(a.Key, "[REDACTED]")
	}
	if urlParsed, err := url.Parse(a.Value.String()); err == nil {
		if _, ok := urlParsed.User.Password(); ok {
			userString := url.UserPassword(urlParsed.User.Username(), "[REDACTED]")
			urlParsed.User = userString
			return slog.String(a.Key, urlParsed.String())
		}
	}
	if a.Key == "error" {
		err, ok := a.Value.Any().(error)
		if !ok {
			return a
		}
		if me, ok := a.Value.Any().(multiError); ok {
			var errAttrs []slog.Attr
			for i, err := range me.Unwrap() {
				errAttrs = append(
					errAttrs,
					slog.Any(fmt.Sprintf("error_%d", i+1), loggerErr.Attrs(err)),
				)
			}
			return slog.GroupAttrs("errors", errAttrs...)
		}
		return slog.GroupAttrs("error", errorAttrs(err)...)
	}
	return a
}
