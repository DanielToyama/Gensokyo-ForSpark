package handlers

// Modified by DanielToyama on 2026-08-18 (Gensokyo-ForSpark fork)

import (
	"encoding/json"

	"github.com/hoshinonyaruko/gensokyo/botstats"
	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/openapi"
)

// GetStatusResponse get_status 响应, 对齐 LLOneBot 文档(OneBot 11/接口列表/系统/bot状态):
// 顶层字段 status/retcode/data/message/wording 全部必填
type GetStatusResponse struct {
	Data    StatusData  `json:"data"`
	Message string      `json:"message"`
	RetCode int         `json:"retcode"`
	Status  string      `json:"status"`
	Wording string      `json:"wording"`
	Echo    interface{} `json:"echo"`
}

// StatusData 对齐 LLOneBot 文档: data.online(是否在线) / data.good(状态是否良好) / data.stat(运行统计)
type StatusData struct {
	Online bool       `json:"online"`
	Good   bool       `json:"good"`
	Stat   Statistics `json:"stat"`
}

// Statistics 对齐 LLOneBot 文档: data.stat 运行统计(全部必填)
type Statistics struct {
	MessageReceived int   `json:"message_received"` // 接收信息总数
	MessageSent     int   `json:"message_sent"`     // 发送信息总数
	LastMessageTime int64 `json:"last_message_time"` // 最后一条消息时间(Unix秒)
	StartupTime     int64 `json:"startup_time"`     // 启动时间(Unix秒)
}

func init() {
	callapi.RegisterHandler("get_status", GetStatus)
}

func GetStatus(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {

	var response GetStatusResponse

	// 真实统计: 收/发消息总数与最后一条消息时间来自 botstats(持久化 db), 启动时间来自进程启动时刻
	messageReceived, messageSent, lastMessageTime, err := botstats.GetStats()
	if err != nil {
		mylog.Printf("get_status错误,获取机器人发信状态错误:%v", err)
	}
	response.Data = StatusData{
		// 能收到该请求即说明机器人连接正常, online/good 直接为 true
		Online: true,
		Good:   true,
		Stat: Statistics{
			MessageReceived: messageReceived,          // 实际数据
			MessageSent:     messageSent,              // 实际数据
			LastMessageTime: lastMessageTime,          // 实际数据
			StartupTime:     botstats.GetStartupTime(), // 实际数据
		},
	}
	response.Message = ""
	response.RetCode = 0
	response.Status = "ok"
	response.Wording = ""
	response.Echo = message.Echo

	outputMap := structToMap(response)

	mylog.Printf("get_status: %+v\n", outputMap)

	err = client.SendMessage(outputMap)
	if err != nil {
		mylog.Printf("Error sending message via client: %v", err)
	}
	//把结果从struct转换为json
	result, err := json.Marshal(response)
	if err != nil {
		mylog.Printf("Error marshaling data: %v", err)
		//todo 符合onebotv11 ws返回的错误码
		return "", nil
	}
	return string(result), nil
}