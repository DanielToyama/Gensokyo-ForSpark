package handlers

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

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

	// 从message中获取group_id和UserID
	groupID := message.Params.GroupID.(string)
	receivedUserID := message.Params.UserID.(string)
	// 根据UserID读取真实的userid
	realUserID, err := idmap.RetrieveRowByIDv2(receivedUserID)
	if err != nil {
		mylog.Printf("Error reading real userID: %v", err)
		return "", nil
	}

	// 读取消息类型
	msgType, err := idmap.ReadConfigv2(groupID, "type")
	if err != nil {
		mylog.Printf("Error reading config for message type: %v", err)
		return "", nil
	}

	// 根据消息类型进行操作
	switch msgType {
	case "group":
		// [新增] 群聊管理: 官方 设置群成员禁言 (POST /v2/groups/{group_openid}/restrict_chat_setting)
		// 说明: 需要机器人拥有群管理员身份; 最大禁言时长为 30 天
		realGroupID, err := idmap.RetrieveRowByIDv2(groupID)
		if err != nil || realGroupID == "" {
			mylog.Printf("setGroupBan(群): 无法反查群openid: %v", err)
			return "", nil
		}
		member := &dto.SetMemberMuteState{
			MemberOpenID: realUserID,
		}
		if message.Params.Duration <= 0 {
			// 解禁
			member.Op = "del"
			member.MuteExpireAt = ""
		} else {
			// 禁言: 到期时间 = 当前时间 + duration 秒
			member.Op = "add"
			member.MuteExpireAt = time.Now().Add(time.Duration(message.Params.Duration) * time.Second).Format(time.RFC3339)
		}
		setting := &dto.SetRestrictChatSettingToCreate{
			Members: []*dto.SetMemberMuteState{member},
		}
		if err := apiv2.SetRestrictChatSetting(context.TODO(), realGroupID, setting); err != nil {
			mylog.Printf("setGroupBan(群): 设置群成员禁言失败: %v", err)
			return "", nil
		}
		mylog.Printf("setGroupBan(群): 禁言成功 group[%v] user[%v] duration[%d]", realGroupID, realUserID, message.Params.Duration)
	case "private":
		mylog.Printf("setGroupBan(频道): 目前暂未适配私聊虚拟群场景的禁言能力")
		return "", nil
	case "guild":
		// 读取ini 通过ChannelID取回之前储存的guild_id (仅频道场景需要)
		guildID, err := idmap.ReadConfigv2(groupID, "guild_id")
		if err != nil {
			mylog.Printf("Error reading config: %v", err)
			return "", nil
		}
		duration := strconv.Itoa(message.Params.Duration)
		mute := &dto.UpdateGuildMute{
			MuteSeconds: duration,
			UserIDs:     []string{realUserID},
		}
		err = api.MemberMute(context.TODO(), guildID, realUserID, mute)
		if err != nil {
			mylog.Printf("Error muting member: %v", err)
		}
	}
	return "", nil
}