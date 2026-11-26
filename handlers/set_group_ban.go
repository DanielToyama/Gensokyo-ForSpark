package handlers

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)
//
// OneBot v11 标准动作 set_group_ban（同时注册历史别名 get_group_ban）。
// 群聊走官方 v2 restrict_chat_setting；频道走原 Guild MemberMute。

import (
	"context"
	"strconv"
	"time"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	// onebot v11 标准动作名分别是 set_group_ban / set_group_whole_ban
	// 早期版本误注册为 get_group_ban / get_group_whole_ban, 这里同时注册以兼容历史调用
	callapi.RegisterHandler("set_group_ban", SetGroupBan)
	callapi.RegisterHandler("get_group_ban", SetGroupBan)
}

func SetGroupBan(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	groupID := paramStr(message.Params.GroupID)
	receivedUserID := paramStr(message.Params.UserID)
	if groupID == "" || receivedUserID == "" {
		callapi.SendAPIError(client, message, "set_group_ban: 缺少 group_id 或 user_id")
		return "", nil
	}

	// 读取消息类型; 未写入时默认按群处理, 保证标准调用不会静默失败
	msgType, err := idmap.ReadConfigv2(groupID, "type")
	if err != nil {
		mylog.Printf("setGroupBan: 读取群类型失败(%v), 按群处理", err)
		msgType = "group"
	}

	switch msgType {
	case "group":
		// 官方 设置群成员禁言 (POST /v2/groups/{group_openid}/restrict_chat_setting)
		realGroupID, err := resolveGroupOpenID(groupID)
		if err != nil || realGroupID == "" {
			callapi.SendAPIError(client, message, "set_group_ban: 无法反查群 openid")
			return "", nil
		}
		realUserID, err := resolveUserOpenID(receivedUserID)
		if err != nil || realUserID == "" {
			callapi.SendAPIError(client, message, "set_group_ban: 无法反查用户 openid")
			return "", nil
		}

		member := &dto.SetMemberMuteState{MemberOpenID: realUserID}
		duration := message.Params.Duration
		if !message.Params.DurationSet {
			duration = 1800 // OneBot v11 默认禁言 30 分钟
		}
		if duration <= 0 {
			member.Op = "del"
			member.MuteExpireAt = ""
		} else {
			member.Op = "add"
			member.MuteExpireAt = time.Now().Add(time.Duration(duration) * time.Second).Format(time.RFC3339)
		}
		setting := &dto.SetRestrictChatSettingToCreate{
			Members: []*dto.SetMemberMuteState{member},
		}
		if err := apiv2.SetRestrictChatSetting(context.TODO(), realGroupID, setting); err != nil {
			mylog.Printf("setGroupBan(群): 设置群成员禁言失败: %v", err)
			callapi.SendAPIError(client, message, "set_group_ban: "+err.Error())
			return "", nil
		}
		mylog.Printf("setGroupBan(群): 禁言成功 group[%v] user[%v] duration[%d]", realGroupID, realUserID, duration)

	case "private":
		callapi.SendAPIError(client, message, "set_group_ban: 私聊虚拟群场景暂未适配禁言")
		return "", nil

	case "guild":
		// 读取 ini 通过 ChannelID 取回之前储存的 guild_id (仅频道场景需要)
		guildID, err := idmap.ReadConfigv2(groupID, "guild_id")
		if err != nil {
			callapi.SendAPIError(client, message, "set_group_ban: 读取频道 guild_id 失败")
			return "", nil
		}
		realUserID, err := idmap.RetrieveRowByIDv2(receivedUserID)
		if err != nil {
			callapi.SendAPIError(client, message, "set_group_ban: 无法反查频道用户")
			return "", nil
		}
		duration := strconv.Itoa(message.Params.Duration)
		mute := &dto.UpdateGuildMute{
			MuteSeconds: duration,
			UserIDs:     []string{realUserID},
		}
		if err := api.MemberMute(context.TODO(), guildID, realUserID, mute); err != nil {
			mylog.Printf("Error muting member: %v", err)
			callapi.SendAPIError(client, message, "set_group_ban: "+err.Error())
			return "", nil
		}

	default:
		callapi.SendAPIError(client, message, "set_group_ban: 未知会话类型 "+msgType)
		return "", nil
	}

	// OneBot v11 标准成功响应 (data: null)
	return sendGroupMgmtResponse(client, nil, message)
}
