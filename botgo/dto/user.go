package dto

// User 用户
type User struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Avatar           string `json:"avatar"`
	Bot              bool   `json:"bot"`
	UnionOpenID      string `json:"union_openid"`       // 特殊关联应用的 openid
	UnionUserAccount string `json:"union_user_account"` // 机器人关联的用户信息，与union_openid关联的应用是同一个
}

// UserInfo 获取用户基本信息 (GET /v2/users/{user_openid}/info, 2026-11-16)
type UserInfo struct {
	UnionOpenID       string `json:"union_openid"`        // 跨应用关联用户 openid (绑定3方应用时有效)
	Username          string `json:"username"`            // 用户名
	Avatar            string `json:"avatar"`              // 用户头像原图
	AllowProactiveMsg bool   `json:"allow_proactive_msg"` // 是否接收主动推送
}
