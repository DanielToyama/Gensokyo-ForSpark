package dto

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

// 群聊管理（API v2）数据结构
// 对应线上文档: https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/
// 获取群基本信息        GET /v2/groups/{group_openid}/info
// 获取机器人群内状态     GET /v2/groups/{group_openid}/bot_state
// 查询群禁言状态         GET /v2/groups/{group_openid}/restrict_chat_setting
// 设置群成员禁言         POST /v2/groups/{group_openid}/restrict_chat_setting
// 入群申请列表拉取       GET /v2/groups/{group_openid}/join_request_list
// 入群申请审批           POST /v2/groups/{group_openid}/approval_join_request/{member_openid}
// 入群自动审批策略      /v2/groups/join_approval_strategy[/*]

// GroupInfo 群基本信息
type GroupInfo struct {
	GroupOpenID     string   `json:"group_openid"`
	GroupName       string   `json:"group_name"`
	GroupFingerMemo string   `json:"group_finger_memo"` // 群简介
	GroupClassText  string   `json:"group_class_text"`  // 群分类
	GroupTags       []string `json:"group_tags"`        // 群标签列表
	GroupMemberNum  int      `json:"group_member_num"`  // 群成员人数
}

// GroupBotState 机器人群内状态
type GroupBotState struct {
	MemberOpenID       string `json:"member_openid"`        // 机器人的 openid
	JoinedAt           string `json:"joined_at"`            // 入群时间戳（RFC3339格式）
	AllowProactiveMsg  bool   `json:"allow_proactive_msg"`  // 是否接收主动推送
	RecvMsgSetting     string `json:"recv_msg_setting"`     // 接收消息的类型: all/only_mention/mention_and_context
	MemberRole         string `json:"member_role"`          // 群成员角色: member/owner/admin
}

// GlobalMuteRule 群级禁言规则（全员禁言配置）
type GlobalMuteRule struct {
	Mode           string              `json:"mode"`            // none 未开启, always 始终禁言, schedule 定时禁言
	ScheduleRules  []MuteScheduleRule  `json:"schedule_rules"`  // 定时禁言规则列表
	RecurringRules []MuteRecurringRule `json:"recurring_rules"` // 周期禁言规则列表
}

// MuteScheduleRule 定时禁言规则
type MuteScheduleRule struct {
	TaskID  string `json:"task_id"`
	StartAt string `json:"start_at"` // RFC3339
	EndAt   string `json:"end_at"`   // RFC3339
	Enabled bool   `json:"enabled"`
}

// MuteRecurringRule 周期禁言规则
type MuteRecurringRule struct {
	TaskID    string `json:"task_id"`
	Weekdays  []int  `json:"weekdays"`  // 1~7（1=周一）
	StartTime string `json:"start_time"` // HH:mm（北京时间）
	EndTime   string `json:"end_time"`   // HH:mm（北京时间）
	Enabled   bool   `json:"enabled"`
}

// MemberMuteState 被禁言成员状态
type MemberMuteState struct {
	MemberOpenID  string `json:"member_openid"`
	MuteExpireAt  string `json:"mute_expire_at"` // RFC3339
	Username      string `json:"username"`
	UnionOpenID   string `json:"union_openid"`
}

// GroupRestrictChatSetting 群禁言状态（查询返回值）
type GroupRestrictChatSetting struct {
	GlobalRule *GlobalMuteRule  `json:"global_rule"`
	Members    []MemberMuteState `json:"members"`
}

// SetMemberMuteState 设置用户禁言的单项（请求体）
type SetMemberMuteState struct {
	Op           string `json:"op"` // add 增加禁言, update 更新禁言到期时间, del 解除禁言
	MemberOpenID string `json:"member_openid"`
	MuteExpireAt string `json:"mute_expire_at,omitempty"` // RFC3339; op=del 时可传空串
}

// SetRestrictChatSettingToCreate 设置群禁言请求体
// 说明: 官方当前文档中 POST /restrict_chat_setting 仅声明 members 字段（成员级禁言），
// 全员禁言的 global_rule 仅在查询接口返回；此处额外保留 GlobalRule 字段以便兼容平台后续能力。
type SetRestrictChatSettingToCreate struct {
	GlobalRule *GlobalMuteRule      `json:"global_rule,omitempty"`
	Members    []*SetMemberMuteState `json:"members,omitempty"`
}

// VerifyInfo 用户入群验证方式
type VerifyInfo struct {
	Method         string      `json:"method"` // verify_message / admin_review_qa
	VerifyMessage  string      `json:"verify_message"`
	ReviewQAList   []ReviewQA  `json:"review_qa_list"`
}

// ReviewQA 入群问答
type ReviewQA struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// JoinRequest 入群申请
type JoinRequest struct {
	JoinRequestID  string      `json:"join_request_id"`
	RiskTips       string      `json:"risk_tips"` // warning_tips / top_tips
	UnionOpenID    string      `json:"union_openid"`
	MemberOpenID   string      `json:"member_openid"`
	Username       string      `json:"username"`
	ApplyAt        string      `json:"apply_at"` // RFC3339
	ApplySource    string      `json:"apply_source"` // self_apply 主动申请, invited 被邀请
	InvitedBy      string      `json:"invited_by"`
	Bot            bool        `json:"bot"`
	VerifyInfo     *VerifyInfo `json:"verify_info"`
}

// JoinRequestList 入群申请列表（分页）
type JoinRequestList struct {
	List       []*JoinRequest `json:"list"`
	NextCursor string         `json:"next_cursor"` // 空串表示已到末页
}

// ApproveJoinRequestToCreate 入群申请审批请求体
type ApproveJoinRequestToCreate struct {
	Op                   string `json:"op"` // approve 通过, decline 拒绝
	JoinRequestID        string `json:"join_request_id"`
	RejectReason         string `json:"reject_reason,omitempty"`
	AddToMemberBlacklist bool   `json:"add_to_member_blacklist,omitempty"`
}

// JoinApprovalStrategyToCreate 创建入群自动审批策略请求体
type JoinApprovalStrategyToCreate struct {
	GroupOpenIDs []string `json:"group_openids,omitempty"` // 与 GroupIDs 二选一必填
	GroupIDs     []string `json:"group_ids,omitempty"`     // QQ群号列表（字符串避免精度问题）
	IsEnable     string   `json:"is_enable,omitempty"`     // on-启用 off-关闭, 默认 on
	ExpireAt     string   `json:"expire_at,omitempty"`     // RFC3339; 不传默认一年过期
	Remark       string   `json:"remark,omitempty"`
}

// JoinApprovalStrategy 入群自动审批策略（查询返回）
type JoinApprovalStrategy struct {
	StrategyID         string   `json:"strategy_id"`
	GroupOpenIDs       []string `json:"group_openids"`
	GroupIDs           []string `json:"group_ids"`
	WhitelistUserCount int      `json:"whitelist_user_count"` // 估算值
	IsEnable           string   `json:"is_enable"`
	ExpireAt           string   `json:"expire_at"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
	Remark             string   `json:"remark"`
}

// JoinApprovalStrategyList 入群自动审批策略列表（分页）
type JoinApprovalStrategyList struct {
	Strategies []*JoinApprovalStrategy `json:"strategies"`
	NextCursor string                  `json:"next_cursor"`
}

// JoinApprovalStrategyResult 创建/修改策略的返回
type JoinApprovalStrategyResult struct {
	StrategyID string `json:"strategy_id"`
	IsEnable   string `json:"is_enable"`
	ExpireAt   string `json:"expire_at"`
}

// GroupAction 修改策略时的关联群增删操作
type GroupAction struct {
	Op           string   `json:"op"` // add 新增关联群, del 删除关联群
	GroupOpenIDs []string `json:"group_openids,omitempty"`
	GroupIDs     []string `json:"group_ids,omitempty"`
}

// UpdateJoinApprovalStrategyToCreate 修改入群自动审批策略请求体
type UpdateJoinApprovalStrategyToCreate struct {
	IsEnable   string       `json:"is_enable,omitempty"` // on-启用 off-关闭
	ExpireAt   string       `json:"expire_at,omitempty"`
	GroupAction *GroupAction `json:"group_action,omitempty"`
	Remark     string       `json:"remark,omitempty"`
}

// WhitelistUsersToCreate 修改策略白名单号码请求体
type WhitelistUsersToCreate struct {
	Op            string   `json:"op"` // add 新增号码, del 删除号码
	WhitelistUsers []string `json:"whitelist_users"` // QQ号码列表(字符串)
}

// WhitelistUpdateResult 白名单修改返回
type WhitelistUpdateResult struct {
	StrategyID         string `json:"strategy_id"`
	WhitelistUserCount int    `json:"whitelist_user_count"`
	UpdatedAt          string `json:"updated_at"`
}

// ***************** 群聊管理事件 *****************

// GroupMemberAddEvent 群成员加入事件 (GROUP_MEMBER_ADD)
type GroupMemberAddEvent struct {
	Timestamp    interface{} `json:"timestamp"`    // Unix 秒
	GroupOpenID  string      `json:"group_openid"`
	MemberOpenID string      `json:"member_openid"`
	UserOpenID   string      `json:"user_openid"` // 跨应用统一标识, 可能为空
}

// GroupMemberRemoveEvent 群成员退出事件 (GROUP_MEMBER_REMOVE)
type GroupMemberRemoveEvent struct {
	Timestamp    interface{} `json:"timestamp"`    // Unix 秒
	GroupOpenID  string      `json:"group_openid"`
	MemberOpenID string      `json:"member_openid"`
	UserOpenID   string      `json:"user_openid"` // 可能为空
}

// GroupJoinRequestEvent 用户申请加群事件 (GROUP_JOIN_REQUEST)
// 注意: 只有当机器人是群管理员时才可以收到此事件。
type GroupJoinRequestEvent struct {
	GroupOpenID    string     `json:"group_openid"`
	JoinRequestID  string     `json:"join_request_id"`
	RiskTips       string     `json:"risk_tips"`
	UnionOpenID    string     `json:"union_openid"`
	MemberOpenID   string     `json:"member_openid"`
	Username       string     `json:"username"`
	ApplyAt        string     `json:"apply_at"`     // RFC3339
	ApplySource    string     `json:"apply_source"` // self_apply 主动申请, invited 被邀请
	InvitedBy      string     `json:"invited_by"`
	Bot            bool       `json:"bot"`
	VerifyInfo     *VerifyInfo `json:"verify_info"`
	AutoApproved   *AutoApproved `json:"auto_approved,omitempty"` // 只在事件中携带
}

// AutoApproved 自动审批通过的扩展信息
type AutoApproved struct {
	StrategyID string `json:"strategy_id"`
}

// SubscribeMsgTemplateResult 订阅模板授权结果
type SubscribeMsgTemplateResult struct {
	TemplateID       int    `json:"template_id"`
	CustomTemplateID string `json:"custom_template_id"`
	Op               int    `json:"op"` // 1=允许订阅, 2=拒绝订阅
	SubscribeID      string `json:"subscribe_id"`
	SubscribeTS      int64  `json:"subscribe_ts"`
	UpdateTS         int64  `json:"update_ts"`
}

// SubscribeMessageStatusEvent 订阅消息授权状态变更事件 (SUBSCRIBE_MESSAGE_STATUS)
type SubscribeMessageStatusEvent struct {
	GroupOpenID string                       `json:"group_openid"`
	OpenID      string                       `json:"openid"`
	Result      []SubscribeMsgTemplateResult `json:"result"`
}