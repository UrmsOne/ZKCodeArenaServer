# Judge 服务独立部署指南

## 📦 关于 docker-compose.judge.yml

这是一个**独立的** Judge 服务部署配置文件，可以：
- ✅ 单独部署 Judge 服务
- ✅ 不影响其他服务（app、mongo、redis）
- ✅ 快速重建和测试
- ✅ 与现有网络集成或独立运行

---

## 🚀 使用场景

### 场景1：独立部署（测试用）

适用于：只想测试 judge 服务，不需要其他服务

```bash
# 启动 judge 服务
docker-compose -f docker-compose.judge.yml up -d

# 查看日志
docker-compose -f docker-compose.judge.yml logs -f

# 停止服务
docker-compose -f docker-compose.judge.yml down
```

### 场景2：重建 judge 服务（不影响其他服务）

适用于：已有运行的 app、mongo、redis，只想重建 judge

```bash
# 停止 judge 服务
docker-compose -f docker-compose.judge.yml down

# 重新构建
docker-compose -f docker-compose.judge.yml build

# 启动服务
docker-compose -f docker-compose.judge.yml up -d
```

### 场景3：与现有服务集成

适用于：judge 需要加入已存在的 docker 网络

**修改 `docker-compose.judge.yml`：**

```yaml
networks:
  zk-arena-network:
    external: true  # 使用已存在的网络
```

然后部署：

```bash
docker-compose -f docker-compose.judge.yml up -d
```

---

## 📋 完整部署流程

### 方式一：快速部署

```bash
# 1. 进入项目目录
cd /path/to/ZKCodeArenaServer

# 2. 构建并启动
docker-compose -f docker-compose.judge.yml up -d --build

# 3. 验证
docker exec -it zk-arena-judge java -version
```

### 方式二：分步部署

```bash
# 1. 构建镜像
docker-compose -f docker-compose.judge.yml build

# 2. 启动服务
docker-compose -f docker-compose.judge.yml up -d

# 3. 查看日志
docker-compose -f docker-compose.judge.yml logs -f judge

# 4. 验证安装
docker exec -it zk-arena-judge java -version
```

---

## 🔧 常用命令

### 服务管理

```bash
# 启动
docker-compose -f docker-compose.judge.yml up -d

# 停止
docker-compose -f docker-compose.judge.yml stop

# 重启
docker-compose -f docker-compose.judge.yml restart

# 停止并删除
docker-compose -f docker-compose.judge.yml down

# 停止并删除（包括卷）
docker-compose -f docker-compose.judge.yml down -v
```

### 构建和更新

```bash
# 重新构建
docker-compose -f docker-compose.judge.yml build

# 强制重建（不使用缓存）
docker-compose -f docker-compose.judge.yml build --no-cache

# 重新构建并启动
docker-compose -f docker-compose.judge.yml up -d --build
```

### 日志和调试

```bash
# 查看日志
docker-compose -f docker-compose.judge.yml logs

# 实时日志
docker-compose -f docker-compose.judge.yml logs -f

# 查看最近 100 行
docker-compose -f docker-compose.judge.yml logs --tail 100

# 进入容器
docker-compose -f docker-compose.judge.yml exec judge bash
```

### 状态查看

```bash
# 查看服务状态
docker-compose -f docker-compose.judge.yml ps

# 查看配置
docker-compose -f docker-compose.judge.yml config

# 验证配置文件
docker-compose -f docker-compose.judge.yml config --quiet
```

---

## 🌐 网络配置

### 独立网络模式（默认）

```yaml
networks:
  zk-arena-network:
    driver: bridge
```

**特点：**
- ✅ 完全独立，不依赖其他服务
- ✅ 适合单独测试
- ❌ 无法与 app 服务通信

**访问方式：**
- 宿主机：`http://localhost:5050`

### 共享网络模式

**修改配置：**
```yaml
networks:
  zk-arena-network:
    external: true
```

**特点：**
- ✅ 可与其他服务通信
- ✅ 适合集成部署
- ⚠️ 需要网络已存在

**访问方式：**
- 宿主机：`http://localhost:5050`
- 容器内：`http://judge:5050`

### 创建共享网络

如果网络不存在，先创建：

```bash
# 创建网络
docker network create zk-arena-network

# 然后启动服务
docker-compose -f docker-compose.judge.yml up -d
```

---

## 📊 验证部署

### 基础验证

```bash
# 1. 检查容器状态
docker-compose -f docker-compose.judge.yml ps

# 2. 验证 Java
docker exec -it zk-arena-judge java -version

# 3. 验证 JAVA_HOME
docker exec -it zk-arena-judge echo $JAVA_HOME

# 4. 测试所有编译器
docker exec -it zk-arena-judge bash -c "
  echo 'Java:' && java -version 2>&1 | head -1 &&
  echo 'C++:' && g++ --version | head -1 &&
  echo 'Python:' && python3 --version &&
  echo 'Go:' && go version
"
```

### 功能测试

```bash
# 测试 Java 编译运行
docker exec -it zk-arena-judge bash -c "
  cd /tmp &&
  echo 'public class HelloWorld { public static void main(String[] args) { System.out.println(\"Hello Java 17\"); }}' > HelloWorld.java &&
  javac HelloWorld.java &&
  java HelloWorld
"

# 测试评测接口
curl http://localhost:5050/version
```

---

## 🔄 多环境部署

### 创建不同环境的配置

#### 开发环境：`docker-compose.judge.dev.yml`

```yaml
version: '3.8'

services:
  judge:
    build:
      context: ./deploy/judge
      dockerfile: Dockerfile
    container_name: zk-arena-judge-dev
    ports:
      - "5050:5050"
    environment:
      - JUDGE_DEBUG=1  # 开启调试
    networks:
      - zk-arena-dev-network

networks:
  zk-arena-dev-network:
    driver: bridge
```

#### 生产环境：`docker-compose.judge.prod.yml`

```yaml
version: '3.8'

services:
  judge:
    build:
      context: ./deploy/judge
      dockerfile: Dockerfile
    container_name: zk-arena-judge-prod
    ports:
      - "5050:5050"
    environment:
      - JUDGE_DEBUG=0  # 关闭调试
    restart: always    # 自动重启
    networks:
      - zk-arena-network

networks:
  zk-arena-network:
    external: true
```

### 使用不同环境

```bash
# 开发环境
docker-compose -f docker-compose.judge.dev.yml up -d

# 生产环境
docker-compose -f docker-compose.judge.prod.yml up -d
```

---

## 🔗 与完整部署集成

### 同时使用多个配置文件

```bash
# 启动基础服务（mongo, redis）
docker-compose -f docker-compose.yml up -d mongo redis

# 单独启动 judge
docker-compose -f docker-compose.judge.yml up -d

# 或者合并使用
docker-compose -f docker-compose.yml -f docker-compose.judge.yml up -d
```

### 覆盖配置

创建 `docker-compose.override.yml`：

```yaml
version: '3.8'

services:
  judge:
    environment:
      - JUDGE_DEBUG=1
      - CUSTOM_VAR=value
```

使用：
```bash
docker-compose -f docker-compose.judge.yml -f docker-compose.override.yml up -d
```

---

## 🛠️ 故障排查

### 问题1：容器启动失败

```bash
# 查看详细日志
docker-compose -f docker-compose.judge.yml logs judge

# 检查配置
docker-compose -f docker-compose.judge.yml config

# 验证 Dockerfile
docker build -f deploy/judge/Dockerfile -t test-judge deploy/judge/
```

### 问题2：网络问题

```bash
# 检查网络
docker network ls
docker network inspect zk-arena-network

# 创建网络
docker network create zk-arena-network
```

### 问题3：端口冲突

```bash
# 检查端口占用
netstat -tulpn | grep 5050
lsof -i :5050

# 修改端口映射（在 yml 中）
ports:
  - "5051:5050"  # 改为其他端口
```

### 问题4：构建失败

```bash
# 清理并重建
docker-compose -f docker-compose.judge.yml down
docker system prune -a
docker-compose -f docker-compose.judge.yml build --no-cache
```

---

## 📝 配置参数说明

### 环境变量

| 变量 | 说明 | 默认值 | 可选值 |
|------|------|--------|--------|
| `JUDGE_DEBUG` | 调试模式 | `1` | `0`（关闭）, `1`（开启） |
| `JAVA_HOME` | Java路径 | `/usr/lib/jvm/java-17-openjdk-amd64` | - |

### 端口映射

| 宿主机端口 | 容器端口 | 说明 |
|-----------|---------|------|
| `5050` | `5050` | Judge 评测服务 |

### 数据卷

| 卷名 | 挂载点 | 说明 |
|------|--------|------|
| `judge_data` | `/judge` | 评测临时文件 |

---

## 🎯 最佳实践

### 1. 开发环境

```bash
# 使用独立配置，方便快速重建
docker-compose -f docker-compose.judge.yml up -d --build
```

### 2. 生产环境

```bash
# 与其他服务集成，使用共享网络
# 修改 yml 中 networks 为 external: true
docker-compose -f docker-compose.judge.yml up -d
```

### 3. 快速重建

```bash
# 不影响数据卷
docker-compose -f docker-compose.judge.yml up -d --build --force-recreate
```

### 4. 完全清理

```bash
# 删除容器、网络、卷
docker-compose -f docker-compose.judge.yml down -v
docker rmi zkcodearenaserver-judge
```

---

## 📚 相关文档

- 完整部署：`deploy/judge/README.md`
- 快速开始：`deploy/judge/QUICKSTART.md`
- 命令参考：`JUDGE_DEPLOYMENT_COMMANDS.md`

---

## 💡 小技巧

### 创建快捷别名

在 `~/.bashrc` 或 `~/.zshrc` 中添加：

```bash
alias judge-up='docker-compose -f docker-compose.judge.yml up -d'
alias judge-down='docker-compose -f docker-compose.judge.yml down'
alias judge-logs='docker-compose -f docker-compose.judge.yml logs -f'
alias judge-rebuild='docker-compose -f docker-compose.judge.yml up -d --build --force-recreate'
```

使用：
```bash
source ~/.bashrc
judge-up        # 启动
judge-logs      # 查看日志
judge-rebuild   # 重建
```

---

**提示：** 这个独立配置文件非常适合单独测试和快速迭代 Judge 服务！


