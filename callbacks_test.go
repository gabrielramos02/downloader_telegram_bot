package main

import (
	"testing"
)

func TestParseCallbackData(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantAction string
		wantHash   string
		wantOK     bool
	}{
		{"cancel action", "cancel:abc123", "cancel", "abc123", true},
		{"info action", "info:xyz", "info", "xyz", true},
		{"refresh action", "refresh:zzz", "refresh", "zzz", true},
		{"empty hash", "cancel:", "cancel", "", true},
		{"unknown action", "delete:abc", "", "", false},
		{"action without colon", "info", "", "", false},
		{"empty data", "", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, hash, ok := parseCallbackData(tt.in)
			if action != tt.wantAction || hash != tt.wantHash || ok != tt.wantOK {
				t.Errorf("parseCallbackData(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.in, action, hash, ok, tt.wantAction, tt.wantHash, tt.wantOK)
			}
		})
	}
}
