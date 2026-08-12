package app

import (
	"strings"
	"testing"

	"github.com/superturkey650/go-qbittorrent/qbt"
)

func TestClassifyURL(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantURL    string
		wantScheme string
		wantErr    string
	}{
		{
			name:       "valid magnet",
			in:         "magnet:?xt=urn:btih:abc&dn=file",
			wantURL:    "magnet:?xt=urn:btih:abc&dn=file",
			wantScheme: "magnet",
		},
		{
			name:       "http scheme supported",
			in:         "http://example.com",
			wantURL:    "http://example.com",
			wantScheme: "http",
		},
		{
			name:       "https scheme mapped to http",
			in:         "https://example.com",
			wantURL:    "https://example.com",
			wantScheme: "http",
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
			url, scheme, err := parseURL(tt.in)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf(
						"classifyURL(%q) expected error containing %q, got nil",
						tt.in,
						tt.wantErr,
					)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("classifyURL(%q) error = %q, want contains %q", tt.in, err, tt.wantErr)
				}
				if url != "" || scheme != "" {
					t.Errorf("classifyURL(%q) = (%q, %q), want empty on error", tt.in, url, scheme)
				}
				return
			}
			if err != nil {
				t.Fatalf("classifyURL(%q) unexpected error: %v", tt.in, err)
			}
			if url != tt.wantURL {
				t.Errorf("classifyURL(%q) url = %q, want %q", tt.in, url, tt.wantURL)
			}
			if scheme != tt.wantScheme {
				t.Errorf("classifyURL(%q) scheme = %q, want %q", tt.in, scheme, tt.wantScheme)
			}
		})
	}
}

func TestExtractHashFromMagnet(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr string
	}{
		{
			name: "valid btih hash lowercased",
			in:   "magnet:?xt=urn:btih:ABC123DEF&dn=file",
			want: "abc123def",
		},
		{
			name: "uppercase hash lowercased",
			in:   "magnet:?xt=urn:btih:ABCDEF",
			want: "abcdef",
		},
		{
			name: "empty hash",
			in:   "magnet:?xt=urn:btih:",
			want: "",
		},
		{
			name:    "no xt parameter",
			in:      "magnet:?dn=file",
			wantErr: "no BTIH hash found in magnet URL",
		},
		{
			name:    "xt without btih prefix",
			in:      "magnet:?xt=urn:sha1:xxxx",
			wantErr: "no BTIH hash found in magnet URL",
		},
		{
			name: "second xt is btih",
			in:   "magnet:?xt=urn:sha1:xxxx&xt=urn:btih:deadbeef",
			want: "deadbeef",
		},
		{
			name: "ampersand inside hash is stripped",
			in:   "magnet:?xt=urn:btih:abc%26xyz",
			want: "abc",
		},
		{
			name:    "malformed url",
			in:      "http://[::1]:badport",
			wantErr: "invalid port",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractHashFromMagnet(tt.in)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf(
						"extractHashFromMagnet(%q) expected error containing %q, got nil",
						tt.in,
						tt.wantErr,
					)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf(
						"extractHashFromMagnet(%q) error = %q, want contains %q",
						tt.in,
						err,
						tt.wantErr,
					)
				}
				return
			}
			if err != nil {
				t.Fatalf("extractHashFromMagnet(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("extractHashFromMagnet(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsTorrentInProgress(t *testing.T) {
	tests := []struct {
		name  string
		state string
		prog  float64
		want  bool
	}{
		{"downloading in progress", "downloading", 0.5, true},
		{"downloading complete", "downloading", 1.0, false},
		{"stalledUP stops", "stalledUP", 0.5, false},
		{"error stops", "error", 0.5, false},
		{"empty state no progress", "", 0.0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			torrent := qbt.TorrentInfo{State: tt.state, Progress: tt.prog}
			if got := isTorrentInProgress(torrent); got != tt.want {
				t.Errorf("isTorrentInProgress(%v) = %v, want %v", torrent, got, tt.want)
			}
		})
	}
}
