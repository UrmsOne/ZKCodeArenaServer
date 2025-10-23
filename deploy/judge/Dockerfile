FROM criyle/go-judge:latest

# 更新包管理器并安装编译环境
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    openjdk-17-jdk \
    g++ \
    gcc \
    python3 \
    python3-pip \
    golang-go && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# 设置 JAVA_HOME 环境变量
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
ENV PATH=$JAVA_HOME/bin:$PATH

# 验证安装
RUN java -version && \
    javac -version && \
    g++ --version && \
    gcc --version && \
    python3 --version && \
    go version

# 暴露评测服务端口
EXPOSE 5050

# 启动评测服务
CMD ["/bin/executorserver", "-http-addr=:5050"]

