package app

import (
	"maps"
	"strings"
	"testing"
)

func TestIsCommand(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"slash command", "/start", true},
		{"bare slash", "/", true},
		{"plain text", "hola", false},
		{"empty text", "", false},
		{"leading space", " /start", false},
		{"slash in middle", "a/b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCommand(tt.in); got != tt.want {
				t.Errorf("isCommand(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	fullEnv := map[string]string{
		"BOT_TOKEN": "tok",
		"ENV":       "prod",
		"GP_URL":    "http://gopeed:9999",
		"GP_TOKEN":  "gopeed-secret",
	}
	t.Run("all env vars set", func(t *testing.T) {
		cfg, err := LoadConfig(fullEnv)
		if err != nil {
			t.Fatalf("loadConfig unexpected error: %v", err)
		}
		want := Config{
			BotToken: "tok",
			Env:      "prod",
			GPURL:    "http://gopeed:9999",
			GPToken:  "gopeed-secret",
		}
		if cfg != want {
			t.Errorf("loadConfig = %+v, want %+v", cfg, want)
		}
	})
	t.Run("missing BOT_TOKEN", func(t *testing.T) {
		env := maps.Clone(fullEnv)
		delete(env, "BOT_TOKEN")
		if _, err := LoadConfig(env); err == nil || !strings.Contains(err.Error(), "BOT_TOKEN") {
			t.Errorf("expected error mentioning BOT_TOKEN, got %v", err)
		}
	})
	t.Run("missing ENV", func(t *testing.T) {
		env := maps.Clone(fullEnv)
		delete(env, "ENV")
		if _, err := LoadConfig(env); err == nil || !strings.Contains(err.Error(), "ENV") {
			t.Errorf("expected error mentioning ENV, got %v", err)
		}
	})
	t.Run("missing GP_URL", func(t *testing.T) {
		env := maps.Clone(fullEnv)
		delete(env, "GP_URL")
		if _, err := LoadConfig(env); err == nil || !strings.Contains(err.Error(), "GP_URL") {
			t.Errorf("expected error mentioning GP_URL, got %v", err)
		}
	})
	t.Run("missing GP_TOKEN", func(t *testing.T) {
		env := maps.Clone(fullEnv)
		delete(env, "GP_TOKEN")
		if _, err := LoadConfig(env); err == nil || !strings.Contains(err.Error(), "GP_TOKEN") {
			t.Errorf("expected error mentioning GP_TOKEN, got %v", err)
		}
	})
	t.Run("empty env reports first missing", func(t *testing.T) {
		if _, err := LoadConfig(
			map[string]string{},
		); err == nil ||
			!strings.Contains(err.Error(), "BOT_TOKEN") {
			t.Errorf("expected error mentioning BOT_TOKEN, got %v", err)
		}
	})
}

func TestValidateEnvVars(t *testing.T) {
	fullEnv := map[string]string{
		"BOT_TOKEN": "tok",
		"ENV":       "prod",
		"GP_URL":    "http://gopeed:9999",
		"GP_TOKEN":  "gopeed-secret",
	}
	t.Run("all set", func(t *testing.T) {
		if err := validateEnvVars(fullEnv); err != nil {
			t.Errorf("validateEnvVars unexpected error: %v", err)
		}
	})
	t.Run("missing ENV", func(t *testing.T) {
		env := maps.Clone(fullEnv)
		delete(env, "ENV")
		if err := validateEnvVars(env); err == nil || !strings.Contains(err.Error(), "ENV") {
			t.Errorf("expected error mentioning ENV, got %v", err)
		}
	})
	t.Run("missing GP_TOKEN", func(t *testing.T) {
		env := maps.Clone(fullEnv)
		delete(env, "GP_TOKEN")
		if err := validateEnvVars(env); err == nil || !strings.Contains(err.Error(), "GP_TOKEN") {
			t.Errorf("expected error mentioning GP_TOKEN, got %v", err)
		}
	})
}
