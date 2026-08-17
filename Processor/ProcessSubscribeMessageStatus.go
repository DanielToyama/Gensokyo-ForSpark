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

// OnebotSubscribeMessageNotice 订阅消息授权状态变更的 onebot 扩展 notice 事件
type OnebotSubscribeMessageNotice struct {
	GroupID    int64  `json:"group_id,omitempty"`
	NoticeType string `json:"notice_type"` // subscribe_message_status
	PostType   string `json:"post_type"`   // notice
	SelfID     int64  `json:"self_id"`
	Time       int64  `json:"time"`
	OpenID     string `json:"openid,omitempty"` // 用户 OpenID（个人订阅场景）
	Data       *dto.SubscribeMessageStatusEvent `json:"data,omitempty"`
	RealUserID  string `json:"real_user_id,omitempty"`  //当前真实uid
	RealGroupID string `json:"real_group_id,omitempty"` //当前真实gid
}

// ProcessSubscribeMessageStatus 处理订阅消息授权状态变更事件 (SUBSCRIBE_MESSAGE_STATUS, Intent GROUP_AND_C2C_EVENT 1<<25)
// 用于判断用户是否允许/拒绝接收某个订阅消息模板
// 转换成 onebot 扩展 notice: notice_type=subscribe_message_status
func (p *Processors) ProcessSubscribeMessageStatus(data *dto.SubscribeMessageStatusEvent) error {
	var GroupID64 int64
	var err error
	if data.GroupOpenID != "" {
		GroupID64, err = idmap.StoreIDv2(data.GroupOpenID)
		if err != nil {
			mylog.Errorf("failed to convert GroupOpenID to int: %v", err)
			return nil
		}
	}

	var selfid64 int64
	if config.GetUseUin() {
		selfid64 = config.GetUinint64()
	} else {
		selfid64 = int64(p.Settings.AppID)
	}

	notice := OnebotSubscribeMessageNotice{
		GroupID:    GroupID64,
		NoticeType: "subscribe_message_status",
		PostType:   "notice",
		SelfID:     selfid64,
		Time:       time.Now().Unix(),
		OpenID:     data.OpenID,
		Data:       data,
	}
	//增强配置
	if !config.GetNativeOb11() {
		if data.GroupOpenID != "" {
			notice.RealGroupID = data.GroupOpenID
		}
		if data.OpenID != "" {
			notice.RealUserID = data.OpenID
		}
	}

	noticeMap := structToMap(notice)
	//上报信息到onebotv11应用端(正反ws)
	go p.BroadcastMessageToAll(noticeMap, p.Apiv2, data)

	mylog.Printf("订阅消息授权状态变更: group[%v] openid[%v] 结果数[%d]", data.GroupOpenID, data.OpenID, len(data.Result))
	return nil
}