# 企业微信消息格式支持说明

## 问题说明

**错误码 44004**：文本内容为空（text content is empty）

这个错误是因为代码只支持简化格式的消息，而企业微信官方使用的是嵌套格式。

## 修复内容

现在代码同时支持两种消息格式：

### 格式1：简化格式（向后兼容）

```json
{
  "msg_type": "text",
  "msg": "这是一条测试消息"
}
```

### 格式2：企业微信官方格式（新增支持）

#### 文本消息
```json
{
  "msg_type": "text",
  "text": {
    "content": "你的快递已到，请携带工卡前往邮件中心领取。\n出发前可查看<a href=\"https://work.weixin.qq.com\">邮件中心视频实况</a>，聪明避开排队。"
  }
}
```

#### Markdown消息
```json
{
  "msg_type": "markdown",
  "markdown": {
    "content": "# 快递通知\n你的快递已到，请前往邮件中心领取。\n\n**时间**：今天 14:00\n**地点**：邮件中心\n\n[查看视频](https://work.weixin.qq.com)"
  }
}
```

## 使用方法

### 方法1：使用 curl

```bash
# 官方格式 - 文本消息
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "msg_type": "text",
    "text": {
      "content": "你的快递已到，请携带工卡前往邮件中心领取。\n出发前可查看<a href=\"https://work.weixin.qq.com\">邮件中心视频实况</a>，聪明避开排队。"
    }
  }'

# 官方格式 - Markdown消息
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "msg_type": "markdown",
    "markdown": {
      "content": "# 快递通知\n你的快递已到，请前往邮件中心领取。"
    }
  }'

# 简化格式（仍然支持）
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "msg_type": "text",
    "msg": "这是一条测试消息"
  }'
```

### 方法2：使用 Docker

拉取最新镜像（重新构建后）：

```bash
docker pull gowfqk/go-wecomchan:latest
```

## 代码修改说明

### 修改1：扩展 RequestBody 结构体

```go
type RequestBody struct {
    Sendkey  string          `json:"sendkey"`
    Msg      string          `json:"msg"`               // 简化格式：文本/Markdown内容
    MsgType  string          `json:"msg_type"`
    ToUser   string          `json:"touser,omitempty"`
    AgentId  string          `json:"agentid,omitempty"`
    Text     *Msg            `json:"text,omitempty"`    // 官方格式：文本消息
    Markdown *Markdown       `json:"markdown,omitempty"` // 官方格式：Markdown消息
}
```

### 修改2：智能解析消息内容

```go
// 优先使用简化格式（向后兼容）
if requestBody.Msg != "" {
    msgContent = requestBody.Msg
    log.Println("使用body传参（简化格式）")
} else {
    // 使用官方格式（text.content 或 markdown.content）
    if requestBody.Text != nil && requestBody.Text.Content != "" {
        msgContent = requestBody.Text.Content
        log.Println("使用body传参（官方格式 - text）")
    } else if requestBody.Markdown != nil && requestBody.Markdown.Content != "" {
        msgContent = requestBody.Markdown.Content
        log.Println("使用body传参（官方格式 - markdown）")
    }
}
```

## 日志输出

当发送消息时，日志会显示使用的格式：

```
使用body传参（简化格式）
使用body传参（官方格式 - text）
使用body传参（官方格式 - markdown）
警告：未找到消息内容
```

## 版本信息

- 提交ID: e1229c2
- 提交信息: feat: 支持企业微信官方消息格式（text.content和markdown.content）
- 修改文件: wecomchan.go

## 兼容性

✅ 完全向后兼容，所有使用简化格式的现有代码无需修改即可继续工作。
✅ 新增支持企业微信官方格式，可以直接复制官方示例代码使用。
