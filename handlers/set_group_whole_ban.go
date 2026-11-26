package handlers

// Modified by DanielToyama on 2026-08-17 (Gensokyo-ForSpark fork)
//
// OneBot v11 标准动作 set_group_whole_ban（同时注册历史别名 get_group_whole_ban）。
//
// 官方限制: POST /v2/groups/{group_openid}/restrict_chat_setting 只声明 members
// (成员级禁言), 不提供 global_rule 写入; global_rule 只能在查询接口读到。
// 因此实现策略:
//   1. 先尝试 global_rule (平台后续若开放则直接使用, 未来新入群成员也被禁言);
//   2. 失败后:
//      - 默认(whole_ban_via_members=false): 直接返回失败并说明官方未开放该能力, 不做任何批量写入;
//      - 实验性开关开启时: 降级为"按成员批量禁言" —— 拉取群成员列表 -> 过滤群主/管理员/机器人
//        -> 每 20 个一批调用 members[].op=add; 解禁则从查询结果里取当前被禁言成员批量 op=del。
//        注意: 降级实现只能处理"当前已在群内"的成员, 之后新入群的成员不会被自动禁言。

import (
	"context"
	"fmt"
	"time"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/config"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

const (
	// wholeBanBatchSize 官方 members 单次上限 20
	wholeBanBatchSize = 20
	// wholeBanDefaultDuration 降级实现未显式指定 duration 时的禁言时长(7 天)
	wholeBanDefaultDuration = 7 * 24 * 3600
)

func init() {
	// onebot v11 标准动作名, 同时注册旧名称兼容
	callapi.RegisterHandler("set_group_whole_ban", SetGroupWholeBan)
	callapi.RegisterHandler("get_group_whole_ban", SetGroupWholeBan)
}

func SetGroupWholeBan(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	groupID := paramStr(message.Params.GroupID)
	if groupID == "" {
		callapi.SendAPIError(client, message, "set_group_whole_ban: 缺少 group_id")
		return "", nil
	}

	msgType, err := idmap.ReadConfigv2(groupID, "type")
	if err != nil {
		mylog.Printf("setGroupWholeBan: 读取群类型失败(%v), 按群处理", err)
		msgType = "group"
	}

	enable := message.Params.Enable
	if !message.Params.EnableSet {
		enable = true // OneBot v11 默认开启全员禁言
	}

	switch msgType {
	case "group":
		realGroupID, err := resolveGroupOpenID(groupID)
		if err != nil || realGroupID == "" {
			callapi.SendAPIError(client, message, "set_group_whole_ban: 无法反查群 openid")
			return "", nil
		}

		mode := "none"
		if enable {
			mode = "always"
		}
		// 步骤 1: 官方 global_rule (当前文档未声明可写, 失败则降级)
		if err := apiv2.SetRestrictChatSetting(context.TODO(), realGroupID, &dto.SetRestrictChatSettingToCreate{
			GlobalRule: &dto.GlobalMuteRule{Mode: mode},
		}); err == nil {
			mylog.Printf("setGroupWholeBan(群): 官方 global_rule 设置成功 group[%v] mode[%v]", realGroupID, mode)
			return sendGroupMgmtResponse(client, nil, message)
		} else {
			mylog.Printf("setGroupWholeBan(群): 官方 global_rule 不支持写入(%v)", err)
		}

		// 步骤 2: 降级实现仅在实验性开关开启时执行, 避免默认触发批量写操作
		if !config.GetWholeBanViaMembers() {
			callapi.SendAPIError(client, message,
				"set_group_whole_ban: 官方未开放全员禁言写入接口(restrict_chat_setting 仅支持成员级禁言); "+
					"如需实验性降级(批量禁言/解禁群成员)请在 config.yml 开启 whole_ban_via_members, 风险自负")
			return "", nil
		}

		mylog.Printf("setGroupWholeBan(群): [实验性] whole_ban_via_members 已开启, 执行成员级批量降级")

		if enable {
			duration := message.Params.Duration
			if !message.Params.DurationSet || duration <= 0 {
				duration = wholeBanDefaultDuration
			}
			muted, skipped, err := muteAllGroupMembers(apiv2, realGroupID, duration)
			if err != nil {
				mylog.Printf("setGroupWholeBan(群): 降级全员禁言失败: %v", err)
				callapi.SendAPIError(client, message, "set_group_whole_ban: "+err.Error())
				return "", nil
			}
			mylog.Printf("setGroupWholeBan(群): 已禁言 %d 人, 跳过群主/管理员/机器人 %d 人, 时长 %ds", muted, skipped, duration)
		} else {
			unmuted, err := unmuteAllGroupMembers(apiv2, realGroupID)
			if err != nil {
				mylog.Printf("setGroupWholeBan(群): 降级解除全员禁言失败: %v", err)
				callapi.SendAPIError(client, message, "set_group_whole_ban: "+err.Error())
				return "", nil
			}
			mylog.Printf("setGroupWholeBan(群): 已解除 %d 人的禁言", unmuted)
		}

	case "private":
		callapi.SendAPIError(client, message, "set_group_whole_ban: 私聊虚拟群场景暂未适配全员禁言")
		return "", nil

	case "guild":
		guildID, err := idmap.ReadConfigv2(groupID, "guild_id")
		if err != nil {
			callapi.SendAPIError(client, message, "set_group_whole_ban: 读取频道 guild_id 失败")
			return "", nil
		}
		duration := "0"
		if enable {
			duration = "604800" // OneBot 全员禁言只有开/关, 频道场景先按 7 天
		}
		mute := &dto.UpdateGuildMute{MuteSeconds: duration}
		if err := api.GuildMute(context.TODO(), guildID, mute); err != nil {
			mylog.Printf("Error setting whole guild mute: %v", err)
			callapi.SendAPIError(client, message, "set_group_whole_ban: "+err.Error())
			return "", nil
		}

	default:
		callapi.SendAPIError(client, message, "set_group_whole_ban: 未知会话类型 "+msgType)
		return "", nil
	}

	// OneBot v11 标准成功响应 (data: null)
	return sendGroupMgmtResponse(client, nil, message)
}

// muteAllGroupMembers 拉取全部群成员并分批禁言(跳过群主/管理员/机器人)
// 返回: 已提交禁言人数, 跳过人数, 错误
func muteAllGroupMembers(apiv2 openapi.OpenAPI, groupOpenID string, duration int) (int, int, error) {
	members, err := fetchAllGroupMembers(apiv2, groupOpenID)
	if err != nil {
		return 0, 0, fmt.Errorf("拉取群成员列表失败(需官方 2026-09-03 内邀权限): %v", err)
	}
	if len(members) == 0 {
		return 0, 0, fmt.Errorf("群成员列表为空, 无法执行全员禁言")
	}

	expireAt := time.Now().Add(time.Duration(duration) * time.Second).Format(time.RFC3339)
	batch := make([]*dto.SetMemberMuteState, 0, wholeBanBatchSize)
	muted, skipped := 0, 0

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		setting := &dto.SetRestrictChatSettingToCreate{Members: batch}
		if err := apiv2.SetRestrictChatSetting(context.TODO(), groupOpenID, setting); err != nil {
			return err
		}
		muted += len(batch)
		batch = batch[:0]
		return nil
	}

	for _, m := range members {
		if m == nil || m.MemberOpenID == "" {
			continue
		}
		// 官方不允许禁言群主/管理员/机器人
		if m.Bot || m.MemberRole == "owner" || m.MemberRole == "admin" {
			skipped++
			continue
		}
		batch = append(batch, &dto.SetMemberMuteState{
			Op:           "add",
			MemberOpenID: m.MemberOpenID,
			MuteExpireAt: expireAt,
		})
		if len(batch) >= wholeBanBatchSize {
			if err := flush(); err != nil {
				return muted, skipped, fmt.Errorf("批量禁言失败(已成功 %d 人): %v", muted, err)
			}
		}
	}
	if err := flush(); err != nil {
		return muted, skipped, fmt.Errorf("批量禁言失败(已成功 %d 人): %v", muted, err)
	}
	return muted, skipped, nil
}

// unmuteAllGroupMembers 查询当前被禁言成员并分批解禁
func unmuteAllGroupMembers(apiv2 openapi.OpenAPI, groupOpenID string) (int, error) {
	setting, err := apiv2.GetRestrictChatSetting(context.TODO(), groupOpenID)
	if err != nil {
		return 0, fmt.Errorf("查询群禁言状态失败: %v", err)
	}
	if setting == nil || len(setting.Members) == 0 {
		return 0, nil
	}

	batch := make([]*dto.SetMemberMuteState, 0, wholeBanBatchSize)
	unmuted := 0
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := apiv2.SetRestrictChatSetting(context.TODO(), groupOpenID, &dto.SetRestrictChatSettingToCreate{Members: batch}); err != nil {
			return err
		}
		unmuted += len(batch)
		batch = batch[:0]
		return nil
	}

	for _, m := range setting.Members {
		if m.MemberOpenID == "" {
			continue
		}
		batch = append(batch, &dto.SetMemberMuteState{
			Op:           "del",
			MemberOpenID: m.MemberOpenID,
			MuteExpireAt: "",
		})
		if len(batch) >= wholeBanBatchSize {
			if err := flush(); err != nil {
				return unmuted, fmt.Errorf("批量解禁失败(已成功 %d 人): %v", unmuted, err)
			}
		}
	}
	if err := flush(); err != nil {
		return unmuted, fmt.Errorf("批量解禁失败(已成功 %d 人): %v", unmuted, err)
	}
	return unmuted, nil
}
