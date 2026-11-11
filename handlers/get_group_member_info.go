package handlers

import (
	"context"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

// 初始化handler，在程序启动时会被调用
func init() {
	callapi.RegisterHandler("get_group_member_info", GetGroupMemberInfo)
}

// GetGroupMemberInfo 获取群成员信息
// 官方接口: GET /v2/groups/{group_openid}/members/{member_openid} (2026-09-03, 内邀白名单)
// OneBot 参数: group_id, user_id
func GetGroupMemberInfo(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	groupID, _ := message.Params.GroupID.(string)
	userIDs := memberIDStrings(message.Params.UserID)
	if groupID == "" || len(userIDs) == 0 || userIDs[0] == "" {
		callapi.SendAPIError(client, message, "get_group_member_info: 缺少 group_id 或 user_id")
		return "", nil
	}

	realGroupID, err := resolveGroupOpenID(groupID)
	if err != nil || realGroupID == "" {
		callapi.SendAPIError(client, message, "get_group_member_info: 无法反查群 openid")
		return "", nil
	}
	realUserID, err := resolveUserOpenID(userIDs[0])
	if err != nil || realUserID == "" {
		callapi.SendAPIError(client, message, "get_group_member_info: 无法反查用户 openid")
		return "", nil
	}

	member, err := apiv2.GetGroupMember(context.TODO(), realGroupID, realUserID)
	if err != nil || member == nil {
		// 官方接口当前为内邀白名单; 无权限时回退到本地 username 缓存, 保证有合法响应
		nick := idmap.RetrieveUsernameByOpenID(realUserID)
		mylog.Printf("get_group_member_info: 官方接口不可用(%v), 回退缓存 nickname[%v]", err, nick)
		member = &dto.GroupMember{
			MemberOpenID: realUserID,
			Username:     nick,
			MemberRole:   "member",
		}
	}

	data := groupMemberToMap(member, message.Params.GroupID)
	return sendGroupMgmtResponse(client, data, message)
}
