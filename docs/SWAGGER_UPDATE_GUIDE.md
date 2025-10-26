# Swagger 文档更新指南

## 问题说明

当您在 Docker 中运行项目后，通过 Swagger UI 访问接口文档时，发现文档显示的是旧版本，新增的接口没有显示出来。

### 问题原因

Docker 镜像构建时使用的是旧的 `docs/` 目录文件，没有重新生成 Swagger 文档。

---

## 解决方案

### 方案 1：重新构建 Docker 镜像（推荐 - 已自动化）

我们已经在 Dockerfile 中添加了自动生成 Swagger 文档的步骤，现在只需要重新构建镜像即可。

#### 操作步骤：

```bash
# 1. 停止现有容器
docker-compose down

# 2. 重新构建镜像（不使用缓存）
docker-compose build --no-cache

# 3. 启动服务
docker-compose up -d

# 4. 查看日志确认启动成功
docker-compose logs -f app
```

#### 如果使用开发环境：

```bash
# 1. 停止现有容器
docker-compose -f docker-compose.dev.yml down

# 2. 重新构建镜像
docker-compose -f docker-compose.dev.yml build --no-cache

# 3. 启动服务
docker-compose -f docker-compose.dev.yml up -d

# 4. 访问 Swagger UI
# http://你的IP:8080/swagger/index.html
```

---

### 方案 2：本地生成后构建（手动方式）

如果您想在本地先生成 Swagger 文档，然后再构建 Docker 镜像：

```bash
# 1. 安装 swag 工具（如果还没安装）
go install github.com/swaggo/swag/cmd/swag@latest

# 2. 生成 Swagger 文档
make swagger
# 或者手动执行
swag init -g cmd/main.go -o docs --parseDependency --parseInternal

# 3. 提交更新的文档到 Git
git add docs/
git commit -m "docs: update swagger documentation"

# 4. 重新构建 Docker 镜像
docker-compose build --no-cache
docker-compose up -d
```

---

## 验证 Swagger 文档是否更新

### 1. 检查文档生成时间

```bash
# 查看 docs 目录文件修改时间
ls -lh docs/
```

### 2. 访问 Swagger UI

```
http://你的服务器IP:8080/swagger/index.html
```

### 3. 检查新接口

在 Swagger UI 中，找到以下新增接口（示例）：

- `POST /api/v1/courses/teachers` - 课程创建者添加老师
- `DELETE /api/v1/courses/teachers` - 课程创建者删除课程老师
- `POST /api/v1/courses/clazzes/teachers` - 班级添加老师
- `DELETE /api/v1/courses/clazzes/teachers` - 班级删除老师
- `POST /api/v1/courses/teacher/query` - 分页查询老师加入的课程

---

## 常见问题排查

### Q1: 执行 `make swagger` 提示 swag 命令不存在

**解决方法：**
```bash
go install github.com/swaggo/swag/cmd/swag@latest

# 确保 GOPATH/bin 在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin
```

### Q2: Docker 构建时 swag init 失败

**可能原因：**
- 网络问题导致 `go install` 失败
- Go 代理配置问题

**解决方法：**
```dockerfile
# 在 Dockerfile 中设置 Go 代理
ENV GOPROXY=https://goproxy.cn,direct
```

### Q3: 重新构建后文档还是旧的

**排查步骤：**

1. 确认使用了 `--no-cache` 参数
```bash
docker-compose build --no-cache
```

2. 删除旧镜像重新构建
```bash
docker-compose down --rmi all
docker-compose build
docker-compose up -d
```

3. 检查容器内的文档
```bash
docker exec -it <container_name> ls -lh /app/docs/
docker exec -it <container_name> cat /app/docs/swagger.json | head -20
```

### Q4: Swagger UI 显示 404 或无法加载

**检查步骤：**

1. 确认 docs 目录已复制到镜像
```bash
docker exec -it <container_name> ls -la /app/docs/
```

2. 检查 swagger 路由是否注册
```go
// 在 server.go 中应该有
s.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

3. 检查 docs 包是否导入
```go
// 在 main.go 中应该有
_ "zk-code-arena-server/docs"
```

---

## 开发流程建议

### 本地开发

```bash
# 1. 修改代码，添加新接口
vim pkg/app/api-server/server/server_xxx.go

# 2. 添加 Swagger 注释
# @Summary      接口摘要
# @Description  详细描述
# @Tags         标签
# ...

# 3. 生成 Swagger 文档
make swagger

# 4. 本地编译测试
make build
make run

# 5. 访问本地 Swagger
# http://localhost:8080/swagger/index.html
```

### 提交代码

```bash
# 提交时包含生成的文档
git add docs/swagger.json docs/swagger.yaml docs/docs.go
git commit -m "feat: add new API endpoints with swagger docs"
git push
```

### 部署到服务器

```bash
# 在服务器上
cd /path/to/project
git pull

# 重新构建和部署
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# 验证
curl http://localhost:8080/swagger/doc.json
```

---

## Makefile 快捷命令

我们已经添加了以下快捷命令：

```bash
# 生成 Swagger 文档
make swagger

# 构建项目
make build

# 构建 Docker 镜像
make docker-build

# 运行 Docker 容器
make docker-run

# 停止 Docker 容器
make docker-stop

# 查看所有可用命令
make help
```

---

## 自动化建议

### CI/CD 集成

在 CI/CD 流程中添加 Swagger 文档生成步骤：

```yaml
# .github/workflows/deploy.yml 示例
name: Deploy

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.23
      
      - name: Install swag
        run: go install github.com/swaggo/swag/cmd/swag@latest
      
      - name: Generate Swagger docs
        run: swag init -g cmd/main.go -o docs --parseDependency --parseInternal
      
      - name: Build Docker image
        run: docker build -t zk-code-arena-server .
      
      - name: Deploy
        run: # your deployment script
```

---

## 总结

现在 Dockerfile 已经配置为**自动生成 Swagger 文档**，您只需要：

1. **重新构建 Docker 镜像**：`docker-compose build --no-cache`
2. **重启服务**：`docker-compose up -d`
3. **访问 Swagger UI** 验证新接口

这样就能看到最新的 API 文档了！🎉

