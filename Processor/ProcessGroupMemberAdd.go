// 处理收到的回调事件
package Processor

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hoshinonyaruko/gensokyo/config"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
)

// eventTimestampToInt64 统一的官方事件时间戳转换 (timestamp 可能是 string/int64/float64)
func eventTimestampToInt64(v interface{}) (int64, error) {
	switch t := v.(type) {
	case string:
		return strconv.ParseInt(t, 10, 64)
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case float64:
		return int64(t), nil
	case json.Number:
		return t.Int64()
	default:
		return 0, fmt.Errorf("invalid timestamp type: %T", v)
	}
}

// ProcessGroupMemberAdd 处理群成员加入事件 (GROUP_MEMBER_ADD, Intent GROUP_MEMBER_EVENT 1<<24)
// 转换成 onebot v11 notice.group_increase
// 说明: 官方事件不携带操作者信息, 无法区分 approve/invite/manage, 默认 sub_type=approve
func (p *Processors) ProcessGroupMemberAdd(data *dto.GroupMemberAddEvent) error {
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

	timestampInt64, err := eventTimestampToInt64(data.Timestamp)
	if err != nil {
		mylog.Printf("ProcessGroupMemberAdd: 时间戳转换失败: %v", err)
		timestampInt64 = time.Now().Unix()
	}

	var selfid64 int64
	if config.GetUseUin() {
		selfid64 = config.GetUinint64()
	} else {
		selfid64 = int64(p.Settings.AppID)
	}

	Notice := GroupNoticeEvent{
		GroupID:    GroupID64,
		NoticeType: "group_increase",
		OperatorID: 0, //官方事件不携带操作者
		PostType:   "notice",
		SelfID:     selfid64,
		SubType:    "approve", //官方无法区分加入方式, 默认approve
		Time:       timestampInt64,
		UserID:     userid64,
	}
	//增强配置
	if !config.GetNativeOb11() {
		Notice.RealUserID = data.MemberOpenID
		Notice.RealGroupID = data.GroupOpenID
	}

	noticeMap := structToMap(Notice)
	//上报信息到onebotv11应用端(正反ws)
	go p.BroadcastMessageToAll(noticeMap, p.Apiv2, data)

	mylog.Printf("群成员[%v]加入群[%v]", userid64, GroupID64)
	return nil
}