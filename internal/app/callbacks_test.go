package app

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
		{"dd cancel", "dd:cancel:task1", ScopeDD, ActionCancel, "task1", true},
		{"dd info", "dd:info:xyz", ScopeDD, ActionInfo, "xyz", true},
		{"dd refresh", "dd:refresh:zzz", ScopeDD, ActionRefresh, "zzz", true},
		{"dd pause", "dd:pause:task2", ScopeDD, ActionPause, "task2", true},
		{"dd continue", "dd:continue:task3", ScopeDD, ActionContinue, "task3", true},
		{"empty id", "dd:cancel:", ScopeDD, ActionCancel, "", true},
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
