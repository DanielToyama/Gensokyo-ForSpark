package v2

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"context"
	"strconv"

	"github.com/tencent-connect/botgo/dto"
)

// 群聊管理接口实现
// 文档: https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/

// GetGroupInfo 获取群基本信息
func (o *openAPIv2) GetGroupInfo(ctx context.Context, groupOpenID string) (*dto.GroupInfo, error) {
	resp, err := o.request(ctx).
		SetResult(dto.GroupInfo{}).
		SetPathParam("group_id", groupOpenID).
		Get(o.getURL(groupInfoURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.GroupInfo), nil
}

// GetGroupBotState 获取机器人群内状态
func (o *openAPIv2) GetGroupBotState(ctx context.Context, groupOpenID string) (*dto.GroupBotState, error) {
	resp, err := o.request(ctx).
		SetResult(dto.GroupBotState{}).
		SetPathParam("group_id", groupOpenID).
		Get(o.getURL(groupBotStateURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.GroupBotState), nil
}

// GetRestrictChatSetting 查询群禁言状态（全员禁言模式 + 成员级禁言列表）
func (o *openAPIv2) GetRestrictChatSetting(ctx context.Context, groupOpenID string) (*dto.GroupRestrictChatSetting, error) {
	resp, err := o.request(ctx).
		SetResult(dto.GroupRestrictChatSetting{}).
		SetPathParam("group_id", groupOpenID).
		Get(o.getURL(groupRestrictChatSettingURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.GroupRestrictChatSetting), nil
}

// SetRestrictChatSetting 设置群成员禁言（单次设置不能超过10个）
// 注意: 官方当前文档仅声明 members 字段；GlobalRule(全员禁言) 为兼容字段。
func (o *openAPIv2) SetRestrictChatSetting(ctx context.Context, groupOpenID string, setting *dto.SetRestrictChatSettingToCreate) error {
	_, err := o.request(ctx).
		SetPathParam("group_id", groupOpenID).
		SetBody(setting).
		Post(o.getURL(groupRestrictChatSettingURI))
	return err
}

// GetJoinRequestList 拉取入群申请列表（分页）
// cursor 分页游标，首次请求可不传或传空串; limit 单页数量，默认 20，最大 100
func (o *openAPIv2) GetJoinRequestList(ctx context.Context, groupOpenID, cursor string, limit int) (*dto.JoinRequestList, error) {
	request := o.request(ctx).
		SetResult(dto.JoinRequestList{}).
		SetPathParam("group_id", groupOpenID)
	if cursor != "" {
		request = request.SetQueryParam("cursor", cursor)
	}
	if limit > 0 {
		request = request.SetQueryParam("limit", strconv.Itoa(limit))
	}
	resp, err := request.Get(o.getURL(groupJoinRequestListURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.JoinRequestList), nil
}

// GetGroupMembers 获取群成员列表（分页，每页最多 30 条）
func (o *openAPIv2) GetGroupMembers(ctx context.Context, groupOpenID, cursor string) (*dto.GroupMemberList, error) {
	request := o.request(ctx).
		SetResult(dto.GroupMemberList{}).
		SetPathParam("group_id", groupOpenID)
	if cursor != "" {
		request = request.SetQueryParam("cursor", cursor)
	}
	resp, err := request.Get(o.getURL(groupMembersURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.GroupMemberList), nil
}

// GetGroupMember 获取单个群成员信息
func (o *openAPIv2) GetGroupMember(ctx context.Context, groupOpenID, memberOpenID string) (*dto.GroupMember, error) {
	resp, err := o.request(ctx).
		SetResult(dto.GroupMember{}).
		SetPathParam("group_id", groupOpenID).
		SetPathParam("member_id", memberOpenID).
		Get(o.getURL(groupMemberURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.GroupMember), nil
}

// BatchRemoveMembers 批量移除群成员（单次最多 20 个）
func (o *openAPIv2) BatchRemoveMembers(ctx context.Context, groupOpenID string, req *dto.BatchRemoveMembersToCreate) (*dto.BatchRemoveMembersResult, error) {
	resp, err := o.request(ctx).
		SetResult(dto.BatchRemoveMembersResult{}).
		SetPathParam("group_id", groupOpenID).
		SetBody(req).
		Post(o.getURL(groupBatchRemoveMembersURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.BatchRemoveMembersResult), nil
}

// GetGroupMemberBlacklist 查询群黑名单（分页）
func (o *openAPIv2) GetGroupMemberBlacklist(ctx context.Context, groupOpenID, cursor string, limit int) (*dto.GroupMemberBlacklist, error) {
	request := o.request(ctx).
		SetResult(dto.GroupMemberBlacklist{}).
		SetPathParam("group_id", groupOpenID)
	if cursor != "" {
		request = request.SetQueryParam("cursor", cursor)
	}
	if limit > 0 {
		request = request.SetQueryParam("limit", strconv.Itoa(limit))
	}
	resp, err := request.Get(o.getURL(groupMemberBlacklistURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.GroupMemberBlacklist), nil
}

// SetGroupMemberBlacklist 群黑名单操作: op=add 加入 / del 移出
func (o *openAPIv2) SetGroupMemberBlacklist(ctx context.Context, groupOpenID string, req *dto.MemberBlacklistToCreate) (*dto.MemberBlacklistResult, error) {
	resp, err := o.request(ctx).
		SetResult(dto.MemberBlacklistResult{}).
		SetPathParam("group_id", groupOpenID).
		SetBody(req).
		Post(o.getURL(groupMemberBlacklistURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.MemberBlacklistResult), nil
}

// ApproveJoinRequest 审批入群申请: op=approve 通过 / decline 拒绝
func (o *openAPIv2) ApproveJoinRequest(ctx context.Context, groupOpenID, memberOpenID string, req *dto.ApproveJoinRequestToCreate) error {
	_, err := o.request(ctx).
		SetPathParam("group_id", groupOpenID).
		SetPathParam("member_id", memberOpenID).
		SetBody(req).
		Post(o.getURL(groupApprovalJoinRequestURI))
	return err
}

// CreateJoinApprovalStrategy 创建入群自动审批策略
// group_openids 与 group_ids 二选一必填; 一个机器人最多 20 个策略
func (o *openAPIv2) CreateJoinApprovalStrategy(ctx context.Context, strategy *dto.JoinApprovalStrategyToCreate) (*dto.JoinApprovalStrategyResult, error) {
	resp, err := o.request(ctx).
		SetResult(dto.JoinApprovalStrategyResult{}).
		SetBody(strategy).
		Post(o.getURL(joinApprovalStrategyURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.JoinApprovalStrategyResult), nil
}

// GetJoinApprovalStrategyList 查询入群自动审批策略列表（按创建时间倒序, 分页）
func (o *openAPIv2) GetJoinApprovalStrategyList(ctx context.Context, cursor string, limit int) (*dto.JoinApprovalStrategyList, error) {
	request := o.request(ctx).
		SetResult(dto.JoinApprovalStrategyList{})
	if cursor != "" {
		request = request.SetQueryParam("cursor", cursor)
	}
	if limit > 0 {
		request = request.SetQueryParam("limit", strconv.Itoa(limit))
	}
	resp, err := request.Get(o.getURL(joinApprovalStrategyURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.JoinApprovalStrategyList), nil
}

// UpdateJoinApprovalStrategy 修改入群自动审批策略
func (o *openAPIv2) UpdateJoinApprovalStrategy(ctx context.Context, strategyID string, strategy *dto.UpdateJoinApprovalStrategyToCreate) (*dto.JoinApprovalStrategyResult, error) {
	resp, err := o.request(ctx).
		SetResult(dto.JoinApprovalStrategyResult{}).
		SetPathParam("strategy_id", strategyID).
		SetBody(strategy).
		Patch(o.getURL(joinApprovalStrategyItemURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.JoinApprovalStrategyResult), nil
}

// DeleteJoinApprovalStrategy 删除入群自动审批策略
func (o *openAPIv2) DeleteJoinApprovalStrategy(ctx context.Context, strategyID string) error {
	_, err := o.request(ctx).
		SetPathParam("strategy_id", strategyID).
		Delete(o.getURL(joinApprovalStrategyItemURI))
	return err
}

// UpdateJoinApprovalStrategyWhitelist 修改入群自动审批策略的白名单号码
// 单次最多 10000 个，号码上限 10W; whitelist_users 使用字符串避免JS精度问题
func (o *openAPIv2) UpdateJoinApprovalStrategyWhitelist(ctx context.Context, strategyID string, whitelist *dto.WhitelistUsersToCreate) (*dto.WhitelistUpdateResult, error) {
	resp, err := o.request(ctx).
		SetResult(dto.WhitelistUpdateResult{}).
		SetPathParam("strategy_id", strategyID).
		SetBody(whitelist).
		Post(o.getURL(joinApprovalStrategyWhitelistURI))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*dto.WhitelistUpdateResult), nil
}

// ExecuteJoinApprovalStrategy 执行入群自动审批策略（全量扫描, 异步约10分钟完成）
func (o *openAPIv2) ExecuteJoinApprovalStrategy(ctx context.Context, strategyID string) error {
	_, err := o.request(ctx).
		SetPathParam("strategy_id", strategyID).
		Post(o.getURL(joinApprovalStrategyExecuteURI))
	return err
}
