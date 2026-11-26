package handlers

// 图文卡片消息 (msg_type=8, 2026-10-28)
// 官方: 群聊 POST /v2/groups/{group_openid}/messages, 单聊 POST /v2/users/{user_openid}/messages
// card.type=tuwen, content 需包含 title/description/url/pic_url, 其中 url/pic_url 必须 https

import (
	"context"
	"fmt"
	"strings"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/config"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("send_group_card_msg", HandleSendGroupCardMsg)
	callapi.RegisterHandler("send_private_card_msg", HandleSendPrivateCardMsg)
}

// HandleSendGroupCardMsg 发送群图文卡片
// 参数: group_id, title, description, url, pic_url, message_id(可选, 被动回复)
func HandleSendGroupCardMsg(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	groupID, _ := message.Params.GroupID.(string)
	if groupID == "" {
		callapi.SendAPIError(client, message, "send_group_card_msg: 缺少 group_id")
		return "", nil
	}
	card, err := buildCard(message)
	if err != nil {
		callapi.SendAPIError(client, message, "send_group_card_msg: "+err.Error())
		return "", nil
	}
	realGroupID, err := resolveGroupOpenID(groupID)
	if err != nil || realGroupID == "" {
		callapi.SendAPIError(client, message, "send_group_card_msg: 无法反查群 openid")
		return "", nil
	}

	msg := &dto.MessageToCreate{
		MsgType: 8,
		Card:    card,
		MsgID:   resolveReplyMsgID(message.Params.MessageID),
	}
	if msg.MsgID != "" {
		msg.MsgSeq = 1
	}
	resp, err := apiv2.PostGroupMessage(context.TODO(), realGroupID, msg)
	if err != nil {
		mylog.Printf("send_group_card_msg: 发送失败: %v", err)
		callapi.SendAPIError(client, message, fmt.Sprintf("send_group_card_msg: %v", err))
		return "", nil
	}
	messageID := ""
	if resp != nil && resp.Message != nil {
		messageID = resp.Message.ID
	}
	return sendGroupMgmtResponse(client, map[string]interface{}{"message_id": messageID}, message)
}

// HandleSendPrivateCardMsg 发送单聊图文卡片
// 参数: user_id, title, description, url, pic_url, message_id(可选, 被动回复)
func HandleSendPrivateCardMsg(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	ids := memberIDStrings(message.Params.UserID)
	if len(ids) == 0 || ids[0] == "" {
		callapi.SendAPIError(client, message, "send_private_card_msg: 缺少 user_id")
		return "", nil
	}
	card, err := buildCard(message)
	if err != nil {
		callapi.SendAPIError(client, message, "send_private_card_msg: "+err.Error())
		return "", nil
	}
	realUserID, err := resolveUserOpenID(ids[0])
	if err != nil || realUserID == "" {
		callapi.SendAPIError(client, message, "send_private_card_msg: 无法反查用户 openid")
		return "", nil
	}

	msg := &dto.MessageToCreate{
		MsgType: 8,
		Card:    card,
		MsgID:   resolveReplyMsgID(message.Params.MessageID),
	}
	if msg.MsgID != "" {
		msg.MsgSeq = 1
	}
	resp, err := apiv2.PostC2CMessage(context.TODO(), realUserID, msg)
	if err != nil {
		mylog.Printf("send_private_card_msg: 发送失败: %v", err)
		callapi.SendAPIError(client, message, fmt.Sprintf("send_private_card_msg: %v", err))
		return "", nil
	}
	messageID := ""
	if resp != nil && resp.Message != nil {
		messageID = resp.Message.ID
	}
	return sendGroupMgmtResponse(client, map[string]interface{}{"message_id": messageID}, message)
}

func buildCard(message callapi.ActionMessage) (*dto.Card, error) {
	title := paramStr(message.Params.Title)
	desc := paramStr(message.Params.Description)
	link := paramStr(message.Params.URL)
	pic := paramStr(message.Params.PicURL)
	if title == "" || desc == "" || link == "" || pic == "" {
		return nil, fmt.Errorf("title/description/url/pic_url 均不能为空")
	}
	if !strings.HasPrefix(strings.ToLower(link), "https://") || !strings.HasPrefix(strings.ToLower(pic), "https://") {
		return nil, fmt.Errorf("url 与 pic_url 必须为 https")
	}
	return &dto.Card{
		Type: "tuwen",
		Content: dto.CardContent{
			Title:       title,
			Description: desc,
			URL:         link,
			PicURL:      pic,
		},
	}, nil
}

// resolveReplyMsgID 应用端 message_id -> 官方 msg_id
// 2026-11-17 起也支持群消息的 msg_id 用于单聊被动回复
func resolveReplyMsgID(v interface{}) string {
	mid := paramStr(v)
	if mid == "" {
		return ""
	}
	if !config.GetStringOb11() {
		if realID, err := idmap.RetrieveRowByCachev2(mid); err == nil && realID != "" {
			return realID
		}
	}
	return mid
}
