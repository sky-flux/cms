# 阶段一：构建
# 使用特定版本标签，避免意外的 breaking changes
FROM golang:1.26.0-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 优先复制依赖文件（利用层缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-w -s" -o /build/cms ./cmd/cms

# 阶段二：运行
FROM alpine:3.21

# 只安装运行时必需的包
RUN apk add --no-cache ca-certificates tzdata curl && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

# 从构建阶段复制二进制和 entrypoint
COPY --from=builder /build/cms .
COPY --from=builder --chown=app:app /build/entrypoint.sh .
RUN chmod +x entrypoint.sh && chown -R app:app /app

USER app

EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["./entrypoint.sh"]
CMD ["./cms", "serve"]
