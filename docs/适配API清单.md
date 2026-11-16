# 适配 API 清单（Gensokyo-ForSpark Fork）

> 本文档只记录 **DanielToyama 在本 fork 中新增或专门适配** 的 OneBot API / 事件 / HTTP 路由，不重复上游已有能力。
>
> 对应实现文件均在 `handlers/`、`Processor/`、`server/`；完整改动背景见 [DanielToyama更新记录.md](./DanielToyama更新记录.md)。
>
> **通用约定**
> - 请求/响应遵循 OneBot v11 结构：`{"action":"...","params":{...},"echo":"..."}`。
> - 应用端看到的 `group_id` / `user_id` 默认是 Gensokyo 的 **idmap 压缩数字 ID**，内部会反查官方 openid；`string_ob11: true` 时直接使用官方原始 ID。
> - 成功响应统一为 `{"status":"ok","retcode":0,"data":...,"message":"success","echo":...}`；失败经 `callapi.SendAPIError` 返回 `{"status":"failed","retcode":1400,"data":{"message":"..."},"echo":...}`。
> - 群管理类接口均要求机器人是 **群管理员**，否则官方接口返回权限错误（仅记录日志）。

---

## 一、新增拓展 API

### 1. 入群审批

#### `set_group_add_request` — 处理加群请求（审批入群申请）

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 否* | 群 ID；与 `user_id` 同时提供时走常规反查 |
| `user_id` | string/int | 否* | 申请人 ID |
| `flag` | string | 是 | 入群申请 ID（`join_request_id`），来自 `request.group` 事件 |
| `approve` | bool | 否 | `true` = 通过；不传/`false` = 拒绝（兼容只传 `approve` 的 SparkBridge groupRequest 插件） |
| `refuse` | bool | 否 | 兼容字段：`true` = 拒绝；不传 `approve` 时同样按拒绝处理 |
| `reason` | string | 否 | 拒绝理由 |

\* `group_id`/`user_id` 可省略：Gensokyo 会用 `flag` 从事件缓存（`ProcessGroupJoinRequest` 写入的 `join_request_id → 群/成员 openid`）反查。

- 官方接口：`POST /v2/groups/{group_openid}/approval_join_request/{member_openid}`
- 响应：`{"status":"ok","retcode":0,"data":{}}`
- 注意：只传 `refuse=true` 而不传 `approve` 时同样按拒绝处理；官方审批不可撤回。

#### `get_group_join_request_list` — 拉取入群申请列表

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID |
| `cursor` | string | 否 | 分页游标，首页留空 |
| `limit` | int | 否 | 单页数量，默认 20，最大 100 |

- 官方接口：`GET /v2/groups/{group_openid}/join_request_list`
- 响应 `data`：`{"list":[...],"next_cursor":"..."}`；每条申请在官方字段之外附加 **`comment`** 增强字段（问答验证/验证消息的统一展示文本）。

### 2. 群禁言

#### `set_group_ban` — 群成员禁言 / 解禁（群聊分支）

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID |
| `user_id` | string/int | 是 | 被禁言成员 ID |
| `duration` | int | 是 | 禁言秒数；`0` = 解禁；官方最大 30 天 |

- 官方接口：`POST /v2/groups/{group_openid}/restrict_chat_setting`（`members[].op=add/del`）
- 说明：本 fork 补上了原实现从未执行的群聊分支；频道分支沿用上游 `MemberMute`。

#### `set_group_whole_ban` — 全员禁言开关

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID |
| `enable` | bool | 是 | `true` = 全员禁言（`global_rule.mode=always`）；`false` = 解除（`none`） |

- 官方接口同上（`global_rule.mode`）。
- 注意：官方文档当前只声明 `members` 可写，`global_rule` 为 **best-effort 透传**，平台不支持时会返回错误（仅日志）。

#### `get_group_restrict_chat_setting` — 查询群禁言状态

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID |

- 官方接口：`GET /v2/groups/{group_openid}/restrict_chat_setting`
- 响应 `data`：`global_rule`（全员禁言模式）+ `members`（成员级禁言列表）。

#### `set_group_restrict_chat_setting` — 设置群禁言（官方原始透传）

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID |
| `members` | array | 否 | 成员级禁言列表，单项 `{op, member_openid, mute_expire_at}`；`op` = `add`/`update`/`del` |
| `enable` | bool | 否 | 未传 `members` 时视为全员禁言开关：`true`=`always`，`false`=`none` |

- 官方接口：`POST /v2/groups/{group_openid}/restrict_chat_setting`
- 约束：成员级禁言单次最多 10 个，最大禁言时长 30 天；`mute_expire_at` 为 RFC3339。

### 3. 机器人群内状态

#### `get_group_bot_state` — 获取机器人群内状态

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID |

- 官方接口：`GET /v2/groups/{group_openid}/bot_state`
- 响应 `data`：`member_openid`、`joined_at`、`allow_proactive_msg`、`recv_msg_setting`、`member_role`。

### 4. 入群自动审批策略（官方 API v2 20260810 新增）

官方接口族：`/v2/groups/join_approval_strategy[/{strategy_id}[/whitelist_users|/execute]]`

| 动作 | 参数 | 说明 |
|---|---|---|
| `create_join_approval_strategy` | `group_openids` 或 `group_ids`（二选一）、`is_enable`（`on`/`off`，默认 `on`）、`expire_at`（RFC3339，默认一年）、`remark` | 创建策略，返回策略信息（含 `strategy_id`） |
| `get_join_approval_strategy_list` | `cursor`、`limit`（默认 20，最大 100） | 查询策略列表 |
| `update_join_approval_strategy` | `strategy_id`、`is_enable`、`expire_at`、`remark`、`group_action_op`（`add`/`del`）、`group_action_group_openids` / `group_action_group_ids` | 修改策略；关联群变更用 `group_action_*` |
| `delete_join_approval_strategy` | `strategy_id` | 删除策略 |
| `update_join_approval_strategy_whitelist` | `strategy_id`、`whitelist_op`（`add`/`del`）、`whitelist_users`（QQ 号数组） | 修改白名单号码 |
| `execute_join_approval_strategy` | `strategy_id` | 立即全量扫描执行（异步，约 10 分钟） |

- 响应：创建/查询/修改返回官方 `data`；删除/执行返回 `{"time": Unix秒}`。
- 说明：`group_openids` 与 `group_ids` 二选一；白名单号码是真实 QQ 号，不是 openid。

### 5. 陌生人信息

#### `get_stranger_info` — 获取陌生人信息

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `user_id` | string/int | 是 | 用户 ID |

- 响应 `data`：`{"user_id":<int>,"nickname":"<缓存昵称>","sex":"","age":0,"qqLevel":9999}`
- 背景：官方 API 没有陌生人资料/QQ 等级接口。`nickname` 来自 `idmap` 的用户名缓存（每条群消息、加群申请会缓存官方 `username`，7 天有效）；`sex/age` 返回占位；`qqLevel` 固定 `9999`，用于放行 SparkBridge groupRequest 等插件的等级门槛（插件把 0 级当「隐私未公开」直接拒绝）。

---

## 二、标准 OneBot API 的适配与增强

| API | 适配内容 |
|---|---|
| `delete_msg` | ① 响应改为标准 `data:null`（原误用 `get_status` 数据）；② 只传 `message_id` 时从 `msgmap` 反查撤回对象，支持群 / 私聊 / 频道 / 频道私信；③ 出站消息入库后**可撤回机器人自己发的消息**；④ `message_id` 缺失不再 panic，idmap 反查失败回退原值。详见 [撤回消息.md](./撤回消息.md) |
| `set_group_add_request` | 支持只传 `flag`，从入群申请事件缓存反查群/成员 openid；`approve` 语义对齐 OneBot（`true` 通过，否则拒绝）。见上文 |
| `set_group_ban` | 补上群聊分支：官方 `restrict_chat_setting` 成员级禁言；`duration=0` 解禁。同时注册标准动作名（旧版误注册为 `get_group_ban`） |
| `set_group_whole_ban` | 补上群聊分支：`global_rule.mode=always/none`（best-effort）。同时注册标准动作名（旧版误注册为 `get_group_whole_ban`） |
| `set_group_kick` | 接入官方 `batch_remove_members`：支持单 `user_id` 或 `user_ids` 批量（单次最多 20），可同步加入群黑名单；同义扩展名 `batch_remove_group_members` |
| `get_group_member_list` / `get_group_member_info` | 接入官方 2026-09-03 群成员接口；列表自动翻页，失败回退本地缓存/username 缓存 |
| `get_group_info` | 群类型返回**官方真实群信息**（`group_name` / `group_memo` / `member_count`），接口失败回退原模拟数据（「测试群」） |
| `get_status` | 重写为 LLOneBot bot 状态结构：`data.{online,good,stat}` + 顶层 `wording`；`stat` 含真实 `message_received` / `message_sent` / `last_message_time` / `startup_time` |
| `send_group_msg` / `send_msg` / `send_private_msg` 等发送链路 | ① 出站 at 段转 `@昵称` 文本（昵称缓存优先）；② `at_markdown` 开启时含 at 的群纯文本升级 markdown 真 at；③ `fake_reply` 开启时 `reply` 段文本伪造；④ 主动消息 `allow_proactive_msg`；⑤ 发送成功后出站消息入库，供撤回/fakeReply |

### 兼容别名

| 别名 | 等价于 | 说明 |
|---|---|---|
| `get_group_ban` | `set_group_ban` | 上游早期误注册名，保留兼容历史调用 |
| `get_group_whole_ban` | `set_group_whole_ban` | 同上 |

---

## 三、新增事件映射

| 官方事件 | OneBot 事件 | 说明 |
|---|---|---|
| `GROUP_MEMBER_ADD` | `notice.group_increase` | `sub_type=approve`（官方事件无操作者字段） |
| `GROUP_MEMBER_REMOVE` | `notice.group_decrease` | `sub_type=leave` |
| `GROUP_JOIN_REQUEST` | `request.group` | `sub_type=add`，`flag=join_request_id`；`comment` 支持问答验证标准格式 |
| `SUBSCRIBE_MESSAGE_STATUS` | `notice.subscribe_message_status`（扩展） | 订阅消息授权状态变更 |

需要在 `text_intent` 中开启对应处理器：

```yaml
text_intent:
  - "GroupMessageEventHandler"            # 群全量消息（需官方权限）
  - "GroupATMessageEventHandler"          # 群 @ 消息
  - "C2CMessageEventHandler"              # 群私聊
  - "GroupMemberAddEventHandler"          # 群成员加入
  - "GroupMemberRemoveEventHandler"       # 群成员退出
  - "GroupJoinRequestEventHandler"        # 用户申请加群（机器人需群管理员）
  - "SubscribeMessageStatusEventHandler"  # 订阅消息授权状态变更
```

---

## 四、HTTP 路由（非 OneBot action）

### `GET /g` — qlogo 兼容头像路由

| 参数 | 必填 | 说明 |
|---|---|---|
| `b` | 是 | 固定 `qq` |
| `nk` | 是 | 应用端视角的数字 ID（idmap 行 ID，即虚构 QQ 号） |
| `s` | 否 | 头像尺寸，默认 `640` |
| `redirect` | 否 | `1` = 302 到官方头像地址；缺省 = 代理下载并直接返回图片字节流 |

- 示例：`http://<gsk地址>:<端口>/g?b=qq&nk=800512121&s=640`
- 状态码：参数非法 400；用户不存在 404；头像源异常 502；成功返回图片（`Cache-Control: public, max-age=86400`）。
- 实现：`server/avatarProxy.go`，路由注册在 `main.go`。

---

## 五、参数补充说明（callapi）

以下字段为 fork 为群管理 API 新增到 `callapi.ParamsContent`：

`approve`、`refuse`、`reason`、`flag`、`cursor`、`limit`、`strategy_id`、`group_openids`、`group_ids`、`is_enable`、`expire_at`、`remark`、`whitelist_users`、`whitelist_op`、`group_action_op`、`group_action_group_openids`、`group_action_group_ids`、`members`。

其中 `members[]` 单项结构：

```json
{
  "op": "add",
  "member_openid": "OPENID",
  "mute_expire_at": "2026-09-01T12:00:00+08:00"
}
```

---

## 六、消息段扩展：@全体成员（2026-10-26 官方新特性）

官方 2026-10-26 起，群聊的**文本消息**和 **markdown 消息**支持 `<qqbot-at-everyone />` 标签 @ 全部成员。本 fork 已把 OneBot 惯例写法转换为官方标签：

| 应用端发送 | 适配层转换 |
|---|---|
| 字符串消息中的 `[CQ:at,qq=all]` | `<qqbot-at-everyone />` |
| 数组段 `{"type":"at","data":{"qq":"all"}}` | `<qqbot-at-everyone />` |
| `qq=everyone`（兼容写法） | `<qqbot-at-everyone />` |

示例：

```json
{
  "action": "send_group_msg",
  "params": {
    "group_id": 123456,
    "message": "维护通知：服务器将在 10 分钟后重启 [CQ:at,qq=all]"
  }
}
```

**限制（官方服务端校验，适配层只能透传）**：

- 仅群聊生效；私聊/频道不支持
- 机器人必须是该群**管理员**
- 每个群**每天只能发送 1 次**，超限会返回官方错误
- 纯文本消息无需开启 `at_markdown`；如果同一条消息还包含普通 `[CQ:at,qq=数字]` 且 `at_markdown` 开启，整条消息会升级为 markdown，`<qqbot-at-everyone />` 在 markdown 中同样有效
- 富媒体消息（图文/语音）不升级 markdown，`@全体成员` 会还原为 `@全体成员` 文本（官方富媒体通道不支持该标签）
- **失败回退**：若官方拒绝整条消息（无权限、管理员未授权、当日次数用完 `304019` 等），适配层会自动去掉 `<qqbot-at-everyone />` 标签后重试一次，保证正文能发出；超时类错误不重试，避免消息实际已送达时重复发送

### 消息形态与 @ 支持

| 消息形态 | 真 @ | 说明 |
|---|---|---|
| markdown 消息（含 markdown 图片 / `qqbot-media://`） | ✅ | `<qqbot-at-user>` / `<qqbot-at-everyone />` 保留 |
| 纯文本 | ⚠️ | 单成员 at 不渲染，`at_markdown` 开启时升级 markdown；`@全体成员` 文本直接支持 |
| 富媒体（`RichMediaMessage` / `MsgType=7`，图片/语音/视频） | ❌ | 官方富媒体通道不支持 at 标签，适配层还原为 `@昵称` / `@全体成员` 文本 |

---

## 七、群成员管理（2026-09-03，内邀白名单）

> 官方 2026-09-03 新增的群成员管理接口目前为**内邀接入**：仅白名单机器人可用，无权限时返回错误码 `11253`（应用无接口访问权限），需联系平台运营申请。以下接口均已适配；非白名单机器人调用会返回标准失败响应。

### 1. `get_group_member_list` — 获取群成员列表

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID（应用端压缩 ID；`string_ob11` 时为群 openid） |
| `cursor` | string | 否 | 分页游标，首页留空 |

- 官方接口：`GET /v2/groups/{group_openid}/members`
- 每页最多 30 条，适配层自动翻页聚合后返回完整列表
- 响应 `data` 为成员数组：`user_id`、`group_id`、`nickname`、`card`、`role`、`join_time`、`member_openid`、`bot` 等
- 官方接口失败（含无 11253 权限）时回退本地 idmap-pro 缓存

### 2. `get_group_member_info` — 获取群成员信息

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID |
| `user_id` | string/int | 是 | 成员 ID |

- 官方接口：`GET /v2/groups/{group_openid}/members/{member_openid}`
- 响应 `data` 为单个成员对象（字段同列表）
- 官方接口失败时回退 `idmap` 的 username 缓存构造最小成员信息

### 3. `set_group_kick` — 群成员移除（OneBot 标准动作，支持批量）

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID |
| `user_id` | string/int | 否* | 单个成员 ID（OneBot 标准写法） |
| `user_ids` | []string | 否* | 批量成员 ID（扩展） |
| `member_openids` | []string | 否* | 直接传官方 openid（扩展） |
| `add_to_member_blacklist` | bool | 否 | 移除时同步加入群黑名单 |
| `reject_add_request` | bool | 否 | OneBot 兼容字段，等同 `add_to_member_blacklist` |

\* 三者至少提供一个；单次最多 20 个成员。

- 官方接口：`POST /v2/groups/{group_openid}/batch_remove_members`
- 同义扩展 action：`batch_remove_group_members`
- 响应 `data`：`remove_members_result`、`add_to_member_blacklist_fail_openids`、`add_to_member_blacklist_fail_user_ids`
- 频率 30 QPM；无内邀权限返回 11253

### 4. `get_group_member_blacklist` — 查询群黑名单

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID |
| `cursor` | string | 否 | 分页游标 |
| `limit` | int | 否 | 单页数量，默认 20，最大 100 |

- 官方接口：`GET /v2/groups/{group_openid}/member_blacklist`
- 响应 `data`：`{"users":[{user_id,nickname,member_openid,union_openid,banned_at,bot}],"next_cursor":"..."}`
- 频率 30 QPM

### 5. `add_group_member_blacklist` / `remove_group_member_blacklist` — 群黑名单操作

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `group_id` | string/int | 是 | 群 ID |
| `user_id` | string/int | 否* | 单个成员 ID |
| `user_ids` | []string | 否* | 批量成员 ID |
| `member_openids` | []string | 否* | 直接传官方 openid |

\* 至少提供一个；单次最多 20 个成员。

- 官方接口：`POST /v2/groups/{group_openid}/member_blacklist`
  - `add_group_member_blacklist` → `op=add`（要求成员当前不在群内）
  - `remove_group_member_blacklist` → `op=del`
- 响应 `data`：`{"op":"add|del","fail_openids":[...],"fail_user_ids":[...]}`
- 频率 60 QPM；无内邀权限返回 11253
