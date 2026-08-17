// 处理收到的回调事件
package Processor

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"time"

	"github.com/hoshinonyaruko/gensokyo/config"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
)

// OnebotGroupJoinRequest 用户申请加群事件转换后的 onebot v11 request.group 事件 (带增强字段)
type OnebotGroupJoinRequest struct {
	Comment     string `json:"comment"` // 验证消息内容
	Flag        string `json:"flag"`    // join_request_id, 审批(set_group_add_request)时回传
	GroupID     int64  `json:"group_id"`
	PostType    string `json:"post_type"`
	RequestType string `json:"request_type"`
	SelfID      int64  `json:"self_id"`
	SubType     string `json:"sub_type"`
	Time        int64  `json:"time"`
	UserID      int64  `json:"user_id"`
	RealUserID  string `json:"real_user_id,omitempty"`  //当前真实uid
	RealGroupID string `json:"real_group_id,omitempty"` //当前真实gid

	// 增强字段
	Username     string `json:"username,omitempty"`                  // 申请人昵称
	RiskTips     string `json:"risk_tips,omitempty"`                 // 安全提示语
	ApplySource  string `json:"apply_source,omitempty"`              // self_apply 主动申请, invited 被邀请
	InvitedBy    string `json:"invited_by,omitempty"`                // 邀请人openid
	VerifyMethod string `json:"verify_method,omitempty"`             // 入群验证方式
	Bot          bool   `json:"bot,omitempty"`                       // 是否为机器人账号
	AutoApproved string `json:"auto_approved_strategy_id,omitempty"` // 自动审批通过的策略ID
}

// ProcessGroupJoinRequest 处理用户申请加群事件 (GROUP_JOIN_REQUEST, Intent GROUP_AND_C2C_EVENT 1<<25)
// 转换成 onebot v11 request.group (sub_type=add), flag 为 join_request_id
// 说明: 只有当机器人是群管理员时才可以收到此事件
func (p *Processors) ProcessGroupJoinRequest(data *dto.GroupJoinRequestEvent) error {
	var userid64 int64
	var GroupID64 int64
	var err error
	if config.GetIdmapPro() {
		GroupID64, userid64, err = idmap.StoreIDv2Pro(data.GroupOpenID, data.MemberOpenID)
		if err != nil {
			mylog.Errorf("Error storing ID: %v", err)
		}
	} else {
		GroupID64, err = idmap.StoreIDv2(data.GroupOpenID)
		if err != nil {
			mylog.Errorf("failed to convert GroupOpenID to int: %v", err)
			return nil
		}
		userid64, err = idmap.StoreIDv2(data.MemberOpenID)
		if err != nil {
			mylog.Printf("Error storing ID: %v", err)
			return nil
		}
	}

	var selfid64 int64
	if config.GetUseUin() {
		selfid64 = config.GetUinint64()
	} else {
		selfid64 = int64(p.Settings.AppID)
	}

	// [新增] 缓存申请人昵称(官方事件携带的 username), 供 get_stranger_info 反查展示
	idmap.StoreUsernameV2(data.MemberOpenID, data.Username)
	// [新增] 缓存入群申请 id -> (群,成员) openid, 供 set_group_add_request 只传 flag 时审批
	idmap.StoreJoinRequestV2(data.JoinRequestID, data.GroupOpenID, data.MemberOpenID)

	comment := ""
	if data.VerifyInfo != nil {
		comment = data.VerifyInfo.VerifyMessage
	}
	if comment == "" && data.ApplySource == "invited" {
		comment = "被邀请入群"
	}

	req := OnebotGroupJoinRequest{
		Comment:     comment,
		Flag:        data.JoinRequestID, // 审批接口回传
		GroupID:     GroupID64,
		PostType:    "request",
		RequestType: "group",
		SelfID:      selfid64,
		SubType:     "add",
		Time:        time.Now().Unix(),
		UserID:      userid64,
	}
	//增强配置
	if !config.GetNativeOb11() {
		req.RealUserID = data.MemberOpenID
		req.RealGroupID = data.GroupOpenID
	}
	//额外信息
	req.Username = data.Username
	req.RiskTips = data.RiskTips
	req.ApplySource = data.ApplySource
	req.InvitedBy = data.InvitedBy
	req.Bot = data.Bot
	if data.VerifyInfo != nil {
		req.VerifyMethod = data.VerifyInfo.Method
	}
	if data.AutoApproved != nil {
		req.AutoApproved = data.AutoApproved.StrategyID
	}

	reqMap := structToMap(req)
	//上报信息到onebotv11应用端(正反ws)
	go p.BroadcastMessageToAll(reqMap, p.Apiv2, data)

	mylog.Printf("用户[%v](%v)申请加入群[%v] flag[%v]", userid64, data.Username, GroupID64, data.JoinRequestID)
	return nil
}
