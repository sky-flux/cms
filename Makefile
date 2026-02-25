.PHONY: setup dev dev-backend dev-frontend test test-backend test-frontend lint build clean migrate-up migrate-down migrate-status test-perf-smoke test-perf test-perf-public test-all

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
	go build -ldflags="-w -s" -o ./tmp/cms ./cmd/cms
	cd web && bun run build

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
