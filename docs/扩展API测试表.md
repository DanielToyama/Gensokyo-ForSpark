# Gensokyo-ForSpark 扩展 API 测试表 / 用例

> 测试工具随仓库提供：`tools/onclient`（Go 编写的极简 OneBot v11 WebSocket 客户端），
> **不依赖 SparkBridge / Aemeath 等任何框架**，直接连 Gensokyo 的正向 WS 发 action 并打印响应。
>
> 对应实现与参数说明见 [官方更新兼容-202611.md](./官方更新兼容-202611.md)、[适配API清单.md](./适配API清单.md)。

---

## 一、环境准备

### 1. Gensokyo 侧

```yaml
settings:
  app_id: 你的appid
  uin: 机器人QQ号
  client_secret: "你的客户端密钥"
  port: "15630"
  enable_ws_server: true
  ws_server_path: "nil"        # 监听根路径；非 nil 时 URL 要带 /<path>
  ws_server_token: "testtoken" # 测试客户端用同一个 token
```

- 机器人需为测试群 **管理员**，否则群管理类接口会失败
- 群成员管理/黑名单/批量移除属于官方**内邀白名单**，无权限时返回 `11253`
- `get_user_info` / 群消息私聊回复需要测试用户与机器人是好友

### 2. 测试工具

工具在 `tools/onclient`，不需要额外依赖（仓库已含 `gorilla/websocket`）：

```bash
# 单次调用
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action get_group_info -params '{"group_id":187781763}'

# 只读套件（T01–T10、T24、T28）
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -group 187781763 -user 800512121 \
    -script tools/onclient/suite.example.json

# 含危险操作（禁言/踢人/黑名单/发消息），需显式 -allow-danger
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -group 187781763 -user 800512121 \
    -set victim_id=800512122 -set message_id=1234567890 \
    -allow-danger \
    -script tools/onclient/suite.example.json

# 交互模式（不带 -action/-script）
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken
> get_group_member_list {"group_id":187781763}
> get_status
```

参数说明：

| 参数 | 说明 |
|---|---|
| `-url` | 正向 WS 地址，含 `ws_server_path`（`nil` 时为根路径） |
| `-token` | `ws_server_token`，以 `Authorization: Bearer` 发送 |
| `-group` / `-user` | 套件变量 `group_id` / `user_id` |
| `-set k=v` | 额外套件变量，可重复（`victim_id`、`message_id`、`flag` 等） |
| `-action` / `-params` | 单次调用 |
| `-script` | 运行套件文件 |
| `-allow-danger` | 允许执行套件里标记 `danger` 的写操作，默认跳过 |
| `-timeout` | 单次请求超时秒数，默认 10 |

### 3. 目标账号怎么拿

Gensokyo 上报的 `user_id` 是 idmap 压缩后的虚拟号，不要手抄。先跑一次：

```bash
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action get_group_member_list -params '{"group_id":187781763}'
```

从返回的 `data[]` 里复制成员的 `user_id`（配合 `nickname`/`role` 辨认）作为 `-user` 或 `-set victim_id=...`。

---

## 二、测试矩阵

> `PASS` 表示请求结构正确且业务结果符合预期；`环境受限` 表示因 11253/无好友/无管理员等官方限制失败，但返回了合法错误响应。

| ID | API / 动作 | 类型 | 测试方式 | 预期结果 |
|---|---|---|---|---|
| T01 | `get_group_info` | OneBot 标准 | 套件 `-script` | `status=ok`，`data.group_id` 正确；绑定三方应用时含 `union_group_openid` |
| T02 | `get_group_member_list` | OneBot 标准 | 套件 `-script` | `data` 为数组；成员含 `user_id`、`nickname`、`card`、`role`、`member_openid` |
| T03 | `get_group_member_info` | OneBot 标准 | 套件 `-script` | `data.user_id` 为指定用户；含 `nickname`/`card`/`role` |
| T04 | `get_group_restrict_chat_setting` | 扩展 | 套件 `-script` | `data.global_rule` + `data.members`（无管理员/权限则失败） |
| T05 | `get_group_bot_state` | 扩展 | 套件 `-script` | `data.member_role`、`allow_proactive_msg` 等字段 |
| T06 | `get_group_join_request_list` | 扩展 | 套件 `-script` | `data.list` + `next_cursor`；无管理员/白名单则失败 |
| T07 | `get_group_member_blacklist` | 扩展 | 套件 `-script` | `data.users` + `next_cursor`；无白名单则 11253 |
| T08 | `get_user_info` | 扩展（官方 2026-11-16） | 套件 `-script` | 好友时 `source=official`，含 `nickname`/`avatar`/`union_openid`；非好友回退 `source=cache` |
| T09 | `get_stranger_info` | OneBot 标准 | 套件 `-script` | `data.user_id/nickname/avatar`；`qqLevel=9999` 占位 |
| T10 | `get_status` | OneBot 扩展 | 套件 `-script` | `data.online/good/stat`，`status=ok` |
| T11 | `set_group_ban` 默认 duration | OneBot 标准 | `-action` 或套件(danger) | 省略 duration 时按 OneBot 默认 **1800 秒**禁言 |
| T12 | `set_group_ban` 指定时长 | OneBot 标准 | 套件(danger) | 禁言 60 秒；`data: null` |
| T13 | `set_group_ban` 解禁 | OneBot 标准 | 套件(danger) | `duration=0` 解禁；`data: null` |
| T14 | `set_group_whole_ban` 开启 | OneBot 标准 | 套件(danger) | 官方无写接口：默认返回失败并提示实验开关；开启 `whole_ban_via_members` 后批量禁言成员 |
| T15 | `set_group_whole_ban` 关闭 | OneBot 标准 | 套件(danger) | 同上；实验模式下从查询结果批量解禁 |
| T16 | `set_group_kick` 单人 | OneBot 标准 | 套件(danger) | 移除成功；`data: null`；失败返回合法错误 |
| T17 | `set_group_kick` 批量 | 扩展参数 | 套件(danger) | `user_ids` 最多 20 个；`data: null` |
| T18 | `add_group_member_blacklist` | 扩展 | 套件(danger) | `data.fail_openids` 为空或列出失败项 |
| T19 | `remove_group_member_blacklist` | 扩展 | 套件(danger) | 移出黑名单 |
| T20 | `send_private_msg` 群消息私聊回复 | OneBot 标准 | 套件(danger) | 用群消息 `message_id` 私聊回复；好友关系下成功 |
| T21 | `send_group_card_msg` | 扩展（官方 2026-10-28） | 套件(danger) | 群内收到图文卡片 |
| T22 | `send_private_card_msg` | 扩展（官方 2026-10-28） | 套件(danger) | 单聊收到图文卡片（需好友） |
| T23 | `set_group_add_request` 默认 approve | OneBot 标准 | 套件(danger, 需 flag) | 不传 approve 时默认通过；`approve=false` 拒绝 |
| T24 | `get_join_approval_strategy_list` | 扩展 | 套件 `-script` | 返回策略列表（白名单接口） |
| T25 | `@全体成员` 失败回退 | 扩展行为 | 套件(danger) | 无权限/次数用完时先失败、自动去标签重试；正文最终送达，不含 `@全体成员` |
| T26 | `delete_msg` 撤回机器人自己消息 | OneBot 标准 | 套件(danger, 需 message_id) | 撤回成功；`data: null` |
| T27 | `send_msg` 显式 message_type | OneBot 标准 | 套件(danger) | 群消息按 group 路由；`data.message_id` 返回 |
| T28 | `get_group_restrict_chat_setting` 复查 | 扩展 | 套件 `-script` | 禁言后能看到 `members` 变化 |

> T11/T14/T23 依赖 OneBot 默认值兼容（`duration` 默认 1800、`enable` 默认 true、`approve` 默认 true），已在适配层处理。

---

## 三、详细用例

### T01–T10 / T24 / T28 只读套件

```bash
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -group 187781763 -user 800512121 \
    -script tools/onclient/suite.example.json
```

输出形如：

```text
T01 get_group_info                          PASS retcode=0 data={"group_id":...,"group_name":"..."}
T02 get_group_member_list                   PASS retcode=0 data=[{"user_id":...}]
...
结果: PASS=12 FAIL=0 SKIP=16
```

默认会跳过所有 `danger` 项，所以一次就能验证只读接口。重点关注：

- T02：成员数量、`user_id`、`member_openid`
- T03：指定用户的 `nickname` / `card`
- T04：`global_rule.mode`
- T05：`member_role`
- T08：`source` 是 `official` 还是 `cache`
- T09：`avatar`、`qqLevel=9999`

### T11–T15 禁言

```bash
# T11 不传 duration，应禁言 1800 秒
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action set_group_ban -params '{"group_id":187781763,"user_id":800512122}'

# T12 禁言 60 秒
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action set_group_ban -params '{"group_id":187781763,"user_id":800512122,"duration":60}'

# T13 解禁
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action set_group_ban -params '{"group_id":187781763,"user_id":800512122,"duration":0}'

# T14/T15 全员禁言
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action set_group_whole_ban -params '{"group_id":187781763,"enable":true}'
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action set_group_whole_ban -params '{"group_id":187781763,"enable":false}'
```

验收：

- 对应成员的禁言状态变化（可在手机 QQ 群成员资料或 T28 复查）
- 单成员禁言每次响应应为 `status=ok`、`data: null`

> `set_group_whole_ban` **不是**普通可用能力：官方 `restrict_chat_setting` 只支持成员级禁言，没有全员禁言写接口。
> 默认行为是返回失败并提示开启 `whole_ban_via_members`；该开关是**实验性**的，会批量禁言/解禁当前群成员（每批 20 人，跳过群主/管理员/机器人），需群成员列表内邀权限，新入群成员不受影响。测试时建议只在测试群开启。

### T16–T17 踢人

```bash
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action set_group_kick -params '{"group_id":187781763,"user_id":800512122,"reject_add_request":false}'

go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action set_group_kick -params '{"group_id":187781763,"user_ids":["800512122","800512123"]}'
```

验收：

- 成员从群成员列表消失
- 单次批量不超过 20 个
- 无内邀权限时返回 `11253` 错误（属于环境限制，不判插件缺陷）

### T18–T19 黑名单

```bash
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action get_group_member_blacklist -params '{"group_id":187781763}'

go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action add_group_member_blacklist -params '{"group_id":187781763,"user_id":800512122}'

go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action remove_group_member_blacklist -params '{"group_id":187781763,"user_id":800512122}'
```

验收：

- `fail_openids` 为空表示成功
- 加入黑名单要求用户当前不在群内

### T20 群消息私聊回复

先让一个好友用户在群里发一条消息，然后从 gsk 日志或事件里取得该群消息的 `message_id`（应用端视角的虚拟 ID），执行：

```bash
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action send_private_msg \
    -params '{"user_id":800512121,"message_id":1234567890,"message":"Gensokyo 群消息私聊回复测试"}'
```

验收：

- 机器人私聊发送给该消息发送者
- 与群内回复共用官方 `msg_id` 次数；非好友会失败

### T21–T22 图文卡片

```bash
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action send_group_card_msg \
    -params '{"group_id":187781763,"title":"测试","description":"描述","url":"https://example.com/a","pic_url":"https://example.com/b.png"}'

go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action send_private_card_msg \
    -params '{"user_id":800512121,"message_id":1234567890,"title":"测试","description":"描述","url":"https://example.com/a","pic_url":"https://example.com/b.png"}'
```

验收：

- 群/单聊出现标题、描述、图片、跳转链接
- `url`/`pic_url` 非 https 时适配层直接拒绝

### T23 入群审批

拿到一条待审批申请的 `flag`（来自 `request.group` 事件或 `get_group_join_request_list`）：

```bash
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action set_group_add_request -params '{"flag":"申请ID","approve":true}'

go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action set_group_add_request -params '{"flag":"申请ID","approve":false,"reason":"不认识"}'
```

验收：

- 通过后用户进群；拒绝后申请消失
- 只传 `flag` 也能从事件缓存反查群/用户

### T25 @全体成员失败回退

```bash
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action send_group_msg \
    -params '{"group_id":187781763,"message":"回退测试 [CQ:at,qq=all] 正文必须送达"}'
```

验收：

- 日志出现 `@全体成员发送失败(...)，去掉@全体成员后重试正文`
- 群里最终收到「回退测试 正文必须送达」，不显示 `@全体成员`，也不显示原始标签
- 不能因为 @ 失败导致整条消息丢失

### T26 撤回机器人自己的消息

先发一条消息拿到返回的 `message_id`，然后：

```bash
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action delete_msg -params '{"message_id":1234567890}'
```

验收：

- 机器人的消息被撤回（或官方接口回失败，但响应结构合法）

### T27 send_msg 显式 message_type

```bash
go run ./tools/onclient -url ws://127.0.0.1:15630 -token testtoken \
    -action send_msg \
    -params '{"message_type":"group","group_id":187781763,"message":[{"type":"text","data":{"text":"send_msg 测试"}}]}'
```

验收：

- 按群消息发送，不需要依赖 echo/idmap 推断
- 返回 `data.message_id`

---

## 四、套件变量

`tools/onclient/suite.example.json` 支持以下变量，通过 `-group/-user/-set` 传入：

| 变量 | 传入方式 | 用于 |
|---|---|---|
| `${group_id}` | `-group 187781763` | 所有群相关 API |
| `${user_id}` | `-user 800512121` | 自己的成员信息、私聊、卡片 |
| `${victim_id}` | `-set victim_id=800512122` | 禁言/踢人/黑名单的目标（**请用可牺牲的测试账号**） |
| `${message_id}` | `-set message_id=1234567890` | T20/T22/T26 的被动回复或撤回 |
| `${flag}` | `-set flag=申请ID` | T23 入群审批 |

未提供的变量会原样保留（例如 `${message_id}`），从失败输出就能看出漏填。

---

## 五、结果记录模板

| ID | API | 结果（PASS/FAIL/环境受限） | status / retcode | 关键返回/错误 | 备注 |
|---|---|---|---|---|---|
| T01 | get_group_info |  |  |  |  |
| T02 | get_group_member_list |  |  |  |  |
| T03 | get_group_member_info |  |  |  |  |
| T04 | get_group_restrict_chat_setting |  |  |  |  |
| T05 | get_group_bot_state |  |  |  |  |
| T06 | get_group_join_request_list |  |  |  |  |
| T07 | get_group_member_blacklist |  |  |  |  |
| T08 | get_user_info |  |  |  |  |
| T09 | get_stranger_info |  |  |  |  |
| T10 | get_status |  |  |  |  |
| T11 | set_group_ban（默认 1800） |  |  |  |  |
| T12 | set_group_ban（60s） |  |  |  |  |
| T13 | set_group_ban（0 解禁） |  |  |  |  |
| T14 | set_group_whole_ban（on） |  |  |  |  |
| T15 | set_group_whole_ban（off） |  |  |  |  |
| T16 | set_group_kick |  |  |  |  |
| T17 | set_group_kick（batch） |  |  |  |  |
| T18 | add_group_member_blacklist |  |  |  |  |
| T19 | remove_group_member_blacklist |  |  |  |  |
| T20 | send_private_msg（群 msg_id） |  |  |  |  |
| T21 | send_group_card_msg |  |  |  |  |
| T22 | send_private_card_msg |  |  |  |  |
| T23 | set_group_add_request |  |  |  |  |
| T24 | get_join_approval_strategy_list |  |  |  |  |
| T25 | @全体成员回退 |  |  |  |  |
| T26 | delete_msg |  |  |  |  |
| T27 | send_msg message_type |  |  |  |  |
| T28 | get_group_restrict_chat_setting 复查 |  |  |  |  |

---

## 六、已知环境/实现限制

- 2026-09-03 群成员管理接口（成员列表/信息、批量移除、黑名单）为官方**内邀白名单**，无权限 `11253` 属预期，不是工具/适配层缺陷
- `get_user_info` 无好友关系返回 `40184001`，适配层回退缓存
- 图文卡片必须 https；`url`/`pic_url` 无有效资源时官方可能发送失败
- 官方被动回复 50 次/60 分钟、@全体成员每日 1 次等由服务端校验，适配层不重复拦截
- 卡片入站暂不转 CQ 段，原始事件保留 `card` 字段
- **全员禁言**：官方没有写接口，默认不可用；`whole_ban_via_members` 为实验性批量降级，默认关闭
- **虚拟号**：`user_id` 是 idmap 压缩号，先从 `get_group_member_list` 输出里复制，不要手编
