# 修复错误码 40058（miniprogram.appid exceed min length）

## 问题描述

在使用外部联系人消息接口时，即使发送的是 `text` 或 `markdown` 类型的消息，也会收到以下错误：

```
errcode:40058
errmsg:miniprogram.appid exceed min length 1. invalid Request Parameter
```

## 根本原因

企业微信 API 会检查请求中的所有字段，包括未使用的消息类型字段。在之前的实现中，即使只发送 `text` 类型的消息，请求体也会包含所有消息类型的字段（text、image、markdown、link、miniprogram），只是大部分字段为空。

当企业微信 API 检测到 `miniprogram` 字段存在但 `appid` 为空时，会认为这是一个无效的小程序消息请求，从而返回 40058 错误。

### 之前的错误示例

发送文本消息时，实际发送的请求体：

```json
{
  "external_userid": ["wmbaMZEAAAM3CVZx_rFIXyO39ZLIpY0w"],
  "sender": "GuoWenQing",
  "msgtype": "text",
  "text": {
    "content": "测试消息"
  },
  "image": {
    "media_id": ""
  },
  "markdown": {
    "content": ""
  },
  "link": {
    "title": "",
    "desc": "",
    "url": "",
    "thumb_media_id": ""
  },
  "miniprogram": {
    "title": "",
    "appid": "",
    "pagepath": "",
    "thumb_media_id": ""
  }
}
```

即使 `msgtype` 是 `text`，企业微信 API 仍然会检查 `miniprogram` 字段，发现 `appid` 为空就报错。

## 解决方案

修改 `externalContactHandler` 处理器，**根据消息类型动态构建请求数据，只包含当前消息类型相关的字段**。

### 修复后的实现

发送文本消息时，实际发送的请求体：

```json
{
  "external_userid": ["wmbaMZEAAAM3CVZx_rFIXyO39ZLIpY0w"],
  "sender": "GuoWenQing",
  "msgtype": "text",
  "text": {
    "content": "测试消息"
  }
}
```

**不再包含其他消息类型的字段！**

## 代码修改

### 1. 修改 SendExternalMessage 函数签名

将参数类型从 `ExternalMessageData` 改为 `interface{}`，以支持动态数据类型：

```go
// 修改前
func SendExternalMessage(accessToken string, postData ExternalMessageData) string

// 修改后
func SendExternalMessage(accessToken string, postData interface{}) string
```

### 2. 修改 externalContactHandler 处理器

根据消息类型动态构建请求数据：

```go
// 根据消息类型构建请求数据 - 只包含当前消息类型相关的字段
var postData interface{}

switch requestBody.MsgType {
case "text":
    if requestBody.Text != nil {
        postData = map[string]interface{}{
            "external_userid": requestBody.ExternalUserIds,
            "sender":          requestBody.Sender,
            "msgtype":         "text",
            "text": map[string]interface{}{
                "content": requestBody.Text.Content,
            },
        }
    }
case "markdown":
    if requestBody.Markdown != nil {
        postData = map[string]interface{}{
            "external_userid": requestBody.ExternalUserIds,
            "sender":          requestBody.Sender,
            "msgtype":         "markdown",
            "markdown": map[string]interface{}{
                "content": requestBody.Markdown.Content,
            },
        }
    }
case "link":
    if requestBody.Link != nil {
        postData = map[string]interface{}{
            "external_userid": requestBody.ExternalUserIds,
            "sender":          requestBody.Sender,
            "msgtype":         "link",
            "link": map[string]interface{}{
                "title":          requestBody.Link.Title,
                "desc":           requestBody.Link.Description,
                "url":            requestBody.Link.Url,
                "thumb_media_id": requestBody.Link.ThumbMediaId,
            },
        }
    }
case "miniprogram":
    if requestBody.MiniProgram != nil {
        postData = map[string]interface{}{
            "external_userid": requestBody.ExternalUserIds,
            "sender":          requestBody.Sender,
            "msgtype":         "miniprogram",
            "miniprogram": map[string]interface{}{
                "title":           requestBody.MiniProgram.Title,
                "appid":           requestBody.MiniProgram.AppId,
                "pagepath":        requestBody.MiniProgram.PagePath,
                "thumb_media_id":  requestBody.MiniProgram.ThumbMediaId,
            },
        }
    }
}
```

## 测试

### 测试1：发送文本消息

```bash
curl -X POST http://localhost:8080/external \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "204800",
    "external_userid": ["wmbaMZEAAAM3CVZx_rFIXyO39ZLIpY0w"],
    "sender": "GuoWenQing",
    "msgtype": "text",
    "text": {
      "content": "测试消息 - 验证修复"
    }
  }'
```

**预期结果**：成功发送，返回 `{"errcode":0, "errmsg":"ok"}`

### 测试2：发送 Markdown 消息

```bash
curl -X POST http://localhost:8080/external \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "204800",
    "external_userid": ["wmbaMZEAAAM3CVZx_rFIXyO39ZLIpY0w"],
    "sender": "GuoWenQing",
    "msgtype": "markdown",
    "markdown": {
      "content": "**测试Markdown消息**"
    }
  }'
```

**预期结果**：成功发送，返回 `{"errcode":0, "errmsg":"ok"}`

### 测试3：发送小程序消息（包含必需字段）

```bash
curl -X POST http://localhost:8080/external \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "204800",
    "external_userid": ["wmbaMZEAAAM3CVZx_rFIXyO39ZLIpY0w"],
    "sender": "GuoWenQing",
    "msgtype": "miniprogram",
    "miniprogram": {
      "title": "小程序测试",
      "appid": "wx1234567890abcdef",
      "pagepath": "pages/index",
      "thumb_media_id": "MEDIA_ID"
    }
  }'
```

**预期结果**：成功发送，返回 `{"errcode":0, "errmsg":"ok"}`

### 批量测试

使用提供的测试脚本：

```bash
bash /workspace/projects/WeChat/test_external_msg_fix.sh
```

## 影响范围

- **修改的文件**：`wecomchan.go`
- **修改的函数**：
  - `SendExternalMessage`：修改函数签名
  - `externalContactHandler`：修改请求数据构建逻辑
- **向后兼容性**：完全兼容，不影响现有调用方式

## 注意事项

1. **小程序消息必需字段**：如果使用 `miniprogram` 类型，必须提供 `appid` 字段，否则会返回 40058 错误
2. **消息类型匹配**：`msgtype` 字段必须与实际的消息内容字段一致
3. **日志输出**：修复后，日志中的 `postJson` 只包含当前消息类型的相关字段，更加清晰

## 相关文档

- [企业微信外部联系人消息接口](https://developer.work.weixin.qq.com/document/path/92130)
- [错误码查询](https://open.work.weixin.qq.com/devtool/query)
