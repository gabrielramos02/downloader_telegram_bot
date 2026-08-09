package messages

import (
	"fmt"
	"strings"
	"testing"

	gopeed "github.com/gabrielramos02/telegram-bot-go/gopeed-api-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name string
		in   int64
		want string
	}{
		{"zero", 0, "0 B"},
		{"below one unit", 1023, "1023 B"},
		{"exactly one KB", 1024, "1.0KB"},
		{"one and half KB", 1536, "1.5KB"},
		{"one MB", 1048576, "1.0MB"},
		{"one GB", 1073741824, "1.0GB"},
		{"one TB", 1099511627776, "1.0TB"},
		{"negative", -512, "-512 B"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatBytes(tt.in); got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"downloading", "⏬ Downloading"},
		{"stalledDL", "⏳ Stalled DL"},
		{"stalledUP", "⏸️ Stalled UP"},
		{"stoppedDL", "⏹️ Stopped"},
		{"pausedDL", "⏹️ Stopped"},
		{"completed", "✅ Completed"},
		{"uploading", "✅ Completed"},
		{"unknown-state", "unknown-state"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := formatStatus(tt.in); got != tt.want {
				t.Errorf("formatStatus(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatETA(t *testing.T) {
	tests := []struct {
		name     string
		eta      int64
		progress float64
		want     string
	}{
		{"done when progress complete", 5000, 1.0, "Done"},
		{"eta zero", 0, 0.5, "--"},
		{"eta negative", -1, 0.5, "--"},
		{"eta at large threshold", 8640000, 0.5, "--"},
		{"just under large threshold", 8639999, 0.5, "99d"},
		{"two days", 172800, 0.5, "2d"},
		{"days round down", 90000, 0.5, "1d"},
		{"hour and minute", 3660, 0.5, "1h 1m"},
		{"exact hour", 3600, 0.5, "1h 0m"},
		{"minute and seconds", 90, 0.5, "1m 30s"},
		{"under one minute", 45, 0.5, "0m 45s"},
		{"one second", 1, 0.5, "0m 1s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatETA(tt.eta, tt.progress); got != tt.want {
				t.Errorf("formatETA(%d, %v) = %q, want %q", tt.eta, tt.progress, got, tt.want)
			}
		})
	}
}

func TestTruncateFilename(t *testing.T) {
	t.Run("short name unchanged", func(t *testing.T) {
		const name = "short.txt"
		if got := truncateFilename(name, 10); got != name {
			t.Errorf("truncateFilename(%q, 10) = %q, want unchanged", name, got)
		}
	})
	t.Run("boundary length unchanged", func(t *testing.T) {
		const name = "exactlen10"
		if got := truncateFilename(name, 10); got != name {
			t.Errorf("truncateFilename(%q, 10) = %q, want unchanged", name, got)
		}
	})
	t.Run("long name truncated", func(t *testing.T) {
		if got, want := truncateFilename("abcdefghijklmnop", 10), "abcdefg..."; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("multibyte name not split", func(t *testing.T) {
		if got, want := truncateFilename("日本語ファイル名が長いです.txt", 10), "日本語ファイル..."; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestBuildProgressBar(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0.0, "<code>[░░░░░░░░░░]</code> 0.0%"},
		{0.5, "<code>[█████░░░░░]</code> 50.0%"},
		{1.0, "<code>[██████████]</code> 100.0%"},
		{1.5, "<code>[██████████]</code> 150.0%"},
		{-0.5, "<code>[░░░░░░░░░░]</code> -50.0%"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.in), func(t *testing.T) {
			if got := buildProgressBar(tt.in); got != tt.want {
				t.Errorf("buildProgressBar(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct{ in, want string }{
		{"plain text", "plain text"},
		{"<b>bold</b>", "&lt;b&gt;bold&lt;/b&gt;"},
		{"a & b", "a &amp; b"},
		{"<&>", "&lt;&amp;&gt;"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := escapeHTML(tt.in); got != tt.want {
				t.Errorf("escapeHTML(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBuildCancelMarkup(t *testing.T) {
	markup := buildCancelMarkup("abc123")
	if len(markup.InlineKeyboard) != 1 || len(markup.InlineKeyboard[0]) != 1 {
		t.Fatalf("expected exactly one button row with one button, got %v", markup.InlineKeyboard)
	}
	btn := markup.InlineKeyboard[0][0]
	if btn.Text != "❌ Cancelar" {
		t.Errorf("button text = %q, want %q", btn.Text, "❌ Cancelar")
	}
	if btn.CallbackData == nil || *btn.CallbackData != "torrent:cancel:abc123" {
		t.Errorf("callback data = %v, want %q", btn.CallbackData, "torrent:cancel:abc123")
	}
}

func TestBuildInlineKeyboard(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		if markup := buildInlineKeyboard(nil); len(markup.InlineKeyboard) != 0 {
			t.Errorf("expected no rows, got %v", markup.InlineKeyboard)
		}
	})
	t.Run("one row per torrent", func(t *testing.T) {
		torrents := []qbt.TorrentInfo{
			{Name: "First Movie", Hash: "aaa"},
			{Name: "Second Very Long Movie Name", Hash: "bbb"},
		}
		markup := buildInlineKeyboard(torrents)
		if len(markup.InlineKeyboard) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(markup.InlineKeyboard))
		}
		first := markup.InlineKeyboard[0][0]
		if want := "🔍 Ver #1 (First Movie)"; first.Text != want {
			t.Errorf("first button text = %q, want %q", first.Text, want)
		}
		if first.CallbackData == nil || *first.CallbackData != "torrent:info:aaa" {
			t.Errorf("first callback data = %v, want %q", first.CallbackData, "torrent:info:aaa")
		}
		second := markup.InlineKeyboard[1][0]
		if want := "🔍 Ver #2 (Second Very ...)"; second.Text != want {
			t.Errorf("second button text = %q, want %q", second.Text, want)
		}
	})
}

func TestBuildTorrentProgress(t *testing.T) {
	torrent := qbt.TorrentInfo{
		Name:      "Ubuntu <24.04> & More",
		Progress:  0.5,
		TotalSize: 1048576,
		Dlspeed:   1024,
		Upspeed:   512,
		Eta:       3600,
		NumSeeds:  5,
		NumLeechs: 3,
		SavePath:  "/downloads/ubuntu",
		Hash:      "abc",
	}
	msg := BuildTorrentProgress(123, torrent)
	if msg.ChatID != 123 {
		t.Errorf("chat id = %d, want 123", msg.ChatID)
	}
	if msg.ParseMode != tgbotapi.ModeHTML {
		t.Errorf("parse mode = %q, want %q", msg.ParseMode, tgbotapi.ModeHTML)
	}
	for _, want := range []string{
		"<b>📌 Ubuntu &lt;24.04&gt; &amp; More</b>",
		"<code>[█████░░░░░]</code> 50.0%",
		"512.0KB / 1.0MB",
		"⬇️ 1.0KB/s | ⬆️ 512 B/s",
		"1h 0m",
		"🌱 5 | 👤 3",
		"<code>/downloads/ubuntu</code>",
	} {
		if !strings.Contains(msg.Text, want) {
			t.Errorf("message missing %q in:\n%s", want, msg.Text)
		}
	}
	markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
	if !ok {
		t.Fatalf("ReplyMarkup is %T, want InlineKeyboardMarkup", msg.ReplyMarkup)
	}
	btn := markup.InlineKeyboard[0][0]
	if btn.CallbackData == nil || *btn.CallbackData != "torrent:cancel:abc" {
		t.Errorf("callback data = %v, want %q", btn.CallbackData, "torrent:cancel:abc")
	}
}

func TestBuildTorrentInfo(t *testing.T) {
	torrent := qbt.TorrentInfo{
		Name:      "Some <File>",
		State:     "downloading",
		Progress:  0.75,
		Size:      1048576,
		Dlspeed:   2048,
		Upspeed:   0,
		NumSeeds:  10,
		NumLeechs: 2,
		Eta:       3600,
		Hash:      "hash1",
	}
	msg := BuildTorrentInfo(99, torrent)
	if msg.ChatID != 99 || msg.ParseMode != tgbotapi.ModeHTML {
		t.Errorf("chat/parse = %d/%q, want 99/%q", msg.ChatID, msg.ParseMode, tgbotapi.ModeHTML)
	}
	for _, want := range []string{
		"📌 <b>Some &lt;File&gt;</b>",
		"<b>State:</b> ⏬ Downloading",
		"<code>[███████░░░]</code> 75.0%",
		"1.0MB",
		"⬇️ <code>2.0KB/s</code>",
		"10 / 2",
	} {
		if !strings.Contains(msg.Text, want) {
			t.Errorf("message missing %q in:\n%s", want, msg.Text)
		}
	}
	markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
	if !ok {
		t.Fatalf("ReplyMarkup is %T, want InlineKeyboardMarkup", msg.ReplyMarkup)
	}
	seen := map[string]bool{}
	for _, row := range markup.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData != nil {
				seen[*btn.CallbackData] = true
			}
		}
	}
	for _, want := range []string{"torrent:cancel:hash1", "torrent:refresh:hash1"} {
		if !seen[want] {
			t.Errorf("keyboard missing button %q, got %v", want, seen)
		}
	}
}

func TestBuildTorrentList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		msg := BuildTorrentList(1, nil)
		if msg.ChatID != 1 {
			t.Errorf("chat id = %d, want 1", msg.ChatID)
		}
		if !strings.Contains(msg.Text, "<pre>") || !strings.Contains(msg.Text, "</pre>") {
			t.Errorf("expected pre block, got:\n%s", msg.Text)
		}
	})
	t.Run("single torrent row and keyboard", func(t *testing.T) {
		torrents := []qbt.TorrentInfo{
			{
				Name:     "movie.iso",
				Size:     1073741824,
				State:    "downloading",
				Progress: 0.5,
				Dlspeed:  102400,
				Eta:      7200,
				Hash:     "h1",
			},
		}
		msg := BuildTorrentList(1, torrents)
		for _, want := range []string{"movie.iso", "1.0GB", "⏬ Downloading", "100.0KB/s", "2h 0m"} {
			if !strings.Contains(msg.Text, want) {
				t.Errorf("message missing %q in:\n%s", want, msg.Text)
			}
		}
		markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
		if !ok {
			t.Fatalf("ReplyMarkup is %T, want InlineKeyboardMarkup", msg.ReplyMarkup)
		}
		btn := markup.InlineKeyboard[0][0]
		if btn.CallbackData == nil || *btn.CallbackData != "torrent:info:h1" {
			t.Errorf("callback data = %v, want %q", btn.CallbackData, "torrent:info:h1")
		}
	})
}

func TestFormatRow(t *testing.T) {
	row := formatRow("SomeFile.tar.gz", 1048576, "downloading", 0.5, 1024, 3600)
	for _, want := range []string{"SomeFile.tar.gz", "1.0MB", "⏬ Downloading", "1.0KB/s", "1h 0m"} {
		if !strings.Contains(row, want) {
			t.Errorf("row missing %q in:\n%s", want, row)
		}
	}
}

func TestFormatGopeedStatus(t *testing.T) {
	tests := []struct {
		in   gopeed.GopeedStatus
		want string
	}{
		{gopeed.GopeedStatusReady, "🟢 Ready"},
		{gopeed.GopeedStatusRunning, "⏬ Running"},
		{gopeed.GopeedStatusPause, "⏸️ Paused"},
		{gopeed.GopeedStatusWait, "⏳ Waiting"},
		{gopeed.GopeedStatusError, "❌ Error"},
		{gopeed.GopeedStatusDone, "✅ Done"},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		t.Run(string(tt.in), func(t *testing.T) {
			if got := formatGopeedStatus(tt.in); got != tt.want {
				t.Errorf("formatGopeedStatus(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBuildDirectDownloadKeyboard(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		if markup := buildDirectDownloadKeyboard(nil); len(markup.InlineKeyboard) != 0 {
			t.Errorf("expected no rows, got %v", markup.InlineKeyboard)
		}
	})
	t.Run("two buttons per task", func(t *testing.T) {
		tasks := []gopeed.GopeedTask{
			{ID: "task1", Name: "Ubuntu ISO"},
			{ID: "task2", Name: "Very Long Direct Download File Name"},
		}
		markup := buildDirectDownloadKeyboard(tasks)
		if len(markup.InlineKeyboard) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(markup.InlineKeyboard))
		}
		if len(markup.InlineKeyboard[0]) != 2 {
			t.Fatalf("expected 2 buttons in first row, got %d", len(markup.InlineKeyboard[0]))
		}
		infoBtn := markup.InlineKeyboard[0][0]
		if want := "ℹ️ Info"; infoBtn.Text != want {
			t.Errorf("info button text = %q, want %q", infoBtn.Text, want)
		}
		if infoBtn.CallbackData == nil || *infoBtn.CallbackData != "dd:info:task1" {
			t.Errorf("info callback data = %v, want %q", infoBtn.CallbackData, "dd:info:task1")
		}
		deleteBtn := markup.InlineKeyboard[0][1]
		if want := "🗑 Eliminar Ubuntu ISO"; deleteBtn.Text != want {
			t.Errorf("delete button text = %q, want %q", deleteBtn.Text, want)
		}
		if deleteBtn.CallbackData == nil || *deleteBtn.CallbackData != "dd:cancel:task1" {
			t.Errorf("delete callback data = %v, want %q", deleteBtn.CallbackData, "dd:cancel:task1")
		}
		secondDelete := markup.InlineKeyboard[1][1]
		if want := "🗑 Eliminar Very Long Direct ..."; secondDelete.Text != want {
			t.Errorf("second delete button text = %q, want %q", secondDelete.Text, want)
		}
	})
}

func TestBuildDirectDownloads(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		msg := BuildDirectDownloads(1, nil)
		if msg.ChatID != 1 {
			t.Errorf("chat id = %d, want 1", msg.ChatID)
		}
		if msg.ParseMode != tgbotapi.ModeHTML {
			t.Errorf("parse mode = %q, want %q", msg.ParseMode, tgbotapi.ModeHTML)
		}
		if !strings.Contains(msg.Text, "No direct downloads found") {
			t.Errorf("expected empty list message, got:\n%s", msg.Text)
		}
		if msg.ReplyMarkup != nil {
			t.Errorf("expected no reply markup for empty list, got %T", msg.ReplyMarkup)
		}
	})
	t.Run("single task", func(t *testing.T) {
		task := gopeed.GopeedTask{
			ID:   "t1",
			Name: "movie.zip",
			Meta: gopeed.GopeedMeta{
				Req: gopeed.GopeedRequest{URL: "https://example.com/movie.zip"},
			},
			Status: gopeed.GopeedStatusRunning,
			Progress: gopeed.GopeedProgress{
				Downloaded: 1073741824,
				Speed:      1048576,
			},
			Size: 2147483648,
		}
		msg := BuildDirectDownloads(42, []gopeed.GopeedTask{task})
		if msg.ChatID != 42 {
			t.Errorf("chat id = %d, want 42", msg.ChatID)
		}
		if msg.ParseMode != tgbotapi.ModeHTML {
			t.Errorf("parse mode = %q, want %q", msg.ParseMode, tgbotapi.ModeHTML)
		}
		for _, want := range []string{
			"<b>⬇️ Direct Downloads</b>",
			"<b>#1 movie.zip</b>",
			"https://example.com/movie.zip",
			"⏬ Running",
			"1.0GB / 2.0GB",
			"1.0MB/s",
		} {
			if !strings.Contains(msg.Text, want) {
				t.Errorf("message missing %q in:\n%s", want, msg.Text)
			}
		}
		markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
		if !ok {
			t.Fatalf("ReplyMarkup is %T, want InlineKeyboardMarkup", msg.ReplyMarkup)
		}
		if len(markup.InlineKeyboard[0]) != 2 {
			t.Fatalf("expected 2 buttons in first row, got %d", len(markup.InlineKeyboard[0]))
		}
		infoBtn := markup.InlineKeyboard[0][0]
		if infoBtn.CallbackData == nil || *infoBtn.CallbackData != "dd:info:t1" {
			t.Errorf("info callback data = %v, want %q", infoBtn.CallbackData, "dd:info:t1")
		}
		deleteBtn := markup.InlineKeyboard[0][1]
		if deleteBtn.CallbackData == nil || *deleteBtn.CallbackData != "dd:cancel:t1" {
			t.Errorf("delete callback data = %v, want %q", deleteBtn.CallbackData, "dd:cancel:t1")
		}
	})
	t.Run("escapes html in name and url", func(t *testing.T) {
		task := gopeed.GopeedTask{
			ID:   "t2",
			Name: "Ubuntu <24.04> & More",
			Meta: gopeed.GopeedMeta{
				Req: gopeed.GopeedRequest{URL: "https://x.com?a=1&b=2"},
			},
			Status: gopeed.GopeedStatusDone,
			Size:   1,
		}
		msg := BuildDirectDownloads(7, []gopeed.GopeedTask{task})
		for _, want := range []string{
			"Ubuntu &lt;24.04&gt; &amp; More",
			"https://x.com?a=1&amp;b=2",
			"✅ Done",
		} {
			if !strings.Contains(msg.Text, want) {
				t.Errorf("message missing %q in:\n%s", want, msg.Text)
			}
		}
	})
}

func TestFormatDirectDownloadTask(t *testing.T) {
	task := gopeed.GopeedTask{
		ID:   "t3",
		Name: "archive.tar.gz",
		Meta: gopeed.GopeedMeta{
			Req: gopeed.GopeedRequest{URL: "https://example.com/archive.tar.gz"},
		},
		Status: gopeed.GopeedStatusRunning,
		Progress: gopeed.GopeedProgress{
			Downloaded: 268435456,
			Speed:      1048576,
		},
		Size: 1073741824,
	}
	row := formatDirectDownloadTask(3, task)
	for _, want := range []string{
		"<b>#3 archive.tar.gz</b>",
		"https://example.com/archive.tar.gz",
		"⏬ Running",
		"256.0MB / 1.0GB",
		"1.0MB/s",
	} {
		if !strings.Contains(row, want) {
			t.Errorf("row missing %q in:\n%s", want, row)
		}
	}
	t.Run("zero size avoids division by zero", func(t *testing.T) {
		task := gopeed.GopeedTask{
			ID:     "t4",
			Name:   "empty",
			Status: gopeed.GopeedStatusReady,
		}
		row := formatDirectDownloadTask(1, task)
		if !strings.Contains(row, "🟢 Ready") {
			t.Errorf("expected Ready status, got:\n%s", row)
		}
	})
}

func TestComputeGopeedProgress(t *testing.T) {
	tests := []struct {
		name string
		task gopeed.GopeedTask
		size int64
		want float64
	}{
		{
			name: "half done",
			task: gopeed.GopeedTask{Size: 100, Progress: gopeed.GopeedProgress{Downloaded: 50}},
			size: 100,
			want: 0.5,
		},
		{
			name: "completed status with zero size",
			task: gopeed.GopeedTask{Status: gopeed.GopeedStatusDone},
			size: 0,
			want: 1.0,
		},
		{
			name: "zero size not done",
			task: gopeed.GopeedTask{Status: gopeed.GopeedStatusWait},
			size: 0,
			want: 0,
		},
		{
			name: "zero downloaded",
			task: gopeed.GopeedTask{Size: 100},
			size: 100,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := computeGopeedProgress(tt.task, tt.size); got != tt.want {
				t.Errorf("computeGopeedProgress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEffectiveTaskSize(t *testing.T) {
	tests := []struct {
		name string
		task gopeed.GopeedTask
		want int64
	}{
		{
			name: "root size has priority",
			task: gopeed.GopeedTask{Size: 100, Meta: gopeed.GopeedMeta{Res: gopeed.GopeedResource{Size: 200}}},
			want: 100,
		},
		{
			name: "fallback to resource size",
			task: gopeed.GopeedTask{
				Meta: gopeed.GopeedMeta{
					Res: gopeed.GopeedResource{Size: 1597014016},
				},
			},
			want: 1597014016,
		},
		{
			name: "fallback to sum of files",
			task: gopeed.GopeedTask{
				Meta: gopeed.GopeedMeta{
					Res: gopeed.GopeedResource{
						Files: []gopeed.GopeedFileInfo{
							{Size: 500},
							{Size: 300},
						},
					},
				},
			},
			want: 800,
		},
		{
			name: "no size available",
			task: gopeed.GopeedTask{},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := effectiveTaskSize(tt.task); got != tt.want {
				t.Errorf("effectiveTaskSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildDirectDownloadProgressUsesResourceSizeFallback(t *testing.T) {
	task := gopeed.GopeedTask{
		ID:       "dd-fallback",
		Protocol: "http",
		Name:     "archlinux.iso",
		Status:   gopeed.GopeedStatusDone,
		Size:     0,
		Progress: gopeed.GopeedProgress{
			Downloaded: 1597014016,
		},
		Meta: gopeed.GopeedMeta{
			Res: gopeed.GopeedResource{Size: 1597014016},
		},
	}
	msg := BuildDirectDownloadProgress(1, task)
	for _, want := range []string{
		"<code>[██████████]</code> 100.0%",
		"1.5GB / 1.5GB",
		"✅ Done",
	} {
		if !strings.Contains(msg.Text, want) {
			t.Errorf("message missing %q in:\n%s", want, msg.Text)
		}
	}
}

func TestBuildDirectDownloadCancelMarkup(t *testing.T) {
	markup := buildDirectDownloadCancelMarkup("task-abc")
	if len(markup.InlineKeyboard) != 1 || len(markup.InlineKeyboard[0]) != 1 {
		t.Fatalf("expected exactly one button row with one button, got %v", markup.InlineKeyboard)
	}
	btn := markup.InlineKeyboard[0][0]
	if btn.Text != "❌ Cancelar" {
		t.Errorf("button text = %q, want %q", btn.Text, "❌ Cancelar")
	}
	if btn.CallbackData == nil || *btn.CallbackData != "dd:cancel:task-abc" {
		t.Errorf("callback data = %v, want %q", btn.CallbackData, "dd:cancel:task-abc")
	}
}

func TestBuildDirectDownloadProgress(t *testing.T) {
	task := gopeed.GopeedTask{
		ID:       "dd-1",
		Protocol: "http",
		Name:     "Movie < Special > & Edition",
		Status:   gopeed.GopeedStatusRunning,
		Size:     2147483648,
		Progress: gopeed.GopeedProgress{
			Downloaded: 1073741824,
			Speed:      2097152,
		},
		Meta: gopeed.GopeedMeta{
			Req: gopeed.GopeedRequest{URL: "https://example.com/movie?x=1&y=2"},
		},
	}
	msg := BuildDirectDownloadProgress(123, task)
	if msg.ChatID != 123 {
		t.Errorf("chat id = %d, want 123", msg.ChatID)
	}
	if msg.ParseMode != tgbotapi.ModeHTML {
		t.Errorf("parse mode = %q, want %q", msg.ParseMode, tgbotapi.ModeHTML)
	}
	for _, want := range []string{
		"<b>📌 Movie &lt; Special &gt; &amp; Edition</b>",
		"<b>Status:</b> ⏬ Running",
		"<code>[█████░░░░░]</code> 50.0%",
		"1.0GB / 2.0GB",
		"⬇️ <code>2.0MB/s</code>",
		"<b>Protocol:</b> <code>http</code>",
		"https://example.com/movie?x=1&amp;y=2",
	} {
		if !strings.Contains(msg.Text, want) {
			t.Errorf("message missing %q in:\n%s", want, msg.Text)
		}
	}
	markup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
	if !ok {
		t.Fatalf("ReplyMarkup is %T, want InlineKeyboardMarkup", msg.ReplyMarkup)
	}
	btn := markup.InlineKeyboard[0][0]
	if btn.CallbackData == nil || *btn.CallbackData != "dd:cancel:dd-1" {
		t.Errorf("callback data = %v, want %q", btn.CallbackData, "dd:cancel:dd-1")
	}
}

func TestBuildDirectDownloadProgressNoUrl(t *testing.T) {
	task := gopeed.GopeedTask{
		ID:       "dd-2",
		Protocol: "https",
		Name:     "Doc",
		Status:   gopeed.GopeedStatusDone,
		Size:     1048576,
		Progress: gopeed.GopeedProgress{
			Downloaded: 1048576,
			Speed:      0,
		},
	}
	msg := BuildDirectDownloadProgress(7, task)
	if !strings.Contains(msg.Text, "✅ Done") {
		t.Errorf("expected Done status, got:\n%s", msg.Text)
	}
	if strings.Contains(msg.Text, "<b>URL:</b>") {
		t.Errorf("should not contain URL label when URL is missing, got:\n%s", msg.Text)
	}
}
