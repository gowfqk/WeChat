# 快速测试命令

## Docker部署（推荐）

### 1. 拉取最新镜像
```bash
docker pull gowfqk/go-wecomchan:latest
```

### 2. 启动容器
```bash
docker run -d -p 8080:8080 \
  --name wecomchan \
  -e SENDKEY=set_a_sendkey \
  -e WECOM_CID=wwxxxxxxxxx \
  -e WECOM_SECRET=xxxxxxxxxx \
  -e WECOM_AID=xxxxxxxxxx \
  -e CACHE_TYPE=memory \
  gowfqk/go-wecomchan:latest
```

### 3. 查看日志
```bash
docker logs -f wecomchan
```

### 4. 测试API

#### 测试1：简化格式
```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "set_a_sendkey",
    "msg_type": "text",
    "msg": "测试消息"
  }'
```

#### 测试2：官方格式
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

## Go源码部署

### 1. 拉取代码
```bash
cd /path/to/WeChat
git pull origin main
```

### 2. 构建
```bash
go build -o wecomchan
```

### 3. 运行
```bash
export SENDKEY=set_a_sendkey
export WECOM_CID=wwxxxxxxxxx
export WECOM_SECRET=xxxxxxxxxx
export WECOM_AID=xxxxxxxxxx
export CACHE_TYPE=memory
./wecomchan
```

### 4. 测试
```bash
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "set_a_sendkey",
    "msg_type": "text",
    "msg": "测试消息"
  }'
```

## 停止服务

### Docker
```bash
docker stop wecomchan
docker rm wecomchan
```

### Go
```bash
# 按 Ctrl+C
```

## 查看日志

### Docker
```bash
docker logs wecomchan
docker logs -f wecomchan  # 实时查看
```

### Go
```bash
# 日志直接输出到终端
```

## 常见错误

### 40001 - sendkey错误
- 确保请求中包含sendkey
- 确保sendkey与环境变量一致

### 44004 - 内容为空
- 确保JSON格式正确
- 确保包含msg或text.content字段

### 企业微信API错误
- 检查WECOM_CID、WECOM_SECRET、WECOM_AID配置
- 确认网络可以访问企业微信API
