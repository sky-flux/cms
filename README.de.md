<p align="center">
  <h1 align="center">Sky Flux CMS</h1>
  <p align="center">Ein modernes, leistungsstarkes Headless-CMS für Teams und unabhängige Entwickler.</p>
</p>

<p align="center">
  <a href="./README.md">English</a> | <a href="./README.zh-CN.md">简体中文</a> | <a href="./README.zh-TW.md">繁體中文</a>
</p>

## Funktionen

- **Multi-Site** — PostgreSQL-Schema-Isolierung (`site_{slug}`) für vollständige Datentrennung
- **Headless API** — RESTful API-First-Design, nutzbar mit jedem Frontend
- **Dynamisches RBAC** — 4 integrierte Rollen (super/admin/editor/viewer) mit Unterstützung für benutzerdefinierte Rollen
- **Umfangreiche Inhalte** — Beiträge, Kategorien, Tags, Medienverwaltung mit Entwurf/Veröffentlichung/Zeitplanung
- **Modernes Admin-Dashboard** — Astro 5 SSR + React 19 + shadcn/ui
- **Volltextsuche** — Meilisearch mit integrierter CJK-Tokenisierung
- **Web-Installationsassistent** — Ersteinrichtung direkt im Browser
- **Zwei-Faktor-Authentifizierung** — TOTP 2FA-Unterstützung
- **Kommentarsystem** — Integrierte Kommentare mit Moderation
- **SEO-optimiert** — RSS/Atom-Feeds, Sitemap, URL-Weiterleitungen, Entwurfsvorschau

## Technologie-Stack

| Schicht | Technologie |
|---------|------------|
| Backend | Go 1.25+, Gin, uptrace/bun ORM |
| Datenbank | PostgreSQL 18, Redis 8 |
| Suche | Meilisearch |
| Objektspeicher | RustFS (S3-kompatibel) |
| E-Mail | Resend |
| Frontend | Astro 5 SSR, React 19, shadcn/ui, TanStack Query v5, Zustand |
| Authentifizierung | JWT + Refresh Token + TOTP 2FA |

## Schnellstart

### Voraussetzungen

- Go 1.25+
- Docker 27+ & Docker Compose 2+
- [Bun](https://bun.sh) 1.2+

### Installation

```bash
git clone https://github.com/sky-flux/cms.git
cd cms

# Umgebung initialisieren, Dienste starten, Abhängigkeiten installieren
make setup

# Entwicklungsumgebung starten (Backend Hot-Reload + Frontend)
make dev
```

Das Admin-Dashboard ist unter `http://localhost:4321` erreichbar, die API unter `http://localhost:8080`.

Beim ersten Aufruf führt der Web-Installationsassistent durch die Datenbankkonfiguration, Site-Einrichtung und Erstellung des Admin-Kontos.

### Häufige Befehle

```bash
make dev              # Entwicklungsumgebung starten (Backend Hot-Reload + Frontend)
make test             # Alle Tests ausführen
make lint             # Code-Prüfung (golangci-lint + Biome)
make migrate-up       # Datenbankmigrationen ausführen
make migrate-down     # Letzte Migration rückgängig machen
make build            # Produktions-Binary + Frontend erstellen
```

## Projektstruktur

```
sky-flux-cms/
├── cmd/cms/            # Cobra CLI (serve/migrate/version)
├── internal/           # Geschäftsmodule (auth, post, media, rbac, ...)
│   ├── config/         # Viper-Konfiguration
│   ├── database/       # DB + Redis + RustFS-Verbindungen
│   ├── middleware/      # Gin-Middleware-Kette
│   ├── model/          # Gemeinsame Datenmodelle (bun ORM)
│   ├── router/         # Routen-Registrierung
│   └── pkg/            # Gemeinsame Hilfspakete (apperror/jwt/crypto)
├── migrations/         # bun Go-Code-Migrationen
├── web/                # Astro 5 Admin-Dashboard
├── docs/               # Designdokumente
├── docker-compose.yml  # PostgreSQL, Redis, Meilisearch, RustFS
└── Makefile
```

## Designdokumente

Alle Designdokumente befinden sich im Verzeichnis `docs/`:

| Dokument | Beschreibung |
|----------|-------------|
| prd.md | Produktanforderungen und Funktionsumfang |
| architecture.md | Systemarchitektur und technische Entscheidungen |
| api.md | API-Design (OpenAPI 3.1) |
| database.md | Datenbankschema und Indexierungsstrategie |
| story.md | User Stories und Akzeptanzkriterien |
| security.md | Sicherheitsstrategie und Bedrohungsmodell |

## Lizenz

[MIT](./LICENSE)
