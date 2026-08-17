package handlers

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/idmap"
)

// 群聊管理 handler 公共辅助函数

// groupResponse 构造 onebot v11 标准响应 map
func groupResponse(data interface{}, message callapi.ActionMessage) map[string]interface{} {
	resp := map[string]interface{}{
		"data":    data,
		"message": "success",
		"retcode": 0,
		"status":  "ok",
	}
	if message.Echo != nil && message.Echo != "" {
		resp["echo"] = message.Echo
	}
	return resp
}

// resolveGroupOpenID 通过 idmap 反查真实群 openid (int格式 -> openid)
func resolveGroupOpenID(groupID string) (string, error) {
	return idmap.RetrieveRowByIDv2(groupID)
}

// resolveUserOpenID 通过 idmap 反查真实用户 openid (int格式 -> openid)
func resolveUserOpenID(userID string) (string, error) {
	return idmap.RetrieveRowByIDv2(userID)
}