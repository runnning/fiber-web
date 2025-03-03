# 阶段 1: 构建
FROM golang:1.23-alpine AS builder

# 安装构建工具
RUN apk add --no-cache gcc musl-dev

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main ./apps/admin/cmd/api

# 阶段 2: 运行
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
    #&& apk del tzdata

# 创建非 root 用户
#RUN adduser -D -g '' appuser

# 创建必要的目录
RUN mkdir -p /app/config /app/logs
    #&& chown -R appuser:appuser /app

WORKDIR /app

# 从构建阶段复制二进制文件和配置
COPY --from=builder /app/main .
COPY --from=builder /app/apps/admin/cmd/config/config.docker.yaml /app/config/config.yaml

# 使用非 root 用户
#USER appuser

# 环境变量
ENV APP_ENV=production \
    TZ=Asia/Shanghai \
    CONFIG_PATH=/app/config \
    CONFIG_NAME=config

# 暴露端口
EXPOSE 3000

# 启动应用
CMD ["./main"]
