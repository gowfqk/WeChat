# 本地部署测试指南

## 环境要求

- Docker已安装
- 或Go 1.16.5+已安装

## 方式1：使用Docker部署（推荐）

### 步骤1：拉取最新镜像

```bash
docker pull gowfqk/go-wecomchan:latest
```

### 步骤2：启动容器

```bash
docker run -d -p 8080:8080 \
  --name wecomchan \
  -e SENDKEY=set_a_sendkey \
  -e WECOM_CID=你的企业微信公司ID \
  -e WECOM_SECRET=你的企业微信应用Secret \
  -e WECOM_AID=你的企业微信应用ID \
  -e CACHE_TYPE=memory \
  gowfqk/go-wecomchan:latest
```

### 步骤3：查看日志

```bash
docker logs -f wecomchan
```

### 步骤4：测试API

#### 测试1：简化格式（带sendkey）

```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "set_a_sendkey",
    "msg_type": "text",
    "msg": "测试消息 - 简化格式"
  }'
```

#### 测试2：官方格式（带sendkey）

```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "set_a_sendkey",
    "msg_type": "text",
    "text": {
      "content": "测试消息 - 官方格式"
    }
  }'
```

#### 测试3：官方示例

```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "set_a_sendkey",
    "msg_type": "text",
    "text": {
      "content": "你的快递已到，请携带工卡前往邮件中心领取。\n出发前可查看<a href=\"https://work.weixin.qq.com\">邮件中心视频实况</a>，聪明避开排队。"
    }
  }'
```

### 步骤5：分析日志

在`docker logs -f wecomchan`中查看详细日志：

```
========== 收到新请求 ==========
请求方法: POST
请求URL: /wecomchan
Content-Type: application/json
请求体内容: {...}
解析结果 - sendkey: 'xxx', msgType: 'text'
使用body传参（简化格式）
最终参数 - sendkey: 'xxx', msgType: 'text', msgContent: 'xxx'
准备发送的数据: {...}
企业微信发送应用消息接口返回==>{errcode:0, errmsg:ok}
```

## 方式2：使用Go源码部署

### 步骤1：克隆或更新代码

```bash
cd /path/to/WeChat
git pull origin main
```

### 步骤2：构建二进制文件

```bash
go build -o wecomchan
```

### 步骤3：设置环境变量

```bash
export SENDKEY=set_a_sendkey
export WECOM_CID=你的企业微信公司ID
export WECOM_SECRET=你的企业微信应用Secret
export WECOM_AID=你的企业微信应用ID
export CACHE_TYPE=memory
```

### 步骤4：运行程序

```bash
./wecomchan
```

### 步骤5：测试API

同方式1的测试步骤。

## 常见问题排查

### 问题1：容器无法启动

**检查命令**：
```bash
docker logs wecomchan
docker ps -a | grep wecomchan
```

**可能原因**：
- 端口8080已被占用
- 环境变量未设置

**解决方案**：
```bash
# 查看端口占用
netstat -tuln | grep 8080

# 使用其他端口
docker run -d -p 9080:8080 \
  --name wecomchan \
  -e SENDKEY=set_a_sendkey \
  ...
```

### 问题2：错误码40001（invalid sendkey）

**原因**：sendkey验证失败

**解决方案**：
- 确保请求中包含正确的sendkey
- 确保环境变量SENDKEY设置正确
- 检查日志中的sendkey验证信息

```bash
# 检查环境变量
docker inspect wecomchan | grep SENDKEY

# 使用正确的sendkey
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "set_a_sendkey",
    "msg_type": "text",
    "msg": "测试消息"
  }'
```

### 问题3：错误码44004（text content is empty）

**原因**：消息内容为空

**检查日志**：
```bash
docker logs wecomchan | grep "msgContent"
```

**解决方案**：
- 确保请求中包含消息内容
- 检查JSON格式是否正确
- 查看详细调试日志

```bash
# 正确格式
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "set_a_sendkey",
    "msg_type": "text",
    "msg": "测试消息"
  }'
```

### 问题4：企业微信API错误

**检查日志**：
```bash
docker logs wecomchan | grep "企业微信发送应用消息接口返回"
```

**可能原因**：
- 企业微信配置错误（WECOM_CID, WECOM_SECRET, WECOM_AID）
- access_token过期
- 网络问题

**解决方案**：
- 检查企业微信配置是否正确
- 确认网络可以访问企业微信API
- 查看完整的错误信息

## 性能测试

使用ab进行压力测试：

```bash
# 安装ab
# Ubuntu/Debian: sudo apt-get install apache2-utils
# CentOS/RHEL: sudo yum install httpd-tools

# 测试
ab -n 100 -c 10 -p test.json -T application/json \
  http://localhost:8080/wecomchan
```

test.json内容：
```json
{
  "sendkey": "set_a_sendkey",
  "msg_type": "text",
  "msg": "性能测试消息"
}
```

## 停止服务

```bash
# 停止Docker容器
docker stop wecomchan

# 删除容器
docker rm wecomchan

# 停止Go程序
# 按 Ctrl+C
```

## 重启服务

```bash
# 重启Docker容器
docker restart wecomchan

# 查看日志
docker logs -f wecomchan
```

## 升级镜像

```bash
# 拉取最新镜像
docker pull gowfqk/go-wecomchan:latest

# 停止旧容器
docker stop wecomchan
docker rm wecomchan

# 启动新容器
docker run -d -p 8080:8080 \
  --name wecomchan \
  -e SENDKEY=set_a_sendkey \
  -e WECOM_CID=你的企业微信公司ID \
  -e WECOM_SECRET=你的企业微信应用Secret \
  -e WECOM_AID=你的企业微信应用ID \
  -e CACHE_TYPE=memory \
  gowfqk/go-wecomchan:latest
```

## 使用docker-compose（推荐）

创建docker-compose.yml：

```yaml
version: '3.8'
services:
  wecomchan:
    image: gowfqk/go-wecomchan:latest
    container_name: wecomchan
    ports:
      - "8080:8080"
    environment:
      - SENDKEY=set_a_sendkey
      - WECOM_CID=你的企业微信公司ID
      - WECOM_SECRET=你的企业微信应用Secret
      - WECOM_AID=你的企业微信应用ID
      - CACHE_TYPE=memory
    restart: unless-stopped
```

启动：
```bash
docker-compose up -d
```

查看日志：
```bash
docker-compose logs -f
```

停止：
```bash
docker-compose down
```
