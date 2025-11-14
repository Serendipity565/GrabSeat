# 第一阶段：构建 Go 应用
FROM golang:1.24-alpine AS builder

# 设置 Go 代理为七牛云的代理
ENV GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 切换到构建目录
WORKDIR /app

# 复制代码
COPY . .

# 下载依赖并构建二进制
RUN go mod tidy && go build -o grabseat

# ========================
# 第二阶段：生成最终镜像
# ========================
FROM alpine:latest

# 安装时区
RUN apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 设置工作目录
WORKDIR /app

# 拷贝二进制文件
COPY --from=builder /app/grabseat .

# 暴露端口
EXPOSE 8080

# 启动服务
CMD ["./grabseat"]
