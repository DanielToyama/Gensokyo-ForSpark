# Gensokyo-ForSpark

**QQ 官方机器人 → OneBot v11 适配层，为 SparkBridge 群服互通场景定制。**

基于 [hoshinonyaruko/gensokyo](https://github.com/hoshinonyaruko/gensokyo) 的 GPLv3 分叉，由 DanielToyama 独立维护。**无需 QQ 小号挂机**，走官方机器人 API 稳定接入；兼容 SparkBridge、Koishi、NoneBot2、Mirai 等 OneBot v11 生态。

> **第一次使用？** 直接打开 **[配置生成器](https://htmlpreview.github.io/?https://raw.githubusercontent.com/DanielToyama/Gensokyo-Forspark/main/gensokyo-config-gen.html)**，3 分钟生成 `config.yml`；再按 [快速开始](#-快速开始3-分钟) 启动即可，不用手写配置。

## 🚀 快速开始（3 分钟）

### 第 1 步：下载程序

1. 打开 [Releases](https://github.com/DanielToyama/Gensokyo-Forspark/releases)
2. 下载对应平台的文件：
   - Windows：`gensokyo-windows-amd64.exe`
   - Linux x86_64：`gensokyo-linux-amd64`
   - Linux ARM64（如甲骨文/树莓派）：`gensokyo-linux-arm64`
3. 放到一个**空目录**（例如 `D:\Gensokyo`），确保该目录可读写。

### 第 2 步：用配置生成器生成 config.yml（推荐）

**不用手写配置。** 打开配置生成器：

- **在线版（推荐）**：[打开配置生成器](https://htmlpreview.github.io/?https://raw.githubusercontent.com/DanielToyama/Gensokyo-Forspark/main/gensokyo-config-gen.html)
- **本地版**：双击仓库根目录的 `gensokyo-config-gen.html`（用浏览器打开）
- **SparkBridge 内置**：SparkBridge 的「Gensokyo配置生成」页面使用同款

填写页面上的 8 个卡片 → 点 **「生成 config.yml」** → 点 **「下载文件」** 或 **「复制内容」** → 保存到 Gensokyo 程序**同目录**，文件名必须是 `config.yml`。

### 第 3 步：启动

- Windows：双击 `gensokyo-windows-amd64.exe`
- Linux：
  ```bash
  chmod +x gensokyo-linux-amd64
  ./gensokyo-linux-amd64
  ```
- 日志出现监听端口（默认 `15630`）即启动成功
- SparkBridge 正向 WS 地址填 `ws://127.0.0.1:15630`，密码填配置生成器里的「正向WS令牌」

---

## ⚙️ 配置生成器详解

> 配置生成器只问「群服互通真正需要的」选项，其余全部使用推荐默认值。生成的 `config.yml` 直接替换程序同目录的同名文件即可。

### ① 机器人账号（必填）

| 字段 | 填什么 | 从哪里获取 |
|---|---|---|
| 应用ID `app_id` | 机器人应用 ID（纯数字） | QQ 开放平台 → 应用管理 |
| 客户端密钥 `client_secret` | App Secret（用于换取 access_token，唯一鉴权凭据） | QQ 开放平台 → 开发设置 |
| 机器人QQ号 `uin` | 机器人资料卡上的 QQ 号 | 手机 QQ 打开机器人资料卡查看 |

> 三个都必填。`app_id` 与 `uin` 不是一回事：前者是开放平台应用 ID，后者是机器人对外展示的 QQ 号。
>
> **旧版「应用令牌 token」字段已废弃**：官方 API v2 只需 `app_id` + `client_secret` 动态换取 access_token（7200 秒有效），配置里的 `token` 不再参与鉴权，已从配置生成器和模板中移除。
>
> 还没有创建机器人应用？先看 [起步：注册 QQ 开放平台 & 启动 Gensokyo](docs/起步-注册QQ开放平台&启动gensokyo.md)。

### ② 服务器连接（必填）

| 字段 | 推荐值 | 说明 |
|---|---|---|
| 监听端口 `port` | `15630` | 正向 WS / WebUI / API 共用此端口；被占用就换一个，SparkBridge 要同步改 |
| 正向WS令牌 `ws_server_token` | 自己设一个随机字符串 | SparkBridge `base/config.json` 的 `pwd` 必须与它一致 |

### ③ 消息行为（推荐配置）

- **事件订阅 `text_intent`**：默认全选（群 @ 消息、群全量消息、私聊、群成员加入/退出、用户申请加群、订阅消息状态、机器人被拉入/移出群、群开关机器人推送）。
  - **群全量消息**需要官方开放权限；没有权限可取消勾选，只保留群 @ 消息。
  - **群管理事件**（入群审批、群禁言等）需要机器人在群里是**群管理员**，由群主在手机 QQ 群设置里配置。
- **响应模式 `require_mention`**：默认「全量响应（所有消息都处理）」；如果只想让机器人响应 @ 它的消息，选「仅响应 @机器人」。
- **主动消息 `allow_proactive_msg`**：推荐「允许主动发送」；需官方主动消息权限，否则会报 `40034105 主动消息失败, 无权限`。
- **日志级别 `log_level`**：`info`（推荐）；排查问题选 `debug`，长期运行选 `error` 减少日志量。

### ④ 群消息 @ 设置（默认关闭，按下游插件需要开启）

QQ 官方机器人对 @ 的支持要分场景看，不是“完全不支持”：

- **群 markdown 消息可以渲染真实 @**：本 fork 的 `at_markdown` 默认开启，含 at 的群纯文本消息会自动升级为 markdown 发送真 at
- **纯文本通道里的 `<qqbot-at-user>` 标签本身不渲染**（显示原文）；关闭 `at_markdown` 时才降级为 `@昵称` 文本
- 2026-10-26 起**纯文本消息**支持 `<qqbot-at-everyone />` @全体成员

**消息形态与 @ 能力对照**：

| 消息形态 | 能否真 @ | 适配行为 |
|---|---|---|
| markdown 消息（含 markdown 图片 / `qqbot-media://` 资源） | ✅ 可以 | 保留 `<qqbot-at-user>` / `<qqbot-at-everyone />` |
| 纯文本消息 | ⚠️ 单成员 at 不渲染，`@全体成员` 可以 | `at_markdown` 开启时单成员 at 升级 markdown；关闭时降级 `@昵称` |
| 富媒体图片/语音/视频（`RichMediaMessage` / `MsgType=7`） | ❌ 不可以 | at 自动还原为 `@昵称` 文本、`@全体成员` 还原为文本；不升级 markdown，避免原始标签当正文发出 |

> 部分用户没有 markdown 发图条件，只能走富媒体图片/语音/视频通道——这条通道官方不支持 at 标签，所以**无法真 @**，本 fork 会自动降级为文本，不会把 `<qqbot-at-user>` 原始标签发给对方。

下面三个开关默认都关闭，但作用在**完全不同的环节**，需要根据下游插件的行为选择，不能一概认为不需要开：

| 开关 | 作用环节 | 什么时候需要开 |
|---|---|---|
| `remove_at` | 接收侧 | 收到的群 @ 消息开头带 `[CQ:at,qq=机器人]`，而下游插件只识纯指令、不认识 at 开头时开启 |
| `add_at_group` | 接收侧 | 下游插件必须看到 @ 才响应时开启（给收到的群消息前补 @机器人，与 `remove_at` 相反） |
| `remove_bot_at_group` | 发送侧 | 不想让发出去的群消息出现任何普通成员 at 时开启：`at_markdown` 开启时删真 at，关闭时删 `@昵称` 文本；`@全体成员` 不受影响 |

### ⑤ 链接发送方式

QQ 官方机器人对消息里的链接有域名校验，陌生域名可能被拦截或不可点击。三选一：

| 模式 | 说明 |
|---|---|
| **二维码模式（推荐）** | 链接转成二维码图片发出，无需域名/端口/备案 |
| 短链模式 | 链接转成 `你的域名/url/xxx`，需要在下方填 `server_dir` 并在 QQ 后台校验域名 |
| 不转换 | 原样发送，可能被官方拦截或不可点击 |

### ⑥ 维护通知（WS 全部掉线时的兜底回复）

当 Gensokyo 到应用端的 WS 连接**全部失败**时，收到消息会通过官方 API 回一条「维护中」文案：

- **总开关** `downtime_message_enabled`：关闭后完全不回复
- **冷却分钟** `downtime_cooldown`：默认 `10`；群聊/频道是**一对多**场景，同一「群/频道 + 用户」冷却期内最多回一次；`0` = 不冷却；私聊/C2C 是 **1 对 1**，每条都回，不受冷却限制
- **维护文案** `downtime_message`：留空 = 不发送（等同于关闭通知）

### ⑦ 假回复 FakeReply

QQ 官方机器人**不支持** OneBot v11 的 `reply` 引用段（`{"type":"reply","data":{"id":...}}`）。开启后，Gensokyo 会把这类消息**文本伪造**为：

```
回复 @昵称
————
被回复的那条消息原文
[本条消息内容]
```

- 默认开启；收到的每条消息都会按 `message_id` 持久化发送者昵称与原文（存于 `msgmap.db`，保留 7 天）
- 查不到记录时跳过伪造，按普通消息发送
- 关闭 `fake_reply` 后直接忽略 `reply` 段
- 配合 `at_markdown` 开启时，「回复 @昵称」的昵称部分会升级为真实 at

### ⑧ 真实 At（Markdown）AtMarkdown

官方文本链的 at 格式 `<qqbot-at-user id="openid"/>` 在**纯文本**消息里实测**不渲染**（显示原文）；频道 markdown 模板语法 `<at id="openid"></at>` 在群聊自定义 markdown 里实测同样不渲染。但**markdown 消息**可以渲染真实 at。

- **默认开启**（建议先在测试群验证；可在此关闭）
- 开启后，**含 at 的群纯文本消息**自动升级为 markdown 消息（`msg_type=2`）发送真 at
- 带图/语音等富媒体消息、超长消息**不升级**（at 还原为 `@昵称`）
- 关闭后维持 `@昵称` 文本

### 生成之后怎么用

1. 点 **「生成 config.yml」**
2. 点 **「下载文件」** 保存为 `config.yml`，或点 **「复制内容」** 粘贴到新建的 `config.yml`
3. 把 `config.yml` 放到 Gensokyo 程序**同目录**（覆盖旧文件）
4. 重启 Gensokyo

> 生成器没有提供的选项（如 `custom_bot_name` 机器人昵称、`title` 标题、HTTP API 等）可以直接编辑生成的 `config.yml` 手动修改。

### 配置器常见问题

| 问题 | 解决办法 |
|---|---|
| 不知道 client_secret 在哪 | QQ 开放平台 → 你的应用 → 开发设置 |
| 端口被占用 / 启动报错 | 换一个 `port`，并同步修改 SparkBridge 的连接地址 |
| SparkBridge 连不上 | 检查正向 WS 地址 `ws://127.0.0.1:<port>`，以及 `ws_server_token` 与 SparkBridge 的 `pwd` 是否一致 |
| 配置生成了但没生效 | 确认文件名是 `config.yml`、放在程序同目录、并重启了进程 |
| 机器人不回复群消息 | 确认 `text_intent` 勾了对应事件；群全量消息需官方权限；群管理事件需机器人是管理员 |
| 想改机器人昵称 | 生成后手动修改 `custom_bot_name` |

---

## ✨ 核心能力

| 能力 | 说明 |
|---|---|
| **群全量消息** | 订阅后可收到群内每条消息，不要求 @bot（需官方权限） |
| **主动消息发送** | 无被动 `msg_id` 窗口也可直接发送（需官方权限） |
| **群管理 API** | 入群审批、申请列表、群禁言/全员禁言、机器人群内状态、入群自动审批策略 |
| **群管理事件** | 群成员加入/退出、入群申请、订阅消息状态 → OneBot notice/request |
| **FakeReply 假回复** | `reply` 引用段文本伪造「回复 @昵称」 |
| **真实 At（Markdown）** | 含 at 的群文本升级 markdown，渲染真实 @ |
| **表情消息适配** | 入站对齐 OneBot（`subType=1` / 表情名）；出站 face 段降级 `[表情名]` |
| **撤回自己的消息** | 标准 `delete_msg` 只传 `message_id` 即可撤回机器人自己发的消息 |
| **@全体成员** | `[CQ:at,qq=all]` → 官方 `<qqbot-at-everyone />`（2026-10-26 新特性） |
| **头像兼容路由** | `GET /g?b=qq&nk=数字ID&s=尺寸` 兼容旧 qlogo 链接 |

> **@全体成员用法**：OneBot 发 `[CQ:at,qq=all]`（数组段 `{"type":"at","data":{"qq":"all"}}`）即可，适配层自动转换为官方 `<qqbot-at-everyone />`。仅群聊生效，文本消息和 markdown 消息都支持；机器人必须是群管理员，且**每个群每天只能发送 1 次**（官方服务端校验，超限会报错）。

以上能力的实现细节、实测结论和完整 API 参数见 **[DanielToyama 更新记录](docs/DanielToyama更新记录.md)** 与 **[适配 API 清单](docs/适配API清单.md)**。

## 📋 配置参考（手动配置）

不想用配置生成器时，可参考以下最小配置。完整字段说明见根目录 `config.yml` 注释。

```yaml
settings:
  app_id: 1234567890
  uin: 4015000000
  client_secret: "你的客户端密钥"

  text_intent:
    - "GroupATMessageEventHandler"          # 群 @ 消息
    - "GroupMessageEventHandler"            # 群全量消息（需官方权限）
    - "C2CMessageEventHandler"              # 私聊
    - "GroupMemberAddEventHandler"          # 群成员加入
    - "GroupMemberRemoveEventHandler"       # 群成员退出
    - "GroupJoinRequestEventHandler"        # 用户申请加群（需群管理员）
    - "SubscribeMessageStatusEventHandler"  # 订阅消息授权状态变更

  port: "15630"
  enable_ws_server: true
  ws_server_token: "你的正向WS令牌"

  allow_proactive_msg: true   # 允许主动消息
  require_mention: false      # false=全量响应, true=仅@响应
  fake_reply: true            # FakeReply 假回复
  at_markdown: true           # 真实 At（Markdown）

  downtime_message_enabled: true
  downtime_cooldown: 10
  downtime_message: "正在维护中，请稍后再试~"
```

## 🛠 方案说明

### 方案优势（对比普通 QQ 小号挂机方案）

- **官方机器人接入**：机器人是官方开放平台注册的应用，**无需普通 QQ 小号挂机**（不需要手机/电脑保持 QQ 在线、不会被挤下线）
- **不掉线**：走官方 WebSocket 网关 + 自动重连，无登录态过期、无第三方协议风控下线，长期稳定运行
- **全局信息**：消息经官方云端 API 全局收发，与任何客户端/设备解耦，重启进程即恢复
- **不易封号**：官方通道合法合规，不存在腾讯封杀第三方协议的风险

### 方案边界（官方 API v2 限制，请先了解）

- **@ 能力分场景**：群 markdown 消息可渲染真 @（`at_markdown` 默认开启）；纯文本里的单成员 at 不渲染，降级为 `@昵称`；2026-10-26 起纯文本支持 `@全体成员`
- **主动消息频控**：Bot 维度 60 QPM（未认证 30 QPM），单关系 20 QPM，不适合高频刷屏场景
- **官方接口能力有限**：踢人、改群名、拉取成员列表等官方 API 没有，这类功能无法实现
- **群管理操作需管理员**：入群审批、群禁言等需要机器人是群管理员
- **部分资料不可得**：QQ 等级、性别、年龄官方不提供（`get_stranger_info` 返回占位 / QQ 等级恒 9999）
- **链接域名校验**：消息里的链接需过 QQ 域名校验，可用配置生成器的二维码/短链模式规避

### 已知待办

- **回调按钮唤醒机制**（规避主动消息 60/min 频控）：Gensokyo 已有按钮回调 → `event_id` 缓存的基础链路（`ProcessInlineSearch` → `echo.AddEvnetID`），但缺少「无可用 event_id 时自动发送回调按钮消息唤醒」的闭环。若未来群服互通高频推送触发频控，需在 `send_group_msg.go` 补上主动唤醒逻辑。

---

## 📚 文档导航

| 文档 | 内容 |
|---|---|
| [DanielToyama 更新记录](docs/DanielToyama更新记录.md) | 全部 45 笔提交、按能力域整理的完整改动记录（GPLv3） |
| [适配 API 清单](docs/适配API清单.md) | 本 fork 新增/专门适配的 API、参数、事件与 HTTP 路由 |
| [撤回消息](docs/撤回消息.md) | `delete_msg` 完整用法、四种会话对照、常见坑 |
| [API 介绍](docs/api介绍.md) | 完整 API 文件索引（49 个动作） |

## 致谢 / Acknowledgements

- 原项目：[hoshinonyaruko/gensokyo](https://github.com/hoshinonyaruko/gensokyo)（[GPLv3](LICENSE)），原作者 [SanaeFox](https://github.com/Hoshinonyaruko) 及上游所有贡献者
- 本项目在原项目基础上增加了群服互通相关能力，**原版权声明与 LICENSE 全文保留**，各修改文件头部已标注 `Modified by DanielToyama` 修改声明
- 若你使用本项目发现问题，请在本仓库反馈，**不要向上游提交与本项目相关的问题**

## 修改记录（GPLv3）

完整的逐笔提交记录、改动文件与实测结论已整理到 **[docs/DanielToyama更新记录.md](docs/DanielToyama更新记录.md)**（含 2026-08-15 起全部 45 笔提交）。

近期重点：

- **2026-11-10**：支持群聊 @全体成员（`[CQ:at,qq=all]` → `<qqbot-at-everyone />`）；移除已废弃的 `token` 配置；修复配置生成器 YAML 格式错误（注释换行 + 特殊字符转义）
- **2026-11-07**：撤回机器人自己发的消息不再静默失败 —— 出站消息入库 + `delete_msg` 反查扩展
- **2026-09-02**：表情表补全（273 条，id 0~431），入站表情分类对齐 OneBot
- **2026-08-31**：出站 `face` 段降级为 `[表情名]` 占位
- **2026-08-20**：`/g` qlogo 兼容头像路由；`delete_msg` 与群禁言修复
- **2026-08-19**：FakeReply、真实 At（Markdown）默认开启；入群 comment 标准格式
- **2026-08-18**：`get_status` 对齐 LLOneBot；维护通知配置化
- **2026-08-15 ~ 08-17**：fork 建立，群全量消息/主动消息/群管理事件与 API/username 回填

各修改文件头部均保留 `Modified by DanielToyama` 声明，遵循上游 GPLv3。

---

<details>
<summary>📖 附：上游 gensokyo 原版文档（API / CQ 码 / Event 参考）</summary>

<p align="center">
  <a href="https://www.github.com/hoshinonyaruko/gensokyo">
    <img src="images/head.gif" width="200" height="200" alt="gensokyo">
  </a>
</p>
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

Gensokyo文档(施工中):[起步](docs/起步-注册QQ开放平台&启动gensokyo.md)

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

> 拓展 API 可前往 [适配 API 清单](docs/适配API清单.md) 查看

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
| /delete_msg√             | [撤回信息]（支持撤回机器人自己发的消息） |
| /set_group_kick          | [群组踢人]             |
| /set_group_ban√          | [群组单人禁言]         |
| /set_group_whole_ban√    | [群组全员禁言]         |
| /set_group_admin         | [群组设置管理员]       |
| /set_group_card          | [设置群名片（群备注）] |
| /set_group_name          | [设置群名]             |
| /set_group_leave         | [退出群组]             |
| /set_group_special_title | [设置群组专属头衔]     |
| /set_friend_add_request  | [处理加好友请求]       |
| /set_group_add_request√  | [处理加群请求/邀请]（支持只传 flag） |
| /get_login_info√         | [获取登录号信息]       |
| /get_stranger_info√      | [获取陌生人信息]（昵称缓存 + 等级放行） |
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

#### 本 Fork 新增的群管理拓展 API（官方 API v2）

| 拓展 API | 功能 |
| -------- | ---- |
| /get_group_join_request_list | [拉取入群申请列表] |
| /get_group_restrict_chat_setting | [查询群禁言状态] |
| /set_group_restrict_chat_setting | [设置群禁言（官方原始透传）] |
| /get_group_bot_state | [获取机器人群内状态] |
| /create_join_approval_strategy | [创建入群自动审批策略] |
| /get_join_approval_strategy_list | [查询入群自动审批策略列表] |
| /update_join_approval_strategy | [修改入群自动审批策略] |
| /delete_join_approval_strategy | [删除入群自动审批策略] |
| /update_join_approval_strategy_whitelist | [修改策略白名单号码] |
| /execute_join_approval_strategy | [执行策略（全量扫描，异步约 10 分钟）] |

> 兼容别名：`/get_group_ban` = `/set_group_ban`、`/get_group_whole_ban` = `/set_group_whole_ban`（上游早期误注册名，保留兼容）。
>
> 以上 API 的参数、响应与注意事项详见 **[适配 API 清单](docs/适配API清单.md)**。


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

</div>
</details>
