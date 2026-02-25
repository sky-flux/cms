# 阶段一：安装依赖
FROM oven/bun:1.3.10-alpine AS deps

WORKDIR /app

# 先复制依赖清单文件（利用层缓存）
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

# 阶段二：构建
FROM oven/bun:1.3.10-alpine AS builder

WORKDIR /app

# 复制依赖（从 deps 阶段）
COPY --from=deps /app/node_modules ./node_modules
# 复制源码
COPY . .

# 构建参数
ARG PUBLIC_API_URL
ENV PUBLIC_API_URL=${PUBLIC_API_URL}

# 构建应用
RUN bun run build

# 阶段三：运行
FROM oven/bun:1.3.10-alpine AS runtime

# 创建非 root 用户
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# 只复制生产依赖和构建产物
# 不需要复制 devDependencies，减小镜像体积
COPY --from=builder --chown=app:app /app/dist ./dist
COPY --from=builder --chown=app:app /app/node_modules ./node_modules
COPY --from=builder --chown=app:app /app/package.json ./

# 运行时用户
USER app

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:3000/ || exit 1

# 使用运行时阶段的别名，保持向后兼容
CMD ["bun", "./dist/server/entry.mjs"]
