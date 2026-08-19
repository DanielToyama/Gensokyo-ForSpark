

# Gensokyo-ForSpark

**本项目是基于 [hoshinonyaruko/gensokyo](https://github.com/hoshinonyaruko/gensokyo) 的 GPLv3 分叉版本**，针对 SparkBridge 群服互通场景进行了魔改，由 DanielToyama 独立维护，**不再与原上游同步**。

## 致谢 / Acknowledgements

- 原项目：[hoshinonyaruko/gensokyo](https://github.com/hoshinonyaruko/gensokyo)（[GPLv3](LICENSE)），原作者 [SanaeFox](https://github.com/Hoshinonyaruko) 及上游所有贡献者
- 本项目在原项目基础上增加了群服互通相关能力（详见下方说明），**原版权声明与 LICENSE 全文保留**，各修改文件头部已标注 `Modified by DanielToyama` 修改声明
- 若你使用本项目发现问题，请在本仓库反馈，**不要向上游提交与本项目相关的问题**

---

> ##  本 Fork 为针对[Sparkbridge-群服互通机器人](https://sparkbridge.cn)使用场景魔改开发版本
>
> 基于 [hoshinonyaruko/gensokyo](https://github.com/hoshinonyaruko/gensokyo) 上游分支，为满足群服互通场景需求，在官方代码基础上增加了以下修改（**未合并回官方，基本vibe加小测试，代码可靠性未知，请勿向官方提 PR**）：
>
> ### 新增能力
> 1. **群全量消息接收**（`GROUP_MESSAGE_CREATE`）
>    - 订阅后机器人可收到群里每条消息（不要求 @bot）
>    - 与 `GroupATMessageEventHandler`（@消息）可同时开启，内部按消息 ID 去重
> 2. **主动消息发送**（无 `msg_id` 直接发送）
>    - 配置 `allow_proactive_msg: true` 后，无被动窗口时也直接发送
>    - 需官方主动消息权限（否则报 `40034105 主动消息失败, 无权限`）
> 3. **`require_mention` 群消息响应开关**
>    - `false`（默认）= 全量消息都响应；`true` = 仅响应 @bot 消息
>    - 仅对全量事件生效，@ 事件不受影响
> 4. **群聊管理（官方 API v2 群管理板块，20260810+ 新增接口/事件）**
>    - 事件（新增 `GROUP_MEMBER_EVENT (1<<24)`；并入 `GROUP_AND_C2C_EVENT (1<<25)`）：
>      - `GROUP_MEMBER_ADD` → onebot `notice.group_increase`（官方事件无操作者字段, sub_type 默认 approve）
>      - `GROUP_MEMBER_REMOVE` → onebot `notice.group_decrease`（sub_type 默认 leave）
>      - `GROUP_JOIN_REQUEST`（用户申请加群，机器人需群管理员）→ onebot `request.group`(sub_type=add, flag=join_request_id)
>      - `SUBSCRIBE_MESSAGE_STATUS`（订阅消息授权状态变更）→ 扩展 `notice.subscribe_message_status`
>    - 操作（均走官方 v2 群管理接口，调用域名已按官方统一为 `api.bot.qq.com`）：
>      - `get_group_info` 群类型返回**真实群信息**（原先为模拟数据"测试群"）
>      - `set_group_ban` → 官方"设置群成员禁言"（`/restrict_chat_setting`，需群管理员，最长 30 天；duration=0 解禁）
>      - `set_group_whole_ban` → 全员禁言开关（官方文档暂未声明 global_rule 可写，best-effort 透传）
>      - `set_group_add_request` → 官方"入群申请审批"（approve/refuse，flag=join_request_id，reason=拒绝理由）
>      - 拓展API：`get_group_join_request_list`（入群申请列表，分页）、`get_group_restrict_chat_setting` / `set_group_restrict_chat_setting`（禁言状态查询 / 原始透传）、`get_group_bot_state`（机器人群内状态）
>      - 入群自动审批策略 6 个接口：`create/get/update/delete_join_approval_strategy` + `update_join_approval_strategy_whitelist` + `execute_join_approval_strategy`
> 5. **SparkBridge 群服互通适配**
>    - **群消息 at 转换**：官方 API 发送富文本实测不渲染任何 @ 标签，gsk 出站时自动把 at 段转成markdown真at或转换为 **`@昵称`** 文本（若需要发送图片情况下无法使用md，使用昵称缓存优先：每条群消息/加群申请都会缓存官方 username，**持久化 7 天、重启不丢**）；**昵称未知时显示 `@Openid+openid前8位`** 保底（让客户区分这不是真 QQ 号）
>    - **`get_stranger_info` 已实现**：昵称取官方事件缓存的 `username`（加群申请等事件携带）；`sex/age` 官方不提供返回空/0；**`qqLevel` 官方无 QQ 等级数据，固定 9999** 放行等级门槛（防插件按 0 级误拒）
>    - **任何 action 都有合法 JSON 回应**：未注册/出错的 action 回 `{"status":"failed","retcode":1400,"data":{"message":...},"echo":...}`，不再沉默（避免 SparkBridge 等对端等待超时拿 undefined 后 JSON5 解析崩溃）
>    - **`set_group_add_request` 支持只传 flag**：SparkBridge groupRequest 插件审批只传 `flag(join_request_id)`/`approve`/`reason` 时，从事件缓存（`join_request_id → 群/成员 openid`）反查后调官方审批接口；`approve=false` 即拒绝（不再要求 refuse 字段）
> 6. **FakeReply 假回复**（OneBot v11 `reply` 引用段的文本伪造实现）
>    - 官方协议不支持真实引用（参考 OneBot 11 规范 `send_group_msg` 的 `{"type":"reply","data":{"id":...}}` 引用段），但引用能让人辨别"在回复谁"
>    - Gensokyo 把**收到的每条群/私聊/频道消息**按 `message_id` 持久化发送者昵称与原文到本地 **`msgmap.db`**（bbolt，重复键覆盖，保留 7 天自动清理）
>    - 应用端（如 SparkBridge）发送带 `reply` 引用段的消息时，自动文本伪造为 `回复 @昵称` + `————` + 原消息内容，再接本条消息正文；支持与 text/at/image 等段**混合发送**；查不到记录时跳过伪造、按普通消息发送
>    - 配置开关 **`fake_reply`**（默认 `true`）：`false` = 忽略 reply 段
>    - **fakeReply 升级**（`at_markdown` 开启时生效）：前缀"回复 @昵称"的昵称部分替换为官方文本链真实 at 标签 `<qqbot-at-user id="openid"/>`（openid 从 msgmap 记录的 `UserID` 经 idmap 反查），整条消息随 markdown 升级发送，群里渲染出真实的"回复 @某某"；反查失败或私聊/频道场景回退为文本 @昵称
> 7. **真实 At（Markdown）AtMarkdown**（官方文本链 at 标签不渲染 → markdown 消息发真 at）
>    - 官方文本链文档给出的 at 格式 `<qqbot-at-user id="openid"/>`（旧协议 `<@userid>` 已废弃）**实测在纯文本消息不渲染**（显示原文）；频道 markdown 模板语法 `<at id="openid"></at>` 在**群聊自定义 markdown 同样实测不渲染**（2026-08-19 实测），此前 gsk 出站 at 一律转 **`@昵称`**（见能力 5）
>    - 实测：**markdown 消息**可渲染真实 at；且 2026/04/23 起**自定义 markdown 已开放到所有机器人**（单聊/群聊，无需申请模版）
>    - 配置开关 **`at_markdown`**（**默认 `true`**）：开启后，发送**含 at 段的群纯文本消息**时自动升级为 markdown 消息（`msg_type=2`, `markdown.content` 内嵌官方文档"最新格式" `<qqbot-at-user id="openid"/>`，openid 由 idmap 反查），尝试渲染真 at；**带图/语音等富媒体消息不升级**（富媒体通道无 at 渲染，at 还原为 @昵称；图文+at 不合并）
>    - 持续更新，更多与onebot11不兼容之处可反馈，我将尝试支持。
> ### 配置说明
> 🛠️ 不想手写配置？用**可视化配置生成器**：[打开 gensokyo-config-gen.html](https://htmlpreview.github.io/?https://raw.githubusercontent.com/DanielToyama/Gensokyo-spark/main/gensokyo-config-gen.html)（浏览器直接渲染成页面；也可本地双击仓库根目录同名文件，或在 SparkBridge 的「Gensokyo配置生成」页面使用内置同款）
> ```yaml
> text_intent:
>   - "GroupMessageEventHandler"      # 群全量消息（需官方开放权限）
>   - "GroupATMessageEventHandler"    # 群@消息
>   - "C2CMessageEventHandler"        # 群私聊
>   - "GroupMemberAddEventHandler"    # 群成员加入（需 GROUP_MEMBER_EVENT 权限）
>   - "GroupMemberRemoveEventHandler" # 群成员退出
>   - "GroupJoinRequestEventHandler"  # 用户申请加群（机器人需群管理员）
>   - "SubscribeMessageStatusEventHandler" # 订阅消息授权状态变更
> allow_proactive_msg: true           # 允许主动消息
> require_mention: false              # false=全量响应, true=仅@响应
> downtime_message_enabled: true      # 维护通知总开关: false=WS全部掉线时不回复维护文案
> downtime_cooldown: 10               # 维护回复冷却(分钟): 群/频道一对多,同一用户冷却期内最多回一次; 0=不冷却; 私聊/C2C不受影响
> downtime_message: ""                # 维护文案, 留空=不发送
> fake_reply: true                    # FakeReply假回复开关: true=应用端发reply引用段时文本伪造"回复 @昵称\n————\n原内容"; false=忽略reply段
> at_markdown: false                  # 真实At(Markdown)开关: 开启后含at的群文本消息升级为markdown消息发真at(实验功能); false=维持@昵称文本
> ```
>
> ### 修复
> - `ProcessGroupMessage` 类型转换问题：兼容 `WSGroupMessageData`（全量）与 `WSGroupATMessageData`（@）两种事件，避免下游 type switch 匹配失败导致消息被误拦
> - **官方 username 回填 `Sender.nickname/card`**：官方事件 `author.username` 已返回真实昵称（如 `Daniel_户山兔兔`），此前 OneBot 事件的 `sender.nickname/card` 恒为空。现在 `card_nick` 配置优先，否则回退使用官方 username 作为默认昵称（覆盖群消息 + C2C 私聊）
> - **群消息 @ 转文字**：官方 API 无可用真 @ 方案（`<qqbot-at-user id=.../>` 实测显示原文），出站 at 段自动转 `@昵称` 文本（昵称缓存优先；昵称未知时移除 at，见新增能力 5）
> - **入站 @ 转 CQ at**：官方群消息里的 @ 是 `<@openid>` 内嵌标签（旧正则只匹配 `<@!数字>`），现已转成 `[CQ:at,qq=xxx]` 体现在 raw_message/message——xxx 与 gsk 中该成员的 user_id 一致，下游插件可用它指定人（如"@某人 绑定白名单"）
> - **入站表情转 CQ face**：官方群消息里的表情是 `<faceType=1,faceId="264",ext="...">` 内嵌标记，现已转成 `[CQ:face,id=264]`（raw_message）与 `{type:"face",data:{id:"264"}}`（array 段），与 LLOneBot 等实现行为一致；**不管 faceType 是几（系统表情/动态表情/贴纸）一律按 faceId 转**（此前 faceType≠1 会返回空导致整条消息被误判为黑白名单拦截丢弃）
> - **维护通知（`downtime_message` 系列）行为修复/增强**（WS 全部掉线时的兜底回复，改动见 `Processor/Processor.go` 的 `BroadcastMessageToAll`）：① 未配置 `downtime_message` 时不再发送空消息；② 群聊@消息/频道@消息为**一对多**场景，按「群/频道+用户」冷却（时长由 **`downtime_cooldown`** 配置，分钟，`0`=不冷却），同一用户冷却期内只回一次（WS 掉线不再"收到什么回什么"刷屏）；③ **群全量消息（未@bot）与频道不at消息不再触发维护回复**；④ 私聊/C2C/频道私信为 **1 对 1** 场景，**每条都回**，不受冷却限制；⑤ 新增 **`downtime_message_enabled`** 总开关（`false`=完全不回复）；（修改由 GitHub 用户 [DanielToyama](https://github.com/DanielToyama) 完成，遵循上游 GPLv3）
>
> ### 修改记录（GPLv3）
> - **2026-08-18**（DanielToyama）：`get_status` 重写为 LLOneBot 文档结构（bot 状态）
>   - 修改文件：`handlers/get_status.go`、`botstats/botstats.go`、`readme.md`
>   - 变更内容：响应改为 `data.{online, good, stat}` + 顶层 `wording`（对齐 LLOneBot `bot状态` 文档）；`stat` 提供真实统计 `message_received`/`message_sent`/`last_message_time`/`startup_time`（后者为新增，来自进程启动时刻），移除原 go-cqhttp 风格写死的测试数据（`packet_received: 1000` 等）；`online/good` 直接为 `true`（能收到请求即说明连接正常）
> - **2026-08-18**（DanielToyama）：维护通知（`downtime_message` 系列）行为调整 + 新增配置项
>   - 修改文件：`Processor/Processor.go`、`Processor/ProcessGroupMessage.go`、`Processor/ProcessGuildNormalMessage.go`、`structs/structs.go`、`config/config.go`、`template/config_template.go`、根目录 `config.yml`、`gensokyo-config-gen.html`、`readme.md`
>   - 变更内容：见上方「修复」维护通知条目；新增 `downtime_message_enabled`（总开关）与 `downtime_cooldown`（冷却分钟）两个配置项，根目录 `config.yml` 已同步补齐，可视化配置工具 `gensokyo-config-gen.html` 新增「⑥ 维护通知」卡片
> - **2026-08-18**（DanielToyama）：新增 FakeReply 假回复（OneBot v11 `reply` 引用段的文本伪造）
>   - 修改文件：新增 `msgmap/msgmap.go`（message_id→发送者 持久化，bbolt `msgmap.db`）；`main.go`（初始化/关闭）、`Processor/ProcessGroupMessage.go`、`Processor/ProcessC2CMessage.go`、`Processor/ProcessGuildATMessage.go`（收到消息时记录）、`handlers/message_parser.go`（`reply` 段解析与伪造渲染）、`structs/structs.go`、`config/config.go`、`template/config_template.go`、根目录 `config.yml`、`gensokyo-config-gen.html`（新增「⑦ 假回复」卡片）、`readme.md`
>   - 变更内容：见上方「新增能力 6」；新增配置项 `fake_reply`（默认 `true`）；伪造格式 `回复 @昵称\n————\n原内容`，原内容优先取 reply 段的 `data.text`，否则取 msgmap 持久化的原文
> - **2026-08-18**（DanielToyama）：新增真实 At（Markdown）AtMarkdown（实验功能）
>   - 修改文件：`handlers/message_parser.go`（`transformMessageTextAt` 产出 `<at id="openid"></at>`）、`handlers/send_group_msg.go`（`maybeUpgradeToMarkdownAt`/`sendGroupMsgUpgraded`，含 at 的群纯文本消息升级 `msg_type=2` markdown）、`structs/structs.go`、`config/config.go`、`template/config_template.go`、根目录 `config.yml`、`gensokyo-config-gen.html`（新增「⑧ 真实 At」卡片）、`readme.md`
>   - 变更内容：见上方「新增能力 7」；新增配置项 `at_markdown`（默认 `false`）；官方文本链 at 标签 `<qqbot-at-user id="openid"/>` 纯文本实测不渲染，改用 markdown 消息按文档"最新格式"注入试渲染真 at；富媒体（图文/语音等）与超长消息不升级，且会把已注入的 at 标签**还原为 @昵称 文本**（防止原始标签当正文发出）；2026-08-19 实测 `<at id="openid"></at>`（频道模板语法）在群聊自定义 markdown 同样不渲染，故注入格式由 `<at>` 改为 `<qqbot-at-user>` 再实测
> - **2026-08-19**（DanielToyama）：fakeReply 假回复前缀真实 at 升级 + `at_markdown` 默认开启
>   - 修改文件：`handlers/message_parser.go`（`buildFakeReplyPrefix` 新增 `mdAt` 参数、`mdAtForFakeReply` 辅助函数）、`config/config.go`（`GetAtMarkdown` 默认 `true`）、`structs/structs.go`、`template/config_template.go`、根目录 `config.yml`、`gensokyo-config-gen.html`（「⑧ 真实 At」默认勾选开启）、`readme.md`
>   - 变更内容：`at_markdown` **默认开启**；fakeReply 假回复前缀"回复 @昵称"在群消息 + `at_markdown` 开启时升级为官方文本链真实 at 标签 `<qqbot-at-user id="openid"/>`（openid 从 msgmap 记录的 `UserID` 经 idmap 反查），整条消息随 markdown 升级发送，群里渲染真实"回复 @某某"；**私聊/频道无 markdown-at 升级链路**（会显示原始标签），回退为文本 @昵称；**图文+at 富媒体消息**维持原样（官方富媒体通道不渲染 at 标签，at 还原为 @昵称，不做公网图合并）
> - **2026-08-19**（DanielToyama）：Release workflow UPX 安装改直连 GitHub Release 下载（apt 镜像卡死修复）
>   - 修改文件：`.github/workflows/cross_compile.yml`、`readme.md`
>   - 变更内容：`sudo apt-get update` 依赖 azure.archive.ubuntu.com 镜像，该镜像（微软托管 runner 专用）抽风时 apt 无限重试，整个构建卡死在 Compress with UPX 步（2026-08-19 `v2026.08.19.144851` 实测）；改为直连 `github.com/upx/upx` Release 下载 UPX 二进制（走 objects.githubusercontent.com CDN，不依赖 apt 镜像），下载失败才回退 apt（带 `Acquire::Retries=3` 重试上限，不再无限挂起）
> - **2026-08-19**（DanielToyama）：Release workflow UPX 二进制统一用 linux amd64 版（上门修复的 win64 版 bug）
>   - 修改文件：`.github/workflows/cross_compile.yml`、`readme.md`
>   - 变更内容：一次修复在 `matrix.os=windows` 时下载 `upx-4.2.4-win64.zip` 并在 runner 上执行——GitHub runner 是 **Linux 环境，无法运行 Windows 的 PE 程序**，windows 目标任务直接失败；linux amd64 版 UPX **同时支持压缩 PE(.exe)与 ELF**，故所有目标统一下载 linux amd64 版 UPX 并直连 GitHub Release（不走 apt 镜像），apt 仅作兜底（带重试上限）
> - **2026-08-20**（DanielToyama）：修复 `delete_msg` 撤回两处问题（响应结构误用 + 缺参静默不执行）
>   - 修改文件：`handlers/delete_msg.go`、`readme.md`
>   - 变更内容：① 响应误用 `GetStatusResponse`（返回了 get_status 的 `data.good/online/stat`），改为标准 onebot v11 成功响应 `data:null`；② 标准 `delete_msg` 仅传 `message_id`，原实现要求 `group_id/user_id` 非空才执行、缺省时四个撤回分支全部跳过（日志只显示一个可疑响应、实际什么都没撤），现在缺参时从 msgmap 缓存（群/私聊消息入库时记录的 GroupID/UserID）反查撤回对象再执行 `RetractGroupMessage` / `RetractC2CMessage`，并打印反查日志
> - **2026-08-20**（DanielToyama）：修复群禁言被 `guild_id` 配置读取卡死（`set_group_ban` / `set_group_whole_ban` 群聊分支永远到不了）
>   - 修改文件：`handlers/set_group_ban.go`、`handlers/set_group_whole_ban.go`、`readme.md`
>   - 变更内容：两个 handler 把 `idmap.ReadConfigv2(groupID, "guild_id")`（仅频道场景需要）放在 switch 之前——普通群聊从未存过该配置，查询必报 `key 'guild_id' in section '...' does not exist` 并直接 return，**群聊禁言/全员禁言实际从未执行**（应用端只看到报错日志、由于是 fire-and-forget 还看不到失败反馈）；已将 `guild_id` 读取移入 `case "guild"` 分支，群聊分支正常走 `POST /v2/groups/{openid}/restrict_chat_setting`（成员级 `members[].mute_expire_at` / 全员 `global_rule.mode`）
> - **2026-08-19**（DanielToyama）：入群申请 comment 问答格式改为 onebot 标准 `问题：…\n答案：…`
>   - 修改文件：`handlers/group_mgmt_common.go`（`BuildJoinRequestComment`）、`readme.md`
>   - 变更内容：问答验证（`admin_review_qa`）的 comment 展示由 `问:Q 答:A；问:Q 答:A` 改为 **onebot 标准格式** `问题：<问题>\n答案：<答案>`（全角冒号，多组问答依序换行）——官方与其他 onebot 客户端均是此格式，下游按标准格式解析/匹配答案；无回答时省略"答案："行；`request.group` 事件 comment 与 `get_group_join_request_list` 的 comment 增强字段同时生效
> - **2026-08-18**（DanielToyama）：入群审核 comment 增强 + 空值留痕（对应用户反馈"审核内容变成了空"）
>   - 修改文件：`Processor/ProcessGroupJoinRequest.go`、`handlers/group_mgmt_common.go`、`handlers/get_group_join_request_list.go`、`readme.md`
>   - 变更内容：申请词 `comment` 支持**问答验证**（`verify_info.method=admin_review_qa`，官方不回填 `verify_message`，问题和回答在 `verify_info.review_qa_list`）——拼接为 `问:… 答:…` 展示；**`get_group_join_request_list` 每条申请附加 `comment` 增强字段**（原字段不变）；`comment` 最终仍为空时打印**留痕日志**（含完整 `verify_info`），下次复发即可区分"申请人没填验证消息 / 官方事件字段缺失 / 偶发丢字段"
> - **2026-08-18**（DanielToyama）：发布物压缩方式与 CI 对齐（workflow 写法未改动）
>   - 本地 `gensokyo.exe` 改用与 `.github/workflows/cross_compile.yml` 相同的压缩方式：`-s -w` 构建后 `upx --best`，32,862,208 B → 9,634,304 B（29.3%，与原 workflow Release 产物体积一致）
> - **2026-08-18**（DanielToyama）：修复 Release workflow 产物丢失（`.github/workflows/cross_compile.yml`）
>   - flatten 步骤改为拷贝到 `output/release/`：此前把 `gensokyo-linux-amd64` 等二进制直接拷到 `output/`，与同名 artifact 目录冲突导致 `cp` 失败（错误被 `2>/dev/null || true` 吞掉），Release 附件只剩 windows 一个
> - **2026-08-17**（DanielToyama）：fork 建立：群全量消息接收、主动消息、`require_mention`、群管理接口、SparkBridge 互通适配（详见上方新增能力/修复，所有上游 LICENSE/copyright 声明保留）
>
> ### 方案优势（对比普通QQ小号挂机方案）
> - **官方机器人接入**：机器人是官方开放平台注册的应用，**无需普通 QQ 小号挂机**（不需要手机/电脑保持 QQ 在线、不会被挤下线）
> - **不掉线**：走官方 WebSocket 网关 + 自动重连，无登录态过期、无第三方协议风控下线，长期稳定运行
> - **全局信息**：消息经官方云端 API 全局收发，与任何客户端/设备解耦，重启进程即恢复
> - **不易封号**：官方通道合法合规，不存在腾讯封杀第三方协议的风险
>
> ### 方案边界（官方 API v2 限制，请先了解）
> - **无真 @**：@ 仅以文本显示（官方不渲染标签，出站自动转 `@昵称` 文本，见新增能力 5）
> - **主动消息频控**：Bot 维度 60 QPM（未认证 30 QPM），单关系 20 QPM，不适合高频刷屏场景
> - **官方接口能力有限**：踢人、改群名、拉取成员列表等官方 API 没有，这类功能无法实现
> - **群管理操作需管理员**：入群审批、群禁言等需要机器人是群管理员，由群主在手机 QQ 客户端给机器人配置权限
> - **部分资料不可得**：QQ 等级、性别、年龄官方不提供（`get_stranger_info` 返回占位/QQ等级恒 9999）
> - **链接域名校验**：消息里的链接需过 QQ 域名校验，gsk 提供**二维码/短链**两种规避方式（配置生成器⑤）
>
> ### 已知待办
> - **回调按钮唤醒机制**（规避主动消息 60/min 频控）：Gensokyo 已有按钮回调 → `event_id` 缓存的基础链路（`ProcessInlineSearch` → `echo.AddEvnetID`），但缺少"无可用 event_id 时自动发送回调按钮消息唤醒"的闭环。若未来群服互通高频推送触发频控，需在 `send_group_msg.go` 补上主动唤醒逻辑（参考 amsghook 的 CALLBACK_KEYBOARD + event_id 用次管理 5 次淘汰）
>
> ---

<p align="center">
  <a href="https://www.github.com/hoshinonyaruko/gensokyo">
    <img src="images/head.gif" width="200" height="200" alt="gensokyo">
  </a>
</p>
</div>
<div align="center">

# gensokyo

_✨ 基于 [OneBot](https://github.com/howmanybots/onebot/blob/master/README.md) QQ官方机器人Api Golang 原生实现 ✨_  


<p align="center">
  <a href="https://raw.githubusercontent.com/hoshinonyaruko/gensokyo/main/LICENSE">
    <img src="https://img.shields.io/github/license/hoshinonyaruko/gensokyo" alt="license">
  </a>
  <a href="https://github.com/hoshinonyaruko/gensokyo/releases">
    <img src="https://img.shields.io/github/v/release/hoshinonyaruko/gensokyo?color=blueviolet&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/howmanybots/onebot/blob/master/README.md">
    <img src="https://img.shields.io/badge/OneBot-v11-blue?style=flat&logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABABAMAAABYR2ztAAAAIVBMVEUAAAAAAAADAwMHBwceHh4UFBQNDQ0ZGRkoKCgvLy8iIiLWSdWYAAAAAXRSTlMAQObYZgAAAQVJREFUSMftlM0RgjAQhV+0ATYK6i1Xb+iMd0qgBEqgBEuwBOxU2QDKsjvojQPvkJ/ZL5sXkgWrFirK4MibYUdE3OR2nEpuKz1/q8CdNxNQgthZCXYVLjyoDQftaKuniHHWRnPh2GCUetR2/9HsMAXyUT4/3UHwtQT2AggSCGKeSAsFnxBIOuAggdh3AKTL7pDuCyABcMb0aQP7aM4AnAbc/wHwA5D2wDHTTe56gIIOUA/4YYV2e1sg713PXdZJAuncdZMAGkAukU9OAn40O849+0ornPwT93rphWF0mgAbauUrEOthlX8Zu7P5A6kZyKCJy75hhw1Mgr9RAUvX7A3csGqZegEdniCx30c3agAAAABJRU5ErkJggg==" alt="gensokyo">
  </a>
  <a href="https://github.com/hoshinonyaruko/gensokyo/actions">
    <img src="images/badge.svg" alt="action">
  </a>
  <a href="https://goreportcard.com/report/github.com/hoshinonyaruko/gensokyo">
  <img src="https://goreportcard.com/badge/github.com/hoshinonyaruko/gensokyo" alt="GoReportCard">
  </a>
</p>

<p align="center">
  <a href="https://github.com/howmanybots/onebot/blob/master/README.md">文档</a>
  ·
  <a href="https://github.com/hoshinonyaruko/gensokyo/releases">下载</a>
  ·
  <a href="https://github.com/hoshinonyaruko/gensokyo/releases">开始使用</a>
  ·
  <a href="https://github.com/hoshinonyaruko/gensokyo/blob/master/CONTRIBUTING.md">参与贡献</a>
</p>
<p align="center">
  <a href="https://gensokyo.bot">项目主页:gensokyo.bot</a>
</p>

## 引用
- [`tencent-connect/botgo`](https://github.com/tencent-connect/botgo): 本项目引用了此项目,并做了一些改动.

## 介绍
gensokyo兼容 [OneBot-v11](https://github.com/botuniverse/onebot-11) ，并在其基础上做了一些扩展，详情请看 OneBot 的文档。

Gensokyo文档(施工中):[起步](/docs/起步-注册QQ开放平台&启动gensokyo.md)

可将官方的websocket和api转换至onebotv11标准,

支持连接koishi,nonebot2,trss,zerobot,MiraiCQ,hoshino..

支持连接tata,派蒙,炸毛,早苗,yobot...

支持连接Mirai(Overflow)...

可以与支持onebotV11适配器的项目相连接使用.

实现插件开发和用户开发者无需重新开发,复用过往生态的插件和使用体验.

持续完善中.....交流群:196173384

欢迎测试,询问任何有关使用的问题,有问必答,有难必帮~

[Gensokyo临时文档](https://www.yuque.com/km57bt/hlhnxg/mw7gm8dlpccd324e)展开左侧折叠栏,临时文档包含markdown定义、额外api文档等内容

后续会将文档独立，因为语雀文档公开查看无需登录需要vip，故暂时放在我的机器人文档中。临时文档也包含了Gensokyo的完整编译教程。

## 特别鸣谢

- [`mnixry/nonebot-plugin-gocqhttp`](https://github.com/mnixry/nonebot-plugin-gocqhttp/): 本项目采用了mnixry编写的前端,并实现了与它对应的,基于qq官方api的后端api.
- 特别鸣谢[`dk 盾`](https://www.dkdun.cn/),友情赞助服务器资源

### 接口

- [x] HTTP API
- [x] 反向 HTTP POST
- [x] 正向 WebSocket
- [x] 反向 WebSocket

### 拓展支持

> 拓展 API 可前往 [文档](docs/cqhttp.md) 查看

- [x] 连接多个ws地址
- [x] 将频道虚拟成群事件
- [x] 将私信虚拟成频道或群事件
- [x] webui,可以在webui修改配置,查看频道列表,发送信息
- [x] 方便过审的指令黑白名单
- [x] 自动url转换(自备域名)
- [x] 可自定义图片压缩\图床服务
- [x] 可编辑的数据库
- [x] 支持array和信息段
- [x] 文字,图片,语音,视频,MD,支持多种类型发送
- [x] 支持全域,频道,频道私聊,群,群私聊
- [x] 主动信息失败自动转被动,提高信息传达可靠性
- [x] 提前于官方支持群列表 群成员 api
- [x] 完善的重连,健壮的连接能力.
- [x] 支持[CQ:markdown,data=] Markdown发送
- [x] [`markdown文档`](https://www.yuque.com/km57bt/hlhnxg/ddkv4a2lgcswitei)
- [x] 持续更新~


### 实现

<details>
<summary>已实现 CQ 码</summary>

#### 符合 OneBot 标准的 CQ 码

| CQ 码        | 功能                        |
| ------------ | --------------------------- |
| [CQ:face]    | [QQ 表情]                   |
| [CQ:record]  | [语音]                      |
| [CQ:video]   | [短视频]                    |
| [CQ:at]      | [@某人]                     |
| [CQ:share]   | [链接分享]                  |
| [CQ:music]   | [音乐分享] [音乐自定义分享] |
| [CQ:reply]   | [回复]                      |
| [CQ:forward] | [合并转发]                  |
| [CQ:node]    | [合并转发节点]              |
| [CQ:xml]     | [XML 消息]                  |
| [CQ:json]    | [JSON 消息]                 |

todo,正在施工中

#### 拓展 CQ 码及与 OneBot 标准有略微差异的 CQ 码

| 拓展 CQ 码     | 功能                              |
| -------------- | --------------------------------- |
| [CQ:image]     | [图片]                            |
| [CQ:poke]      | [戳一戳]                          |
| [CQ:node]      | [合并转发消息节点]                |
| [CQ:markdown]  | [markdown卡片收发] |
| [CQ:tts]       | [文本转语音]                      |


</details>

<details>
<summary>已实现 API</summary>

#### 符合 OneBot 标准的 API

| API                      | 功能                   |
| ------------------------ | ---------------------- |
| /send_private_msg√        | [发送私聊消息]         |
| /send_group_msg√         | [发送群消息]           |
| /send_guild_channel_msg√ | [发送频道消息]         |
| /send_msg√               | [发送消息]             |
| /delete_msg              | [撤回信息]             |
| /set_group_kick          | [群组踢人]             |
| /set_group_ban√          | [群组单人禁言]         |
| /set_group_whole_ban√    | [群组全员禁言]         |
| /set_group_admin         | [群组设置管理员]       |
| /set_group_card          | [设置群名片（群备注）] |
| /set_group_name          | [设置群名]             |
| /set_group_leave         | [退出群组]             |
| /set_group_special_title | [设置群组专属头衔]     |
| /set_friend_add_request  | [处理加好友请求]       |
| /set_group_add_request   | [处理加群请求/邀请]    |
| /get_login_info√         | [获取登录号信息]       |
| /get_stranger_info       | [获取陌生人信息]       |
| /get_friend_list√        | [获取好友列表]         |
| /get_group_info√          | [获取群/频道信息]     |
| /get_group_list√         | [获取群列表]           |
| /get_group_member_info√  | [获取群成员信息]       |
| /get_group_member_list√  | [获取群成员列表]       |
| /get_group_honor_info    | [获取群荣誉信息]       |
| /can_send_image√         | [检查是否可以发送图片] |
| /can_send_record         | [检查是否可以发送语音] |
| /get_version_info√       | [获取版本信息]         |
| /set_restart√             | [重启 gensokyo]       |
| /.handle_quick_operation | [对事件执行快速操作]   |


#### 拓展 API 及与 OneBot 标准有略微差异的 API

| 拓展 API                    | 功能                   |
| --------------------------- | ---------------------- |
| /set_group_portrait         | [设置群头像]           |
| /get_image                  | [获取图片信息]         |
| /get_msg                    | [获取消息]             |
| /get_forward_msg            | [获取合并转发内容]     |
| /send_group_forward_msg√     | [发送合并转发(群)]     |
| /.get_word_slices           | [获取中文分词]         |
| /.ocr_image                 | [图片 OCR]             |
| /get_group_system_msg       | [获取群系统消息]       |
| /get_group_file_system_info | [获取群文件系统信息]   |
| /get_group_root_files       | [获取群根目录文件列表] |
| /get_group_files_by_folder  | [获取群子目录文件列表] |
| /get_group_file_url         | [获取群文件资源链接]   |
| /get_status√                 | [获取状态]             |


</details>

<details>
<summary>已实现 Event</summary>

#### 符合 OneBot 标准的 Event（部分 Event 比 OneBot 标准多上报几个字段，不影响使用）

| 事件类型 | Event            |
| -------- | ---------------- |
| 消息事件 | [私聊信息]√       |
| 消息事件 | [群消息]√         |
| 通知事件 | [群文件上传]     |
| 通知事件 | [群管理员变动]   |
| 通知事件 | [群成员减少]     |
| 通知事件 | [群成员增加]     |
| 通知事件 | [群禁言]         |
| 通知事件 | [好友添加]       |
| 通知事件 | [群消息撤回]     |
| 通知事件 | [好友消息撤回]   |
| 通知事件 | [群内戳一戳]     |
| 通知事件 | [群红包运气王]   |
| 通知事件 | [群成员荣誉变更] |
| 请求事件 | [加好友请求]     |
| 请求事件 | [加群请求/邀请]  |


#### 拓展 Event

| 事件类型 | 拓展 Event       |
| -------- | ---------------- |
| 通知事件 | [好友戳一戳]     |
| 通知事件 | [群内戳一戳]     |
| 通知事件 | [群成员名片更新] |
| 通知事件 | [接收到离线文件] |


</details>

## 关于 ISSUE

以下 ISSUE 会被直接关闭

- 提交 BUG 不使用 Template
- 询问已知问题
- 提问找不到重点
- 重复提问

> 请注意, 开发者并没有义务回复您的问题. 您应该具备基本的提问技巧。  
> 有关如何提问，请阅读[《提问的智慧》](https://github.com/ryanhanwu/How-To-Ask-Questions-The-Smart-Way/blob/main/README-zh_CN.md)

## 性能

10mb内存占用 端口错开可多开 稳定运行无报错
