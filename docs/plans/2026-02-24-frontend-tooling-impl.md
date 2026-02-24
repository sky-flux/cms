# Frontend Tooling Setup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Install frontend dependencies (TanStack Query, Zustand, react-i18next, Biome) and replace Prettier with Biome for lint + format.

**Architecture:** Pure tooling setup — add packages, configure Biome, update scripts, update docs. No feature code.

**Tech Stack:** Biome 2.x, TanStack Query v5, Zustand, react-i18next, i18next

---

### Task 1: Install Runtime Dependencies

**Files:**
- Modify: `web/package.json`

**Step 1: Install TanStack Query + Zustand + i18n packages**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun add @tanstack/react-query zustand react-i18next i18next
```

Expected: packages added to `dependencies` in package.json, bun.lock updated.

**Step 2: Verify installation**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun pm ls --depth 0 | grep -E "tanstack|zustand|i18next"
```

Expected: all 4 packages listed.

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add web/package.json web/bun.lock && git commit -m "feat(web): add TanStack Query, Zustand, react-i18next dependencies"
```

---

### Task 2: Replace Prettier with Biome

**Files:**
- Modify: `web/package.json`

**Step 1: Remove Prettier**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun remove prettier
```

**Step 2: Install Biome as exact dev dependency**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun add -D -E @biomejs/biome
```

**Step 3: Verify Biome is installed**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx biome --version
```

Expected: Biome version number (2.x).

**Step 4: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add web/package.json web/bun.lock && git commit -m "chore(web): replace Prettier with Biome"
```

---

### Task 3: Configure Biome

**Files:**
- Create: `web/biome.json`

**Step 1: Create biome.json**

Create `web/biome.json` with:

```json
{
  "$schema": "https://biomejs.dev/schemas/2.0.6/schema.json",
  "files": {
    "include": ["src/**/*.ts", "src/**/*.tsx", "src/**/*.astro"],
    "ignore": ["dist/", "node_modules/", ".astro/"]
  },
  "formatter": {
    "indentStyle": "space",
    "indentWidth": 2,
    "lineWidth": 100
  },
  "linter": {
    "rules": {
      "recommended": true
    }
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "single",
      "semicolons": "always"
    }
  }
}
```

NOTE: Check the actual installed Biome version to use the correct schema URL. Run `bunx biome --version` in `web/` and adjust `$schema` to match (e.g., `2.0.6` → actual version).

**Step 2: Verify Biome config loads**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx biome check src/
```

Expected: Biome runs and reports any lint/format issues in existing files. No config errors.

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add web/biome.json && git commit -m "chore(web): add Biome configuration"
```

---

### Task 4: Fix Existing Code to Pass Biome

**Files:**
- Modify: `web/src/lib/utils.ts` (double quotes → single quotes, etc.)
- Modify: `web/src/lib/api.ts`
- Potentially other existing .ts/.tsx files

**Step 1: Auto-fix with Biome**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx biome check --fix src/
```

**Step 2: Verify clean**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx biome check src/
```

Expected: 0 errors, 0 warnings.

**Step 3: Verify build still works**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run build
```

Expected: Build succeeds.

**Step 4: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add web/src/ && git commit -m "style(web): format existing code with Biome"
```

---

### Task 5: Update package.json Scripts

**Files:**
- Modify: `web/package.json`

**Step 1: Update scripts**

Replace the `scripts` section in `web/package.json`:

```json
{
  "scripts": {
    "dev": "astro dev",
    "build": "astro build",
    "preview": "astro preview",
    "lint": "biome check src/",
    "lint:fix": "biome check --fix src/",
    "format": "biome format --write src/",
    "typecheck": "astro check",
    "test": "echo 'No tests yet' && exit 0",
    "astro": "astro"
  }
}
```

Changes:
- `lint`: `astro check` → `biome check src/`
- `format`: `prettier --write .` → `biome format --write src/`
- Add: `lint:fix`
- Keep: `typecheck` as `astro check` (TypeScript checking is separate from Biome)

**Step 2: Verify scripts work**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run lint && bun run format && bun run typecheck
```

Expected: All three pass.

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add web/package.json && git commit -m "chore(web): update npm scripts for Biome"
```

---

### Task 6: Update VS Code Settings

**Files:**
- Modify: `web/.vscode/extensions.json`

**Step 1: Add Biome extension recommendation**

Update `web/.vscode/extensions.json`:

```json
{
  "recommendations": [
    "astro-build.astro-vscode",
    "biomejs.biome"
  ],
  "unwantedRecommendations": []
}
```

**Step 2: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add web/.vscode/extensions.json && git commit -m "chore(web): recommend Biome VS Code extension"
```

---

### Task 7: Update Documentation

**Files:**
- Modify: `CLAUDE.md` (2 sections)
- Modify: `docs/standard.md` (add §2.8 or insert lint section)

**Step 1: Update CLAUDE.md frontend section**

In the `### 前端` section, update to reflect current tooling:

```markdown
### 前端

- Astro Islands 架构: 仅交互组件加载 React Runtime
- 状态管理: Zustand (全局) + TanStack Query (服务端状态)
- UI 组件: shadcn/ui (Radix UI + Tailwind CSS V4)
- 错误提示: Sonner Toast
- 界面语言: react-i18next (zh-CN / en)
- 代码质量: Biome (lint + format, 替代 ESLint + Prettier)
```

Also in the tech stack block at the top, ensure it says:

```
前端: Astro 5 SSR + React 19 + shadcn/ui + TanStack Query v5 + Zustand + Tailwind V4
```

And in 常用命令 section, `make lint` and `make fmt` descriptions should note Biome.

**Step 2: Add lint/format section to standard.md**

After §2.6 导入排序 (line 1246), before §2.7 前端测试规范, insert a new section:

```markdown
### 2.7 代码质量工具 (Biome)

项目使用 [Biome](https://biomejs.dev/) 统一 lint + format（替代 ESLint + Prettier）。

#### 配置

- 配置文件: `web/biome.json`
- Scope: `src/**/*.{ts,tsx,astro}`
- 规则集: recommended（内置最佳实践）

#### 常用命令

```bash
bun run lint        # 检查 lint + format 问题
bun run lint:fix    # 自动修复
bun run format      # 仅格式化
bun run typecheck   # TypeScript 类型检查 (astro check)
```

#### 规则说明

| 规则 | 说明 |
|------|------|
| 缩进 | 2 spaces |
| 行宽 | 100 字符 |
| 引号 | 单引号 (`'`) |
| 分号 | 始终添加 |
| 导入排序 | Biome 内置 `organizeImports`（遵循 §2.6 排序约定） |
```

Renumber existing §2.7 前端测试规范 → §2.8, §2.8 集成测试 → §2.9.

**Step 3: Commit**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && git add CLAUDE.md docs/standard.md && git commit -m "docs: update CLAUDE.md and standard.md for Biome tooling"
```

---

### Task 8: Update Makefile lint target

**Files:**
- Modify: `Makefile` (lines 44-50)

**Step 1: Update Makefile**

The Makefile `lint` and `fmt` targets already delegate to `bun run lint` and `bun run format`, which now point to Biome. Verify this works:

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && make lint
```

Expected: golangci-lint runs for Go, Biome runs for web. Both pass.

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms && make fmt
```

Expected: gofmt runs for Go, Biome format runs for web. Both pass.

No Makefile changes needed — the indirection through `bun run` scripts handles the Prettier→Biome swap automatically.

---

### Task 9: Final Verification

**Step 1: Full clean build**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && rm -rf node_modules && bun install && bun run lint && bun run typecheck && bun run build
```

Expected: All pass with zero errors.

**Step 2: Verify package.json final state**

Confirm:
- `prettier` NOT in dependencies or devDependencies
- `@biomejs/biome` in devDependencies
- `@tanstack/react-query`, `zustand`, `react-i18next`, `i18next` in dependencies

**Step 3: Update project memory**

Update MEMORY.md to reflect:
- Biome replaces Prettier
- TanStack Query, Zustand, react-i18next installed
- Tailwind V4 confirmed
