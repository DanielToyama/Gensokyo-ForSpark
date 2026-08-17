package handlers

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	// 原始透传接口: 直接按官方请求体格式设置群禁言(成员级/全员级)
	callapi.RegisterHandler("set_group_restrict_chat_setting", HandleSetGroupRestrictChatSetting)
}

// HandleSetGroupRestrictChatSetting 设置群禁言 (拓展API, 原始透传)
// 参数: group_id, members[{op, member_openid, mute_expire_at}...]
//       或者 enable=true/false 表示全员禁言 always/none (官方文档暂未声明 global_rule 可写, 透传尝试)
// 官方接口: POST /v2/groups/{group_openid}/restrict_chat_setting
// 说明: 机器人需拥有群管理员身份; 成员级禁言单次不能超过10个, 最大禁言时长为 30 天
func HandleSetGroupRestrictChatSetting(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	realGroupID, err := resolveGroupOpenID(message.Params.GroupID.(string))
	if err != nil || realGroupID == "" {
		mylog.Printf("set_group_restrict_chat_setting: 无法反查群openid: %v", err)
		return "", nil
	}

	setting := &dto.SetRestrictChatSettingToCreate{}
	for _, m := range message.Params.Members {
		setting.Members = append(setting.Members, &dto.SetMemberMuteState{
			Op:           m.Op,
			MemberOpenID: m.MemberOpenID,
			MuteExpireAt: m.MuteExpireAt,
		})
	}
	// 未指定 members 时, 视为全员禁言开关操作
	if len(setting.Members) == 0 {
		mode := "none"
		if message.Params.Enable {
			mode = "always"
		}
		setting.GlobalRule = &dto.GlobalMuteRule{Mode: mode}
	}

	if err := apiv2.SetRestrictChatSetting(context.TODO(), realGroupID, setting); err != nil {
		mylog.Printf("set_group_restrict_chat_setting: 设置群禁言失败: %v", err)
		return "", nil
	}

	t := time.Now()
	resp := groupResponse(map[string]interface{}{"time": t.Unix()}, message)
	if err := client.SendMessage(resp); err != nil {
		mylog.Printf("set_group_restrict_chat_setting: 发送响应失败: %v", err)
	}
	result, err := json.Marshal(resp)
	if err != nil {
		return "", nil
	}
	return string(result), nil
}