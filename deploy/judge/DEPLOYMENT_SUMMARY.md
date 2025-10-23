# Judge 服务部署总结 📝

## 🎯 任务目标

在 go-judge 沙箱 Docker 容器中安装 Java 17 (OpenJDK)，以支持 Java 代码评测。

## ✅ 已完成的工作

### 1. 创建自定义 Dockerfile

**文件：** `deploy/judge/Dockerfile`

**内容：**
- 基于 `criyle/go-judge:latest` 镜像
- 安装 OpenJDK 17
- 安装其他编译环境（g++, gcc, python3, go）
- 配置 JAVA_HOME 环境变量
- 验证所有编译器安装成功

### 2. 修改 Docker Compose 配置

**已修改文件：**
- ✅ `docker-compose.dev.yml` (开发环境)
- ✅ `docker-compose.yml` (生产环境)

**修改内容：**
```yaml
judge:
  build:
    context: ./deploy/judge
    dockerfile: Dockerfile
  environment:
    - JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
```

### 3. 创建部署文档

- ✅ `deploy/judge/README.md` - 完整部署指南
- ✅ `deploy/judge/QUICKSTART.md` - 快速开始指南
- ✅ `deploy/judge/DEPLOYMENT_SUMMARY.md` - 部署总结（本文件）

### 4. 创建自动化脚本

- ✅ `scripts/rebuild-judge.sh` - 一键重建脚本

---

## 📋 部署步骤总览

### 方式一：一键部署（推荐）⭐

```bash
# 在 Linux 服务器上执行

# 1. 进入项目目录
cd /path/to/ZKCodeArenaServer

# 2. 添加执行权限
chmod +x scripts/rebuild-judge.sh

# 3. 运行脚本
# 开发环境
./scripts/rebuild-judge.sh dev

# 生产环境
./scripts/rebuild-judge.sh
```

### 方式二：手动部署

#### Windows 本地（同步文件）
```powershell
# 1. 将以下文件上传到服务器：
- deploy/judge/Dockerfile
- docker-compose.dev.yml
- docker-compose.yml
- scripts/rebuild-judge.sh
```

#### Linux 服务器
```bash
# 1. 停止旧容器
docker stop zk-arena-judge-dev
docker rm zk-arena-judge-dev

# 2. 构建新镜像
docker-compose -f docker-compose.dev.yml build judge

# 3. 启动服务
docker-compose -f docker-compose.dev.yml up -d judge

# 4. 验证
docker exec -it zk-arena-judge-dev java -version
```

---

## 🔍 配置对齐说明

### config.yaml 中的配置

```yaml
Judge:
  SandboxURL: "http://localhost:5050"  # 或 http://judge:5050
  
  Languages:
    java:
      name: "Java 17"
      compile:
        args: ["/usr/bin/javac", "-encoding", "UTF-8", "Main.java"]
        env:
          - "PATH=/usr/bin:/bin"
          - "JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64"
```

### Dockerfile 中的安装

```dockerfile
RUN apt-get install -y openjdk-17-jdk
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
```

### 验证路径一致性

```bash
# 执行后应显示 Java 17
docker exec -it zk-arena-judge-dev java -version

# 执行后应显示 /usr/lib/jvm/java-17-openjdk-amd64
docker exec -it zk-arena-judge-dev echo $JAVA_HOME
```

---

## 🌐 网络配置说明

### 开发环境网络

**网络名：** `zk-arena-dev-network`

**服务通信：**
```
app (8080) ──→ judge:5050
         └──→ mongo:27017
         └──→ redis:6379
```

**访问方式：**
- 从宿主机：`http://localhost:5050`
- 从容器内：`http://judge:5050`

### 生产环境网络

**网络名：** `zk-arena-network`

**服务通信：**
```
nginx (80/443) ──→ app:8080 ──→ judge:5050
                           └──→ mongo:27017
                           └──→ redis:6379
```

---

## 📦 涉及的文件清单

### 新建文件
```
deploy/judge/
  ├── Dockerfile                    # 自定义镜像定义
  ├── README.md                     # 详细部署指南
  ├── QUICKSTART.md                 # 快速开始
  └── DEPLOYMENT_SUMMARY.md         # 本文件

scripts/
  └── rebuild-judge.sh              # 重建脚本
```

### 修改文件
```
docker-compose.dev.yml              # 开发环境配置
docker-compose.yml                  # 生产环境配置
```

### 保持不变
```
conf/config.yaml                    # 无需修改（已对齐）
```

---

## 💡 技术要点

### 1. 为什么选择 OpenJDK 17？

- ✅ 与 `config.yaml` 配置一致
- ✅ 长期支持版本（LTS）
- ✅ 性能和安全性更好
- ✅ Debian/Ubuntu 官方源支持

### 2. 为什么需要自定义镜像？

原始 `criyle/go-judge` 镜像不包含 JDK，需要：
- 安装完整的 JDK（包含 javac 编译器）
- 配置 JAVA_HOME 环境变量
- 保证每次容器启动都有 Java 环境

### 3. 网络配置为什么重要？

- 容器之间通过 Docker 网络通信
- 使用服务名（如 `judge`）作为主机名
- 端口映射只影响宿主机访问

---

## ⚠️ 注意事项

### 1. 首次构建时间

- ⏱️ 预计 5-10 分钟
- 📦 镜像大小约 600-800MB
- 🌐 需要稳定的网络连接

### 2. 磁盘空间

- 镜像：~700MB
- 容器运行：~100MB
- 日志和数据：根据使用量

### 3. 端口占用

确保以下端口未被占用：
- `5050` - Judge 服务
- `8080` - 应用服务
- `27017` - MongoDB
- `6379` - Redis

### 4. 权限要求

- Docker 需要 root 或 sudo 权限
- `privileged: true` 是评测沙箱必需的

---

## 🧪 测试验证

### 基础验证

```bash
# 1. 容器运行状态
docker ps | grep judge

# 2. Java 版本
docker exec -it zk-arena-judge-dev java -version

# 3. 编译器检查
docker exec -it zk-arena-judge-dev bash -c "
  java -version 2>&1 &&
  javac -version 2>&1 &&
  g++ --version &&
  python3 --version &&
  go version
"
```

### 功能测试

```bash
# 测试简单的 Java 程序编译
docker exec -it zk-arena-judge-dev bash -c "
  echo 'public class Test { public static void main(String[] args) { System.out.println(\"Hello\"); }}' > Test.java &&
  javac Test.java &&
  java Test
"

# 应输出: Hello
```

---

## 🚀 部署后检查清单

部署完成后，请确认：

- [ ] Judge 容器正常运行
- [ ] Java 17 安装成功（`java -version`）
- [ ] JAVA_HOME 配置正确
- [ ] 网络连接正常（app 可以访问 judge:5050）
- [ ] 评测功能正常（通过 API 测试）
- [ ] 日志无异常错误

---

## 📞 问题排查

### 容器无法启动

```bash
# 查看日志
docker logs zk-arena-judge-dev

# 检查配置
docker-compose -f docker-compose.dev.yml config
```

### Java 未找到

```bash
# 进入容器检查
docker exec -it zk-arena-judge-dev bash
which java
which javac
echo $JAVA_HOME
```

### 网络连接问题

```bash
# 从 app 容器测试连接
docker exec -it zk-code-arena-server-dev curl http://judge:5050/version
```

### 编译/运行超时

修改 `conf/config.yaml` 增加时间限制。

---

## 📚 相关文档

- 完整部署指南：`deploy/judge/README.md`
- 快速开始：`deploy/judge/QUICKSTART.md`
- 重建脚本：`scripts/rebuild-judge.sh`
- 项目配置：`conf/config.yaml`

---

## 🎉 总结

通过本次部署，Judge 评测服务已经：

✅ 支持 Java 17 代码编译和运行  
✅ 支持 C/C++、Python、Go 等多语言  
✅ 配置与 config.yaml 完全对齐  
✅ 集成到 Docker Compose 编排中  
✅ 提供完整的文档和脚本支持  

**现在可以进行 Java 代码评测了！** 🚀

