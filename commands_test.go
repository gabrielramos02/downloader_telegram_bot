package main

import (
	"testing"
)

func TestCommandAction(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		want   string
		wantOK bool
	}{
		{"start command", "/start", startCommand, true},
		{"get torrents command", "/get_torrents", getTorrentsCommand, true},
		{"unknown command", "/unknown", "", false},
		{"empty command", "", "", false},
		{"bare slash", "/", "", false},
		{"trailing space", "/start ", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := commandAction(tt.in)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("commandAction(%q) = (%q, %v), want (%q, %v)", tt.in, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
