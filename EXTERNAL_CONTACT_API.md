# 外部联系人消息接口

## 概述

go-wecomchan现在支持企业微信外部联系人消息接口，可以向企业外部的客户发送消息。

## 新增功能

### 1. 新增API端点

```
POST /external
```

### 2. 支持的消息类型

- ✅ 文本消息 (text)
- ✅ 图片消息 (image)
- ✅ Markdown消息 (markdown)
- ✅ 链接消息 (link)
- ✅ 小程序消息 (miniprogram)

### 3. 请求格式

#### 通用参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sendkey | string | ✅ | 验证密钥，需与环境变量SENDKEY一致 |
| external_userid | []string | ✅ | 外部联系人userid列表，最多1000个 |
| sender | string | ✅ | 发送企业成员的userid |
| msgtype | string | ✅ | 消息类型：text/image/markdown/link/miniprogram |

### 4. 消息格式示例

#### 4.1 文本消息

```json
{
  "sendkey": "204800",
  "external_userid": ["wxid_xxx", "wxid_yyy"],
  "sender": "zhangsan",
  "msgtype": "text",
  "text": {
    "content": "您好，这是一条测试消息"
  }
}
```

**Python示例**：
```python
import requests
import json

url = "http://localhost:8080/external"
headers = {"Content-Type": "application/json"}
data = {
    "sendkey": "204800",
    "external_userid": ["wxid_xxx"],
    "sender": "zhangsan",
    "msgtype": "text",
    "text": {
        "content": "您好，这是一条测试消息"
    }
}

response = requests.post(url, headers=headers, json=data)
print(response.json())
```

#### 4.2 图片消息

```json
{
  "sendkey": "204800",
  "external_userid": ["wxid_xxx"],
  "sender": "zhangsan",
  "msgtype": "image",
  "image": {
    "media_id": "MEDIA_ID"
  }
}
```

#### 4.3 Markdown消息

```json
{
  "sendkey": "204800",
  "external_userid": ["wxid_xxx"],
  "sender": "zhangsan",
  "msgtype": "markdown",
  "markdown": {
    "content": "# 📢 通知\n\n这是一条Markdown格式的消息。\n\n**重要提醒**：请及时查看"
  }
}
```

#### 4.4 链接消息

```json
{
  "sendkey": "204800",
  "external_userid": ["wxid_xxx"],
  "sender": "zhangsan",
  "msgtype": "link",
  "link": {
    "title": "标题",
    "desc": "描述信息",
    "url": "https://example.com",
    "thumb_media_id": "THUMB_MEDIA_ID"
  }
}
```

#### 4.5 小程序消息

```json
{
  "sendkey": "204800",
  "external_userid": ["wxid_xxx"],
  "sender": "zhangsan",
  "msgtype": "miniprogram",
  "miniprogram": {
    "title": "标题",
    "appid": "wx1234567890abcdef",
    "pagepath": "pages/index/index",
    "thumb_media_id": "THUMB_MEDIA_ID"
  }
}
```

## 使用方法

### 1. 启动服务

```bash
# 使用Go
export SENDKEY=204800
export WECOM_CID=你的企业微信公司ID
export WECOM_SECRET=你的企业微信应用Secret
export WECOM_AID=你的企业微信应用ID
export CACHE_TYPE=memory
./wecomchan

# 使用Docker
docker run -d -p 8080:8080 \
  -e SENDKEY=204800 \
  -e WECOM_CID=你的企业微信公司ID \
  -e WECOM_SECRET=你的企业微信应用Secret \
  -e WECOM_AID=你的企业微信应用ID \
  -e CACHE_TYPE=memory \
  gowfqk/go-wecomchan:latest
```

### 2. 发送外部联系人消息

```bash
curl -X POST http://localhost:8080/external \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "204800",
    "external_userid": ["wxid_xxx"],
    "sender": "zhangsan",
    "msgtype": "text",
    "text": {
      "content": "您好，这是一条测试消息"
    }
  }'
```

## 错误码说明

| 错误码 | 说明 | 解决方案 |
|--------|------|---------|
| 40001 | sendkey错误 | 检查sendkey是否与环境变量SENDKEY一致 |
| 40002 | JSON格式错误 | 检查请求体格式是否正确 |
| 40003 | external_userid为空 | 提供至少一个外部联系人userid |
| 40004 | sender为空 | 提供发送企业成员的userid |

## 对比：应用消息 vs 外部联系人消息

| 特性 | 应用消息 (/wecomchan) | 外部联系人消息 (/external) |
|------|---------------------|-------------------------|
| 接收对象 | 企业内部成员 | 企业外部客户 |
| 接收者字段 | touser | external_userid |
| 发送者字段 | 无（应用本身） | sender（企业成员） |
| 支持的消息类型 | text, image, markdown | text, image, markdown, link, miniprogram |
| 企业微信API | /cgi-bin/message/send | /cgi-bin/externalcontact/message/send |

## 注意事项

1. **权限要求**：需要企业微信管理员配置"联系我"或"客户联系"权限
2. **external_userid获取**：需要通过企业微信的"配置客户联系"或"外部联系人管理"API获取
3. **sender限制**：sender必须是企业内部成员，且具有发送外部联系人消息的权限
4. **数量限制**：每次最多发送给1000个外部联系人
5. **频率限制**：需要遵守企业微信的API调用频率限制

## 完整示例

### 发送文本消息

```bash
curl -X POST http://localhost:8080/external \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "204800",
    "external_userid": ["wmxxxxxxxxxxxxxxxx", "wmyyyyyyyyyyyyyyyyyy"],
    "sender": "zhangsan",
    "msgtype": "text",
    "text": {
      "content": "您好，感谢您关注我们！\n\n如有任何问题，请随时联系我们。"
    }
  }'
```

### 发送Markdown消息

```bash
curl -X POST http://localhost:8080/external \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "204800",
    "external_userid": ["wmxxxxxxxxxxxxxxxx"],
    "sender": "zhangsan",
    "msgtype": "markdown",
    "markdown": {
      "content": "# 📢 重要通知\n\n尊敬的客户：\n\n我们很高兴地通知您，新功能已上线！\n\n## 新功能亮点\n\n- ✅ 更加简洁的界面\n- ✅ 更快的响应速度\n- ✅ 更好的用户体验\n\n## 如何使用\n\n1. 登录您的账户\n2. 进入\"设置\"页面\n3. 启用新功能\n\n---\n\n如有任何问题，请联系我们的客服团队。"
    }
  }'
```

## 相关文档

- [完整使用说明](README.md)
- [消息格式支持说明](MESSAGE_FORMAT.md)
- [Markdown消息指南](MARKDOWN_GUIDE.md)
