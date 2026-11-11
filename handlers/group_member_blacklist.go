package handlers

// 群黑名单扩展 API (2026-09-03, 内邀白名单)
// 官方接口:
//   GET  /v2/groups/{group_openid}/member_blacklist        查询
//   POST /v2/groups/{group_openid}/member_blacklist        op=add 加入 / del 移出

import (
	"context"
	"fmt"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("get_group_member_blacklist", HandleGetGroupMemberBlacklist)
	callapi.RegisterHandler("add_group_member_blacklist", HandleAddGroupMemberBlacklist)
	callapi.RegisterHandler("remove_group_member_blacklist", HandleRemoveGroupMemberBlacklist)
}

// HandleGetGroupMemberBlacklist 查询群黑名单
// 参数: group_id, cursor(可选), limit(可选, 默认20最大100)
func HandleGetGroupMemberBlacklist(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	groupID, _ := message.Params.GroupID.(string)
	if groupID == "" {
		callapi.SendAPIError(client, message, "get_group_member_blacklist: 缺少 group_id")
		return "", nil
	}
	realGroupID, err := resolveGroupOpenID(groupID)
	if err != nil || realGroupID == "" {
		callapi.SendAPIError(client, message, "get_group_member_blacklist: 无法反查群 openid")
		return "", nil
	}

	list, err := apiv2.GetGroupMemberBlacklist(context.TODO(), realGroupID, message.Params.Cursor, message.Params.Limit)
	if err != nil {
		mylog.Printf("get_group_member_blacklist: 查询失败: %v", err)
		callapi.SendAPIError(client, message, fmt.Sprintf("get_group_member_blacklist: 官方接口失败(可能无内邀权限 11253): %v", err))
		return "", nil
	}

	users := make([]map[string]interface{}, 0)
	nextCursor := ""
	if list != nil {
		nextCursor = list.NextCursor
		for _, user := range list.Users {
			users = append(users, blacklistUserToMap(user))
		}
	}
	data := map[string]interface{}{
		"users":       users,
		"next_cursor": nextCursor,
	}
	return sendGroupMgmtResponse(client, data, message)
}

// HandleAddGroupMemberBlacklist 加入群黑名单 (要求成员当前不在群内)
func HandleAddGroupMemberBlacklist(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	return handleGroupMemberBlacklistOp(client, apiv2, message, "add")
}

// HandleRemoveGroupMemberBlacklist 移出群黑名单
func HandleRemoveGroupMemberBlacklist(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	return handleGroupMemberBlacklistOp(client, apiv2, message, "del")
}

func handleGroupMemberBlacklistOp(client callapi.Client, apiv2 openapi.OpenAPI, message callapi.ActionMessage, op string) (string, error) {
	groupID, _ := message.Params.GroupID.(string)
	if groupID == "" {
		callapi.SendAPIError(client, message, "group_member_blacklist: 缺少 group_id")
		return "", nil
	}
	openIDs := make([]string, 0, 4)
	seen := make(map[string]bool)
	for _, id := range message.Params.MemberOpenIDs {
		if id != "" && !seen[id] {
			seen[id] = true
			openIDs = append(openIDs, id)
		}
	}
	ids := memberIDStrings(message.Params.UserID)
	ids = append(ids, message.Params.UserIDs...)
	for _, id := range ids {
		if id == "" {
			continue
		}
		openID, err := resolveUserOpenID(id)
		if err != nil || openID == "" {
			callapi.SendAPIError(client, message, fmt.Sprintf("group_member_blacklist: 无法反查用户[%s]的openid", id))
			return "", nil
		}
		if !seen[openID] {
			seen[openID] = true
			openIDs = append(openIDs, openID)
		}
	}
	if len(openIDs) == 0 {
		callapi.SendAPIError(client, message, "group_member_blacklist: 缺少 user_id / user_ids / member_openids")
		return "", nil
	}
	if len(openIDs) > 20 {
		callapi.SendAPIError(client, message, "group_member_blacklist: 单次最多操作 20 个成员")
		return "", nil
	}

	realGroupID, err := resolveGroupOpenID(groupID)
	if err != nil || realGroupID == "" {
		callapi.SendAPIError(client, message, "group_member_blacklist: 无法反查群 openid")
		return "", nil
	}

	result, err := apiv2.SetGroupMemberBlacklist(context.TODO(), realGroupID, &dto.MemberBlacklistToCreate{
		Op:            op,
		MemberOpenIDs: openIDs,
	})
	if err != nil {
		mylog.Printf("group_member_blacklist(%s): 操作失败: %v", op, err)
		callapi.SendAPIError(client, message, fmt.Sprintf("group_member_blacklist(%s): 官方接口失败(可能无内邀权限 11253): %v", op, err))
		return "", nil
	}

	failOpenIDs := []string{}
	failUserIDs := []interface{}{}
	if result != nil {
		failOpenIDs = result.FailOpenIDs
		for _, openID := range result.FailOpenIDs {
			failUserIDs = append(failUserIDs, appUserIDValue(openID))
		}
	}
	return sendGroupMgmtResponse(client, map[string]interface{}{
		"op":            op,
		"fail_openids":  failOpenIDs,
		"fail_user_ids": failUserIDs,
	}, message)
}
