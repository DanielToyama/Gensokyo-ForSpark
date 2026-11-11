package handlers

// 群成员管理 (2026-09-03) 公共辅助

import (
	"strconv"
	"time"

	"github.com/hoshinonyaruko/gensokyo/config"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/tencent-connect/botgo/dto"
)

// appUserIDValue 官方 openid -> 应用端可见 user_id
// string_ob11 模式直接返回 openid; 否则用 idmap 压缩为数字行 ID
func appUserIDValue(openid string) interface{} {
	if openid == "" {
		return ""
	}
	if config.GetStringOb11() {
		return openid
	}
	if id, err := idmap.StoreIDv2(openid); err == nil {
		return id
	}
	return openid
}

// groupMemberToMap 把官方 GroupMember 转为 OneBot get_group_member_list / get_group_member_info 的成员对象
func groupMemberToMap(member *dto.GroupMember, groupID interface{}) map[string]interface{} {
	if member == nil {
		return map[string]interface{}{}
	}
	role := member.MemberRole
	if role == "" {
		role = "member"
	}
	var joinTime int64
	if t, err := time.Parse(time.RFC3339, member.JoinedAt); err == nil {
		joinTime = t.Unix()
	}
	return map[string]interface{}{
		"user_id":           appUserIDValue(member.MemberOpenID),
		"group_id":          groupID,
		"nickname":          member.Username,
		"card":              member.Username,
		"sex":               "0",
		"age":               0,
		"area":              "0",
		"join_time":         joinTime,
		"last_sent_time":    0,
		"level":             "0",
		"role":              role,
		"unfriendly":        false,
		"title":             "0",
		"title_expire_time": 0,
		"card_changeable":   false,
		"shut_up_timestamp": 0,
		// 额外字段, 方便应用端需要时直接使用官方 openid
		"member_openid": member.MemberOpenID,
		"bot":           member.Bot,
	}
}

// blacklistUserToMap 把官方黑名单用户转为应用端对象
func blacklistUserToMap(user *dto.GroupMemberBlacklistUser) map[string]interface{} {
	if user == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"user_id":       appUserIDValue(user.MemberOpenID),
		"nickname":      user.Username,
		"member_openid": user.MemberOpenID,
		"union_openid":  user.UnionOpenID,
		"banned_at":     user.BannedAt,
		"bot":           user.Bot,
	}
}

// memberIDStrings 从 callapi 的 interface{} 参数中提取 ID 列表
// 支持 string / float64 / int / []string / []interface{}
func memberIDStrings(v interface{}) []string {
	switch x := v.(type) {
	case string:
		if x != "" {
			return []string{x}
		}
	case float64:
		return []string{strconv.FormatInt(int64(x), 10)}
	case int:
		return []string{strconv.Itoa(x)}
	case int64:
		return []string{strconv.FormatInt(x, 10)}
	case []string:
		return x
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, item := range x {
			out = append(out, memberIDStrings(item)...)
		}
		return out
	}
	return nil
}
