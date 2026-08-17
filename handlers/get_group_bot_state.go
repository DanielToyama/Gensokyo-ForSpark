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
	callapi.RegisterHandler("get_group_bot_state", HandleGetGroupBotState)
}

// HandleGetGroupBotState 获取机器人群内状态 (拓展API)
// 参数: group_id
// 官方接口: GET /v2/groups/{group_openid}/bot_state
// 返回: member_openid, joined_at, allow_proactive_msg, recv_msg_setting, member_role
func HandleGetGroupBotState(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	realGroupID, err := resolveGroupOpenID(message.Params.GroupID.(string))
	if err != nil || realGroupID == "" {
		mylog.Printf("get_group_bot_state: 无法反查群openid: %v", err)
		return "", nil
	}

	state, err := apiv2.GetGroupBotState(context.TODO(), realGroupID)
	if err != nil {
		mylog.Printf("get_group_bot_state: 获取机器人群内状态失败: %v", err)
		return "", nil
	}

	resp := groupResponse(state, message)
	if err := client.SendMessage(resp); err != nil {
		mylog.Printf("get_group_bot_state: 发送响应失败: %v", err)
	}
	result, err := json.Marshal(resp)
	if err != nil {
		return "", nil
	}
	return string(result), nil
}