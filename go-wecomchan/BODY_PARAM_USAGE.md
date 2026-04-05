# Body传参使用说明

## 更新内容

已将go-wecomchan升级，支持通过请求体（body）传递JSON格式参数，同时保持向后兼容原有的URL参数传参方式。

## 新增的RequestBody结构

```go
type RequestBody struct {
    Sendkey  string `json:"sendkey"`
    Msg      string `json:"msg"`
    MsgType  string `json:"msg_type"`
    ToUser   string `json:"touser,omitempty"`  // 可选：覆盖默认的接收人
    AgentId  string `json:"agentid,omitempty"` // 可选：覆盖默认的应用ID
}
```

## 使用方式

### 方式1：通过Body传参（推荐）

**请求示例：**

```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "your_sendkey",
    "msg": "这是一条测试消息",
    "msg_type": "text",
    "touser": "@all",
    "agentid": "1000001"
  }'
```

**参数说明：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| sendkey | string | 是 | 验证密钥，与环境变量SENDKEY保持一致 |
| msg | string | 是 | 消息内容 |
| msg_type | string | 否 | 消息类型，默认为"text"，可选值：text、markdown、image |
| touser | string | 否 | 接收人，默认为环境变量WECOM_TOUID的值 |
| agentid | string | 否 | 应用ID，默认为环境变量WECOM_AID的值 |

### 方式2：通过URL参数传参（向后兼容）

**请求示例：**

```bash
curl -X POST "http://localhost:8080/wecomchan?sendkey=your_sendkey&msg=这是一条测试消息&msg_type=text&touser=@all&agentid=1000001"
```

### 方式3：混合使用（不推荐）

如果body和URL参数同时存在，优先使用body中的参数值。

## 消息类型

### 文本消息（text）

```json
{
  "sendkey": "your_sendkey",
  "msg": "这是一条文本消息",
  "msg_type": "text"
}
```

### Markdown消息（markdown）

支持企业微信的markdown格式消息，可以使用以下语法：

- **标题**：使用 `#` 标记
- **加粗**：使用 `**` 包裹
- **字体颜色**：使用 `<font color="info">文字</font>`
- **链接**：使用 `[文字](URL)`
- **引用**：使用 `>` 标记
- **代码**：使用 `` ` `` 包裹

**请求示例：**

```json
{
  "sendkey": "your_sendkey",
  "msg": "# 任务提醒\n\n**任务名称**：系统维护\n\n**时间**：2024-01-01 02:00-04:00\n\n**内容**：<font color=\"warning\">服务器将进行例行维护，请提前保存工作</font>\n\n详情请查看：[维护公告](https://example.com/notice)",
  "msg_type": "markdown"
}
```

**curl示例：**

```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "your_sendkey",
    "msg": "# 消息标题\n\n这是**加粗**的文字\n\n这是<font color=\"info\">蓝色文字</font>",
    "msg_type": "markdown"
  }'
```

**支持的markdown标签：**

| 标签 | 说明 | 示例 |
|------|------|------|
| 标题 | 使用1-6个# | `# 标题一` |
| 加粗 | 使用**包裹 | `**加粗文字**` |
| 引用 | 使用>开头 | `> 引用内容` |
| 字体颜色 | 使用font标签 | `<font color=\"info\">文字</font>` |
| 链接 | 使用标准markdown链接 | `[文字](URL)` |
| 代码 | 使用反引号 | `` `代码` `` |

**可用颜色值：**

- `info` - 灰色
- `comment` - 灰色
- `warning` - 橙色

### 图片消息（image）

发送图片消息需要使用multipart/form-data格式：

```bash
curl -X POST http://localhost:8080/wecomchan \
  -F "sendkey=your_sendkey" \
  -F "msg_type=image" \
  -F "media=@/path/to/image.jpg"
```

**注意：** 图片上传功能仍需使用表单方式，不支持JSON body方式。

## 优先级规则

1. **body传参** > **URL参数** > **环境变量默认值**

如果body中的某个字段为空字符串或不存在，则会尝试从URL参数中获取；如果URL参数也没有，则使用环境变量中的默认值。

## 兼容性说明

- ✅ 完全向后兼容原有的URL参数传参方式
- ✅ 新增的body传参方式不影响现有代码
- ✅ 支持灵活的参数组合方式

## 错误处理

如果JSON格式不正确，系统会自动回退到URL参数方式，并输出日志：

```
JSON解析失败: xxx，回退到URL参数
```

## 测试示例

### 测试body传参（文本消息）

```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{"sendkey":"test_key","msg":"测试body传参","msg_type":"text"}'
```

### 测试URL参数传参（文本消息）

```bash
curl -X POST "http://localhost:8080/wecomchan?sendkey=test_key&msg=测试URL参数&msg_type=text"
```

### 测试Markdown消息

```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "test_key",
    "msg": "# 通知\n\n**任务**: 系统升级\n\n**时间**: 今晚22:00\n\n<font color=\"warning\">请提前做好准备</font>",
    "msg_type": "markdown"
  }'
```

### 测试自定义接收人

```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "test_key",
    "msg": "发送给指定用户",
    "msg_type": "text",
    "touser": "zhangsan"
  }'
```

## 编译和运行

```bash
# 编译
go build

# 运行
./wecomchan

# 或直接运行
go run .
```

## Docker部署

```bash
# 构建镜像
docker build -t go-wecomchan .

# 运行容器
docker run -d -p 8080:8080 \
  -e SENDKEY=your_sendkey \
  -e WECOM_CID=your_cid \
  -e WECOM_SECRET=your_secret \
  -e WECOM_AID=your_aid \
  go-wecomchan
```

## 更新日志

- v2.2.0 - 新增markdown消息类型支持
  - 支持企业微信markdown格式消息
  - 支持标题、加粗、颜色、链接等markdown语法
- v2.1.0 - 新增body传参支持，保持向后兼容
  - 支持通过JSON格式传递参数
  - 支持自定义接收人和应用ID
