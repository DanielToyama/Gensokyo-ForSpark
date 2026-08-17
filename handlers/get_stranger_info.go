package handlers

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"fmt"
	"strconv"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("get_stranger_info", GetStrangerInfo)
}

// GetStrangerInfo 获取陌生人信息 (onebot v11 get_stranger_info)
//
// QQ官方机器人API v2 没有任何"陌生人资料/QQ等级"类接口, 因此:
//   - nickname: 从 idmap 用户名缓存取 (加群申请等事件里官方会带 username, Processor 侧已缓存)
//   - sex/age:  官方不提供, 返回 "" / 0 占位 (必须返回合法 JSON, 否则对端插件 JSON5 解析崩溃)
//   - qqLevel:  官方不提供, 固定返回 9999。目的: SparkBridge groupRequest 等插件的
//     "QQ等级门槛"判断会永远放行(插件把 0 级当作"隐私未公开"并直接拒绝, 返回 9999 避免误拒所有人)。
func GetStrangerInfo(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	rawUserID := message.Params.UserID

	// user_id 可能是 JSON 数字(float64)或字符串
	var userIDStr string
	switch v := rawUserID.(type) {
	case nil:
		userIDStr = ""
	case float64:
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		userIDStr = fmt.Sprintf("%v", v)
	}

	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	// 反查官方 openid -> 用户缓存昵称
	nickname := ""
	if userIDStr != "" {
		if openid, err := idmap.RetrieveRowByIDv2(userIDStr); err == nil && openid != "" {
			nickname = idmap.RetrieveUsernameByOpenID(openid)
		}
	}

	data := map[string]interface{}{
		"user_id":  userID,
		"nickname": nickname,
		"sex":      "",
		"age":      0,
		"qqLevel":  9999, // 官方API无QQ等级数据, 固定放行等级门槛(见函数头注释)
	}
	resp := groupResponse(data, message)
	if err := client.SendMessage(resp); err != nil {
		mylog.Printf("get_stranger_info: 发送响应失败: %v", err)
	}
	mylog.Printf("get_stranger_info: user_id=%d nickname=%s", userID, nickname)
	return "", nil
}
