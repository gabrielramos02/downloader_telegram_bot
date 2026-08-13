package messages

import (
	"fmt"
	"strings"
	"testing"

	gopeed "github.com/gabrielramos02/gopeed-api-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func TestFormatURLLink(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain url", "https://example.com/file.iso",
			`<a href="https://example.com/file.iso">Link</a>`},
		{"query with ampersand", "https://x.com?a=1&b=2",
			`<a href="https://x.com?a=1&amp;b=2">Link</a>`},
		{"url with html chars", "https://x.com/<a>&b",
			`<a href="https://x.com/&lt;a&gt;&amp;b">Link</a>`},
		{"magnet wrapped in code", "magnet:?xt=urn:btih:abc&dn=file",
			`<code>magnet:?xt=urn:btih:abc&amp;dn=file</code>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatURLLink(tt.in); got != tt.want {
				t.Errorf("formatURLLink(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
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
			{ID: "task1", Name: "Ubuntu ISO", Status: gopeed.GopeedStatusRunning},
			{ID: "task2", Name: "Very Long Direct Download File Name", Status: gopeed.GopeedStatusDone},
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
		cancelBtn := markup.InlineKeyboard[0][1]
		if want := "❌ Cancelar Ubuntu ISO"; cancelBtn.Text != want {
			t.Errorf("cancel button text = %q, want %q", cancelBtn.Text, want)
		}
		if cancelBtn.CallbackData == nil || *cancelBtn.CallbackData != "dd:cancel:task1" {
			t.Errorf("cancel callback data = %v, want %q", cancelBtn.CallbackData, "dd:cancel:task1")
		}
		deleteBtn := markup.InlineKeyboard[1][1]
		if want := "🗑 Eliminar Very Long Direct ..."; deleteBtn.Text != want {
			t.Errorf("delete button text = %q, want %q", deleteBtn.Text, want)
		}
		if deleteBtn.CallbackData == nil || *deleteBtn.CallbackData != "dd:delete:task2" {
			t.Errorf("delete callback data = %v, want %q", deleteBtn.CallbackData, "dd:delete:task2")
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

func TestBuildDirectDownloadInfoMarkup(t *testing.T) {
	t.Run("running shows pause then cancel", func(t *testing.T) {
		markup := buildDirectDownloadInfoMarkup(gopeed.GopeedTask{
			ID:     "t1",
			Status: gopeed.GopeedStatusRunning,
		})
		if len(markup.InlineKeyboard) != 1 || len(markup.InlineKeyboard[0]) != 2 {
			t.Fatalf("expected one row with two buttons, got %v", markup.InlineKeyboard)
		}
		left := markup.InlineKeyboard[0][0]
		if left.Text != "⏸️ Pausar" {
			t.Errorf("left button text = %q, want %q", left.Text, "⏸️ Pausar")
		}
		if left.CallbackData == nil || *left.CallbackData != "dd:pause:t1" {
			t.Errorf("left callback data = %v, want %q", left.CallbackData, "dd:pause:t1")
		}
		right := markup.InlineKeyboard[0][1]
		if right.CallbackData == nil || *right.CallbackData != "dd:cancel:t1" {
			t.Errorf("right callback data = %v, want %q", right.CallbackData, "dd:cancel:t1")
		}
	})
	t.Run("paused shows continue then cancel", func(t *testing.T) {
		markup := buildDirectDownloadInfoMarkup(gopeed.GopeedTask{
			ID:     "t2",
			Status: gopeed.GopeedStatusPause,
		})
		if len(markup.InlineKeyboard) != 1 || len(markup.InlineKeyboard[0]) != 2 {
			t.Fatalf("expected one row with two buttons, got %v", markup.InlineKeyboard)
		}
		left := markup.InlineKeyboard[0][0]
		if left.Text != "▶️ Continuar" {
			t.Errorf("left button text = %q, want %q", left.Text, "▶️ Continuar")
		}
		if left.CallbackData == nil || *left.CallbackData != "dd:continue:t2" {
			t.Errorf("left callback data = %v, want %q", left.CallbackData, "dd:continue:t2")
		}
		right := markup.InlineKeyboard[0][1]
		if right.CallbackData == nil || *right.CallbackData != "dd:cancel:t2" {
			t.Errorf("right callback data = %v, want %q", right.CallbackData, "dd:cancel:t2")
		}
	})
	t.Run("done shows delete", func(t *testing.T) {
		markup := buildDirectDownloadInfoMarkup(gopeed.GopeedTask{
			ID:     "t4",
			Status: gopeed.GopeedStatusDone,
		})
		if len(markup.InlineKeyboard) != 1 || len(markup.InlineKeyboard[0]) != 1 {
			t.Fatalf("expected one row with one button, got %v", markup.InlineKeyboard)
		}
		btn := markup.InlineKeyboard[0][0]
		if btn.Text != "🗑️ Eliminar" {
			t.Errorf("button text = %q, want %q", btn.Text, "🗑️ Eliminar")
		}
		if btn.CallbackData == nil || *btn.CallbackData != "dd:delete:t4" {
			t.Errorf("callback data = %v, want %q", btn.CallbackData, "dd:delete:t4")
		}
	})
	t.Run("other status only shows cancel", func(t *testing.T) {
		for _, status := range []gopeed.GopeedStatus{
			gopeed.GopeedStatusReady,
			gopeed.GopeedStatusWait,
			gopeed.GopeedStatusError,
		} {
			markup := buildDirectDownloadInfoMarkup(gopeed.GopeedTask{ID: "t3", Status: status})
			if len(markup.InlineKeyboard) != 1 || len(markup.InlineKeyboard[0]) != 1 {
				t.Errorf("status %s: expected one row with one button, got %v", status, markup.InlineKeyboard)
				continue
			}
			btn := markup.InlineKeyboard[0][0]
			if btn.CallbackData == nil || *btn.CallbackData != "dd:cancel:t3" {
				t.Errorf("status %s: callback data = %v, want %q", status, btn.CallbackData, "dd:cancel:t3")
			}
		}
	})
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
	if len(markup.InlineKeyboard) != 1 || len(markup.InlineKeyboard[0]) != 2 {
		t.Fatalf("expected one row with two buttons, got %v", markup.InlineKeyboard)
	}
	pauseBtn := markup.InlineKeyboard[0][0]
	if pauseBtn.Text != "⏸️ Pausar" {
		t.Errorf("first button text = %q, want %q", pauseBtn.Text, "⏸️ Pausar")
	}
	if pauseBtn.CallbackData == nil || *pauseBtn.CallbackData != "dd:pause:dd-1" {
		t.Errorf("pause callback data = %v, want %q", pauseBtn.CallbackData, "dd:pause:dd-1")
	}
	cancelBtn := markup.InlineKeyboard[0][1]
	if cancelBtn.Text != "❌ Cancelar" {
		t.Errorf("second button text = %q, want %q", cancelBtn.Text, "❌ Cancelar")
	}
	if cancelBtn.CallbackData == nil || *cancelBtn.CallbackData != "dd:cancel:dd-1" {
		t.Errorf("cancel callback data = %v, want %q", cancelBtn.CallbackData, "dd:cancel:dd-1")
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
