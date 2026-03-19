.PHONY: setup dev dev-backend dev-frontend test test-backend test-frontend lint build clean migrate-up migrate-down migrate-status test-perf-smoke test-perf test-perf-public test-all docker-build docker-push docker-prod-up docker-prod-down docker-prod-logs docker-local-up docker-local-down docker-local-logs docker-local-reset templ-generate templ-watch css-build css-watch

# ──────────────────────────────────────
# 开发环境
# ──────────────────────────────────────

setup:
	@test -f .env || cp .env.example .env
	docker compose up -d --wait
	go mod download
	cd web && bun install
	@echo "Setup complete. Run 'make dev' to start."

dev:
	@make -j2 dev-backend dev-frontend

dev-backend:
	air

dev-frontend:
	cd web && bun dev

# ──────────────────────────────────────
# 测试
# ──────────────────────────────────────

test: test-backend test-frontend

test-backend:
	go test -v -race -count=1 ./...

test-frontend:
	cd web && bun run test

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-perf-smoke:
	k6 run web/performance/smoke.ts

test-perf:
	k6 run web/performance/full-load.ts

test-perf-public:
	k6 run web/performance/scenarios/public-api.ts

test-all: test test-perf-smoke

# ──────────────────────────────────────
# 代码质量
# ──────────────────────────────────────

lint:
	golangci-lint run ./...
	cd web && bun run lint

fmt:
	gofmt -w .
	cd web && bun run format

# ──────────────────────────────────────
# 数据库迁移
# ──────────────────────────────────────

migrate-up:
	go run ./cmd/cms migrate up

migrate-down:
	go run ./cmd/cms migrate down

migrate-status:
	go run ./cmd/cms migrate status

# ──────────────────────────────────────
# 构建
# ──────────────────────────────────────

build:
	templ generate ./web/templates/
	tailwindcss -i web/styles/input.css -o web/static/app.css --minify
	go build -ldflags="-w -s" -o ./tmp/cms ./cmd/cms
	cd web && bun run build

# ──────────────────────────────────────
# Web 公共站点 (Templ + HTMX + Tailwind)
# ──────────────────────────────────────

templ-generate:
	templ generate ./web/templates/

templ-watch:
	templ generate --watch ./web/templates/

css-build:
	tailwindcss -i web/styles/input.css -o web/static/app.css --minify

css-watch:
	tailwindcss -i web/styles/input.css -o web/static/app.css --watch

# ──────────────────────────────────────
# 清理
# ──────────────────────────────────────

clean:
	docker compose stop
	rm -rf tmp/ coverage.out coverage.html
	cd web && rm -rf node_modules/.cache dist/

reset: clean
	docker compose down -v
	$(MAKE) setup

# ──────────────────────────────────────
# Docker 构建与推送
# ──────────────────────────────────────

# 构建镜像（使用 BuildKit 并行构建）
docker-build: docker-build-parallel

docker-push:
	docker tag cms-backend:latest ghcr.io/sky-flux/cms-backend:latest
	docker tag cms-frontend:latest ghcr.io/sky-flux/cms-frontend:latest
	docker push ghcr.io/sky-flux/cms-backend:latest
	docker push ghcr.io/sky-flux/cms-frontend:latest

docker-prod-up:
	docker compose -f docker-compose.prod.yml up -d

docker-prod-down:
	docker compose -f docker-compose.prod.yml down

docker-prod-logs:
	docker compose -f docker-compose.prod.yml logs -f

# ──────────────────────────────────────
# 本地 Docker 测试 (完整容器化)
# ──────────────────────────────────────

docker-local-up:
	@test -f .env || (echo "Creating .env from .env.example" && cp .env.example .env)
	docker compose -f docker-compose.yml -f docker-compose.local.yml up -d --build
	@echo ""
	@echo "========================================"
	@echo "🚀 Sky Flux CMS (Local Docker) 启动完成!"
	@echo "========================================"
	@echo ""
	@echo "访问地址 (通过 Caddy 反向代理):"
	@echo "  前端:   http://localhost:3000"
	@echo "  API:    http://localhost:3000/api/*"
	@echo ""
	@echo "首次访问会自动进入安装向导"
	@echo ""
	@echo "查看日志: make docker-local-logs"
	@echo "停止服务: make docker-local-down"
	@echo "重置环境: make docker-local-reset"
	@echo ""

docker-local-down:
	docker compose -f docker-compose.yml -f docker-compose.local.yml down

docker-local-logs:
	docker compose -f docker-compose.yml -f docker-compose.local.yml logs -f

docker-local-reset:
	docker compose -f docker-compose.yml -f docker-compose.local.yml down -v
	@echo "Volumes 已清理，运行 'make docker-local-up' 重新开始"

# ──────────────────────────────────────
# Docker 构建（兼容模式）
# ──────────────────────────────────────

# 并行构建前后端镜像
docker-build-parallel:
	@echo "Building backend and frontend in parallel..."
	$(MAKE) -j2 docker-build-backend docker-build-frontend

docker-build-backend:
	docker build -t cms-backend:latest .

docker-build-frontend:
	docker build -t cms-frontend:latest --build-arg PUBLIC_API_URL=/api ./web
