package v1

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"context"
	"strconv"

	"github.com/tencent-connect/botgo/dto"
)

// 群聊管理接口实现 (v1 实现与 v2 相同, 均调用 /v2/ 路径)
// 文档: https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/

// GetGroupInfo 获取群基本信息
func (o *openAPI) GetGroupInfo(ctx context.Context, groupOpenID string) (*dto.GroupInfo, error) {
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
func (o *openAPI) GetGroupBotState(ctx context.Context, groupOpenID string) (*dto.GroupBotState, error) {
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
func (o *openAPI) GetRestrictChatSetting(ctx context.Context, groupOpenID string) (*dto.GroupRestrictChatSetting, error) {
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
func (o *openAPI) SetRestrictChatSetting(ctx context.Context, groupOpenID string, setting *dto.SetRestrictChatSettingToCreate) error {
	_, err := o.request(ctx).
		SetPathParam("group_id", groupOpenID).
		SetBody(setting).
		Post(o.getURL(groupRestrictChatSettingURI))
	return err
}

// GetJoinRequestList 拉取入群申请列表（分页）
func (o *openAPI) GetJoinRequestList(ctx context.Context, groupOpenID, cursor string, limit int) (*dto.JoinRequestList, error) {
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

// ApproveJoinRequest 审批入群申请: op=approve 通过 / decline 拒绝
func (o *openAPI) ApproveJoinRequest(ctx context.Context, groupOpenID, memberOpenID string, req *dto.ApproveJoinRequestToCreate) error {
	_, err := o.request(ctx).
		SetPathParam("group_id", groupOpenID).
		SetPathParam("member_id", memberOpenID).
		SetBody(req).
		Post(o.getURL(groupApprovalJoinRequestURI))
	return err
}

// CreateJoinApprovalStrategy 创建入群自动审批策略
func (o *openAPI) CreateJoinApprovalStrategy(ctx context.Context, strategy *dto.JoinApprovalStrategyToCreate) (*dto.JoinApprovalStrategyResult, error) {
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
func (o *openAPI) GetJoinApprovalStrategyList(ctx context.Context, cursor string, limit int) (*dto.JoinApprovalStrategyList, error) {
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
func (o *openAPI) UpdateJoinApprovalStrategy(ctx context.Context, strategyID string, strategy *dto.UpdateJoinApprovalStrategyToCreate) (*dto.JoinApprovalStrategyResult, error) {
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
func (o *openAPI) DeleteJoinApprovalStrategy(ctx context.Context, strategyID string) error {
	_, err := o.request(ctx).
		SetPathParam("strategy_id", strategyID).
		Delete(o.getURL(joinApprovalStrategyItemURI))
	return err
}

// UpdateJoinApprovalStrategyWhitelist 修改入群自动审批策略的白名单号码
func (o *openAPI) UpdateJoinApprovalStrategyWhitelist(ctx context.Context, strategyID string, whitelist *dto.WhitelistUsersToCreate) (*dto.WhitelistUpdateResult, error) {
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
func (o *openAPI) ExecuteJoinApprovalStrategy(ctx context.Context, strategyID string) error {
	_, err := o.request(ctx).
		SetPathParam("strategy_id", strategyID).
		Post(o.getURL(joinApprovalStrategyExecuteURI))
	return err
}