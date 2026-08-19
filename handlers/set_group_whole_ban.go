package handlers

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"context"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	// onebot v11 标准动作名, 同时注册旧名称兼容
	callapi.RegisterHandler("set_group_whole_ban", SetGroupWholeBan)
	callapi.RegisterHandler("get_group_whole_ban", SetGroupWholeBan)
}

func SetGroupWholeBan(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	// 从message中获取group_id
	groupID := message.Params.GroupID.(string)
	// 读取消息类型
	msgType, err := idmap.ReadConfigv2(groupID, "type")
	if err != nil {
		mylog.Printf("Error reading config for message type: %v", err)
		return "", nil
	}

	// 根据消息类型进行操作
	switch msgType {
	case "group":
		// [新增] 群聊管理: 全员禁言开关
		// 官方 设置群禁言 (POST /v2/groups/{group_openid}/restrict_chat_setting) 当前文档仅声明
		// members(成员级禁言) 字段, 全员禁言 global_rule 仅在查询接口返回; 此处尝试通过 global_rule
		// 的 mode 字段设置 always/none, 若平台暂不支持将返回错误(透传到上层日志)。
		realGroupID, err := idmap.RetrieveRowByIDv2(groupID)
		if err != nil || realGroupID == "" {
			mylog.Printf("setGroupWholeBan(群): 无法反查群openid: %v", err)
			return "", nil
		}
		mode := "none"
		if message.Params.Enable {
			mode = "always"
		}
		setting := &dto.SetRestrictChatSettingToCreate{
			GlobalRule: &dto.GlobalMuteRule{Mode: mode},
		}
		if err := apiv2.SetRestrictChatSetting(context.TODO(), realGroupID, setting); err != nil {
			mylog.Printf("setGroupWholeBan(群): 设置全员禁言失败: %v", err)
			return "", nil
		}
		mylog.Printf("setGroupWholeBan(群): 全员禁言设置成功 group[%v] mode[%v]", realGroupID, mode)
	case "private":
		mylog.Printf("setGroupWholeBan(频道): 目前暂未适配私聊虚拟群场景的禁言能力")
		return "", nil
	case "guild":
		// 读取ini 通过ChannelID取回之前储存的guild_id (仅频道场景需要)
		guildID, err := idmap.ReadConfigv2(groupID, "guild_id")
		if err != nil {
			mylog.Printf("Error reading config: %v", err)
			return "", nil
		}
		var duration string
		if message.Params.Enable {
			duration = "604800" // 7天: 60 * 60 * 24 * 7 onebot的全体禁言只有禁言和解开,先尝试7天
		} else {
			duration = "0"
		}

		mute := &dto.UpdateGuildMute{
			MuteSeconds: duration,
		}
		err = api.GuildMute(context.TODO(), guildID, mute)
		if err != nil {
			mylog.Printf("Error setting whole guild mute: %v", err)
		}
		return "", nil
	}
	return "", nil
}