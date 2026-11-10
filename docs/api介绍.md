# API 介绍

目前支持的 API 列表（共 49 个动作，含 2 个兼容别名）。

具体 API 的定义请参考 [OneBot-11](https://github.com/botuniverse/onebot-11)。

> **本 fork 新增 / 专门适配的 API** 的参数、响应、官方接口与注意事项，详见 **[适配 API 清单](./适配API清单.md)**；
> 全部改动背景见 **[DanielToyama 更新记录](./DanielToyama更新记录.md)**。

## 一、本 Fork 新增或专门适配的 API

| 动作 | 实现文件 | 说明 |
|---|---|---|
| `/delete_msg` | delete_msg.go | 标准响应 + msgmap 反查（群/私聊/频道/频道私信）+ 支持撤回机器人自己发的消息 |
| `/get_group_bot_state` | get_group_bot_state.go | 获取机器人群内状态（官方 v2） |
| `/get_group_info` | get_group_info.go | 群类型返回官方真实群信息，失败回退模拟数据 |
| `/get_group_join_request_list` | get_group_join_request_list.go | 拉取入群申请列表（官方 v2），附加 `comment` 增强字段 |
| `/get_group_restrict_chat_setting` | get_group_restrict_chat_setting.go | 查询群禁言状态（官方 v2） |
| `/get_status` | get_status.go | 对齐 LLOneBot bot 状态结构（`data.online/good/stat` + `wording`） |
| `/get_stranger_info` | get_stranger_info.go | 昵称取 username 缓存；`sex/age` 占位；`qqLevel` 固定 9999 |
| `/set_group_add_request` | set_group_add_request.go | 入群审批，支持只传 `flag` 从事件缓存反查 |
| `/set_group_ban` | set_group_ban.go | 群聊成员禁言接入官方 `restrict_chat_setting`；`duration=0` 解禁 |
| `/get_group_ban` | set_group_ban.go | **兼容别名**，等价 `/set_group_ban`（上游早期误注册名） |
| `/set_group_whole_ban` | set_group_whole_ban.go | 群聊全员禁言（`global_rule.mode`，best-effort） |
| `/get_group_whole_ban` | set_group_whole_ban.go | **兼容别名**，等价 `/set_group_whole_ban` |
| `/set_group_restrict_chat_setting` | set_group_restrict_chat_setting.go | 设置群禁言（官方原始透传：`members[]` 或 `enable`） |
| `/create_join_approval_strategy` | join_approval_strategy.go | 创建入群自动审批策略 |
| `/get_join_approval_strategy_list` | join_approval_strategy.go | 查询入群自动审批策略列表 |
| `/update_join_approval_strategy` | join_approval_strategy.go | 修改入群自动审批策略 |
| `/delete_join_approval_strategy` | join_approval_strategy.go | 删除入群自动审批策略 |
| `/update_join_approval_strategy_whitelist` | join_approval_strategy.go | 修改策略白名单号码 |
| `/execute_join_approval_strategy` | join_approval_strategy.go | 执行策略（全量扫描，异步约 10 分钟） |
| `/send_group_msg`、`/send_msg`、`/send_private_msg`、`/send_guild_channel_msg`、`/send_group_msg_async`、`/send_group_msg_raw`、`/send_private_msg_async`、`/send_private_msg_sse`、`/send_private_msg_wakeup`、`/send_msg_async`、`/send_group_forward_msg` | send_*.go | 发送链路统一适配：at 转文本/真 at、FakeReply、主动消息、出站消息入库（供撤回） |

## 二、上游已有 API

| 动作 | 实现文件 |
|---|---|
| `/.handle_quick_operation` | handle_quick_operation.go |
| `/.handle_quick_operation_async` | handle_quick_operation_async.go |
| `/get_avatar` | get_avatar.go |
| `/get_friend_list` | get_friend_list.go |
| `/get_group_list` | get_group_list.go |
| `/get_group_member_info` | get_group_member_info.go |
| `/get_group_member_list` | get_group_member_list.go |
| `/get_guild_channel_list` | get_guild_channel_list.go |
| `/get_guild_list` | get_guild_list.go |
| `/get_guild_service_profile` | get_guild_service_profile.go |
| `/get_login_info` | get_login_info.go |
| `/get_online_clients` | get_online_clients.go |
| `/get_robot_share_link` | get_robot_share_link.go |
| `/get_version_info` | get_version_info.go |
| `/mark_msg_as_read` | mark_msg_as_read.go |
| `/put_interaction` | put_interaction.go |
| `/send_guild_channel_forum` | send_guild_channel_forum.go |
| `/send_guild_private_msg` | send_guild_private_msg.go |
| `/send_to_group` | send_group_msg.go |

## 三、相关文档

- [适配 API 清单](./适配API清单.md) —— 本 fork 新增/适配 API 的完整参数说明
- [撤回消息](./撤回消息.md) —— `delete_msg` 完整用法、四种会话对照、常见坑
- [DanielToyama 更新记录](./DanielToyama更新记录.md) —— 全部 45 笔提交与改动文件索引
