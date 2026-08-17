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

// ProcessGroupMemberRemove 处理群成员退出事件 (GROUP_MEMBER_REMOVE, Intent GROUP_MEMBER_EVENT 1<<24)
// 转换成 onebot v11 notice.group_decrease
// 说明: 官方事件不携带操作者信息, 无法区分 leave/kick, 默认 sub_type=leave
func (p *Processors) ProcessGroupMemberRemove(data *dto.GroupMemberRemoveEvent) error {
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
		mylog.Printf("ProcessGroupMemberRemove: 时间戳转换失败: %v", err)
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
		NoticeType: "group_decrease",
		OperatorID: 0, //官方事件不携带操作者
		PostType:   "notice",
		SelfID:     selfid64,
		SubType:    "leave", //官方无法区分退出方式, 默认leave
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

	mylog.Printf("群成员[%v]退出群[%v]", userid64, GroupID64)
	return nil
}