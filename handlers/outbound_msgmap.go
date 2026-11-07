package handlers

// Modified by DanielToyama on 2026-09-02 (Gensokyo-ForSpark fork)
//
// 出站消息入库 —— 让标准 onebot v11 `delete_msg`(只传 message_id)也能撤回
// 机器人自己发送的消息。
//
// 背景: msgmap 原本只记录机器人「收到」的消息(ProcessGroupMessage / ProcessC2CMessage
// / ProcessGuildATMessage 里各埋了一处), 机器人自己发出去的消息从未入表。于是应用端
// 用标准写法 {"action":"delete_msg","params":{"message_id":X}} 撤回自己的消息时,
// delete_msg.go 的 msgmap 反查落空 -> 四个撤回分支全跳过 -> 响应依然 {"status":"ok"},
// 应用端以为撤回成功, 实际什么都没发生(静默失败)。
//
// 这里在「发送成功回执」阶段把 message_id -> 会话对象写进 msgmap(Outbound=true),
// 五个回执函数(SendResponse / SendResponseSB / SendGuildResponse / SendC2CResponse /
// SendGuildPrivateResponse)各调一次, 覆盖 群 / C2C / 频道 / 频道私信 全部发送入口
// (send_msg、*_async、*_raw、forward、sse、wakeup 最终都汇到这几个函数)。

import (
	"strconv"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/config"
	"github.com/hoshinonyaruko/gensokyo/msgmap"
)

// paramStr 取出 ParamsContent 里的字符串参数(反序列化时已统一转成 string, 这里再兜一层 nil)
func paramStr(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// recordOutboundMsg 把刚发送成功的消息记入 msgmap(出站标记), 供 delete_msg 反查撤回对象
//
//	message        本次发送请求(取会话 id: group_id / user_id / channel_id / guild_id)
//	officialID     官方真实 message_id(用于 string_ob11 模式, 应用端拿到的就是它)
//	virtualID      应用端可见的压缩行 id(非 string_ob11 模式, 应用端拿到的就是它; 0 表示无)
//	guildIDOverride 频道私信的真实 guild id(仅 SendGuildPrivateResponse 传, 其余传 "")
func recordOutboundMsg(message *callapi.ActionMessage, officialID string, virtualID int64, guildIDOverride string) {
	if message == nil {
		return
	}

	keys := make([]string, 0, 2)
	if officialID != "" {
		keys = append(keys, officialID)
	}
	if virtualID != 0 {
		keys = append(keys, strconv.FormatInt(virtualID, 10))
	}
	if len(keys) == 0 {
		return
	}

	groupID := paramStr(message.Params.GroupID)
	userID := paramStr(message.Params.UserID)
	channelID := paramStr(message.Params.ChannelID)
	guildID := paramStr(message.Params.GuildID)
	if guildIDOverride != "" {
		guildID = guildIDOverride
	}

	info := msgmap.MsgInfo{}
	// 昵称记成应用端可见的机器人名(get_login_info 返回的也是它), 这样应用端引用机器人
	// 自己的消息时, fakeReply 能伪造出「回复 @机器人名」而不是「回复 @未知用户」
	info.Nickname = config.GetCustomBotName()
	switch {
	case groupID != "":
		// 群聊: 反查只需 group_id
		info.MsgType = "group"
		info.GroupID = groupID
	case channelID != "":
		// 频道子频道: 反查只需 channel_id(记为 GroupID, 与入站频道记录的字段位置一致)
		info.MsgType = "guild"
		info.GroupID = channelID
		info.UserID = userID
	case guildID != "":
		// 频道私信: 反查只需 guild_id
		info.MsgType = "guild_private"
		info.GroupID = guildID
		info.UserID = userID
	case userID != "":
		// C2C 好友私聊: 反查只需 user_id(与入站私聊记录一致, GroupID 也填 user_id)
		info.MsgType = "group_private"
		info.GroupID = userID
		info.UserID = userID
	default:
		// 拿不到任何会话 id, 记了也反查不回来
		return
	}

	msgmap.RecordOutbound(keys, info)
}
