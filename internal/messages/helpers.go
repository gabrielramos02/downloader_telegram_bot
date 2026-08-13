package messages

import (
	"fmt"
	"html"
	"strings"
)

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func escapeHTML(text string) string {
	r := strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	return r.Replace(text)
}

func formatURLLink(url string) string {
	escaped := html.EscapeString(url)
	if strings.HasPrefix(strings.ToLower(url), "magnet:") {
		return fmt.Sprintf(`<code>%s</code>`, escaped)

	}
	return fmt.Sprintf(`<a href="%s">Link</a>`, escaped)
}

func buildProgressBar(progress float64) string {
	filled := int(progress * 10)
	if filled > 10 {
		filled = 10
	} else if filled < 0 {
		filled = 0
	}
	return fmt.Sprintf("<code>[%s%s]</code> %.1f%%",
		strings.Repeat("█", filled),
		strings.Repeat("░", 10-filled),
		progress*100,
	)
}

func truncateFilename(name string, maxLen int) string {
	runes := []rune(name)
	if len(runes) <= maxLen {
		return name
	}
	return string(runes[:maxLen-3]) + "..."
}
