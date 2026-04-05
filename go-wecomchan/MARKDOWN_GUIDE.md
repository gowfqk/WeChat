# Markdown消息使用指南

## 概述

go-wecomchan v2.2.0 新增了对企业微信markdown消息的支持，允许发送格式更丰富的消息内容。

## 为什么使用Markdown？

相比于纯文本消息，Markdown消息具有以下优势：

- ✅ **更好的可读性**：通过标题、列表、引用等结构化内容
- ✅ **突出重点**：通过加粗、颜色等方式强调关键信息
- ✅ **专业外观**：格式化的消息看起来更专业、更易阅读
- ✅ **信息丰富**：可以在消息中包含链接、代码块等

## Markdown语法支持

企业微信的markdown支持以下语法：

### 1. 标题

使用 1-6 个 `#` 号表示不同级别的标题：

```markdown
# 一级标题
## 二级标题
### 三级标题
```

**示例：**

```json
{
  "sendkey": "your_sendkey",
  "msg": "# 系统通知\n\n## 维护提醒\n\n服务器将在今晚进行维护",
  "msg_type": "markdown"
}
```

### 2. 加粗

使用 `**` 包裹文字使其加粗：

```markdown
**加粗文字**
```

**示例：**

```json
{
  "sendkey": "your_sendkey",
  "msg": "**重要提醒**：请务必在今晚前完成工作",
  "msg_type": "markdown"
}
```

### 3. 字体颜色

使用 `<font color="颜色值">` 标签设置文字颜色：

**可用颜色值：**

- `info` - 灰色（信息）
- `comment` - 灰色（注释）
- `warning` - 橙色（警告）

```markdown
<font color="info">这是灰色文字</font>
<font color="warning">这是橙色警告</font>
```

**示例：**

```json
{
  "sendkey": "your_sendkey",
  "msg": "系统将于<font color=\"warning\">今晚22:00</font>进行维护，请提前保存工作",
  "msg_type": "markdown"
}
```

### 4. 链接

使用标准markdown链接语法：

```markdown
[链接文字](URL)
```

**示例：**

```json
{
  "sendkey": "your_sendkey",
  "msg": "详情请查看：[点击这里](https://example.com/notice)",
  "msg_type": "markdown"
}
```

### 5. 引用

使用 `>` 开头表示引用：

```markdown
> 这是一段引用内容
```

**示例：**

```json
{
  "sendkey": "your_sendkey",
  "msg": "> 温馨提示：请及时更新密码",
  "msg_type": "markdown"
}
```

### 6. 代码

使用反引号 ``` ` ``` 包裹代码：

```markdown
`代码片段`
```

**示例：**

```json
{
  "sendkey": "your_sendkey",
  "msg": "错误码：`500 Internal Server Error`",
  "msg_type": "markdown"
}
```

## 实际应用示例

### 示例1：系统维护通知

```json
{
  "sendkey": "your_sendkey",
  "msg": "# 系统维护通知\n\n**维护时间**：2024-01-15 22:00 - 23:30\n\n**影响范围**：所有服务\n\n**维护内容**：<font color=\"warning\">服务器升级和数据库优化</font>\n\n请提前做好数据备份，避免在维护期间进行重要操作。\n\n如有疑问，请联系 [IT支持](https://help.example.com)",
  "msg_type": "markdown"
}
```

### 示例2：任务提醒

```json
{
  "sendkey": "your_sendkey",
  "msg": "# 任务提醒\n\n## 待办事项\n\n- [ ] 完成代码审查\n- [ ] 更新文档\n- [ ] 部署到测试环境\n\n**截止时间**：<font color=\"warning\">今日18:00</font>\n\n> 请务必按时完成，以免影响项目进度",
  "msg_type": "markdown"
}
```

### 示例3：错误报告

```json
{
  "sendkey": "your_sendkey",
  "msg": "# 错误报告\n\n**服务名称**：用户服务\n\n**错误类型**：`ConnectionTimeout`\n\n**发生时间**：2024-01-15 10:30:45\n\n**错误信息**：<font color=\"warning\">无法连接到数据库服务器</font>\n\n## 建议操作\n\n1. 检查数据库服务状态\n2. 查看日志文件\n3. 联系运维团队\n\n详情：[查看日志](https://logs.example.com/error/12345)",
  "msg_type": "markdown"
}
```

### 示例4：欢迎消息

```json
{
  "sendkey": "your_sendkey",
  "msg": "# 欢迎使用企业微信通知服务\n\n欢迎使用我们的通知系统！\n\n## 功能介绍\n\n✨ **文本消息**：发送纯文本通知\n📝 **Markdown消息**：发送格式化消息\n🖼️ **图片消息**：发送图片通知\n\n## 快速开始\n\n发送第一条消息：\n```\ncurl -X POST http://your-server/wecomchan -H \"Content-Type: application/json\" -d '{\"sendkey\":\"your_key\",\"msg\":\"Hello!\",\"msg_type\":\"text\"}'\n```\n\n如有问题，请查看：[使用文档](https://docs.example.com)",
  "msg_type": "markdown"
}
```

## 注意事项

1. **转义字符**：在JSON中，特殊字符需要转义：
   - `\n` → 换行
   - `\"` → 双引号
   - `\\` → 反斜杠

2. **内容长度限制**：企业微信对消息内容有长度限制，建议控制在2048字符以内

3. **兼容性**：确保接收方支持markdown显示，否则会以纯文本形式展示

4. **颜色标签**：只支持 `info`、`comment`、`warning` 三种颜色值

## 发送方式

### 通过curl发送

```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "your_sendkey",
    "msg": "# 通知\n\n**内容**：这是一条测试消息",
    "msg_type": "markdown"
  }'
```

### 通过URL参数发送

```bash
curl -X POST "http://localhost:8080/wecomchan?sendkey=your_sendkey&msg=# 标题\n\n**内容**：测试消息&msg_type=markdown"
```

## 对比：文本 vs Markdown

### 纯文本消息

```json
{
  "sendkey": "your_sendkey",
  "msg": "系统维护通知\n维护时间：今晚22:00\n请提前保存工作",
  "msg_type": "text"
}
```

**效果：** 纯文本，无格式

### Markdown消息

```json
{
  "sendkey": "your_sendkey",
  "msg": "# 系统维护通知\n\n**维护时间**：<font color=\"warning\">今晚22:00</font>\n\n请提前保存工作",
  "msg_type": "markdown"
}
```

**效果：** 带标题、加粗、颜色的格式化消息

## 最佳实践

1. **结构清晰**：使用标题和列表组织内容
2. **突出重点**：使用加粗和颜色强调关键信息
3. **简洁明了**：避免过长的消息，控制在合理长度
4. **测试显示**：发送前测试markdown的显示效果
5. **提供链接**：重要信息提供详细文档的链接

## 常见问题

**Q: Markdown消息在所有客户端都支持吗？**
A: 企业微信官方客户端支持，第三方客户端可能不支持，会以纯文本显示。

**Q: 可以使用HTML标签吗？**
A: 除了 `<font>` 标签，其他HTML标签不支持。

**Q: 支持代码块语法高亮吗？**
A: 不支持，只能使用单个反引号包裹代码片段。

**Q: 如何在消息中显示链接而不是URL？**
A: 使用 `[链接文字](URL)` 语法。

## 更新日志

- v2.2.0 - 新增markdown消息类型支持
  - 支持标题、加粗、颜色、链接等语法
  - 提供丰富的格式化消息能力
