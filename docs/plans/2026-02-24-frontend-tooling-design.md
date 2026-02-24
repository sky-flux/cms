# Frontend Tooling Setup Design

> Date: 2026-02-24
> Status: Approved

## Context

web/ 目录已有 Astro 5 + React 19 + Tailwind V4 + shadcn/ui 基础骨架。需要补齐前端工具链：状态管理、国际化、代码质量工具。

## Changes

### Dependencies

**Add:**
- `@tanstack/react-query` — server state (TanStack Query v5)
- `zustand` — client global state
- `react-i18next` + `i18next` — i18n (zh-CN / en)
- `@biomejs/biome` (dev) — lint + format

**Remove:**
- `prettier` — replaced by Biome formatter

**Unchanged:**
- Tailwind V4 (`@tailwindcss/vite ^4.2.1`) already integrated
- shadcn/ui, React 19, Astro 5 unchanged

### Biome Configuration

File: `web/biome.json`

- Scope: `src/**/*.{ts,tsx,astro}`
- Ignore: `dist/`, `node_modules/`, `.astro/`
- Formatter: 2-space indent, 100 line width, single quotes, semicolons
- Linter: recommended rules enabled
- Reference: https://biomejs.dev/guides/big-projects/

### package.json Scripts

```
lint       → biome check src/
lint:fix   → biome check --fix src/
format     → biome format --write src/
typecheck  → astro check
```

Remove old `prettier`-based `format` script.

### Documentation Updates

- CLAUDE.md: add Tailwind V4 to frontend stack, replace Prettier with Biome
- standard.md: update frontend lint section if applicable

### Out of Scope

- CSS/JSON formatting (Tailwind V4 CSS handled by Vite plugin)
- ESLint (fully replaced by Biome)
- Vitest/Playwright setup (deferred to feature development)
- React component/page creation (tooling only)
