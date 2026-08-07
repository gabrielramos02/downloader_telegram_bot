package main

import (
	"testing"
)

func TestParseCallbackData(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantScope  CallbackScope
		wantAction CallbackAction
		wantID     string
		wantOK     bool
	}{
		{"torrent cancel", "torrent:cancel:abc123", ScopeTorrent, ActionCancel, "abc123", true},
		{"torrent info", "torrent:info:xyz", ScopeTorrent, ActionInfo, "xyz", true},
		{"torrent refresh", "torrent:refresh:zzz", ScopeTorrent, ActionRefresh, "zzz", true},
		{"dd cancel", "dd:cancel:task1", ScopeDD, ActionCancel, "task1", true},
		{"empty id", "torrent:cancel:", ScopeTorrent, ActionCancel, "", true},
		{"unknown scope parses ok", "delete:cancel:abc", "delete", "cancel", "abc", true},
		{"missing parts", "info", "", "", "", false},
		{"empty data", "", "", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, action, id, ok := parseCallbackData(tt.in)
			if scope != tt.wantScope || action != tt.wantAction || id != tt.wantID || ok != tt.wantOK {
				t.Errorf("parseCallbackData(%q) = (%q, %q, %q, %v), want (%q, %q, %q, %v)",
					tt.in, scope, action, id, ok, tt.wantScope, tt.wantAction, tt.wantID, tt.wantOK)
			}
		})
	}
}
