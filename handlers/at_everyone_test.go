package handlers

import (
	"strings"
	"testing"

	"github.com/tencent-connect/botgo/dto"
)

// [2026-10-26 官方新特性] 群聊 @全体成员: OneBot CQ 码 -> 官方 <qqbot-at-everyone />
func TestTransformMessageTextAtEveryone(t *testing.T) {
	cases := []string{
		"[CQ:at,qq=all]",
		"[CQ:at,qq=ALL]",
		"[CQ:at,qq=everyone]",
	}
	for _, in := range cases {
		if got := transformMessageTextAt(in, "123456"); got != "<qqbot-at-everyone />" {
			t.Fatalf("transformMessageTextAt(%q) = %q, want <qqbot-at-everyone />", in, got)
		}
	}
}

// @全体成员 在纯文本消息中保持文本发送; 与普通 at 混合时升级 markdown
func TestMaybeUpgradeToMarkdownAtEveryone(t *testing.T) {
	textOnly := maybeUpgradeToMarkdownAt(&dto.MessageToCreate{Content: "<qqbot-at-everyone />"})
	if textOnly.MsgType != 0 || textOnly.Markdown != nil {
		t.Fatalf("纯文本 @全体成员 不应升级 markdown: msgType=%v markdown=%v", textOnly.MsgType, textOnly.Markdown)
	}

	mixed := maybeUpgradeToMarkdownAt(&dto.MessageToCreate{
		Content: `<qqbot-at-everyone /> <qqbot-at-user id="abc" />`,
	})
	if mixed.MsgType != 2 || mixed.Markdown == nil {
		t.Fatalf("@全体成员+普通at 应升级 markdown: msgType=%v markdown=%v", mixed.MsgType, mixed.Markdown)
	}
}

// 富媒体消息(图片/语音/视频, MsgType=7 / Media.FileInfo)不支持 at 标签:
// at 必须还原为文本, 不能升级 markdown, 也不能把原始标签发给官方富媒体通道
func TestMaybeUpgradeToMarkdownAtRichMedia(t *testing.T) {
	rich := maybeUpgradeToMarkdownAt(&dto.MessageToCreate{
		Content: `<qqbot-at-user id="0123456789abcdef0123456789abcdef" /> <qqbot-at-everyone />`,
		MsgType: 7,
		Media:   dto.Media{FileInfo: "file-info"},
	})
	if rich.Markdown != nil || rich.MsgType != 7 {
		t.Fatalf("富媒体消息不应升级 markdown: msgType=%v markdown=%v", rich.MsgType, rich.Markdown)
	}
	if strings.Contains(rich.Content, "qqbot-at-") {
		t.Fatalf("富媒体消息仍残留 at 标签: %q", rich.Content)
	}
	if !strings.Contains(rich.Content, "@全体成员") {
		t.Fatalf("富媒体消息的 @全体成员 应还原为文本: %q", rich.Content)
	}
}
