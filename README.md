# go-wecomchan

通过企业微信向微信推送消息的Go语言解决方案。

## 功能特性

- ✅ 支持多种消息类型：文本（text）、Markdown（markdown）、图片（image）
- ✅ 支持多种传参方式：Body JSON传参、URL参数传参
- ✅ 支持自定义接收人和应用ID
- ✅ 支持多种缓存方式：不缓存、内存缓存、Redis缓存
- ✅ 支持Docker部署
- ✅ 支持多架构构建（amd64、arm64）

## What's New

### v2.3.0 - 最新更新

- ✅ **新增内存缓存支持**
  - 支持内存缓存access_token，无需Redis服务
  - 适合单实例部署，配置简单
  - 通过CACHE_TYPE环境变量控制缓存方式

### v2.2.0

- ✅ **新增Markdown消息类型支持**
  - 支持企业微信markdown格式消息
  - 支持标题、加粗、颜色、链接、引用、代码等语法
  - 查看完整指南：[MARKDOWN_GUIDE.md](MARKDOWN_GUIDE.md)

- ✅ **新增Body传参方式**
  - 支持JSON格式的请求体传参
  - 完全向后兼容URL参数方式
  - 查看使用说明：[BODY_PARAM_USAGE.md](BODY_PARAM_USAGE.md)

### v2.1.0

- 添加 Dockerfile.architecture 使用docker buildx支持构建多架构镜像。

关于docker buildx build 使用方式参考官方文档:

[https://docs.docker.com/engine/reference/commandline/buildx_build/](https://docs.docker.com/engine/reference/commandline/buildx_build/)

## 配置说明

直接使用和构建二进制文件使用需要golang环境，并且网络可以安装依赖。  
docker构建镜像使用，需要安装docker，不依赖golang以及网络。  

## 修改默认值

修改的sendkey，企业微信公司ID 等默认值为你的企业中的相关信息，如不设置运行时和打包后都可通过环境变量传入。

```golang
var Sendkey = GetEnvDefault("SENDKEY", "set_a_sendkey")
var WecomCid = GetEnvDefault("WECOM_CID", "企业微信公司ID")
var WecomSecret = GetEnvDefault("WECOM_SECRET", "企业微信应用Secret")
var WecomAid = GetEnvDefault("WECOM_AID", "企业微信应用ID")
var WecomToUid = GetEnvDefault("WECOM_TOUID", "@all")
var CacheType = GetEnvDefault("CACHE_TYPE", "none") // 可选值: none, memory, redis
var RedisStat = GetEnvDefault("REDIS_STAT", "OFF")
var RedisAddr = GetEnvDefault("REDIS_ADDR", "localhost:6379")
var RedisPassword = GetEnvDefault("REDIS_PASSWORD", "")
```

## 直接使用

如果没有添加默认值，需要先引入环境变量，以SENDKEY为例：

`export SENDKEY=set_a_sendkey`
依次引入环境变量后，执行
`go run .`

## build命令构建二进制文件使用

1. 构建命令
`go build`

2. 启动
`./wecomchan`

## 构建docker镜像使用（推荐，不依赖golang，不依赖网络）

1. 构建镜像
`docker build -t go-wecomchan .`

2. 构建多架构镜像
```bash
docker buildx build --platform linux/amd64,linux/arm64 -t go-wecomchan:latest .
```

3. 修改默认值后启动镜像
`docker run -dit -p 8080:8080 go-wecomchan`

4. 通过环境变量启动镜像并启用redis

```bash
docker run -dit -e SENDKEY=set_a_sendkey \
-e WECOM_CID=企业微信公司ID \
-e WECOM_SECRET=企业微信应用Secret \
-e WECOM_AID=企业微信应用ID \
-e WECOM_TOUID="@all" \
-e REDIS_STAT=ON \
-e REDIS_ADDR="localhost:6379" \
-e REDIS_PASSWORD="" \
-p 8080:8080 go-wecomchan
```

**可选：启用Redis缓存**

如果需要启用Redis缓存access_token，可以添加以下环境变量：

```bash
docker run -dit -e SENDKEY=set_a_sendkey \
-e WECOM_CID=企业微信公司ID \
-e WECOM_SECRET=企业微信应用Secret \
-e WECOM_AID=企业微信应用ID \
-e WECOM_TOUID="@all" \
-e REDIS_STAT=ON \
-e REDIS_ADDR="redis-server:6379" \
-e REDIS_PASSWORD="your_redis_password" \
-p 8080:8080 go-wecomchan
```

5. 环境变量说明

|名称|描述|
|---|---|
|SENDKEY|发送时用来验证的key|
|WECOM_CID|企业微信公司ID|
|WECOM_SECRET|企业微信应用Secret|
|WECOM_AID|企业微信应用ID|
|WECOM_TOUID|需要发送给的人，详见[企业微信官方文档](https://work.weixin.qq.com/api/doc/90000/90135/90236#%E6%96%87%E6%9C%AC%E6%B6%88%E6%81%AF)|
|CACHE_TYPE|缓存类型，可选值：`none`（不缓存）、`memory`（内存缓存）、`redis`（Redis缓存），默认为`none`|
|REDIS_STAT|是否启用Redis缓存token，`ON`-启用，`OFF`-不启用（仅在CACHE_TYPE=redis时有效）|
|REDIS_ADDR|Redis服务器地址，如不启用Redis缓存可不设置|
|REDIS_PASSWORD|Redis的连接密码，如不启用Redis缓存可不设置|

## 缓存方式说明

### 1. 不使用缓存（默认）

适用于低频调用场景，每次请求都会重新获取access_token。

```bash
docker run -d -p 8080:8080 \
  -e SENDKEY=your_sendkey \
  -e WECOM_CID=your_cid \
  -e WECOM_SECRET=your_secret \
  -e WECOM_AID=your_aid \
  go-wecomchan
```

### 2. 内存缓存（推荐单实例使用）

适用于单实例部署，无需额外服务，缓存有效期约2小时。

```bash
docker run -d -p 8080:8080 \
  -e SENDKEY=your_sendkey \
  -e WECOM_CID=your_cid \
  -e WECOM_SECRET=your_secret \
  -e WECOM_AID=your_aid \
  -e CACHE_TYPE=memory \
  go-wecomchan
```

**优点**：
- ✅ 无需额外服务
- ✅ 配置简单
- ✅ 性能优于不缓存

**缺点**：
- ❌ 服务重启后缓存丢失
- ❌ 多实例部署时缓存不共享

### 3. Redis缓存（推荐多实例使用）

适用于多实例部署，多个实例共享缓存，需要单独部署Redis服务。

```bash
docker run -d -p 8080:8080 \
  -e SENDKEY=your_sendkey \
  -e WECOM_CID=your_cid \
  -e WECOM_SECRET=your_secret \
  -e WECOM_AID=your_aid \
  -e CACHE_TYPE=redis \
  -e REDIS_STAT=ON \
  -e REDIS_ADDR="redis-server:6379" \
  -e REDIS_PASSWORD="your_redis_password" \
  go-wecomchan
```

**优点**：
- ✅ 多实例共享缓存
- ✅ 缓存持久化
- ✅ 适合高并发场景

**缺点**：
- ❌ 需要额外部署Redis服务

## 使用docker-compose 部署

修改docker-compose.yml 文件内的环境变量，之后执行：

```bash
docker-compose up -d
```

注意：docker-compose.yml已简化，不再包含Redis服务。如需启用Redis缓存access_token，请单独部署Redis并设置环境变量REDIS_STAT=ON、REDIS_ADDR、REDIS_PASSWORD。

## 调用方式

### 方式1：Body JSON传参（推荐）

```bash
# 推送文本消息
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "your_sendkey",
    "msg": "这是一条测试消息",
    "msg_type": "text"
  }'

# 推送Markdown消息
curl -X POST http://localhost:8080/wecomchan \
  -H "Content-Type: application/json" \
  -d '{
    "sendkey": "your_sendkey",
    "msg": "# 通知\n\n**内容**: 系统将于今晚维护",
    "msg_type": "markdown"
  }'

# 推送图片消息（使用multipart/form-data）
curl -X POST http://localhost:8080/wecomchan \
  -F "sendkey=your_sendkey" \
  -F "msg_type=image" \
  -F "media=@test.jpg"
```

### 方式2：URL参数传参（向后兼容）

```bash
# 推送文本消息
curl -X GET "http://localhost:8080/wecomchan?sendkey={你的sendkey}&msg={你的文本消息}&msg_type=text"

# 推送Markdown消息
curl -X GET "http://localhost:8080/wecomchan?sendkey={你的sendkey}&msg={markdown内容}&msg_type=markdown"

# 推送图片消息
curl -X POST 'http://localhost:8080/wecomchan?sendkey={你的sendkey}&msg_type=image' \
--form 'media=@"test.jpg"'
```

### 参数说明

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| sendkey | string | 是 | 验证密钥，与环境变量SENDKEY保持一致 |
| msg | string | 是 | 消息内容 |
| msg_type | string | 否 | 消息类型，默认为"text"，可选值：text、markdown、image |
| touser | string | 否 | 接收人，默认为环境变量WECOM_TOUID的值 |
| agentid | string | 否 | 应用ID，默认为环境变量WECOM_AID的值 |

详细使用说明请查看：[BODY_PARAM_USAGE.md](BODY_PARAM_USAGE.md)

## 已完成功能

- [x] Dockerfile 打包镜像(不依赖网络环境)
- [x] 通过环境变量传递企业微信id，secret等，镜像一次构建多次使用
- [x] docker-compose 快速部署
- [x] 支持多种消息类型（文本、Markdown、图片）
- [x] 支持Body JSON传参方式
- [x] 支持自定义接收人和应用ID
- [x] 支持多架构Docker镜像构建

## 鸣谢

本项目基于 [easychen/wecomchan](https://github.com/easychen/wecomchan) 项目进行开发和改进。

原项目 [easychen/wecomchan](https://github.com/easychen/wecomchan) 是一个开源的企业微信消息推送解决方案，提供了多语言实现（PHP、Go、Python等）和多种云平台部署方案。

感谢原项目作者 easychen 的开源贡献！

### 原项目地址

- GitHub: https://github.com/easychen/wecomchan
- 许可证: MIT

### 本项目改进

在原项目基础上，本项目进行了以下改进：

- ✅ 新增Markdown消息类型支持
- ✅ 新增Body JSON传参方式
- ✅ 简化项目结构，专注于Go版本
- ✅ 更新和完善文档
- ✅ 优化代码结构