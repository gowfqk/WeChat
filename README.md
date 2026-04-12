# go-push

通过企业微信、钉钉、飞书、Telegram、Bark、Server酱、PushPlus 向用户推送消息的 Go 语言解决方案。

> 本项目基于 [wecomchan](https://github.com/easychen/wecomchan) 重构而来。

## 功能特性

**推送渠道**：企业微信 · 钉钉 · 飞书 · Telegram · Bark · Server酱 · PushPlus

- ✅ 企业微信应用消息（文本 / Markdown / 图片）
- ✅ 企业微信邮件（多人收发、抄送密送、附件）
- ✅ 钉钉机器人 Webhook（Markdown + 加签）
- ✅ 飞书机器人 Webhook（富文本卡片）
- ✅ Telegram Bot API（Markdown）
- ✅ Bark iOS 推送
- ✅ Server酱 / PushPlus
- ✅ 统一 `/push` 接口，`channel` 字段切换渠道
- ✅ 缓存策略：无缓存 / 内存 / Redis
- ✅ 健康检查：`/healthz`、`/readyz`
- ✅ Docker 多架构部署（amd64、arm64）

## 快速开始

### Docker Compose（推荐）

```yaml
services:
  go-push:
    image: gowfqk/go-push:latest
    container_name: go-push
    ports:
      - "8080:8080"
    environment:
      # 必填
      - SENDKEY=your_sendkey
      - WECOM_CID=your_corpid
      - WECOM_SECRET=your_secret
      - WECOM_AID=your_agentid
      # 按需配置其他渠道
      - WECOM_TOUID=@all
      - CACHE_TYPE=memory
      # - BARK_URL=https://api.day.app/your_key
      # - DINGTALK_TOKEN=your_token
      # - DINGTALK_SECRET=your_secret
      # - FEISHU_WEBHOOK=https://open.feishu.cn/open-apis/bot/v2/hook/xxx
      # - TELEGRAM_BOT_TOKEN=your_bot_token
      # - TELEGRAM_CHAT_ID=your_chat_id
      # - SERVERCHAN_KEY=your_key
      # - PUSHPLUS_TOKEN=your_token
    restart: unless-stopped
```

### 本地运行

```bash
cp .env.example .env
# 编辑 .env 填入配置
source .env
go run .
```

## API 接口

### 统一推送 `/push`

所有渠道走同一个接口，通过 `channel` 区分。不填 `channel` 默认走企业微信。

**公共参数**：

| 参数 | 必填 | 说明 |
|------|------|------|
| `sendkey` | 是 | 验证密钥 |
| `channel` | 否 | 推送渠道，见下表 |
| `title` | 否 | 标题（部分渠道显示为通知标题） |
| `content` | 是 | 消息正文 |

**channel 可选值**：

| channel | 渠道 | 额外参数 | 说明 |
|---------|------|----------|------|
| `wecom` | 企业微信（默认） | `touser` `agentid` `msg_type` | msg_type 支持 text / markdown |
| `bark` | Bark | `bark_url` | bark_url 可选，不填用环境变量 |
| `dingtalk` | 钉钉 | `ding_secret` | ding_secret 可选，用于加签 |
| `feishu` | 飞书 | - | 以卡片消息发送 |
| `telegram` | Telegram | `chat_id` | chat_id 可选，支持 Markdown |
| `serverchan` | Server酱 | - | title → 标题，content → 详情 |
| `pushplus` | PushPlus | - | 同上 |

**示例**：

```bash
# Bark
curl -X POST http://localhost:8080/push \
  -H 'Content-Type: application/json' \
  -d '{"sendkey":"xxx","channel":"bark","title":"部署通知","content":"v2.0 上线完成"}'

# 钉钉
curl -X POST http://localhost:8080/push \
  -H 'Content-Type: application/json' \
  -d '{"sendkey":"xxx","channel":"dingtalk","title":"告警","content":"CPU **90%**"}'

# Telegram
curl -X POST http://localhost:8080/push \
  -H 'Content-Type: application/json' \
  -d '{"sendkey":"xxx","channel":"telegram","title":"备份","content":"已完成","chat_id":"123456"}'
```

### 企业微信应用消息 `/wecomchan`

保留原有接口，兼容 wecomchan 项目（路径不变）。

```bash
# JSON Body
curl -X POST http://localhost:8080/wecomchan \
  -H 'Content-Type: application/json' \
  -d '{"sendkey":"xxx","msg_type":"text","text":{"content":"你好"}}'

# URL 参数
curl "http://localhost:8080/wecomchan?sendkey=xxx&msg=你好&msgtype=text"
```

支持 `text`、`markdown`、`image`（需 multipart 上传 media 文件）三种消息类型。

### 企业微信邮件 `/mail`

```json
{
  "sendkey": "your_sendkey",
  "to": {
    "emails": ["user@example.com"],
    "userids": ["william"]
  },
  "cc": {
    "emails": ["cc@example.com"]
  },
  "bcc": {
    "emails": ["secret@example.com"]
  },
  "subject": "周报 - 第 15 周",
  "content": "<h2>本周完成</h2><ul><li>接口优化</li></ul>",
  "content_type": "html",
  "attachment_list": [
    {
      "file_name": "report.pdf",
      "content": "BASE64_CONTENT"
    }
  ],
  "enable_id_trans": 1
}
```

| 参数 | 必填 | 说明 |
|------|------|------|
| `sendkey` | 是 | 验证密钥 |
| `to` | 是 | 收件人（`emails` 数组或 `userids` 数组，至少一个） |
| `cc` | 否 | 抄送 |
| `bcc` | 否 | 密送 |
| `subject` | 是 | 邮件主题 |
| `content` | 是 | 邮件正文 |
| `content_type` | 否 | `html`（默认）或 `text` |
| `attachment_list` | 否 | 附件，每项含 `file_name` 和 `content`（base64） |
| `enable_id_trans` | 否 | id 转译：`0`（默认）或 `1` |

### 健康检查

```bash
curl http://localhost:8080/healthz   # 存活检查
curl http://localhost:8080/readyz    # 就绪检查（含 access_token 验证）
```

## 环境变量

### 通用

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SENDKEY` | 请求验证密钥 | `set_a_sendkey` |
| `CACHE_TYPE` | 缓存：`none` / `memory` / `redis` | `none` |

### 企业微信

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `WECOM_CID` | 企业微信公司 ID | - |
| `WECOM_SECRET` | 应用 Secret | - |
| `WECOM_AID` | 应用 AgentId | - |
| `WECOM_TOUID` | 默认接收人 | `@all` |

### 钉钉

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DINGTALK_TOKEN` | 机器人 access_token | - |
| `DINGTALK_SECRET` | 加签密钥（可选） | - |

### 飞书

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `FEISHU_WEBHOOK` | 机器人 Webhook URL | - |

### Telegram

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `TELEGRAM_BOT_TOKEN` | Bot Token | - |
| `TELEGRAM_CHAT_ID` | 默认 Chat ID | - |

### Bark / Server酱 / PushPlus

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `BARK_URL` | Bark 服务地址，如 `https://api.day.app/key` | - |
| `SERVERCHAN_KEY` | Server酱 Key | - |
| `PUSHPLUS_TOKEN` | PushPlus Token | - |

### Redis

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `REDIS_STAT` | 开关：`ON` / `OFF` | `OFF` |
| `REDIS_ADDR` | 地址 | `localhost:6379` |
| `REDIS_PASSWORD` | 密码 | - |

> 使用 Redis 缓存时，必须将 `REDIS_STAT` 设为 `ON`。

## 项目结构

```
.
├── main.go             # 入口，路由注册
├── handlers.go         # HTTP 处理器（/wecomchan / /mail / /push / /healthz）
├── channels.go         # 各推送渠道实现 + 统一分发
├── wecom_api.go        # 企业微信 API（token / 消息 / 邮件 / 素材上传）
├── config.go           # 环境变量与 API 地址
├── types.go            # 结构体定义
├── utils.go            # 工具函数
├── utils_test.go       # 单元测试
├── docker-compose.yml  # Docker Compose 配置
├── Dockerfile
├── .env.example        # 环境变量示例
├── go.mod / go.sum
└── .github/workflows/  # CI（Docker 镜像构建）
```

## 测试

```bash
go test -v           # 单元测试
go test -bench=.     # 基准测试
```

## 构建

```bash
go build                                          # 本地
docker build -t gowfqk/go-push:latest .      # Docker
```

## 许可证

MIT
