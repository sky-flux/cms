#!/bin/sh
set -e

# 读取 Docker Secrets 文件并注入环境变量
for var in DB_PASSWORD REDIS_PASSWORD JWT_SECRET TOTP_ENCRYPTION_KEY \
           MEILI_MASTER_KEY RUSTFS_ACCESS_KEY RUSTFS_SECRET_KEY RESEND_API_KEY; do
    file_var="${var}_FILE"
    eval file_path="\${$file_var:-}"
    if [ -f "${file_path}" ]; then
        export "$var"="$(cat "${file_path}")"
    fi
done

SERVER_URL="${SERVER_URL:-http://localhost:8080}"

# 启动服务器进程（在后台）
echo "Starting server in background..."
./cms serve &
SERVER_PID=$!

# 等待服务器启动
echo "Waiting for server at ${SERVER_URL}..."
for i in $(seq 1 60); do
    if curl -sf "${SERVER_URL}/health" > /dev/null 2>&1; then
        echo "Server is ready."
        break
    fi
    if [ $i -eq 60 ]; then
        echo "Server failed to start within 60 seconds."
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi
    sleep 1
done

# 自动运行数据库迁移（仅在 AUTO_MIGRATE=true 时执行）
if [ "${AUTO_MIGRATE:-false}" = "true" ]; then
    echo "Running database migrations..."
    ./cms migrate up
    echo "Migrations complete."
fi

# 自动安装（仅在 AUTO_SETUP=true 时执行）
if [ "${AUTO_SETUP:-false}" = "true" ]; then
    echo "Running auto-setup..."

    # 检查是否已安装
    CHECK_RESP=$(curl -sf -X POST "${SERVER_URL}/api/v1/setup/check" 2>&1)
    if [ $? -ne 0 ]; then
        echo "Setup check endpoint failed, skipping setup."
    else
        # 解析 JSON 响应判断是否已安装
        INSTALLED=$(echo "$CHECK_RESP" | grep -o '"installed":[^,}]*' | grep -o 'true\|false' || echo "false")
        if [ "$INSTALLED" = "true" ]; then
            echo "System already installed, skipping setup."
        else
            # 执行自动安装
            echo "Creating initial site and admin user..."
            SETUP_RESP=$(curl -sf -X POST "${SERVER_URL}/api/v1/setup/initialize" \
                -H "Content-Type: application/json" \
                -d "{
                    \"site_name\": \"${SEED_SITE_NAME:-My Site}\",
                    \"site_slug\": \"${SEED_SITE_SLUG:-sky-flux}\",
                    \"site_url\": \"${SEED_SITE_URL:-http://localhost:3000}\",
                    \"super_email\": \"${SEED_SUPER_EMAIL:-admin@example.com}\",
                    \"super_password\": \"${SEED_SUPER_PASSWORD:-admin123}\",
                    \"super_name\": \"${SEED_SUPER_NAME:-Administrator}\",
                    \"locale\": \"${SEED_LOCALE:-zh-CN}\"
                }" 2>&1)

            if [ $? -eq 0 ] && echo "$SETUP_RESP" | grep -q '"success":true'; then
                echo "Setup complete!"
            else
                echo "Setup failed: ${SETUP_RESP}"
            fi
        fi
    fi
fi

# 等待后台服务器进程
wait $SERVER_PID
