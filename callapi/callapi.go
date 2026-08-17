package callapi

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"encoding/json"
	"fmt"

	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/openapi"
)

// onebot发来的action调用信息
type ActionMessage struct {
	Action      string        `json:"action"`
	Params      ParamsContent `json:"params"`
	Echo        interface{}   `json:"echo,omitempty"`
	PostType    string        `json:"post_type,omitempty"`
	MessageType string        `json:"message_type,omitempty"`
}

func (a *ActionMessage) UnmarshalJSON(data []byte) error {
	type Alias ActionMessage

	var rawEcho json.RawMessage
	temp := &struct {
		*Alias
		Echo *json.RawMessage `json:"echo,omitempty"`
	}{
		Alias: (*Alias)(a),
		Echo:  &rawEcho,
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if rawEcho != nil {
		var lastErr error

		var intValue int
		if lastErr = json.Unmarshal(rawEcho, &intValue); lastErr == nil {
			a.Echo = intValue
			return nil
		}

		var strValue string
		if lastErr = json.Unmarshal(rawEcho, &strValue); lastErr == nil {
			a.Echo = strValue
			return nil
		}

		var arrValue []interface{}
		if lastErr = json.Unmarshal(rawEcho, &arrValue); lastErr == nil {
			a.Echo = arrValue
			return nil
		}

		var objValue map[string]interface{}
		if lastErr = json.Unmarshal(rawEcho, &objValue); lastErr == nil {
			a.Echo = objValue
			return nil
		}

		return fmt.Errorf("unable to unmarshal echo: %v", lastErr)
	}

	return nil
}

// params类型
type ParamsContent struct {
	BotQQ     string      `json:"botqq,omitempty"`
	ChannelID interface{} `json:"channel_id,omitempty"`
	GuildID   interface{} `json:"guild_id,omitempty"`
	GroupID   interface{} `json:"group_id,omitempty"`   // 每一种onebotv11实现的字段类型都可能不同
	MessageID interface{} `json:"message_id,omitempty"` // 用于撤回信息
	Message   interface{} `json:"message,omitempty"`    // 这里使用interface{}因为它可能是多种类型
	Messages  interface{} `json:"messages,omitempty"`   // 坑爹转发信息
	UserID    interface{} `json:"user_id,omitempty"`    // 这里使用interface{}因为它可能是多种类型
	Duration  int         `json:"duration,omitempty"`   // 可选的整数
	Enable    bool        `json:"enable,omitempty"`     // 可选的布尔值
	// handle quick operation
	Context      Context   `json:"context,omitempty"`       // context 字段
	Operation    Operation `json:"operation,omitempty"`     // operation 字段
	CallbackData string    `json:"callback_data,omitempty"` // 新增: 用于接收 GenerateURLLink 的参数

	// [新增] 群聊管理相关参数 (set_group_add_request / join_request_list / strategy 等)
	Approve                 bool                   `json:"approve,omitempty"`                    // 入群审批: 是否通过
	Refuse                  bool                   `json:"refuse,omitempty"`                     // 入群审批: 是否拒绝
	Reason                  string                 `json:"reason,omitempty"`                     // 入群审批: 拒绝理由
	Flag                    string                 `json:"flag,omitempty"`                       // 入群审批: 申请ID(join_request_id)
	Cursor                  string                 `json:"cursor,omitempty"`                     // 分页游标
	Limit                   int                    `json:"limit,omitempty"`                      // 分页数量
	StrategyID              string                 `json:"strategy_id,omitempty"`                // 策略ID
	GroupOpenIDs            []string               `json:"group_openids,omitempty"`              // 关联群 openid 列表
	GroupIDs                []string               `json:"group_ids,omitempty"`                  // 关联 QQ 群号列表
	IsEnable                string                 `json:"is_enable,omitempty"`                  // on-启用 off-关闭
	ExpireAt                string                 `json:"expire_at,omitempty"`                  // 过期时间 RFC3339
	Remark                  string                 `json:"remark,omitempty"`                     // 策略备注
	WhitelistUsers          []string               `json:"whitelist_users,omitempty"`            // 白名单QQ号码列表
	WhitelistOp             string                 `json:"whitelist_op,omitempty"`               // 白名单操作: add 新增号码, del 删除号码
	GroupActionOp           string                 `json:"group_action_op,omitempty"`            // 关联群操作: add/del
	GroupActionGroupOpenIDs []string               `json:"group_action_group_openids,omitempty"` // 关联群操作-openid 列表
	GroupActionGroupIDs     []string               `json:"group_action_group_ids,omitempty"`     // 关联群操作-群号列表
	Members                 []MemberMuteStateParam `json:"members,omitempty"`                    // 群禁言成员列表(原始扩展API用)
}

// MemberMuteStateParam 设置群成员禁言的单项参数
type MemberMuteStateParam struct {
	Op           string `json:"op"` // add 增加禁言, update 更新禁言到期时间, del 解除禁言
	MemberOpenID string `json:"member_openid"`
	MuteExpireAt string `json:"mute_expire_at,omitempty"`
}

// Context 结构体用于存储 context 字段相关信息
type Context struct {
	Avatar      string `json:"avatar,omitempty"`       // 用户头像链接
	Font        int    `json:"font,omitempty"`         // 字体（假设是整数类型）
	MessageID   int    `json:"message_id,omitempty"`   // 消息 ID
	MessageSeq  int    `json:"message_seq,omitempty"`  // 消息序列号
	MessageType string `json:"message_type,omitempty"` // 消息类型
	PostType    string `json:"post_type,omitempty"`    // 帖子类型
	SubType     string `json:"sub_type,omitempty"`     // 子类型
	Time        int64  `json:"time,omitempty"`         // 时间戳
	UserID      int    `json:"user_id,omitempty"`      // 用户 ID
	GroupID     int    `json:"group_id,omitempty"`     // 群号
}

// Operation 结构体用于存储 operation 字段相关信息
type Operation struct {
	Reply    string `json:"reply,omitempty"`     // 回复内容
	AtSender bool   `json:"at_sender,omitempty"` // 是否 @ 发送者
}

// 自定义一个ParamsContent的UnmarshalJSON 让GroupID同时兼容str和int
func (p *ParamsContent) UnmarshalJSON(data []byte) error {
	type Alias ParamsContent
	aux := &struct {
		GroupID   interface{} `json:"group_id"`
		UserID    interface{} `json:"user_id"`
		MessageID interface{} `json:"message_id"`
		ChannelID interface{} `json:"channel_id"`
		GuildID   interface{} `json:"guild_id"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.GroupID.(type) {
	case nil: // 当GroupID不存在时
		p.GroupID = ""
	case float64: // JSON的数字默认被解码为float64
		p.GroupID = fmt.Sprintf("%.0f", v) // 将其转换为字符串，忽略小数点后的部分
	case string:
		p.GroupID = v
	default:
		return fmt.Errorf("GroupID has unsupported type")
	}

	switch v := aux.UserID.(type) {
	case nil: // 当UserID不存在时
		p.UserID = ""
	case float64: // JSON的数字默认被解码为float64
		p.UserID = fmt.Sprintf("%.0f", v) // 将其转换为字符串，忽略小数点后的部分
	case string:
		p.UserID = v
	default:
		return fmt.Errorf("UserID has unsupported type")
	}

	switch v := aux.MessageID.(type) {
	case nil: // 当UserID不存在时
		p.MessageID = ""
	case float64: // JSON的数字默认被解码为float64
		p.MessageID = fmt.Sprintf("%.0f", v) // 将其转换为字符串，忽略小数点后的部分
	case string:
		p.MessageID = v
	default:
		return fmt.Errorf("MessageID has unsupported type")
	}

	switch v := aux.ChannelID.(type) {
	case nil: // 当ChannelID不存在时
		p.ChannelID = ""
	case float64: // JSON的数字默认被解码为float64
		p.ChannelID = fmt.Sprintf("%.0f", v) // 将其转换为字符串，忽略小数点后的部分
	case string:
		p.ChannelID = v
	default:
		return fmt.Errorf("MessageID has unsupported type")
	}

	switch v := aux.GuildID.(type) {
	case nil: // 当GuildID不存在时
		p.GuildID = ""
	case float64: // JSON的数字默认被解码为float64
		p.GuildID = fmt.Sprintf("%.0f", v) // 将其转换为字符串，忽略小数点后的部分
	case string:
		p.GuildID = v
	default:
		return fmt.Errorf("MessageID has unsupported type")
	}

	return nil
}

// Message represents a standardized structure for the incoming messages.
type Message struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
	Echo   interface{}            `json:"echo,omitempty"`
}

// 这是一个接口,在wsclient传入client但不需要引用wsclient包,避免循环引用,复用wsserver和client逻辑
type Client interface {
	SendMessage(message map[string]interface{}) error
}

// 为了解决processor和server循环依赖设计的接口
type WebSocketServerClienter interface {
	SendMessage(message map[string]interface{}) error
	Close() error
}

// 根据action订阅handler处理api
type HandlerFunc func(client Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, messgae ActionMessage) (string, error)

var handlers = make(map[string]HandlerFunc)

// RegisterHandler registers a new handler for a specific action.
func RegisterHandler(action string, handler HandlerFunc) {
	handlers[action] = handler
}

// CallAPIFromDict 处理信息 by calling the 对应的 handler.
func CallAPIFromDict(client Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message ActionMessage) string {
	handler, ok := handlers[message.Action]
	if !ok {
		// [修复] 未注册的 action 必须回一个合法的 onebot 错误响应, 不能沉默/返回空串。
		// 否则对端(如 SparkBridge)等不到响应会超时, 插件拿 undefined 直接 JSON5 解析崩溃。
		mylog.Println("Unsupported action:", message.Action)
		SendAPIError(client, message, "Unsupported action: "+message.Action)
		return ""
	}

	jsonString, err := handler(client, api, apiv2, message)
	if err != nil {
		// 处理错误: 同样回合法错误响应, 避免对端挂起
		mylog.Println("Error handling action:", message.Action, "Error:", err)
		SendAPIError(client, message, fmt.Sprintf("%s: %v", message.Action, err))
		return ""
	}

	return jsonString
}

// SendAPIError 向对端发送 onebot v11 标准错误响应 (retcode 1400)
// 未注册的 action / handler 出错 都会走这里, 保证任何请求都有合法 JSON 回应,
// 避免对端(SparkBridge 等)等待超时拿到 undefined 后 JSON5 解析崩溃。
func SendAPIError(client Client, message ActionMessage, errMsg string) {
	resp := map[string]interface{}{
		"status":  "failed",
		"retcode": 1400,
		"data": map[string]interface{}{
			"message": errMsg,
		},
	}
	if message.Echo != nil && message.Echo != "" {
		resp["echo"] = message.Echo
	}
	if err := client.SendMessage(resp); err != nil {
		mylog.Printf("发送错误响应失败: %v", err)
	}
}
