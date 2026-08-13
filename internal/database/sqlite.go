package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/gabrielramos02/telegram-bot-go/sql/schema"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func OpenDB(dsn string) (*sql.DB, error) {
	parsed, err := parseDBURL(dsn)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	goose.SetBaseFS(schema.FS)
	err := goose.SetDialect("sqlite3")
	if err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}
	err = goose.Up(db, ".")
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func parseDBURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme != "" && u.Scheme != "sqlite" {
		return "", fmt.Errorf("unsupported DB scheme %q", u.Scheme)
	}
	path := strings.TrimPrefix(u.Path, "/")
	if path == "" {
		return "", fmt.Errorf("empty database path")
	}
	return "file:" + path + "?_pragma=foreign_keys(1)", nil
}
