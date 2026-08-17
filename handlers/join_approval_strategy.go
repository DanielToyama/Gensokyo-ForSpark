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
	// 入群自动审批策略接口 (官方 API v2 20260810 新增)
	// 文档: /v2/groups/join_approval_strategy[/*]
	callapi.RegisterHandler("create_join_approval_strategy", HandleCreateJoinApprovalStrategy)
	callapi.RegisterHandler("get_join_approval_strategy_list", HandleGetJoinApprovalStrategyList)
	callapi.RegisterHandler("update_join_approval_strategy", HandleUpdateJoinApprovalStrategy)
	callapi.RegisterHandler("delete_join_approval_strategy", HandleDeleteJoinApprovalStrategy)
	callapi.RegisterHandler("update_join_approval_strategy_whitelist", HandleUpdateJoinApprovalStrategyWhitelist)
	callapi.RegisterHandler("execute_join_approval_strategy", HandleExecuteJoinApprovalStrategy)
}

// HandleCreateJoinApprovalStrategy 创建入群自动审批策略
// 参数: group_openids/group_ids 二选一, is_enable(on/off,默认on), expire_at(RFC3339,不传默认一年), remark
func HandleCreateJoinApprovalStrategy(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	strategy := &dto.JoinApprovalStrategyToCreate{
		GroupOpenIDs: message.Params.GroupOpenIDs,
		GroupIDs:     message.Params.GroupIDs,
		IsEnable:     message.Params.IsEnable,
		ExpireAt:     message.Params.ExpireAt,
		Remark:       message.Params.Remark,
	}
	result, err := apiv2.CreateJoinApprovalStrategy(context.TODO(), strategy)
	if err != nil {
		mylog.Printf("create_join_approval_strategy: 创建失败: %v", err)
		return "", nil
	}
	return sendGroupMgmtResponse(client, result, message)
}

// HandleGetJoinApprovalStrategyList 查询入群自动审批策略列表
// 参数: cursor(可选), limit(可选,默认20最大100)
func HandleGetJoinApprovalStrategyList(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	list, err := apiv2.GetJoinApprovalStrategyList(context.TODO(), message.Params.Cursor, message.Params.Limit)
	if err != nil {
		mylog.Printf("get_join_approval_strategy_list: 查询列表失败: %v", err)
		return "", nil
	}
	return sendGroupMgmtResponse(client, list, message)
}

// HandleUpdateJoinApprovalStrategy 修改入群自动审批策略
// 参数: strategy_id, is_enable(可选), expire_at(可选), remark(可选),
//       group_action_op(add/del), group_action_group_openids/group_action_group_ids(可选)
func HandleUpdateJoinApprovalStrategy(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	strategy := &dto.UpdateJoinApprovalStrategyToCreate{
		IsEnable: message.Params.IsEnable,
		ExpireAt: message.Params.ExpireAt,
		Remark:   message.Params.Remark,
	}
	if message.Params.GroupActionOp != "" {
		strategy.GroupAction = &dto.GroupAction{
			Op:           message.Params.GroupActionOp,
			GroupOpenIDs: message.Params.GroupActionGroupOpenIDs,
			GroupIDs:     message.Params.GroupActionGroupIDs,
		}
	}
	result, err := apiv2.UpdateJoinApprovalStrategy(context.TODO(), message.Params.StrategyID, strategy)
	if err != nil {
		mylog.Printf("update_join_approval_strategy: 修改失败: %v", err)
		return "", nil
	}
	return sendGroupMgmtResponse(client, result, message)
}

// HandleDeleteJoinApprovalStrategy 删除入群自动审批策略
// 参数: strategy_id
func HandleDeleteJoinApprovalStrategy(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	if err := apiv2.DeleteJoinApprovalStrategy(context.TODO(), message.Params.StrategyID); err != nil {
		mylog.Printf("delete_join_approval_strategy: 删除失败: %v", err)
		return "", nil
	}
	t := time.Now()
	return sendGroupMgmtResponse(client, map[string]interface{}{"time": t.Unix()}, message)
}

// HandleUpdateJoinApprovalStrategyWhitelist 修改入群自动审批策略的白名单号码
// 参数: strategy_id, whitelist_op(add/del), whitelist_users([...] QQ号码)
func HandleUpdateJoinApprovalStrategyWhitelist(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	whitelist := &dto.WhitelistUsersToCreate{
		Op:            message.Params.WhitelistOp,
		WhitelistUsers: message.Params.WhitelistUsers,
	}
	result, err := apiv2.UpdateJoinApprovalStrategyWhitelist(context.TODO(), message.Params.StrategyID, whitelist)
	if err != nil {
		mylog.Printf("update_join_approval_strategy_whitelist: 修改白名单失败: %v", err)
		return "", nil
	}
	return sendGroupMgmtResponse(client, result, message)
}

// HandleExecuteJoinApprovalStrategy 执行入群自动审批策略 (全量扫描, 异步约10分钟)
// 参数: strategy_id
func HandleExecuteJoinApprovalStrategy(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	if err := apiv2.ExecuteJoinApprovalStrategy(context.TODO(), message.Params.StrategyID); err != nil {
		mylog.Printf("execute_join_approval_strategy: 执行失败: %v", err)
		return "", nil
	}
	t := time.Now()
	return sendGroupMgmtResponse(client, map[string]interface{}{"time": t.Unix()}, message)
}

// sendGroupMgmtResponse 构造并发送标准 onebot 响应
func sendGroupMgmtResponse(client callapi.Client, data interface{}, message callapi.ActionMessage) (string, error) {
	resp := groupResponse(data, message)
	if err := client.SendMessage(resp); err != nil {
		mylog.Printf("join_approval_strategy: 发送响应失败: %v", err)
	}
	result, err := json.Marshal(resp)
	if err != nil {
		return "", nil
	}
	return string(result), nil
}