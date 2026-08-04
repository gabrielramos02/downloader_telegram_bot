package main

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/gabrielramos02/telegram-bot-go/internal/loggerErr"
	pkgerr "github.com/pkg/errors"
)

func groupStringAttrs(t *testing.T, a slog.Attr) map[string]string {
	t.Helper()
	m := map[string]string{}
	if a.Value.Kind() != slog.KindGroup {
		return m
	}
	for _, attr := range a.Value.Group() {
		m[attr.Key] = attr.Value.String()
	}
	return m
}

func TestReplaceAttrSensitiveKeys(t *testing.T) {
	for _, key := range sensitiveKeys {
		t.Run(key, func(t *testing.T) {
			got := replaceAttr(nil, slog.String(key, "value"))
			if got.Key != key {
				t.Errorf("key = %q, want %q", got.Key, key)
			}
			if got.Value.String() != "[REDACTED]" {
				t.Errorf("value = %q, want %q", got.Value.String(), "[REDACTED]")
			}
		})
	}
}

func TestReplaceAttr(t *testing.T) {
	t.Run("sensitive key takes precedence over URL redaction", func(t *testing.T) {
		got := replaceAttr(nil, slog.String("user", "https://alice:secret@example.com"))
		if got.Value.String() != "[REDACTED]" {
			t.Errorf("value = %q, want %q", got.Value.String(), "[REDACTED]")
		}
	})
	t.Run("sensitive key is case sensitive", func(t *testing.T) {
		a := slog.String("Password", "value")
		if got := replaceAttr(nil, a); got.Value.String() != "value" {
			t.Errorf("value = %q, want %q", got.Value.String(), "value")
		}
	})
	t.Run("non sensitive key unchanged", func(t *testing.T) {
		a := slog.String("name", "value")
		if got := replaceAttr(nil, a); got.Key != "name" || got.Value.String() != "value" {
			t.Errorf("got (%q, %q), want (name, value)", got.Key, got.Value.String())
		}
	})
	t.Run("URL with password redacted but username kept", func(t *testing.T) {
		got := replaceAttr(nil, slog.String("endpoint", "https://alice:secret@example.com/path"))
		if want := "https://alice:%5BREDACTED%5D@example.com/path"; got.Value.String() != want {
			t.Errorf("value = %q, want %q", got.Value.String(), want)
		}
	})
	t.Run("URL with user but no password unchanged", func(t *testing.T) {
		a := slog.String("endpoint", "https://alice@example.com")
		if got := replaceAttr(nil, a); got.Value.String() != "https://alice@example.com" {
			t.Errorf("value = %q, want unchanged", got.Value.String())
		}
	})
	t.Run("URL without user unchanged", func(t *testing.T) {
		a := slog.String("endpoint", "https://example.com")
		if got := replaceAttr(nil, a); got.Value.String() != "https://example.com" {
			t.Errorf("value = %q, want unchanged", got.Value.String())
		}
	})
	t.Run("unparseable value unchanged", func(t *testing.T) {
		a := slog.String("endpoint", "http://[::1]:badport")
		if got := replaceAttr(nil, a); got.Value.String() != "http://[::1]:badport" {
			t.Errorf("value = %q, want unchanged", got.Value.String())
		}
	})
	t.Run("error key with non-error value unchanged", func(t *testing.T) {
		a := slog.Any("error", "notanerror")
		got := replaceAttr(nil, a)
		if got.Key != "error" || got.Value.Any() != "notanerror" {
			t.Errorf("got (%q, %v), want (error, notanerror)", got.Key, got.Value.Any())
		}
	})
	t.Run("plain error becomes group", func(t *testing.T) {
		got := replaceAttr(nil, slog.Any("error", errors.New("boom")))
		if got.Key != "error" || got.Value.Kind() != slog.KindGroup {
			t.Fatalf("expected group named error, got %+v", got)
		}
		m := groupStringAttrs(t, got)
		if m["message"] != "boom" {
			t.Errorf("message = %q, want boom", m["message"])
		}
		if _, hasStack := m["stack_trace"]; hasStack {
			t.Error("unexpected stack_trace for plain error")
		}
	})
	t.Run("error with loggerErr attrs keeps them", func(t *testing.T) {
		err := loggerErr.WithAttrs(errors.New("boom"), "torrent", "abc")
		got := replaceAttr(nil, slog.Any("error", err))
		m := groupStringAttrs(t, got)
		if m["message"] != "boom" || m["torrent"] != "abc" {
			t.Errorf("group = %v, want message=boom torrent=abc", m)
		}
	})
	t.Run("multiError becomes errors group", func(t *testing.T) {
		err1 := loggerErr.WithAttrs(errors.New("a"), "name", "one")
		err2 := loggerErr.WithAttrs(errors.New("b"), "name", "two")
		got := replaceAttr(nil, slog.Any("error", errors.Join(err1, err2)))
		if got.Key != "errors" || got.Value.Kind() != slog.KindGroup {
			t.Fatalf("expected group named errors, got %+v", got)
		}
		group := map[string]map[string]string{}
		for _, attr := range got.Value.Group() {
			group[attr.Key] = groupStringAttrs(t, attr)
		}
		if group["error_1"]["name"] != "one" || group["error_2"]["name"] != "two" {
			t.Errorf("groups = %v, want error_1.name=one error_2.name=two", group)
		}
	})
}

func TestErrorAttrs(t *testing.T) {
	t.Run("plain error has message only", func(t *testing.T) {
		attrs := errorAttrs(errors.New("boom"))
		if attrs[0].Key != "message" || attrs[0].Value.String() != "boom" {
			t.Errorf("first attr = %+v, want message=boom", attrs[0])
		}
		for _, a := range attrs {
			if a.Key == "stack_trace" {
				t.Error("unexpected stack_trace for plain error")
			}
		}
	})
	t.Run("pkg/errors error includes stack trace", func(t *testing.T) {
		attrs := errorAttrs(pkgerr.New("boom"))
		found := false
		for _, a := range attrs {
			if a.Key == "stack_trace" {
				found = true
				if a.Value.String() == "" {
					t.Error("empty stack_trace")
				}
			}
		}
		if !found {
			t.Errorf("expected stack_trace attr, got %v", attrs)
		}
	})
	t.Run("error with loggerErr attrs", func(t *testing.T) {
		attrs := errorAttrs(loggerErr.WithAttrs(errors.New("boom"), "torrent", "abc"))
		m := map[string]string{}
		for _, a := range attrs {
			m[a.Key] = a.Value.String()
		}
		if m["message"] != "boom" || m["torrent"] != "abc" {
			t.Errorf("attrs = %v, want message=boom torrent=abc", m)
		}
	})
}
