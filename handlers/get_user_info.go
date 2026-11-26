package handlers

// 获取用户基本信息 (GET /v2/users/{user_openid}/info, 2026-11-16)
// 官方接口要求用户与机器人有好友关系, 否则返回 40184001

import (
	"context"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("get_user_info", HandleGetUserInfo)
}

// HandleGetUserInfo 获取用户基本信息 (扩展 API)
// 参数: user_id (应用端压缩 ID; string_ob11 时为 user_openid)
func HandleGetUserInfo(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	rawUserID := message.Params.UserID
	ids := memberIDStrings(rawUserID)
	if len(ids) == 0 || ids[0] == "" {
		callapi.SendAPIError(client, message, "get_user_info: 缺少 user_id")
		return "", nil
	}
	userOpenID, err := resolveUserOpenID(ids[0])
	if err != nil || userOpenID == "" {
		callapi.SendAPIError(client, message, "get_user_info: 无法反查用户 openid")
		return "", nil
	}

	info, err := apiv2.GetUserInfo(context.TODO(), userOpenID)
	if err != nil || info == nil {
		// 无好友关系(40184001)或无权限时回退 username 缓存
		nick := idmap.RetrieveUsernameByOpenID(userOpenID)
		mylog.Printf("get_user_info: 官方接口不可用(%v), 回退缓存 nickname[%v]", err, nick)
		return sendGroupMgmtResponse(client, map[string]interface{}{
			"user_id":             appUserIDValue(userOpenID),
			"nickname":            nick,
			"avatar":              "",
			"union_openid":        "",
			"allow_proactive_msg": false,
			"source":              "cache",
		}, message)
	}

	return sendGroupMgmtResponse(client, map[string]interface{}{
		"user_id":             appUserIDValue(userOpenID),
		"nickname":            info.Username,
		"avatar":              info.Avatar,
		"union_openid":        info.UnionOpenID,
		"allow_proactive_msg": info.AllowProactiveMsg,
		"source":              "official",
	}, message)
}
