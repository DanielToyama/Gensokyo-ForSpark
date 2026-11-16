package handlers

import (
	"strings"
	"testing"
)

func TestReplaceUrlDots(t *testing.T) {
	in := "访问 https://example.com/a.b?x=1.2 或 example.org"
	got := replaceUrlDots(in)
	if strings.Contains(got, "example.com") || strings.Contains(got, "example.org") {
		t.Fatalf("URL dots not replaced: %q", got)
	}
	if !strings.Contains(got, "example·com") || !strings.Contains(got, "example·org") {
		t.Fatalf("unexpected replacement: %q", got)
	}
	// 非 URL 的普通句点不受影响
	plain := "版本 1.2.3，请看说明。"
	if got := replaceUrlDots(plain); got != plain {
		t.Fatalf("plain text changed: %q -> %q", plain, got)
	}
}
