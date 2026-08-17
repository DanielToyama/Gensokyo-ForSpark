package handlers

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("set_group_add_request", HandleSetGroupAddRequest)
}

// HandleSetGroupAddRequest 处理加群请求(入群申请审批)
// onebot v11 标准参数: group_id, user_id, flag(join_request_id), approve/refuse, reason
// 官方接口: POST /v2/groups/{group_openid}/approval_join_request/{member_openid}
// 说明: 机器人需拥有群管理员身份才能审批
func HandleSetGroupAddRequest(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	// group_id/user_id 可能为空: SparkBridge groupRequest 插件审批时只传 flag/approve/reason
	groupID := ""
	if v, ok := message.Params.GroupID.(string); ok {
		groupID = v
	}
	userID := ""
	if v, ok := message.Params.UserID.(string); ok {
		userID = v
	}
	flag := message.Params.Flag

	var realGroupID, realUserID string
	var err error
	if groupID != "" && userID != "" {
		// 常规路径: 用 group_id/user_id 反查 openid
		realGroupID, err = idmap.RetrieveRowByIDv2(groupID)
		if err != nil || realGroupID == "" {
			msg := fmt.Sprintf("set_group_add_request: 无法反查群openid group[%v] err[%v]", groupID, err)
			mylog.Println(msg)
			callapi.SendAPIError(client, message, msg)
			return "", nil
		}
		realUserID, err = idmap.RetrieveRowByIDv2(userID)
		if err != nil || realUserID == "" {
			msg := fmt.Sprintf("set_group_add_request: 无法反查用户openid user[%v] err[%v]", userID, err)
			mylog.Println(msg)
			callapi.SendAPIError(client, message, msg)
			return "", nil
		}
	} else {
		// 兜底路径: 插件只传 flag(join_request_id), 从事件缓存反查 (群,成员) openid
		realGroupID, realUserID = idmap.RetrieveJoinRequestV2(flag)
		if realGroupID == "" || realUserID == "" {
			msg := fmt.Sprintf("set_group_add_request: 缺少 group_id/user_id 且缓存中无此 flag[%v]", flag)
			mylog.Println(msg)
			callapi.SendAPIError(client, message, msg)
			return "", nil
		}
	}

	req := &dto.ApproveJoinRequestToCreate{
		JoinRequestID: flag,
		RejectReason:  message.Params.Reason,
	}
	// onebot v11 语义: approve=true 通过; approve=false 拒绝(拒绝理由走 reason)
	// 兼容只传 approve 的实现(如 SparkBridge groupRequest 插件), 不再要求 refuse 字段
	if message.Params.Approve {
		req.Op = "approve"
	} else {
		req.Op = "decline"
	}

	mylog.Printf("set_group_add_request: 审批申请 group[%v] user[%v] op[%v] flag[%v]", realGroupID, realUserID, req.Op, flag)
	if err := apiv2.ApproveJoinRequest(context.TODO(), realGroupID, realUserID, req); err != nil {
		msg := fmt.Sprintf("set_group_add_request: 审批失败: %v", err)
		mylog.Println(msg)
		callapi.SendAPIError(client, message, msg)
		return "", nil
	}

	t := time.Now()
	response := map[string]interface{}{
		"data":    map[string]interface{}{},
		"message": "success",
		"retcode": 0,
		"status":  "ok",
		"time":    t.Unix(),
	}
	if message.Echo != nil && message.Echo != "" {
		response["echo"] = message.Echo
	}
	if err := client.SendMessage(response); err != nil {
		mylog.Printf("set_group_add_request: 发送响应失败: %v", err)
	}
	result, err := json.Marshal(response)
	if err != nil {
		return "", nil
	}
	return string(result), nil
}
