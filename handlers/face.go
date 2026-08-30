package handlers

// Modified by DanielToyama on 2026-08-31 (Gensokyo-ForSpark fork)
//
// face 段出站转换 —— 最终结论。
//
// 官方 bot 能力表(2026-08-31, https://bot.q.qq.com/wiki/develop/api-v2/):
// 消息类型只有 文本 / Markdown / 图片 / 视频 / 语音 / 文件 / 结构化卡片 / Embed /
// 表情表态 / 引用消息 —— **没有"表情"作为消息内容**这一项(表情表态=给消息点赞,
// 不是发送表情)。因此 onebot v11 的 face 段(数字表情)在官方 bot 通道**从根上
// 不存在**:
//   - <emoji:ID> 纯文本通道实测显示原文;
//   - <emoji:ID> markdown 通道实测同样不渲染;
//   - Unicode emoji 字符方案实测行不通。
//
// 三条路全部证伪后, face 段降级为可读占位 "[表情:<id>]", 保证消息不丢内容、
// 不中断后续段, 也不触发 markdown 升级(避免消息变成气泡样式却仍无表情)。

func faceIDToEmoji(id string) string {
	if id == "" {
		return ""
	}
	return "[表情:" + id + "]"
}
