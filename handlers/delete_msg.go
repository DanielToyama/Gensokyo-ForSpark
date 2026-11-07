package handlers

// Modified by DanielToyama on 2026-08-20 (Gensokyo-ForSpark fork)
// 修复1: 响应结构误用 GetStatusResponse(serialize出 good/online/stat 的 get_status 数据),
//        改为标准 onebot v11 delete_msg 成功响应 data:null
// 修复2: 标准 delete_msg 仅传 message_id; 缺省 group_id/user_id 时从 msgmap 缓存
//        (入站消息入库时记录 GroupID/UserID/MsgType) 反查撤回对象, 不再"静默什么都不做"
//
// Modified by DanielToyama on 2026-09-02 (Gensokyo-ForSpark fork)
// 修复3: 让「撤回机器人自己发的消息」真正可用:
//   ① 配合 msgmap.RecordOutbound(出站发送成功后把 message_id→会话对象入库), 标准
//      delete_msg 只传 message_id 就能撤回自己发的消息 —— 原来 msgmap 只记录「收到的」
//      消息, 自己发出去的查不到, 四个分支全跳过、响应却依然 ok(应用端以为撤回了);
//   ② 反查扩展到 channel_id(频道)与 guild_id(频道私信), 不再只有群/私聊两条;
//   ③ message_id 缺失时不再 panic(原实现直接 .(string) 断言), 改为记录日志后走标准响应;
//      缓存里查不到真实 id 时回退用原始 message_id(兼容应用端直传官方 id 的情形)。

import (
	"context"
	"encoding/json"

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
	orignalMsgID := paramStr(message.Params.MessageID)
	if orignalMsgID == "" {
		mylog.Printf("delete_msg: 缺少 message_id, 无法撤回(响应仍按标准成功返回, 请检查应用端参数)")
	}

	// 不是stringob11才需要转换
	if !config.GetStringOb11() {
		// 如果从内存取
		if config.GetMemoryMsgid() {
			//还原msgid
			RealMsgID, _ = echo.GetCacheIDFromMemoryByRowID(orignalMsgID)
		} else {
			//还原msgid
			RealMsgID, err = idmap.RetrieveRowByCachev2(orignalMsgID)
			if err != nil {
				mylog.Printf("error retrieving real RChannelID: %v", err)
			}
		}
		// [修复3③] 缓存里没有映射时回退用原始 id, 避免把空 id 发给官方接口
		if RealMsgID == "" && orignalMsgID != "" {
			RealMsgID = orignalMsgID
			mylog.Printf("delete_msg: 缓存未找到 message_id[%v] 的映射, 回退使用原始 id", orignalMsgID)
		}
	} else {
		RealMsgID = orignalMsgID
	}

	// [修复3③] 缺 message_id 时不要继续拼空参数
	if RealMsgID == "" {
		return sendDeleteMsgResponse(client, message, "缺少 message_id")
	}

	//重新赋值
	message.Params.MessageID = RealMsgID

	//撤回频道信息(子频道)
	var channelTarget string
	if cid := paramStr(message.Params.ChannelID); cid != "" {
		channelTarget = cid
	} else if info, ok := msgmap.LookupMessage(orignalMsgID); ok && info.Outbound && info.MsgType == "guild" && info.GroupID != "" {
		// [修复3②] 标准 delete_msg 只传 message_id: 从 msgmap 反查频道(出站消息入库时记录)
		channelTarget = info.GroupID
		mylog.Printf("delete_msg: 缺省channel_id, 从msgmap反查到频道[%v]", info.GroupID)
	}
	if channelTarget != "" {
		// 使用RetrieveRowByIDv2还原真实的ChannelID
		RChannelID, rerr := idmap.RetrieveRowByIDv2(channelTarget)
		if rerr != nil {
			mylog.Printf("error retrieving real RChannelID: %v", rerr)
			RChannelID = channelTarget // 兜底: 反查失败时用原值(可能是官方 id 直传)
		}
		message.Params.ChannelID = RChannelID
		err = api.RetractMessage(context.TODO(), RChannelID, RealMsgID, openapi.RetractMessageOptionHidetip)
		if err != nil {
			mylog.Printf("Error retracting channel message: %v", err)
		}
	}

	//撤回频道私信
	var guildTarget string
	if gid := paramStr(message.Params.GuildID); gid != "" {
		guildTarget = gid
	} else if info, ok := msgmap.LookupMessage(orignalMsgID); ok && info.Outbound && info.MsgType == "guild_private" && info.GroupID != "" {
		// [修复3②] 缺省 guild_id: 从 msgmap 反查(出站消息入库时记的是真实 guild_id)
		// 这里很复杂 要取的话需要调用internal-api 根据情况还原，虚拟成群就用群（channel-id）还原完整channel-id，
		// 然后internal-api读配置获取guild-id ，虚拟成私信就用userid还原完整userid，然后读channel-id然后读guild-id
		// 因为GuildID本身不直接出现在ob11事件里。
		guildTarget = info.GroupID
		mylog.Printf("delete_msg: 缺省guild_id, 从msgmap反查到频道私信[%v]", info.GroupID)
	}
	if guildTarget != "" {
		err := api.RetractDMMessage(context.TODO(), guildTarget, RealMsgID, openapi.RetractMessageOptionHidetip)
		if err != nil {
			mylog.Printf("Error retracting DM message: %v", err)
		}
	}

	//撤回群信息
	var groupTarget string
	if gid := paramStr(message.Params.GroupID); gid != "" {
		groupTarget = gid
	} else if info, ok := msgmap.LookupMessage(orignalMsgID); ok && info.MsgType == "group" && info.GroupID != "" {
		// [修复2] 标准 delete_msg 只传 message_id: 从 msgmap 反查群对象
		// (入站记录 + 修复3的出站记录都覆盖, 所以撤自己的消息也能反查到)
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
		err = api.RetractGroupMessage(context.TODO(), groupTarget, RealMsgID, openapi.RetractMessageOptionHidetip)
		if err != nil {
			mylog.Printf("Error retracting group message: %v", err)
		}
	}

	//撤回C2C私信消息列表
	var userTarget string
	if uid := paramStr(message.Params.UserID); uid != "" {
		userTarget = uid
	} else if info, ok := msgmap.LookupMessage(orignalMsgID); ok && info.MsgType == "group_private" && info.UserID != "" {
		// [修复2] 私聊消息缺省 user_id: 从 msgmap 反查用户对象
		// [修复3②] 限定 MsgType == group_private, 避免频道记录(带 user_id)被误判为 C2C
		userTarget = info.UserID
		mylog.Printf("delete_msg: 缺省user_id, 从msgmap反查到用户[%v]", info.UserID)
	}
	if userTarget != "" {
		var UserID string
		//还原真实的userid
		UserID, err = idmap.RetrieveRowByIDv2(userTarget)
		if err != nil {
			mylog.Printf("Error reading config: %v", err)
			UserID = userTarget // 兜底: 反查失败时用原值
		}
		message.Params.UserID = UserID
		err = api.RetractC2CMessage(context.TODO(), UserID, RealMsgID, openapi.RetractMessageOptionHidetip)
		if err != nil {
			mylog.Printf("Error retracting C2C message: %v", err)
		}
	}

	return sendDeleteMsgResponse(client, message, "")
}

// sendDeleteMsgResponse [修复1] 标准 onebot v11 成功响应: data 为 null
// (原实现误用 GetStatusResponse, 返回了 get_status 的 good/online/stat 数据,
// 会让应用端误以为收到了 bot 状态)
//
// 说明: 官方撤回接口失败只在日志里打印(见各分支 mylog.Printf), 响应结构保持标准不变,
// 应用端若要判断成败需自行检查日志或改用返回内容。
func sendDeleteMsgResponse(client callapi.Client, message callapi.ActionMessage, warnMsg string) (string, error) {
	outputMap := map[string]interface{}{
		"data":    nil,
		"message": warnMsg,
		"retcode": 0,
		"status":  "ok",
	}
	if message.Echo != nil {
		outputMap["echo"] = message.Echo
	}

	mylog.Printf("delete_msg: %+v\n", outputMap)

	if err := client.SendMessage(outputMap); err != nil {
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
