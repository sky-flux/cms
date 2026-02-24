<p align="center">
  <h1 align="center">Sky Flux CMS</h1>
  <p align="center">A modern, high-performance headless CMS for teams and indie developers.</p>
</p>

<p align="center">
  <a href="./README.zh-CN.md">简体中文</a> | <a href="./README.zh-TW.md">繁體中文</a> | <a href="./README.de.md">Deutsch</a>
</p>

## Features

- **Multi-Site** — PostgreSQL schema isolation (`site_{slug}`) for complete data separation
- **Headless API** — RESTful API-first design, consume from any frontend
- **Dynamic RBAC** — Built-in roles (super/admin/editor/viewer) with custom role support
- **Rich Content** — Posts, categories, tags, media management with draft/publish/schedule workflow
- **Modern Admin** — Astro 5 SSR + React 19 + shadcn/ui dashboard
- **Full-Text Search** — Meilisearch with built-in CJK tokenization
- **Web Installer** — Browser-based setup wizard for first-time configuration
- **2FA Support** — TOTP two-factor authentication
- **Comments** — Built-in comment system with moderation
- **SEO Ready** — RSS/Atom feeds, sitemap, URL redirects, draft preview

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25+, Gin, uptrace/bun ORM |
| Database | PostgreSQL 18, Redis 8 |
| Search | Meilisearch |
| Storage | RustFS (S3-compatible) |
| Email | Resend |
| Frontend | Astro 5 SSR, React 19, shadcn/ui, TanStack Query v5, Zustand |
| Auth | JWT + Refresh Token + TOTP 2FA |

## Quick Start

### Prerequisites

- Go 1.25+
- Docker 27+ & Docker Compose 2+
- [Bun](https://bun.sh) 1.2+

### Setup

```bash
git clone https://github.com/sky-flux/cms.git
cd cms

# Initialize environment, start services, install dependencies
make setup

# Start development (backend + frontend)
make dev
```

The admin dashboard will be available at `http://localhost:4321`, and the API at `http://localhost:8080`.

On first visit, the web installer will guide you through database setup, site configuration, and admin account creation.

### Common Commands

```bash
make dev              # Start dev environment (backend hot-reload + frontend)
make test             # Run all tests
make lint             # Run linters (golangci-lint + Biome)
make migrate-up       # Run database migrations
make migrate-down     # Rollback last migration
make build            # Build production binary + frontend
```

## Project Structure

```
sky-flux-cms/
├── cmd/cms/            # Cobra CLI (serve/migrate/version)
├── internal/           # Business modules (auth, post, media, rbac, ...)
│   ├── config/         # Viper configuration
│   ├── database/       # DB + Redis + RustFS connections
│   ├── middleware/      # Gin middleware chain
│   ├── model/          # Shared data models (bun ORM)
│   ├── router/         # Route registration
│   └── pkg/            # Shared utilities (apperror/jwt/crypto)
├── migrations/         # bun Go code migrations
├── web/                # Astro 5 admin dashboard
├── docs/               # Design documents
├── docker-compose.yml  # PostgreSQL, Redis, Meilisearch, RustFS
└── Makefile
```

## Documentation

Design documents are available in the `docs/` directory:

| Document | Description |
|----------|-------------|
| prd.md | Product requirements & feature scope |
| architecture.md | System architecture & technical decisions |
| api.md | API design (OpenAPI 3.1) |
| database.md | Database schema & indexing strategy |
| story.md | User stories & acceptance criteria |
| security.md | Security strategy & threat model |

## License

[MIT](./LICENSE)
