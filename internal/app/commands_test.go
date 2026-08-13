package app

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestHandleCommandUnknown(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"unknown command", "/unknown"},
		{"empty command", ""},
		{"bare slash", "/"},
		{"trailing space", "/start "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 123}}
			if err := handleCommand(msg, tt.command); err == nil {
				t.Errorf("handleCommand(%q) = nil, want error", tt.command)
			}
		})
	}
}

func TestCommandHandlersRegistered(t *testing.T) {
	expected := []string{"/start", "/get_direct_downloads", "/get_storage_info"}
	for _, command := range expected {
		if _, exists := commandHandlers[command]; !exists {
			t.Errorf("expected handler for %q to be registered", command)
		}
	}
}
