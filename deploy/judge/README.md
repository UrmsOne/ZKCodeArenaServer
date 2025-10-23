# Judge 评测服务部署指南

## 📋 说明

本目录包含带有 Java 17 编译环境的自定义 go-judge 镜像配置。

## 🔧 已安装的编译环境

- **Java 17** (OpenJDK)
- **C++ 17** (g++)
- **C11** (gcc)
- **Python 3**
- **Go**

## 🚀 部署步骤

### 1. 删除旧的 Judge 容器

在Linux服务器上执行：

```bash
# 停止并删除旧容器
docker stop zk-arena-judge-dev
docker rm zk-arena-judge-dev

# （可选）删除旧镜像释放空间
docker rmi criyle/go-judge:latest
```

### 2. 重新构建并启动服务

```bash
# 进入项目根目录
cd /path/to/ZKCodeArenaServer

# 构建自定义 judge 镜像（首次构建约5-10分钟）
docker-compose -f docker-compose.dev.yml build judge

# 启动所有服务
docker-compose -f docker-compose.dev.yml up -d

# 或者只启动 judge 服务
docker-compose -f docker-compose.dev.yml up -d judge
```

### 3. 验证安装

```bash
# 查看容器状态
docker ps | grep judge

# 验证 Java 安装
docker exec -it zk-arena-judge-dev java -version
docker exec -it zk-arena-judge-dev javac -version

# 验证 JAVA_HOME
docker exec -it zk-arena-judge-dev echo $JAVA_HOME

# 查看所有编译器版本
docker exec -it zk-arena-judge-dev bash -c "java -version && g++ --version && python3 --version"

# 查看启动日志
docker logs zk-arena-judge-dev
```

### 4. 预期输出

Java 版本应该显示：

```
openjdk version "17.0.x" 2023-xx-xx
OpenJDK Runtime Environment (build 17.0.x+x)
OpenJDK 64-Bit Server VM (build 17.0.x+x, mixed mode, sharing)
```

## 🔄 更新镜像

如果需要更新镜像或添加新的依赖：

```bash
# 1. 修改 Dockerfile
vim deploy/judge/Dockerfile

# 2. 停止并删除旧容器
docker-compose -f docker-compose.dev.yml down judge

# 3. 强制重新构建（不使用缓存）
docker-compose -f docker-compose.dev.yml build --no-cache judge

# 4. 启动服务
docker-compose -f docker-compose.dev.yml up -d judge
```

## 📝 注意事项

1. **首次构建时间较长**：下载 JDK 和其他依赖需要 5-10 分钟
2. **磁盘空间**：自定义镜像约 600-800MB
3. **网络要求**：需要稳定的网络连接下载依赖包
4. **配置一致性**：确保 `conf/config.yaml` 中的路径与镜像中的一致

## 🐛 故障排查

### 问题1：容器启动失败

```bash
# 查看详细日志
docker logs zk-arena-judge-dev

# 检查配置
docker-compose -f docker-compose.dev.yml config
```

### 问题2：Java 未找到

```bash
# 进入容器检查
docker exec -it zk-arena-judge-dev bash
which java
ls -la /usr/lib/jvm/
```

### 问题3：编译超时

检查 `conf/config.yaml` 中的时间限制配置是否合理。

## 🌐 网络配置

Judge 服务在 `zk-arena-dev-network` 网络中，可与以下服务通信：

- **app** (应用服务) - 通过 `http://judge:5050` 访问
- **mongo** (数据库)
- **redis** (缓存)

## 📦 生产环境部署

如需在生产环境部署，修改 `docker-compose.yml` 文件，应用相同的配置：

```yaml
  judge:
    build:
      context: ./deploy/judge
      dockerfile: Dockerfile
    container_name: zk-arena-judge-prod
    # ... 其他配置
```

