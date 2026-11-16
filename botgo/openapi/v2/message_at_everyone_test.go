package v2

import (
	"errors"
	"strings"
	"testing"

	"github.com/tencent-connect/botgo/dto"
)

func TestStripAtEveryone(t *testing.T) {
	original := &dto.MessageToCreate{Content: "维护通知 <qqbot-at-everyone /> 10分钟后重启"}
	stripped, ok := stripAtEveryone(original)
	if !ok {
		t.Fatal("expected stripAtEveryone to change message")
	}
	got := stripped.(*dto.MessageToCreate)
	if strings.Contains(got.Content, "qqbot-at-everyone") {
		t.Fatalf("fallback content still contains tag: %q", got.Content)
	}
	if !strings.Contains(got.Content, "维护通知") || !strings.Contains(got.Content, "10分钟后重启") {
		t.Fatalf("fallback content lost text: %q", got.Content)
	}
	// 不能修改原对象, 调用方重试/队列仍可能需要原消息
	if !strings.Contains(original.Content, "qqbot-at-everyone") {
		t.Fatalf("stripAtEveryone mutated original: %q", original.Content)
	}
}

func TestStripAtEveryoneFromMarkdown(t *testing.T) {
	original := &dto.MessageToCreate{
		Markdown: &dto.Markdown{Content: "内容 <qqbot-at-everyone /> 结束"},
	}
	stripped, ok := stripAtEveryone(original)
	if !ok {
		t.Fatal("expected stripAtEveryone to change markdown message")
	}
	got := stripped.(*dto.MessageToCreate)
	if got.Markdown == nil || strings.Contains(got.Markdown.Content, "qqbot-at-everyone") {
		t.Fatalf("markdown fallback still contains tag: %+v", got.Markdown)
	}
	if original.Markdown.Content == got.Markdown.Content {
		t.Fatal("expected markdown content to be copied and changed")
	}
}

func TestIsTimeoutError(t *testing.T) {
	if !isTimeoutError(errors.New("Post \"https://x\": context deadline exceeded")) {
		t.Fatal("deadline error should be treated as timeout")
	}
	if isTimeoutError(errors.New(`{"code":40054,"message":"no permission"}`)) {
		t.Fatal("permission error should not be treated as timeout")
	}
}
