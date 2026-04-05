# GitHub Actions 自动部署说明

本项目已配置GitHub Actions，可以自动构建Docker镜像并推送到Docker Hub。

## 配置步骤

### 1. 准备Docker Hub账号

确保你已经注册了Docker Hub账号：https://hub.docker.com/

### 2. 创建访问令牌（Access Token）

1. 登录 Docker Hub
2. 点击右上角头像 → Account Settings
3. 左侧菜单选择 Security → New Access Token
4. 输入Token描述（如：GitHub Actions）
5. 点击 Generate
6. **重要**：复制生成的Token（只显示一次）

### 3. 配置GitHub仓库Secrets

在你的GitHub仓库中配置以下Secrets：

1. 进入GitHub仓库页面
2. 点击 Settings → Secrets and variables → Actions
3. 点击 New repository secret
4. 添加以下两个Secret：

| Secret名称 | 说明 | 示例值 |
|-----------|------|--------|
| `DOCKER_USERNAME` | Docker Hub用户名 | `your_dockerhub_username` |
| `DOCKER_PASSWORD` | Docker Hub访问令牌 | `dckr_pat_xxxxxx` |

### 4. 验证配置

完成配置后，每次推送到main分支或打tag时，会自动触发构建。

## 触发条件

### 自动触发

以下情况会自动触发构建和推送：

1. **推送到main分支**
   - 推送代码到main分支时自动构建
   - 镜像标签：`latest`

2. **创建标签（Tag）**
   - 创建以 `v` 开头的标签时自动构建
   - 镜像标签：`v1.0.0`, `1.0.0`, `1.0`, `latest`

### 手动触发

1. 进入GitHub仓库的 Actions 页面
2. 选择 "Build and Push Docker Image" 工作流
3. 点击 "Run workflow"
4. 选择分支并运行

## 支持的平台

- `linux/amd64` - x86_64架构（Intel/AMD）
- `linux/arm64` - ARM64架构（树莓派、Apple Silicon等）

## 查看构建状态

### 在GitHub上查看

1. 进入仓库的 Actions 页面
2. 查看最近的构建记录
3. 点击具体构建查看详细日志

### 在Docker Hub上查看

访问：https://hub.docker.com/r/gowfqk/go-wecomchan/tags

## 使用镜像

### 拉取镜像

```bash
# 拉取最新版本
docker pull gowfqk/go-wecomchan:latest

# 拉取指定版本
docker pull gowfqk/go-wecomchan:v2.3.0

# 拉取主次版本
docker pull gowfqk/go-wecomchan:v2.3
```

### 运行容器

```bash
# 使用内存缓存（推荐）
docker run -d -p 8080:8080 \
  -e SENDKEY=your_sendkey \
  -e WECOM_CID=your_cid \
  -e WECOM_SECRET=your_secret \
  -e WECOM_AID=your_aid \
  -e CACHE_TYPE=memory \
  gowfqk/go-wecomchan:latest

# 使用docker-compose
version: '3'
services:
  go-wecomchan:
    image: gowfqk/go-wecomchan:latest
    environment:
      - SENDKEY=your_sendkey
      - WECOM_CID=your_cid
      - WECOM_SECRET=your_secret
      - WECOM_AID=your_aid
      - CACHE_TYPE=memory
    ports:
      - 8080:8080
    restart: unless-stopped
```

## 常见问题

### Q1: 构建失败，提示认证错误

**原因**: Docker Hub凭证配置错误

**解决**:
1. 检查Secrets中的 `DOCKER_USERNAME` 和 `DOCKER_PASSWORD` 是否正确
2. 确认Token没有过期
3. 重新生成Token并更新Secrets

### Q2: 如何修改镜像名称？

修改 `.github/workflows/docker.yml` 中的环境变量：

```yaml
env:
  DOCKER_IMAGE: your-dockerhub-username/go-wecomchan
```

### Q3: 如何跳过自动构建？

在commit message中添加 `[skip ci]` 或 `[ci skip]`：

```bash
git commit -m "Update README [skip ci]"
```

### Q4: 如何手动删除旧的镜像版本？

1. 登录 Docker Hub
2. 进入仓库页面
3. 点击 Tags
4. 选择要删除的标签
5. 点击 Delete

### Q5: 构建需要多长时间？

- 初次构建：约5-8分钟
- 使用缓存：约3-5分钟
- 受镜像大小和网络速度影响

## 工作流详情

### 主要步骤

1. **Checkout代码** - 拉取仓库代码
2. **设置QEMU** - 支持多架构构建
3. **设置Buildx** - Docker构建工具
4. **登录Docker Hub** - 使用Secrets中的凭证
5. **提取元数据** - 自动生成标签
6. **构建和推送** - 构建多架构镜像并推送

### 缓存优化

使用GitHub Actions缓存，加速后续构建：

```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

## 安全建议

1. **定期更新Token** - 建议每3-6个月更新一次访问令牌
2. **使用专用Token** - 为GitHub Actions创建专用Token
3. **限制Token权限** - 只授予必要的读写权限
4. **监控构建日志** - 定期检查构建日志，确保没有泄露敏感信息
5. **使用私有仓库** - 敏感项目建议使用私有Docker Hub仓库

## 相关链接

- GitHub Actions文档: https://docs.github.com/en/actions
- Docker Hub: https://hub.docker.com/
- 本项目镜像: https://hub.docker.com/r/gowfqk/go-wecomchan
