package handlers

// 群成员列表: 官方 v2 接口优先, 失败回退本地 idmap 缓存

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

// GetGroupMemberListGroup 获取群成员列表 (OneBot get_group_member_list)
// 官方接口: GET /v2/groups/{group_openid}/members (2026-09-03, 内邀白名单, 每页最多30条)
func GetGroupMemberListGroup(client callapi.Client, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	groupID, _ := message.Params.GroupID.(string)
	realGroupID, err := resolveGroupOpenID(groupID)
	if err == nil && realGroupID != "" {
		members, ferr := fetchAllGroupMembers(apiv2, realGroupID)
		if ferr == nil {
			data := make([]map[string]interface{}, 0, len(members))
			for _, member := range members {
				data = append(data, groupMemberToMap(member, message.Params.GroupID))
			}
			return sendGroupMgmtResponse(client, data, message)
		}
		mylog.Printf("get_group_member_list: 官方接口不可用(%v), 回退本地缓存", ferr)
	} else if err != nil {
		mylog.Printf("get_group_member_list: 无法反查群openid(%v), 回退本地缓存", err)
	}
	return getGroupMemberListLocal(client, message)
}

// fetchAllGroupMembers 按 cursor 拉取全部群成员 (每页最多30条, 上限保护防止死循环)
func fetchAllGroupMembers(apiv2 openapi.OpenAPI, groupOpenID string) ([]*dto.GroupMember, error) {
	var all []*dto.GroupMember
	cursor := ""
	for page := 0; page < 200; page++ {
		list, err := apiv2.GetGroupMembers(context.TODO(), groupOpenID, cursor)
		if err != nil {
			return nil, err
		}
		if list == nil {
			break
		}
		all = append(all, list.Members...)
		if list.NextCursor == "" || list.NextCursor == cursor {
			break
		}
		cursor = list.NextCursor
	}
	return all, nil
}

// getGroupMemberListLocal 旧逻辑: 从 idmap-pro 缓存的群成员构造列表
func getGroupMemberListLocal(client callapi.Client, message callapi.ActionMessage) (string, error) {
	groupID, _ := message.Params.GroupID.(string)
	mylog.Printf("getGroupMemberList(group): 官方接口不可用, 从本地缓存获取(请在config打开idmap-pro)")
	userIDs, err := idmap.FindSubKeysByIdPro(groupID)
	if err != nil {
		mylog.Printf("Error retrieving user IDs: %v", err)
		return "", nil
	}
	groupIDInt, _ := strconv.ParseUint(groupID, 10, 64)
	yesterday := int32(time.Now().AddDate(0, 0, -1).Unix())
	members := make([]MemberList, 0, len(userIDs))
	for _, userID := range userIDs {
		userIDInt, _ := strconv.ParseUint(userID, 10, 64)
		members = append(members, MemberList{
			UserID:          userIDInt,
			GroupID:         groupIDInt,
			Nickname:        "主人",
			Card:            "主人",
			Sex:             "0",
			Age:             0,
			Area:            "0",
			JoinTime:        yesterday,
			LastSentTime:    0,
			Level:           "0",
			Role:            "member",
			Unfriendly:      false,
			Title:           "0",
			TitleExpireTime: 0,
			CardChangeable:  false,
			ShutUpTimestamp: 0,
		})
	}
	return sendGroupMgmtResponse(client, members, message)
}
