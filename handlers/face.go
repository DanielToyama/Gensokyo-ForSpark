package handlers

// Modified by DanielToyama on 2026-08-20 (Gensokyo-ForSpark fork)
//
// face 段出站转换: QQ 官方 bot 发系统表情的正确姿势是在 content 里内嵌
// <emoji:ID> 标签(仅支持 type=1 系统表情), 参考官方文档:
// https://bot.q.qq.com/wiki/develop/api-v2/openapi/emoji/model.html
// (botgo dto.Emoji(id) = "<emoji:%d>")。
//
// onebot v11 的 face 段 id 与 QQ 官方 type=1 系统表情编号一致(1=微笑, 2=撇嘴...),
// 因此出站时把 face id 直接转成 <emoji:id> 文本拼进 content 即可, QQ 客户端
// 会渲染为原生系统表情。此前实现(Unicode emoji 映射)方向不对, 已废弃。
//
// 注意: 动态表情/魔法表情等非 type=1 系统表情官方通道不支持, 非纯数字 id
// 回退为可读占位 "[表情:<id>]", 保证消息不丢内容、不中断后续段。

import (
	"regexp"
)

var reDigitsOnly = regexp.MustCompile(`^\d+$`)

func faceIDToEmoji(id string) string {
	if id == "" {
		return ""
	}
	if !reDigitsOnly.MatchString(id) {
		return "[表情:" + id + "]"
	}
	return "<emoji:" + id + ">"
}
