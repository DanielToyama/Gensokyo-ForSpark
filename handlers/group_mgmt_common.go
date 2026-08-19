package handlers

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)

import (
	"encoding/json"
	"strings"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/tencent-connect/botgo/dto"
)

// 群聊管理 handler 公共辅助函数

// [DanielToyama] BuildJoinRequestComment 计算入群申请的展示文本(onebot request.group 的 comment / 申请列表条目的 comment 增强)
// - 问答验证(admin_review_qa): 官方不回填 verify_message, 问题和用户回答在 review_qa_list, 拼为 "问:Q 答:A；问:Q 答:A"
// - 其余方式: 取 verify_message 原文
// - apply_source=invited(被邀请): 为空时兜底 "被邀请入群"
func BuildJoinRequestComment(method, verifyMessage, applySource string, qaList []dto.ReviewQA) string {
	comment := ""
	switch method {
	case "admin_review_qa":
		if len(qaList) > 0 {
			parts := make([]string, 0, len(qaList))
			for _, qa := range qaList {
				t := "问:" + qa.Question
				if qa.Answer != "" {
					t += " 答:" + qa.Answer
				}
				parts = append(parts, t)
			}
			comment = strings.Join(parts, "；")
		}
	default:
		comment = verifyMessage
	}
	if comment == "" && applySource == "invited" {
		comment = "被邀请入群"
	}
	return comment
}

// [DanielToyama] joinRequestToMap 把官方 JoinRequest 条目转为 map 并附加 comment(增强字段, 原字段不变)
func joinRequestToMap(jr *dto.JoinRequest) map[string]interface{} {
	m := map[string]interface{}{}
	if jr == nil {
		return m
	}
	if b, err := json.Marshal(jr); err == nil {
		_ = json.Unmarshal(b, &m)
	}
	method, verifyMsg := "", ""
	var qaList []dto.ReviewQA
	if jr.VerifyInfo != nil {
		method = jr.VerifyInfo.Method
		verifyMsg = jr.VerifyInfo.VerifyMessage
		qaList = jr.VerifyInfo.ReviewQAList
	}
	m["comment"] = BuildJoinRequestComment(method, verifyMsg, jr.ApplySource, qaList)
	return m
}

// groupResponse 构造 onebot v11 标准响应 map
func groupResponse(data interface{}, message callapi.ActionMessage) map[string]interface{} {
	resp := map[string]interface{}{
		"data":    data,
		"message": "success",
		"retcode": 0,
		"status":  "ok",
	}
	if message.Echo != nil && message.Echo != "" {
		resp["echo"] = message.Echo
	}
	return resp
}

// resolveGroupOpenID 通过 idmap 反查真实群 openid (int格式 -> openid)
func resolveGroupOpenID(groupID string) (string, error) {
	return idmap.RetrieveRowByIDv2(groupID)
}

// resolveUserOpenID 通过 idmap 反查真实用户 openid (int格式 -> openid)
func resolveUserOpenID(userID string) (string, error) {
	return idmap.RetrieveRowByIDv2(userID)
}