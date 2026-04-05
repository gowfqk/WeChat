# GitHub Actions 自动部署说明

本项目已配置GitHub Actions，可以自动构建Docker镜像并推送到Docker Hub。

## 配置步骤

### 1. 准备Docker Hub账号

确保你已经注册了Docker Hub账号：https://hub.docker.com/

如果你还没有账号，请：
1. 访问 https://hub.docker.com/signup
2. 填写用户名、邮箱和密码
3. 完成邮箱验证
4. 登录你的账户

### 2. 创建访问令牌（Access Token）

#### 为什么需要访问令牌？

- **安全**：避免直接使用账户密码
- **可撤销**：随时可以删除令牌，不影响账户
- **有限权限**：令牌只能用于特定用途
- **可追踪**：每个令牌都有唯一标识

#### 创建步骤

1. **登录 Docker Hub**
   - 访问 https://hub.docker.com/
   - 点击右上角 "Sign In"
   - 输入用户名和密码登录

2. **进入安全设置**
   - 点击右上角头像 → "Account Settings"
   - 在左侧菜单中找到 "Security"
   - 点击 "New Access Token"

3. **创建令牌**
   - **Access Token Description**: 输入描述（如：`GitHub Actions - go-wecomchan`）
   - **Access permissions**: 选择权限（推荐选择 "Read, Write, Delete"）
   - 点击 "Generate"

4. **复制令牌** ⚠️ 重要
   - 令牌只会显示一次
   - 格式类似：`dckr_pat_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`
   - 立即复制并保存到安全的地方

5. **保存令牌**
   - 建议使用密码管理器保存
   - 不要保存在代码、文档或聊天记录中
   - 不要与他人分享

### 3. 配置GitHub仓库Secrets

#### 什么是GitHub Secrets？

GitHub Secrets是一种安全的存储敏感信息的方式，可以：
- 加密存储密码、令牌等敏感数据
- 在Actions中安全引用，不会在日志中暴露
- 支持在组织级别共享

#### 配置步骤（详细）

1. **进入仓库设置**
   - 打开你的GitHub仓库页面：https://github.com/gowfqk/WeChat
   - 点击顶部导航栏的 "Settings"
   - 在左侧菜单中选择 "Secrets and variables" → "Actions"

2. **添加第一个Secret：DOCKER_USERNAME**
   - 点击 "New repository secret"
   - **Name**: 输入 `DOCKER_USERNAME`（必须完全一致）
   - **Secret**: 输入你的Docker Hub用户名（例如：`gowfqk`）
   - 点击 "Add secret"

3. **添加第二个Secret：DOCKER_PASSWORD**
   - 再次点击 "New repository secret"
   - **Name**: 输入 `DOCKER_PASSWORD`（必须完全一致）
   - **Secret**: 粘贴你刚才创建的Docker Hub访问令牌
   - 点击 "Add secret"

4. **验证配置**
   - 在Secrets页面应该能看到两个配置项：
     ```
     ✓ DOCKER_USERNAME      Updated at [时间]
     ✓ DOCKER_PASSWORD      Updated at [时间]
     ```

#### 配置示例

| Secret名称 | 说明 | 示例值 | 注意事项 |
|-----------|------|--------|---------|
| `DOCKER_USERNAME` | Docker Hub用户名 | `gowfqk` | 不是邮箱，是用户名 |
| `DOCKER_PASSWORD` | Docker Hub访问令牌 | `dckr_pat_xxxxxx` | 前缀为dckr_pat_ |

#### ⚠️ 安全注意事项

1. **永远不要在代码中硬编码密码**
   ```yaml
   # ❌ 错误
   password: "my_password_123"

   # ✅ 正确
   password: ${{ secrets.DOCKER_PASSWORD }}
   ```

2. **定期更换访问令牌**
   - 建议每3-6个月更换一次
   - 如果怀疑泄露，立即撤销并重新生成

3. **使用不同的令牌**
   - 为不同项目使用不同的令牌
   - 便于追踪和管理

4. **限制令牌权限**
   - 只授予必要的权限
   - 不要使用完全权限的令牌

5. **检查日志泄露**
   - 定期检查Actions日志
   - 确保没有敏感信息泄露

### 4. 验证配置

#### 方法1：查看Actions日志

1. 推送代码到main分支：
   ```bash
   git add .
   git commit -m "test: trigger build"
   git push origin main
   ```

2. 查看构建状态：
   - 访问：https://github.com/gowfqk/WeChat/actions
   - 找到最近的构建
   - 查看详细日志

3. 检查登录步骤：
   ```
   Run docker/login-action@v3
   Login Succeeded
   ```
   - 如果看到 "Login Succeeded"，说明配置成功
   - 如果看到错误，检查Secrets配置是否正确

#### 方法2：本地测试登录

```bash
# 1. 退出之前的登录（如果有）
docker logout

# 2. 测试登录
echo "YOUR_ACCESS_TOKEN" | docker login -u YOUR_USERNAME --password-stdin

# 3. 预期输出
# Login Succeeded
```

#### 方法3：手动触发工作流

1. 访问：https://github.com/gowfqk/WeChat/actions
2. 点击左侧 "Build and Push Docker Image"
3. 点击右上角 "Run workflow"
4. 选择分支，点击 "Run workflow"
5. 查看运行结果

## 触发条件

### 自动触发

以下情况会自动触发构建和推送：

1. **推送到main分支**
   ```bash
   git push origin main
   ```
   - 推送代码到main分支时自动构建
   - 镜像标签：`latest`
   - 适用于日常开发

2. **创建标签（Tag）**
   ```bash
   git tag v2.3.0
   git push origin v2.3.0
   ```
   - 创建以 `v` 开头的标签时自动构建
   - 镜像标签：`v2.3.0`, `2.3.0`, `2.3`, `latest`
   - 适用于发布版本

### 手动触发

1. 进入GitHub仓库的 Actions 页面
2. 选择 "Build and Push Docker Image" 工作流
3. 点击 "Run workflow"
4. 选择分支并运行

### 跳过构建

如果不想触发构建，在commit message中添加：

```bash
git commit -m "Update README [skip ci]"
```

## 支持的平台

- `linux/amd64` - x86_64架构（Intel/AMD）
- `linux/arm64` - ARM64架构（树莓派、Apple Silicon等）

## 查看构建状态

### 在GitHub上查看

1. 进入仓库的 Actions 页面：https://github.com/gowfqk/WeChat/actions
2. 查看最近的构建记录
3. 点击具体构建查看详细日志

### 日志解读

**成功构建的日志**：
```
✅ Checkout code
✅ Set up QEMU
✅ Set up Docker Buildx
✅ Log in to Docker Hub  ← 检查这里是否成功
✅ Extract metadata
✅ Build and push Docker image
```

**常见错误**：

1. **登录失败**
   ```
   Error: Login: unsuccessful: incorrect username or password
   ```
   检查 `DOCKER_USERNAME` 和 `DOCKER_PASSWORD` 是否正确

2. **权限不足**
   ```
   Error: insufficient permissions to push image
   ```
   检查访问令牌权限是否包含 "Write" 权限

3. **仓库不存在**
   ```
   Error: repository access denied
   ```
   检查镜像名称是否正确，确认你有推送权限

### 在Docker Hub上查看

访问：https://hub.docker.com/r/gowfqk/go-wecomchan/tags

可以看到：
- 所有构建的镜像版本
- 镜像大小
- 构建时间
- 架构信息

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

**错误信息**：
```
Error response from daemon: unauthorized: incorrect username or password
```

**原因**：
- Docker Hub凭证配置错误
- 访问令牌过期或失效
- 用户名或密码输入错误

**解决方法**：
1. 检查Secrets配置
   ```
   Settings → Secrets and variables → Actions
   - DOCKER_USERNAME: 确认是用户名（不是邮箱）
   - DOCKER_PASSWORD: 确认是有效的访问令牌
   ```

2. 测试本地登录
   ```bash
   echo "YOUR_TOKEN" | docker login -u YOUR_USERNAME --password-stdin
   ```

3. 重新生成访问令牌
   - Docker Hub → Account Settings → Security
   - 删除旧令牌
   - 创建新令牌
   - 更新GitHub Secrets

### Q2: 如何修改镜像名称？

**当前配置**：
```yaml
env:
  DOCKER_IMAGE: gowfqk/go-wecomchan
```

**修改步骤**：
1. 编辑 `.github/workflows/docker.yml`
2. 修改环境变量：
   ```yaml
   env:
     DOCKER_IMAGE: your-username/your-image-name
   ```
3. 提交并推送：
   ```bash
   git add .github/workflows/docker.yml
   git commit -m "chore: update docker image name"
   git push origin main
   ```

### Q3: 跳过自动构建

在commit message中添加 `[skip ci]` 或 `[ci skip]`：

```bash
git commit -m "Update README [skip ci]"
```

### Q4: 如何删除旧的镜像版本？

1. 登录 Docker Hub
2. 进入仓库页面：https://hub.docker.com/r/gowfqk/go-wecomchan
3. 点击 "Tags"
4. 选择要删除的标签
5. 点击 "Delete"

**注意**：
- 无法删除 `latest` 标签（只能被覆盖）
- 删除后无法恢复

### Q5: 构建需要多长时间？

| 构建类型 | 时间 | 说明 |
|---------|------|------|
| 初次构建 | 5-8分钟 | 需要下载依赖、编译 |
| 使用缓存 | 3-5分钟 | 利用缓存加速 |
| 小改动 | 2-3分钟 | 只需重新打包 |

**影响时间的因素**：
- 镜像大小
- 网络速度
- GitHub Actions负载

### Q6: 如何查看构建缓存状态？

在GitHub Actions日志中查找：

```
Run docker/build-push-action@v5
  #6 [linux/amd64 2/4] WORKDIR /go/src/app  0.3s done
#7 [linux/arm64 2/4] WORKDIR /go/src/app  0.3s done
#8 [linux/amd64 3/4] COPY go.* ./          0.0s
#9 [linux/arm64 3/4] COPY go.* ./          0.0s
#10 [linux/amd64 4/4] COPY . .            1.5s
#11 [linux/arm64 4/4] COPY . .            1.5s
```

查看 `cache-from` 和 `cache-to` 的输出。

### Q7: 为什么PR不会推送镜像？

这是正常的安全设计：

```yaml
- name: Log in to Docker Hub
  if: github.event_name != 'pull_request'  # PR时跳过登录
  uses: docker/login-action@v3
```

**原因**：
- 避免PR中的恶意代码推送镜像
- 保护Docker Hub凭证安全
- 只在main分支或tag时推送

### Q8: 如何启用自动标记版本？

当前已配置自动标记：

```yaml
tags: |
  type=ref,event=branch           # main → latest
  type=semver,pattern={{version}}  # v1.0.0 → 1.0.0, v1.0.0
  type=semver,pattern={{major}}.{{minor}}  # v1.0.0 → 1.0
  type=raw,value=latest,enable={{is_default_branch}}  # main → latest
```

**使用方法**：
```bash
# 创建版本标签
git tag v2.3.0
git push origin v2.3.0

# 会自动生成以下标签：
# - v2.3.0
# - 2.3.0
# - 2.3
# - latest (因为推送到main分支)
```

## 工作流详情

### 主要步骤

1. **Checkout代码**
   - 拉取仓库代码
   - 版本：actions/checkout@v4

2. **设置QEMU**
   - 支持多架构构建
   - 版本：docker/setup-qemu-action@v3

3. **设置Buildx**
   - Docker构建工具
   - 版本：docker/setup-buildx-action@v3

4. **登录Docker Hub**
   - 使用Secrets中的凭证
   - 版本：docker/login-action@v3
   - PR时跳过此步骤

5. **提取元数据**
   - 自动生成标签
   - 版本：docker/metadata-action@v5

6. **构建和推送**
   - 构建多架构镜像
   - 推送到Docker Hub
   - 版本：docker/build-push-action@v5

### 缓存优化

使用GitHub Actions缓存，加速后续构建：

```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

**缓存位置**：
- GitHub Actions自动管理
- 最大10GB存储空间
- 7天后自动清理

## 安全建议

1. **定期更新Token**
   - 建议每3-6个月更新一次访问令牌
   - 删除不再使用的令牌

2. **使用专用Token**
   - 为GitHub Actions创建专用Token
   - 为不同项目使用不同Token
   - 便于追踪和管理

3. **限制Token权限**
   ```yaml
   # 选择最小必要权限
   Read, Write, Delete  # 完全权限
   Read, Write         # 推荐权限
   Read                # 只读权限
   ```

4. **监控构建日志**
   - 定期检查构建日志
   - 确保没有泄露敏感信息
   - 检查异常行为

5. **使用私有仓库**（如果需要）
   - 敏感项目建议使用私有Docker Hub仓库
   - 或使用自建私有镜像仓库

6. **启用双因素认证**
   - 为Docker Hub账户启用2FA
   - 即使Token泄露，账户仍然安全

## 相关链接

- GitHub Actions文档: https://docs.github.com/en/actions
- Docker Hub: https://hub.docker.com/
- 本项目镜像: https://hub.docker.com/r/gowfqk/go-wecomchan
- 本项目Actions: https://github.com/gowfqk/WeChat/actions
- Docker Hub安全: https://docs.docker.com/security/for-developers/access-tokens/

## 快速参考

### 配置检查清单

- [ ] Docker Hub账号已注册
- [ ] 访问令牌已创建并保存
- [ ] GitHub Secrets已配置
  - [ ] DOCKER_USERNAME
  - [ ] DOCKER_PASSWORD
- [ ] 本地测试登录成功
- [ ] 首次构建成功
- [ ] 镜像可以在Docker Hub上看到

### 常用命令

```bash
# 拉取镜像
docker pull gowfqk/go-wecomchan:latest

# 运行容器
docker run -d -p 8080:8080 \
  -e CACHE_TYPE=memory \
  gowfqk/go-wecomchan:latest

# 查看Actions
# https://github.com/gowfqk/WeChat/actions

# 查看镜像
# https://hub.docker.com/r/gowfqk/go-wecomchan/tags
```

### 故障排查流程

1. 检查Secrets配置
2. 查看Actions日志
3. 本地测试登录
4. 重新生成Token
5. 联系支持

---

**文档版本**: v1.0
**最后更新**: 2024-01-15
**维护者**: gowfqk
