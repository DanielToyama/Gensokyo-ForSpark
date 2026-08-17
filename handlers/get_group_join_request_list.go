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
	callapi.RegisterHandler("get_group_join_request_list", HandleGetGroupJoinRequestList)
}

// HandleGetGroupJoinRequestList 拉取入群申请列表 (拓展API)
// 参数: group_id, cursor(分页游标,可选), limit(单页数量,默认20最大100,可选)
// 官方接口: GET /v2/groups/{group_openid}/join_request_list
// 说明: 机器人需拥有群管理员身份
func HandleGetGroupJoinRequestList(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	realGroupID, err := resolveGroupOpenID(message.Params.GroupID.(string))
	if err != nil || realGroupID == "" {
		mylog.Printf("get_group_join_request_list: 无法反查群openid: %v", err)
		return "", nil
	}

	list, err := apiv2.GetJoinRequestList(context.TODO(), realGroupID, message.Params.Cursor, message.Params.Limit)
	if err != nil {
		mylog.Printf("get_group_join_request_list: 拉取入群申请列表失败: %v", err)
		return "", nil
	}

	resp := groupResponse(list, message)
	if err := client.SendMessage(resp); err != nil {
		mylog.Printf("get_group_join_request_list: 发送响应失败: %v", err)
	}
	result, err := json.Marshal(resp)
	if err != nil {
		return "", nil
	}
	return string(result), nil
}