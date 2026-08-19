package handlers

// Modified by DanielToyama on 2026-08-20 (Gensokyo-ForSpark fork)
// 修复1: 响应结构误用 GetStatusResponse(serialize出 good/online/stat 的 get_status 数据),
//        改为标准 onebot v11 delete_msg 成功响应 data:null
// 修复2: 标准 delete_msg 仅传 message_id; 缺省 group_id/user_id 时从 msgmap 缓存
//        (入站消息入库时记录 GroupID/UserID/MsgType) 反查撤回对象, 不再"静默什么都不做"

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/config"
	"github.com/hoshinonyaruko/gensokyo/echo"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/msgmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("delete_msg", DeleteMsg)
}

func DeleteMsg(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var RealMsgID string
	var err error

	// 应用端传入的原始 message_id, 还原前保留(用于 msgmap 反查撤回对象)
	orignalMsgID := message.Params.MessageID.(string)

	// 不是stringob才需要转换
	if !config.GetStringOb11() {
		// 如果从内存取
		if config.GetMemoryMsgid() {
			//还原msgid
			RealMsgID, _ = echo.GetCacheIDFromMemoryByRowID(message.Params.MessageID.(string))
		} else {
			//还原msgid
			RealMsgID, err = idmap.RetrieveRowByCachev2(message.Params.MessageID.(string))
			if err != nil {
				mylog.Printf("error retrieving real RChannelID: %v", err)
			}
		}
	} else {
		RealMsgID = message.Params.MessageID.(string)
	}

	//重新赋值
	message.Params.MessageID = RealMsgID
	//撤回频道信息
	if message.Params.ChannelID != nil && message.Params.ChannelID != "" {
		var RChannelID string
		var err error
		// 使用RetrieveRowByIDv2还原真实的ChannelID
		RChannelID, err = idmap.RetrieveRowByIDv2(message.Params.ChannelID.(string))
		if err != nil {
			mylog.Printf("error retrieving real RChannelID: %v", err)
		}
		message.Params.ChannelID = RChannelID
		err = api.RetractMessage(context.TODO(), message.Params.ChannelID.(string), message.Params.MessageID.(string), openapi.RetractMessageOptionHidetip)
		if err != nil {
			fmt.Println("Error retracting channel message:", err)
		}

	}

	//撤回频道私信
	if message.Params.GuildID != nil && message.Params.GuildID != "" {
		//这里很复杂 要取的话需要调用internal-api 根据情况还原，虚拟成群就用群（channel-id）还原完整channel-id，
		//然后internal-api读配置获取guild-id ，虚拟成私信就用userid还原完整userid，然后读channel-id然后读guild-id
		//因为GuildID本身不直接出现在ob11事件里。
		err := api.RetractDMMessage(context.TODO(), message.Params.GuildID.(string), message.Params.MessageID.(string), openapi.RetractMessageOptionHidetip)
		if err != nil {
			fmt.Println("Error retracting DM message:", err)
		}

	}

	//撤回群信息
	var groupTarget string
	if message.Params.GroupID != nil && message.Params.GroupID != "" {
		groupTarget = message.Params.GroupID.(string)
	} else if info, ok := msgmap.LookupMessage(orignalMsgID); ok && info.MsgType == "group" && info.GroupID != "" {
		// [修复] 标准 delete_msg 只传 message_id: 从 msgmap 反查群对象
		groupTarget = info.GroupID
		mylog.Printf("delete_msg: 缺省group_id, 从msgmap反查到群[%v]", info.GroupID)
	}
	if groupTarget != "" {
		// 判断是否是原始id
		if len(groupTarget) != 32 {
			var originalGroupID string
			originalGroupID, err := idmap.RetrieveRowByIDv2(groupTarget)
			if err != nil {
				mylog.Printf("Error retrieving original GroupID: %v", err)
			}
			groupTarget = originalGroupID
		}
		err = api.RetractGroupMessage(context.TODO(), groupTarget, message.Params.MessageID.(string), openapi.RetractMessageOptionHidetip)
		if err != nil {
			fmt.Println("Error retracting group message:", err)
		}
	}

	//撤回C2C私信消息列表
	var userTarget string
	if message.Params.UserID != nil && message.Params.UserID != "" {
		userTarget = message.Params.UserID.(string)
	} else if info, ok := msgmap.LookupMessage(orignalMsgID); ok && info.MsgType != "group" && info.UserID != "" {
		// [修复] 私聊消息缺省 user_id: 从 msgmap 反查用户对象
		userTarget = info.UserID
		mylog.Printf("delete_msg: 缺省user_id, 从msgmap反查到用户[%v]", info.UserID)
	}
	if userTarget != "" {
		var UserID string
		//还原真实的userid
		UserID, err := idmap.RetrieveRowByIDv2(userTarget)
		if err != nil {
			mylog.Printf("Error reading config: %v", err)
			return "", nil
		}
		message.Params.UserID = UserID
		err = api.RetractC2CMessage(context.TODO(), message.Params.UserID.(string), message.Params.MessageID.(string), openapi.RetractMessageOptionHidetip)
		if err != nil {
			fmt.Println("Error retracting C2C message:", err)
		}

	}

	// [修复] 标准 onebot v11 成功响应: data 为 null (原实现误用 GetStatusResponse, 返回了
	// get_status 的 good/online/stat 数据, 会让应用端误以为收到了 bot 状态)
	outputMap := map[string]interface{}{
		"data":    nil,
		"message": "",
		"retcode": 0,
		"status":  "ok",
	}
	if message.Echo != nil {
		outputMap["echo"] = message.Echo
	}

	mylog.Printf("delete_msg: %+v\n", outputMap)

	err = client.SendMessage(outputMap)
	if err != nil {
		mylog.Printf("Error sending message via client: %v", err)
	}
	//把结果从struct转换为json
	result, err := json.Marshal(outputMap)
	if err != nil {
		mylog.Printf("Error marshaling data: %v", err)
		return "", nil
	}
	return string(result), nil
}