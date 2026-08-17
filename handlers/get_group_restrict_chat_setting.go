package handlers

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"context"
	"encoding/json"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("get_group_restrict_chat_setting", HandleGetGroupRestrictChatSetting)
}

// HandleGetGroupRestrictChatSetting 查询群禁言状态 (拓展API)
// 参数: group_id
// 官方接口: GET /v2/groups/{group_openid}/restrict_chat_setting
// 返回: global_rule(全员禁言模式) + members(成员级禁言列表)
// 说明: 机器人需拥有群管理员身份
func HandleGetGroupRestrictChatSetting(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	realGroupID, err := resolveGroupOpenID(message.Params.GroupID.(string))
	if err != nil || realGroupID == "" {
		mylog.Printf("get_group_restrict_chat_setting: 无法反查群openid: %v", err)
		return "", nil
	}

	setting, err := apiv2.GetRestrictChatSetting(context.TODO(), realGroupID)
	if err != nil {
		mylog.Printf("get_group_restrict_chat_setting: 查询群禁言状态失败: %v", err)
		return "", nil
	}

	resp := groupResponse(setting, message)
	if err := client.SendMessage(resp); err != nil {
		mylog.Printf("get_group_restrict_chat_setting: 发送响应失败: %v", err)
	}
	result, err := json.Marshal(resp)
	if err != nil {
		return "", nil
	}
	return string(result), nil
}