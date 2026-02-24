<p align="center">
  <h1 align="center">Sky Flux CMS</h1>
  <p align="center">面向中小型團隊及獨立開發者的現代化、高效能 Headless 內容管理系統。</p>
</p>

<p align="center">
  <a href="./README.md">English</a> | <a href="./README.zh-CN.md">简体中文</a> | <a href="./README.de.md">Deutsch</a>
</p>

## 核心特性

- **多站點** — PostgreSQL Schema 隔離（`site_{slug}`），完全資料分離
- **Headless API** — RESTful API-First 設計，任意前端消費
- **動態 RBAC** — 內建 4 個角色（super/admin/editor/viewer），支援自訂角色與權限
- **豐富內容** — 文章、分類、標籤、媒體管理，草稿/發佈/定時發佈工作流
- **現代管理後台** — Astro 5 SSR + React 19 + shadcn/ui
- **全文搜尋** — Meilisearch，內建 CJK 分詞
- **Web 安裝精靈** — 瀏覽器中即可完成首次設定
- **雙因素驗證** — TOTP 2FA 支援
- **留言系統** — 內建留言與審核功能
- **SEO 友善** — RSS/Atom 訂閱、Sitemap、URL 重新導向、草稿預覽

## 技術棧

| 層級 | 技術 |
|------|------|
| 後端 | Go 1.25+、Gin、uptrace/bun ORM |
| 資料庫 | PostgreSQL 18、Redis 8 |
| 搜尋 | Meilisearch |
| 物件儲存 | RustFS（S3 相容） |
| 郵件 | Resend |
| 前端 | Astro 5 SSR、React 19、shadcn/ui、TanStack Query v5、Zustand |
| 驗證 | JWT + Refresh Token + TOTP 2FA |

## 快速開始

### 環境需求

- Go 1.25+
- Docker 27+ & Docker Compose 2+
- [Bun](https://bun.sh) 1.2+

### 安裝

```bash
git clone https://github.com/sky-flux/cms.git
cd cms

# 初始化環境、啟動服務、安裝依賴
make setup

# 啟動開發環境（後端熱重載 + 前端）
make dev
```

管理後台存取 `http://localhost:4321`，API 存取 `http://localhost:8080`。

首次存取時，Web 安裝精靈將引導你完成資料庫設定、站點設定和管理員帳號建立。

### 常用指令

```bash
make dev              # 啟動開發環境（後端熱重載 + 前端）
make test             # 執行全部測試
make lint             # 程式碼檢查（golangci-lint + Biome）
make migrate-up       # 執行資料庫遷移
make migrate-down     # 回滾最近一次遷移
make build            # 建置生產二進位 + 前端
```

## 專案結構

```
sky-flux-cms/
├── cmd/cms/            # Cobra CLI（serve/migrate/version）
├── internal/           # 業務模組（auth、post、media、rbac 等）
│   ├── config/         # Viper 設定載入
│   ├── database/       # DB + Redis + RustFS 連線
│   ├── middleware/      # Gin 中介軟體鏈
│   ├── model/          # 共享資料模型（bun ORM）
│   ├── router/         # 路由註冊
│   └── pkg/            # 共享工具套件（apperror/jwt/crypto）
├── migrations/         # bun Go 程式碼遷移
├── web/                # Astro 5 管理後台
├── docs/               # 設計文件
├── docker-compose.yml  # PostgreSQL、Redis、Meilisearch、RustFS
└── Makefile
```

## 設計文件

所有設計文件位於 `docs/` 目錄：

| 文件 | 內容 |
|------|------|
| prd.md | 產品需求與功能範圍 |
| architecture.md | 系統架構與技術決策 |
| api.md | API 設計（OpenAPI 3.1） |
| database.md | 資料庫 Schema 與索引策略 |
| story.md | 使用者故事與驗收標準 |
| security.md | 安全策略與威脅模型 |

## 開源授權

[MIT](./LICENSE)
