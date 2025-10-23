# Judge 服务快速部署指南 🚀

## 📦 一键部署（推荐）

### Linux 服务器部署

```bash
# 1. 上传项目到服务器
cd /path/to/ZKCodeArenaServer

# 2. 添加执行权限
chmod +x scripts/rebuild-judge.sh

# 3. 运行重建脚本
# 开发环境
./scripts/rebuild-judge.sh dev

# 或生产环境
./scripts/rebuild-judge.sh
```

脚本会自动完成：
- ✅ 停止并删除旧容器
- ✅ 构建带 Java 17 的新镜像
- ✅ 启动服务
- ✅ 验证安装

---

## 🔧 手动部署（详细步骤）

### 开发环境

```bash
# 1. 停止并删除旧容器
docker stop zk-arena-judge-dev
docker rm zk-arena-judge-dev

# 2. 构建镜像
docker-compose -f docker-compose.dev.yml build judge

# 3. 启动服务
docker-compose -f docker-compose.dev.yml up -d judge

# 4. 验证
docker exec -it zk-arena-judge-dev java -version
```

### 生产环境

```bash
# 1. 停止并删除旧容器
docker stop zk-arena-judge
docker rm zk-arena-judge

# 2. 构建镜像
docker-compose -f docker-compose.yml build judge

# 3. 启动服务
docker-compose -f docker-compose.yml up -d judge

# 4. 验证
docker exec -it zk-arena-judge java -version
```

---

## ✅ 验证清单

执行以下命令确保一切正常：

```bash
# 开发环境容器名: zk-arena-judge-dev
# 生产环境容器名: zk-arena-judge

# 1. 检查容器状态
docker ps | grep judge

# 2. 验证 Java
docker exec -it <容器名> java -version
# 应显示: openjdk version "17.0.x"

# 3. 验证 JAVA_HOME
docker exec -it <容器名> echo $JAVA_HOME
# 应显示: /usr/lib/jvm/java-17-openjdk-amd64

# 4. 验证所有编译器
docker exec -it <容器名> bash -c "
  echo 'Java:' && java -version 2>&1 | head -1 &&
  echo 'C++:' && g++ --version | head -1 &&
  echo 'C:' && gcc --version | head -1 &&
  echo 'Python:' && python3 --version &&
  echo 'Go:' && go version
"

# 5. 测试评测接口
curl http://localhost:5050/version
```

---

## 🎯 配置对齐检查

确保以下配置文件中的路径一致：

### conf/config.yaml
```yaml
Judge:
  SandboxURL: "http://localhost:5050"  # 本地测试
  # 或
  SandboxURL: "http://judge:5050"      # Docker 容器内访问
  
  Languages:
    java:
      name: "Java 17"
      compile:
        env:
          - "JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64"
```

### docker-compose.*.yml
```yaml
judge:
  environment:
    - JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
```

---

## 📊 监控和日志

```bash
# 查看实时日志
docker logs -f <容器名>

# 查看最近 100 行日志
docker logs --tail 100 <容器名>

# 进入容器调试
docker exec -it <容器名> bash

# 查看容器资源使用
docker stats <容器名>
```

---

## ⚠️ 常见问题

### Q1: 容器启动失败
```bash
# 查看详细错误
docker logs <容器名>

# 检查端口占用
netstat -tulpn | grep 5050

# 重新构建（不使用缓存）
docker-compose -f docker-compose.*.yml build --no-cache judge
```

### Q2: Java 未找到
```bash
# 检查环境变量
docker exec -it <容器名> env | grep JAVA

# 检查安装路径
docker exec -it <容器名> ls -la /usr/lib/jvm/
```

### Q3: 编译超时
修改 `conf/config.yaml` 增加时间限制：
```yaml
java:
  compile:
    time_limit: 20000000000  # 改为 20 秒
```

### Q4: 镜像构建失败
```bash
# 检查网络连接
ping google.com

# 使用国内镜像源
# 修改 deploy/judge/Dockerfile，在 apt-get 前添加：
# RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list
```

---

## 🔄 更新和维护

### 更新镜像
```bash
# 重新构建
docker-compose -f docker-compose.*.yml build --no-cache judge

# 重启服务
docker-compose -f docker-compose.*.yml up -d judge
```

### 清理无用镜像
```bash
# 查看所有镜像
docker images

# 删除无用镜像
docker image prune -a
```

### 备份配置
```bash
# 备份整个 deploy 目录
tar -czf deploy-backup-$(date +%Y%m%d).tar.gz deploy/
```

---

## 📞 技术支持

如遇问题，请提供以下信息：

1. 容器日志：`docker logs <容器名>`
2. 系统信息：`uname -a && docker version`
3. 配置文件：`conf/config.yaml`
4. 错误截图或错误信息

---

## 🎉 部署完成后

服务部署成功后，judge 服务将：
- 监听 `5050` 端口
- 支持 Java 17、C++、C、Python、Go 代码评测
- 通过 Docker 网络与主应用通信

**下一步：**
- 通过后端 API 测试代码提交功能
- 查看评测日志确认工作正常
- 根据负载调整 `conf/config.yaml` 中的 Workers 配置

