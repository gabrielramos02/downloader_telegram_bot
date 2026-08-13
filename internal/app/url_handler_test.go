package app

import (
	"strings"
	"testing"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantScheme string
		wantErr    string
	}{
		{
			name:       "valid magnet",
			in:         "magnet:?xt=urn:btih:abc&dn=file",
			wantScheme: "magnet",
		},
		{
			name:       "http scheme supported",
			in:         "http://example.com",
			wantScheme: "http",
		},
		{
			name:       "https scheme supported",
			in:         "https://example.com",
			wantScheme: "https",
		},
		{
			name:    "empty string has empty scheme",
			in:      "",
			wantErr: "unsupported URL scheme",
		},
		{
			name:    "invalid URL fails to parse",
			in:      "http://[::1]:badport",
			wantErr: "invalid URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, err := parseURL(tt.in)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("parseURL(%q) expected error containing %q, got nil", tt.in, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("parseURL(%q) error = %q, want contains %q", tt.in, err, tt.wantErr)
				}
				if scheme != "" {
					t.Errorf("parseURL(%q) = %q, want empty on error", tt.in, scheme)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseURL(%q) unexpected error: %v", tt.in, err)
			}
			if scheme != tt.wantScheme {
				t.Errorf("parseURL(%q) scheme = %q, want %q", tt.in, scheme, tt.wantScheme)
			}
		})
	}
}
