# DanielToyama 更新记录（Gensokyo-ForSpark Fork）

> 维护者：DanielToyama（GitHub: [DanielToyama](https://github.com/DanielToyama)，提交邮箱 `jswnlinyu@Outlook.com` / `daniel_linyu@outlook.com`，另有 3 笔使用 GitHub 账号 `Daniel Lin <53430542+DanielToyama@users.noreply.github.com>`）
>
> 统计范围：fork 首笔提交 `f45da73`（2026-08-15）至 `76eee77`（2026-11-07），共 **45 笔提交**，涉及 **53 个文件、+4686 / -419 行**。
>
> 说明：下方日期以 **git 提交时间** 为准；个别提交的代码注释标注了实际修改日期（如 `76eee77` 标注 2026-09-02、`6f9328f` 标注 2026-08-18），统计仍按提交时间排序。
>
> 上游基线：`ce25b67` Beta530 (#529)，2026-01-31。本 fork **不再与原上游同步**，所有修改遵循 GPLv3，修改文件头部均带 `Modified by DanielToyama` 声明。
>
> 配套文档：[适配 API 清单](./适配API清单.md) · [撤回消息](./撤回消息.md) · [API 介绍](./api介绍.md)

---

## 一、总览

本 fork 面向 **SparkBridge 群服互通** 场景，核心目标是让 QQ 官方机器人在 OneBot v11 生态（尤其 SparkBridge）中「像普通 QQ 号一样可用」。全部改动可归为 9 个能力域：

| # | 能力域 | 一句话说明 | 主要文件 |
|---|--------|-----------|----------|
| 1 | 群全量消息接收 | 订阅 `GROUP_MESSAGE_CREATE` 后接收群内每条消息，不要求 @bot；`require_mention` 控制是否只响应 @ | `main.go`、`Processor/ProcessGroupMessage.go`、`botgo/dto/websocket_event.go` |
| 2 | 主动消息发送 | `allow_proactive_msg: true` 时无被动 `msg_id` 也直接发送 | `handlers/send_group_msg.go` 等发送链路、`config/config.go` |
| 3 | 群管理事件 | 群成员加入/退出、入群申请、订阅消息状态变更 → OneBot notice/request 事件 | `Processor/ProcessGroupMember*.go`、`ProcessGroupJoinRequest.go`、`ProcessSubscribeMessageStatus.go` |
| 4 | 群管理 API | 群信息、成员/全员禁言、入群审批、申请列表、禁言查询、机器人群内状态、入群自动审批策略 6 接口 | `handlers/get_group_*.go`、`set_group_*.go`、`join_approval_strategy.go`、`botgo/openapi/v2/group.go` |
| 5 | SparkBridge 互通适配 | 出站 @ 转文本/真 at、昵称缓存、`get_stranger_info`、任何 action 都有合法 JSON 响应 | `handlers/message_parser.go`、`idmap/username.go`、`callapi/callapi.go` |
| 6 | FakeReply 假回复 | OneBot `reply` 引用段无法真引用 → 文本伪造「回复 @昵称 / ———— / 原内容」 | `msgmap/msgmap.go`、`handlers/message_parser.go` |
| 7 | 真实 At（Markdown） | 含 at 的群纯文本消息升级为 `msg_type=2` markdown，内嵌 `<qqbot-at-user>` 渲染真 at | `handlers/send_group_msg.go`、`handlers/message_parser.go` |
| 8 | 表情消息适配 | 入站对齐 OneBot（`subType=1`、表情包降级名字）；出站 face 段降级 `[表情名]` | `handlers/face.go`、`handlers/message_parser.go` |
| 9 | 撤回自己的消息 | 出站消息入库 `msgmap.RecordOutbound`，标准 `delete_msg` 仅传 `message_id` 即可撤回 | `msgmap/msgmap.go`、`handlers/outbound_msgmap.go`、`handlers/delete_msg.go` |

此外还有头像代理路由、`get_status` 重写、维护通知配置化、CI/构建修复等配套改动。

---

## 二、核心能力详解

### 2.1 群全量消息接收与主动消息

**群全量消息**（`f45da73` 起）

- 新增订阅 `GROUP_MESSAGE_CREATE`（`GroupMessageEventHandler`），机器人可收到群内**每条**消息，不要求 @bot。
- 与 `GROUP_AT_MESSAGE_CREATE`（`GroupATMessageEventHandler`）可同时开启；`@` 消息可能两条事件都到达，botgo 层按 `data.ID` 去重。
- `ProcessGroupMessage` 参数由 `*dto.WSGroupATMessageData` 改为底层 `*dto.Message`，内部再转回具体类型，兼容两种事件（修复下游 type switch 匹配失败导致消息被误拦）。
- 新增 `require_mention` 配置：`false`（默认）= 全量消息都响应；`true` = 仅响应 @bot 的消息。仅对全量事件生效，@ 事件本身即代表被 @，不受影响。
- 全量消息未 @bot 时不触发维护通知兜底回复（避免 WS 掉线时「收到什么回什么」）。

**主动消息**

- 新增 `allow_proactive_msg` 配置：无被动 `msg_id` 缓存时直接发送（官方 API 支持无 `msg_id` 的主动消息）。
- 需官方主动消息权限，否则报 `40034105 主动消息失败, 无权限`。
- 默认 `false`，沿用原有被动窗口 / AtoP 逻辑。

### 2.2 群管理事件

对应官方 API v2「群管理」板块（2026-08-10+ 新增接口/事件）：

| 官方事件 | 常量 | OneBot 事件 |
|---|---|---|
| `GROUP_MEMBER_ADD` | `GROUP_MEMBER_EVENT (1<<24)`，并入 `GROUP_AND_C2C_EVENT (1<<25)` | `notice.group_increase`（官方无操作者字段，`sub_type` 默认 `approve`） |
| `GROUP_MEMBER_REMOVE` | 同上 | `notice.group_decrease`（`sub_type` 默认 `leave`） |
| `GROUP_JOIN_REQUEST` | 同上 | `request.group`（`sub_type=add`，`flag=join_request_id`） |
| `SUBSCRIBE_MESSAGE_STATUS` | 同上 | 扩展 `notice.subscribe_message_status` |

- `ProcessGroupJoinRequest` 会缓存 `join_request_id → (群 openid, 成员 openid)`，供 `set_group_add_request` 只传 `flag` 时审批（SparkBridge groupRequest 插件写法）。
- 入群申请 `comment` 支持问答验证（`verify_info.method=admin_review_qa`），按 OneBot 标准格式拼接为 `问题：…\n答案：…`（多组问答换行，无回答省略答案行）；`apply_source=invited` 且无内容时兜底「被邀请入群」。
- `comment` 最终为空时打印留痕日志（含完整 `verify_info`），便于区分申请人未填 / 官方缺字段 / 偶发丢字段。

### 2.3 SparkBridge 互通适配

**出站 @**

- 官方 API 发送富文本实测不渲染任何 @ 标签：文本链 `<qqbot-at-user id="openid"/>` 纯文本显示原文；频道模板 `<at id="openid"></at>` 在群聊自定义 markdown 同样不渲染。
- 因此出站 at 段默认转 **`@昵称` 文本**；昵称优先取缓存（每条群消息/加群申请缓存官方 `username`，持久化 7 天、重启不丢），未知时显示 `@Openid+openid前8位`。
- `at_markdown` 开启时，含 at 的群纯文本消息升级为 markdown 消息发真 at（见 2.5）。
- 对话文本内的 @ / 伪 at 码做转义，避免误解析；新群友进群时机器人 @ 行为优化（拼接/去重更合理）。

**入站 @**

- 官方群消息里的 @ 是 `<@openid>` 内嵌标签，转成 `[CQ:at,qq=xxx]` 体现在 `raw_message`/`message`，`xxx` 与 gsk 内该成员的 `user_id` 一致，下游插件可据此指定人。

**昵称**

- 官方事件 `author.username` 已返回真实昵称（如 `Daniel_户山兔兔`），回填 `Sender.nickname/card`：`card_nick` 配置优先，否则回退官方 `username`（覆盖群消息 + C2C 私聊）。
- `get_stranger_info` 的昵称从 `idmap` 用户名缓存取；`sex/age` 官方不提供返回空/0；`qqLevel` 固定 `9999` 放行等级门槛（防插件按 0 级误拒）。

**任何 action 都有合法 JSON 响应**

- 未注册 action / handler 出错时返回 `{"status":"failed","retcode":1400,"data":{"message":...},"echo":...}`，杜绝对端等待超时拿 `undefined` 后 JSON5 解析崩溃。

### 2.4 FakeReply 假回复

- 官方协议不支持真实引用（OneBot `send_group_msg` 的 `{"type":"reply","data":{"id":...}}`），但引用能让人辨别「在回复谁」。
- Gensokyo 把收到的每条群/私聊/频道消息按 `message_id` 持久化发送者昵称与原文到 `msgmap.db`（bbolt，重复键覆盖，7 天自动清理）。
- 应用端发送带 `reply` 引用段的消息时，自动文本伪造为：
  ```
  回复 @昵称
  ————
  原消息内容

  本条消息正文
  ```
- 支持与 text/at/image 等段混合发送；查不到记录时跳过伪造、按普通消息发送。
- 配置 `fake_reply`（默认 `true`）：`false` = 忽略 reply 段。
- `at_markdown` 开启时，前缀「回复 @昵称」的昵称部分升级为官方文本链真实 at `<qqbot-at-user id="openid"/>`（openid 从 msgmap 的 `UserID` 经 idmap 反查），随 markdown 升级发送；私聊/频道/富媒体场景回退文本 @昵称。

### 2.5 真实 At（Markdown）

- 实测：**markdown 消息**可渲染真实 at；2026-04-23 起自定义 markdown 已开放到所有机器人（单聊/群聊，无需申请模版）。
- 配置 `at_markdown`（**默认 `true`**）：发送**含 at 段的群纯文本消息**时自动升级为 markdown 消息（`msg_type=2`，`markdown.content` 内嵌 `<qqbot-at-user id="openid" />`，openid 由 idmap 反查）。
- **带图/语音等富媒体消息不升级**：富媒体通道无 at 渲染，at 还原为 `@昵称`；图文 + at 不合并。
- 富媒体/超长消息（>2800 字符）会把已注入的 at 标签还原为 `@昵称`，防止原始标签当正文发出。
- 历史演进：`6c807d8` 引入（默认关闭，注入 `<at id="openid"></at>`）→ `02a54e3` 改注入官方文本链 `<qqbot-at-user id="openid"/>`（频道模板语法群聊实测不渲染）→ `6387043` 富媒体还原 → `684467d` 默认开启并让 fakeReply 前缀也升级真 at。

### 2.6 表情消息适配

**入站（对齐 OneBot / LLOneBot）**

- 官方群消息 `faceType=4`（表情包商店）/ `faceType=6`（自定义/收藏表情）本质是图片，其附件在 `image` 段标记 `subType=1`（动画表情），普通图片 `subType=0`；`message` 段数组与 `raw_message` 字符串两条路径都处理。
- `faceType=6` 跳过无意义的 face 标签（真实内容即附件图片）。
- `faceType=4` 官方事件无图片附件（`faceId` 空、`ext` 仅含 `{"text":"表情名"}`，无 `emoji_id→URL` 数据源），降级为**文本名字**（解码 ext，如 `[代班]`/`[跳舞]`）。
- 系统表情仍转 `[CQ:face,id=xxx]` / `{type:"face",data:{id:"264"}}`。

**出站（降级占位）**

- 官方 bot 能力表**没有「表情」这一消息类型**（表情表态 = 给消息点赞，不是发送表情）。
- 三条路均实测证伪：`<emoji:ID>` 纯文本显示原文、`<emoji:ID>` markdown 通道不渲染、Unicode emoji 字符方案行不通。
- 最终 `face` 段统一降级为 **`[表情名]`**（如 `[微笑]`），不丢内容、不中断后续段、不触发 markdown 升级。
- `faceNameMap` 为 QQNT 官方面板 + `face_config` 经典两套编号体系的合并（**273 条，id 0~431**，冲突以 QQNT 为准）；未收录显示 `[表情:id]`。
- 若未来 QQ 开放平台支持表情消息，只需改造 `faceIDToEmoji` 输出官方真实格式，`faceNameMap` 可继续复用。

### 2.7 撤回机器人自己发的消息

- **问题**：标准 OneBot v11 `delete_msg` 只传 `message_id`，而 `msgmap` 只记录机器人**收到**的消息，机器人自己发出的消息从未入表 → 反查落空 → 四个撤回分支全跳过，响应却仍 `{"status":"ok"}`（SparkBridge3 的 `QClient.deleteMsg(id)` 正是这种写法，必踩）。
- **出站入库**：新增 `msgmap.RecordOutbound`（`MsgInfo` 加 `Outbound` 标记，`Nickname` 记 `custom_bot_name`，让 fakeReply 引用机器人自己的消息时显示「回复 @机器人名」）。五个发送回执函数 `SendResponse` / `SendResponseSB` / `SendGuildResponse` / `SendC2CResponse` / `SendGuildPrivateResponse` 各埋一处（`send_msg`、`*_async`、`*_raw`、`forward`、`sse`、`wakeup` 全部汇流于此），把 `message_id`（官方 id + 应用端压缩行 id 双键）→ 会话对象写入，7 天 TTL 与入站一致。
- **反查扩展**：`delete_msg` 缺参反查从「群 / 私聊」扩到「群 / 私聊 / 频道（`channel_id`）/ 频道私信（`guild_id`）」四条，均只对 `Outbound` 记录生效；私聊分支限定 `MsgType=="group_private"`，防止频道记录带 `user_id` 被误判为 C2C。
- **健壮性**：`message_id` 缺失时不再 `.(string)` panic；idmap 反查失败回退原始 id，不再把空值发给官方接口。
- **效果**：`{"action":"delete_msg","params":{"message_id":X}}` 即可撤回机器人自己发的消息，无需应用端传 `group_id`。
- 完整用法见 [撤回消息.md](./撤回消息.md)。

### 2.8 野鸡 qlogo 头像路由

- 部分客户端习惯 `https://q1.qlogo.cn/g?b=qq&nk=<QQ号>&s=640`，但官方 bot 无 QQ 号（出站 avatar 为 `q.qlogo.cn/qqapp/<appid>/<openid>/640`）。
- 新增 HTTP 路由 `GET /g?b=qq&nk=<数字ID>&s=<尺寸>`：把应用端视角的虚构 QQ 号（idmap 数字行 ID）反查为 openid。
- **默认代理模式**：Gensokyo 后台下载官方头像并直接返回图片字节流（透传 `Content-Type`，客户端无需理解重定向），覆盖 axios/浏览器/不跟随 302 的客户端。
- `&redirect=1` 时保留 302 到官方地址。
- `b` 非 `qq` / `nk`、`s` 非纯数字返回 400，反查不到 404，头像源不可达 502，成功响应带 `Cache-Control: public, max-age=86400`。

### 2.9 维护通知（downtime_message）配置化

WS 全部掉线时的兜底回复，改动见 `Processor/Processor.go` 的 `BroadcastMessageToAll`：

1. 未配置 `downtime_message` 时不再发送空消息；
2. 群聊 @ / 频道 @ 为一对多场景，按「群/频道 + 用户」冷却（`downtime_cooldown`，分钟，`0`=不冷却），同一用户冷却期内只回一次；
3. 群全量消息（未 @bot）与频道不 at 消息不再触发维护回复；
4. 私聊 / C2C / 频道私信为一对一场景，每条都回，不受冷却限制；
5. 新增 `downtime_message_enabled` 总开关（`false` = 完全不回复）。

### 2.10 get_status 重写

- 对齐 LLOneBot `bot状态` 文档：顶层 `status/retcode/data/message/wording`，`data.{online, good, stat}`。
- `stat` 提供真实统计：`message_received` / `message_sent` / `last_message_time` / `startup_time`（新增，进程启动时刻），来自 `botstats` 持久化。
- 移除原 go-cqhttp 风格写死测试数据（`packet_received: 1000` 等）；能收到请求即说明连接正常，`online/good` 直接为 `true`。

---

## 三、新增配置项

| 配置 | 默认 | 说明 |
|---|---|---|
| `allow_proactive_msg` | `false` | 允许主动消息：无被动 `msg_id` 缓存时直接发送 |
| `require_mention` | `false` | 群全量消息是否要求 @bot 才响应 |
| `downtime_message_enabled` | `true` | 维护通知总开关：`false` = WS 全部掉线时不回复维护文案 |
| `downtime_cooldown` | `10` | 维护回复冷却（分钟）：群/频道一对多场景同一用户冷却期内最多回一次；`0` = 不冷却；私聊/C2C 不受影响 |
| `fake_reply` | `true` | FakeReply 假回复开关：`true` = 应用端发 `reply` 引用段时文本伪造；`false` = 忽略 |
| `at_markdown` | `true` | 真实 At（Markdown）开关：含 at 的群纯文本消息升级 `msg_type=2` |

对应模板/可视化配置：`template/config_template.go`、根目录 `config.yml`、`gensokyo-config-gen.html`（「⑥ 维护通知」「⑦ 假回复」「⑧ 真实 At」卡片）。

---

> **2026-11-10 工作区变更（待提交）**
>
> - **支持官方 2026-10-26 新特性「群聊 @全体成员」**：OneBot `[CQ:at,qq=all]` / 数组段 `{"type":"at","data":{"qq":"all"}}` 自动转换为官方 `<qqbot-at-everyone />`；文本与 markdown 消息均可用，机器人需群管理员、每群每天限 1 次。
> - **移除已废弃的 `token`（旧应用令牌）配置项**：官方 API v2 只用 `app_id` + `client_secret` 动态换取 access_token（7200 秒有效），`token` 已不参与鉴权；同步清理 `structs`、`config.yml`、配置模板、配置生成器与 `botgo/token` 中的旧字段。
> - **修复配置生成器 YAML 格式错误**：① JS 模板字符串注释里的 `\n` 会被解释为真实换行，导致生成的 `config.yml` 在「假回复」注释处解析失败（已改为单行描述）；② 对 `client_secret`、`ws_server_token`、`server_dir`、`port`、`downtime_message` 等用户输入统一做 YAML 双引号转义，避免引号/反斜杠/换行导致解析失败；同时校验 `app_id`/`uin` 为纯数字。
> - **修正「④ 群消息 @ 设置」说明**：官方并非“完全不支持真 @”——群 markdown 消息可渲染 `<qqbot-at-user>`（`at_markdown` 默认开启），纯文本单成员 at 才降级为 @昵称；三个开关（`remove_at`/`add_at_group`/`remove_bot_at_group`）按下游插件行为选择，模板中 `remove_bot_at_group` 默认改为 `false` 以免误删真 at。
> - **区分消息形态**：markdown 消息（含 markdown 图片）保留真 at；纯文本单成员 at 不渲染（可升级 markdown）；富媒体图片/语音/视频（`RichMediaMessage` / `MsgType=7`，无 markdown 发图条件时使用）不支持 at，自动还原为 `@昵称` / `@全体成员` 文本，绝不把原始标签发给官方富媒体通道。

---

## 四、完整提交记录（倒序）

> 说明：同日多次迭代通常是同一功能的连续修复，按最终效果归并说明；`Daniel Lin` 为同一 GitHub 账号的提交署名。

| 日期 | Commit | 类型 | 内容 |
|---|---|---|---|
| 2026-11-07 | `76eee77` | fix | 撤回机器人自己发的消息静默失效：出站消息入库 `msgmap.RecordOutbound` + `delete_msg` 反查扩展、健壮性修复、单测与文档 |
| 2026-09-02 | `a4830bc` | feat | 表情表补全为 QQNT 官方面板 + `face_config` 经典合并（273 条，id 0~431）；`faceType` 分类入站对齐 OneBot；readme 修改记录重排、新增能力 8/9 |
| 2026-08-31 | `7e6c462` | feat | `face` 段降级占位改为输出表情名 `[大哭]`：`faceNameMap`（1-106 经典表情，后于 `a4830bc` 扩展到 273 条），未收录兜底 `[表情:id]` |
| 2026-08-31 | `e2af699` | fix | 官方 bot 能力表无「表情消息」类型，三条路实测证伪：`face` 段降级 `[表情:id]` 占位、不触发 markdown 升级；`maybeUpgradeToMarkdownAt` 恢复仅处理 at |
| 2026-08-31 | `a4d643a` | fix | `face` 段按官方系统表情模型发 `<emoji:ID>` 内嵌标签（`type=1` 系统表情），废弃 Unicode emoji 映射方案（随后被证伪） |
| 2026-08-31 | `bdf7808` | fix | 出站消息 `face` 段被静默丢弃（`parseMessageContent` 无 `case face`）：face id 转 emoji 文本（未收录回退占位），消息不再丢内容 |
| 2026-08-20 | `bff8edb` | feat | `/g` 头像路由改代理模式：后台下载官方头像直接返回字节流；`redirect=1` 保留 302；文件改名 `avatarProxy.go` |
| 2026-08-20 | `16764fe` | feat | 新增野鸡 qlogo 兼容路由 `GET /g?b=qq&nk=数字ID&s=尺寸`：nk 反查 openid 后 302 到官方头像 |
| 2026-08-20 | `e16cbea` | fix | `delete_msg` 响应误用 `GetStatusResponse` 改为标准 `data:null`；缺省 `group_id/user_id` 时从 msgmap 反查撤回对象（原实现静默不执行） |
| 2026-08-20 | `5bf6b43` | fix | 群禁言/全员禁言被 `guild_id` 配置读取卡死（仅频道需要却放在 switch 前）：群聊分支从未执行过禁言；读取移入 guild 分支 |
| 2026-08-19 | `4ca175c` | feat | 入群申请 comment 问答格式改 OneBot 标准 `问题：<问题>\n答案：<答案>`，与其他 OneBot 客户端一致；多组问答换行、无回答省略答案行 |
| 2026-08-19 | `684467d` | feat | FakeReply 伪回复前缀升级真实 at（群 + `at_markdown` 时 `<qqbot-at-user>`）+ `at_markdown` 默认开启；私聊/频道回退文本 @昵称，富媒体图文维持还原 |
| 2026-08-19 | `02a54e3` | fix | `at_markdown` 注入格式改为官方文本链 `<qqbot-at-user id=openid/>`（`<at>` 频道模板语法群聊实测不渲染）；升级/还原全量日志 |
| 2026-08-19 | `6387043` | fix | `at_markdown` 富媒体消息的 `<at>` 标签还原为 @昵称，防图文/语音把 markdown at 原始标签当正文发出 |
| 2026-08-19 | `6c807d8` | feat | 真实 At（Markdown）`at_markdown`：含 at 的群文本消息升级 `msg_type=2` 内嵌 `<at id=openid></at>` 渲染真 at（实验功能，默认关闭） |
| 2026-08-19 | `b9cc17c` | fix | 入群审核 comment 增强：问答验证（`admin_review_qa`）拼接问/答 + 申请列表附加 comment + 空值留痕日志 |
| 2026-08-19 | `7e6d8cc` | feat | FakeReply 假回复：OneBot v11 `reply` 引用段的文本伪造（官方协议不支持真实引用），新增 `msgmap` 持久化 |
| 2026-08-19 | `ce42966` | fix(ci) | UPX 统一用 linux amd64 版（win64 的 `upx.exe` 无法在 linux runner 执行；linux 版同时支持压 PE 与 ELF） |
| 2026-08-19 | `4640c3e` | fix(ci) | UPX 安装改直连 GitHub Release 下载（apt azure 镜像抽风导致无限重试卡死构建）；失败才回退 apt 带重试上限 |
| 2026-08-19 | `5759aaa` | docs | 更新 readme |
| 2026-08-18 | `e77e083` | feat | `get_status` 看齐 LLOneBot 的 bot 状态 API（`data.{online,good,stat}` + 顶层 `wording`，真实收发统计） |
| 2026-08-18 | `16d2605` | feat | 维护通知（`downtime_message`）行为调整与配置化：新增 `downtime_message_enabled` / `downtime_cooldown`，群/频道一对多冷却、私聊每条都回 |
| 2026-08-18 | `4abead7` | fix(ci) | Release workflow flatten 产物拷贝到 `output/release`，修复 linux 产物丢失（文件名与 artifact 目录同名导致 `cp` 失败） |
| 2026-08-18 | `da70ae1` | chore | 测试 workflow 改动 |
| 2026-08-17 | `c17eae1` | docs | GPLv3 分叉独立化：为所有修改文件补 `Modified by DanielToyama` 声明与来源说明 |
| 2026-08-17 | `7326407` | fix | 优化新群友进群时机器人 @ 的行为（拼接/去重更合理），配合 `idmap/username.go` 昵称缓存 |
| 2026-08-17 | `7670987` / `fc023b1` | fix | 修正 qface 表情出站格式（官方事件表情内嵌标记 → OneBot CQ 码） |
| 2026-08-17 | `b37c5f8` | ci | 修改 workflow 行为 |
| 2026-08-17 | `2ea1ae8` | chore | 清理测试文件（`gensokyo-forspark.exe`、`sensitive_words*`、`white.txt`）并加入 `.gitignore` |
| 2026-08-17 | `7acf65a` | docs | 更新 readme / 配置生成器 |
| 2026-08-16 | `4b9db0e` | feat | 补全群管理事件 API：`GROUP_MEMBER_ADD/REMOVE`、`GROUP_JOIN_REQUEST`、`SUBSCRIBE_MESSAGE_STATUS` 事件与对应 intent/注册 |
| 2026-08-16 | `85d4df7` | feat | 更好的处理 at 行为：新增 `get_stranger_info`、`set_group_add_request` 仅传 `flag` 的缓存反查；对话 @ 处理优化 |
| 2026-08-16 | `688cc86` | fix | 转义对话中的 at，避免文本里的 @ 被误解析 |
| 2026-08-16 | `6f9328f` | chore | 使用与上游一致的压缩编译（`-s -w` + `upx --best`），节省体积 |
| 2026-08-15 | `f45da73` | feat | fork 首笔：支持群全量消息接收（`GROUP_MESSAGE_CREATE`）与主动消息发送（`allow_proactive_msg`） |
| 2026-08-15 | `6fe736c` | fix | 官方 `username` 回填 `Sender.nickname/card`（覆盖群消息 + C2C 私聊） |
| 2026-08-15 | `c142002` / `946796e` / `9dcde75` / `7ad4f9f` | ci | 合并为单 workflow 自动构建发布、自动生成 release notes、修复弃用 action、移除 `paths-ignore` 对 workflow 的忽略 |
| 2026-08-15 | `a2f1d49` / `acbe6ce` / `497ee92` / `6b17bd8` | docs | README 补充自用修改说明、fork 说明、username 修复与回调按钮待办 |

---

## 五、文件变更索引

### 新增文件

| 文件 | 说明 |
|---|---|
| `msgmap/msgmap.go` | message_id → 发送者/会话 持久化（bbolt `msgmap.db`，7 天 TTL），FakeReply 与出站撤回共用 |
| `msgmap/outbound_roundtrip_test.go` | 出站消息入库往返单测 |
| `handlers/face.go` | 表情 id ↔ 名字映射（`faceNameMap` 273 条）、`faceIDToEmoji`、`decodeFaceExtText` |
| `handlers/outbound_msgmap.go` | 出站消息入库辅助 `recordOutboundMsg` |
| `handlers/get_group_bot_state.go` | 机器人群内状态拓展 API |
| `handlers/get_group_join_request_list.go` | 入群申请列表拓展 API |
| `handlers/get_group_restrict_chat_setting.go` | 群禁言状态查询拓展 API |
| `handlers/set_group_restrict_chat_setting.go` | 群禁言原始透传拓展 API |
| `handlers/join_approval_strategy.go` | 入群自动审批策略 6 接口 |
| `handlers/group_mgmt_common.go` | 群管理公共函数（comment 拼接、响应构造、openid 反查） |
| `handlers/set_group_add_request.go` | 入群申请审批（支持只传 flag） |
| `handlers/get_stranger_info.go` | 陌生人信息（昵称缓存 + 占位字段） |
| `idmap/username.go` | 官方 username 缓存（7 天）与 join_request 缓存 |
| `server/avatarProxy.go` | `/g` qlogo 兼容头像路由（代理模式） |
| `Processor/ProcessGroupMemberAdd.go` | 群成员加入事件 → `notice.group_increase` |
| `Processor/ProcessGroupMemberRemove.go` | 群成员退出事件 → `notice.group_decrease` |
| `Processor/ProcessGroupJoinRequest.go` | 入群申请事件 → `request.group` |
| `Processor/ProcessSubscribeMessageStatus.go` | 订阅消息状态 → 扩展 notice |
| `botgo/dto/group.go` | 群管理 DTO |
| `botgo/openapi/v1/group.go` / `botgo/openapi/v2/group.go` | 群管理官方接口封装（v1/v2） |
| `docs/撤回消息.md` | `delete_msg` 完整用法与四种会话对照 |

### 主要修改文件

| 文件 | 改动 |
|---|---|
| `main.go` | 全量消息事件注册、群管理事件注册、msgmap 初始化、`/g` 路由、`require_mention` 分支、webui 移除 |
| `handlers/message_parser.go` | 出站 @ 转文本/真 at、FakeReply 前缀、face 段处理、入站 @/表情转换、5 个发送回执埋点 |
| `handlers/send_group_msg.go` | `maybeUpgradeToMarkdownAt`、`sendGroupMsgUpgraded`、富媒体 at 还原、主动消息 |
| `handlers/delete_msg.go` | 标准响应、msgmap 反查（群/私聊/频道/频道私信）、健壮性 |
| `handlers/get_group_info.go` | 群类型返回官方真实群信息（失败回退模拟数据） |
| `handlers/set_group_ban.go` / `set_group_whole_ban.go` | 群聊禁言/全员禁言接入官方 v2 接口；补注册标准动作名 |
| `handlers/get_status.go` | 重写为 LLOneBot 结构 |
| `handlers/get_group_member_info.go` 等 | 昵称/字段对齐（随主链路） |
| `config/config.go` / `structs/structs.go` / `template/config_template.go` | 6 个新增配置项及读取函数 |
| `callapi/callapi.go` | 群管理参数、`SendAPIError`、任何 action 都有合法响应 |
| `Processor/ProcessGroupMessage.go` | 全量消息兼容、`require_mention`、msgmap 记录、昵称回填、维护通知跳过标记 |
| `Processor/ProcessC2CMessage.go` | 昵称回填、msgmap 记录 |
| `Processor/Processor.go` | 维护通知冷却/总开关、事件分发 |
| `botgo/dto/websocket_event.go` / `websocket_intent.go` / `event/register.go` | 新事件常量与注册 |
| `.github/workflows/cross_compile.yml` | 自动发布、UPX 安装/架构修复、产物路径修复 |

---

## 六、构建与发布

- 本地构建与 CI 一致：`go build -ldflags "-s -w"` 后用 `upx --best` 压缩（32,862,208 B → 9,634,304 B，约 29.3%）。
- Release workflow 统一使用 **linux amd64 版 UPX**：GitHub runner 是 Linux，无法执行 Windows PE 程序；linux 版 UPX 同时支持压缩 PE（`.exe`）与 ELF。
- UPX 安装优先直连 `github.com/upx/upx` Release 下载（不依赖 apt 镜像），失败才回退 apt（`Acquire::Retries=3`）。
- 产物 flatten 到 `output/release/`，避免与同名 artifact 目录冲突导致 `cp` 失败。

---

## 七、实测结论与已知边界

- **真 @**：官方纯文本消息不渲染任何 at 标签；markdown 消息可渲染 `<qqbot-at-user id="openid"/>`（2026-08-19 真群实测）。
- **表情消息**：官方 bot 能力表无「表情作为消息内容」类型；纯文本 / markdown / Unicode 三路均证伪，仅能降级为 `[表情名]`。
- **主动消息频控**：Bot 维度 60 QPM（未认证 30 QPM），单关系 20 QPM，不适合高频刷屏。
- **官方接口能力有限**：踢人、改群名、拉取成员列表等官方 API 没有；群管理操作需机器人是群管理员。
- **资料不可得**：QQ 等级、性别、年龄官方不提供（`get_stranger_info` 返回占位 / QQ 等级恒 9999）。
- **链接域名校验**：消息里的链接需过 QQ 域名校验，gsk 提供二维码/短链两种规避方式（配置生成器⑤）。

---

## 八、维护约定

- 所有修改文件头部标注 `Modified by DanielToyama on YYYY-MM-DD (Gensokyo-ForSpark fork)`。
- 新增功能优先通过配置开关控制，默认值以「不破坏上游行为」为原则（例外：`at_markdown` 因实测收益明确默认开启）。
- 修改记录按日期倒序维护；本文件为完整记录，readme 仅保留摘要与链接。
- 提交信息使用 `type(scope): 摘要 - 细节` 格式，正文说明问题、修复、效果。
