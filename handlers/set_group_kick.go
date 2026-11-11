package handlers

// 群成员批量移除 (OneBot set_group_kick 扩展)
// 官方接口: POST /v2/groups/{group_openid}/batch_remove_members (2026-09-03, 内邀白名单)
// 限制: 单次最多 20 个, 30 QPM

import (
	"context"
	"fmt"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("set_group_kick", HandleSetGroupKick)
	// 语义更明确的别名, 便于应用端直接调用批量接口
	callapi.RegisterHandler("batch_remove_group_members", HandleSetGroupKick)
}

// HandleSetGroupKick 群成员移除
// OneBot v11 标准参数: group_id, user_id, reject_add_request
// 扩展参数: user_ids []string, member_openids []string, add_to_member_blacklist bool
// 单次最多 20 个成员; 只传一个 user_id 时等价 OneBot 单人踢人。
func HandleSetGroupKick(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	groupID, _ := message.Params.GroupID.(string)
	if groupID == "" {
		callapi.SendAPIError(client, message, "set_group_kick: 缺少 group_id")
		return "", nil
	}

	openIDs := make([]string, 0, 4)
	seen := make(map[string]bool)

	// 直接传官方 openid 的情况
	for _, id := range message.Params.MemberOpenIDs {
		if id != "" && !seen[id] {
			seen[id] = true
			openIDs = append(openIDs, id)
		}
	}

	// 应用端 user_id / user_ids: 反查 openid
	ids := memberIDStrings(message.Params.UserID)
	ids = append(ids, message.Params.UserIDs...)
	for _, id := range ids {
		if id == "" {
			continue
		}
		openID, err := resolveUserOpenID(id)
		if err != nil || openID == "" {
			callapi.SendAPIError(client, message, fmt.Sprintf("set_group_kick: 无法反查用户[%s]的openid", id))
			return "", nil
		}
		if !seen[openID] {
			seen[openID] = true
			openIDs = append(openIDs, openID)
		}
	}

	if len(openIDs) == 0 {
		callapi.SendAPIError(client, message, "set_group_kick: 缺少 user_id / user_ids / member_openids")
		return "", nil
	}
	if len(openIDs) > 20 {
		callapi.SendAPIError(client, message, "set_group_kick: 单次最多移除 20 个成员")
		return "", nil
	}

	realGroupID, err := resolveGroupOpenID(groupID)
	if err != nil || realGroupID == "" {
		callapi.SendAPIError(client, message, "set_group_kick: 无法反查群 openid")
		return "", nil
	}

	req := &dto.BatchRemoveMembersToCreate{
		MemberOpenIDs:        openIDs,
		AddToMemberBlacklist: message.Params.AddToMemberBlacklist || message.Params.RejectAddRequest,
	}
	result, err := apiv2.BatchRemoveMembers(context.TODO(), realGroupID, req)
	if err != nil {
		mylog.Printf("set_group_kick: 批量移除失败 group[%v] count[%v]: %v", realGroupID, len(openIDs), err)
		callapi.SendAPIError(client, message, fmt.Sprintf("set_group_kick: 官方接口失败(可能无内邀权限 11253): %v", err))
		return "", nil
	}

	data := map[string]interface{}{
		"remove_members_result":                 "",
		"add_to_member_blacklist_fail_openids":  []string{},
		"add_to_member_blacklist_fail_user_ids": []interface{}{},
	}
	if result != nil {
		data["remove_members_result"] = result.RemoveMembersResult
		failIDs := make([]interface{}, 0, len(result.AddToMemberBlacklistFailOpenIDs))
		for _, openID := range result.AddToMemberBlacklistFailOpenIDs {
			failIDs = append(failIDs, appUserIDValue(openID))
		}
		data["add_to_member_blacklist_fail_openids"] = result.AddToMemberBlacklistFailOpenIDs
		data["add_to_member_blacklist_fail_user_ids"] = failIDs
	}
	return sendGroupMgmtResponse(client, data, message)
}
