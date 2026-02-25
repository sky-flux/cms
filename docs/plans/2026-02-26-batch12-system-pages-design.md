# Batch 12: System Management Pages — Design Document

> **Date**: 2026-02-26
> **Scope**: 9 modules — Users, Roles, Sites, Settings, API Keys, Audit, Comments, Menus, Redirects
> **Strategy**: 5 parallel Agents (TDD), Agent 1 first → Agents 2-5 concurrent

## Agent Decomposition

| Agent | Modules | React Components | Astro Pages | Test Files |
|-------|---------|-----------------|-------------|------------|
| 1 (infra) | system-api + i18n + PermissionTree | 1 | 0 | 1 |
| 2 (user-role) | Users + Roles | 6 | 3 | 2 |
| 3 (site-settings) | Sites + Settings + API Keys | 9 | 4 | 3 |
| 4 (comment-audit) | Comments + Audit | 5 | 2 | 2 |
| 5 (menu-redirect) | Menus + Redirects | 8 | 3 | 2 |
| **Total** | | **29** | **12** | **10** |

## Agent 1 — Infrastructure

### Files
- `web/src/lib/system-api.ts` — API wrappers for all 9 modules
- `web/src/i18n/locales/en.json` — ~120 new `system.*` keys
- `web/src/i18n/locales/zh-CN.json` — matching Chinese translations
- `web/src/components/shared/PermissionTree.tsx` — Checkbox tree for RBAC API/Menu assignment
- `web/src/components/shared/__tests__/PermissionTree.test.tsx`

### system-api.ts Structure
```
usersApi      — list / create / get / update / delete
rolesApi      — list / create / get / update / delete / getApis / setApis / getMenus / setMenus
templatesApi  — list / create / get / update / delete / apply
rbacApi       — listApis / getMyMenus
settingsApi   — get / update
apiKeysApi    — list / create / delete
auditApi      — list
commentsApi   — list / get / updateStatus / togglePin / reply / batchStatus / delete
menusApi      — list / create / get / update / delete / addItem / updateItem / deleteItem / reorderItems
redirectsApi  — list / create / update / delete / batchDelete / import / export
sitesApi      — list / create / get / update / delete / listUsers / assignRole / removeRole
```

### PermissionTree Component
- Tree structure with checkboxes (parent auto-select/deselect children)
- Tri-state: checked / unchecked / indeterminate
- Props: `items: TreeNode[]`, `checkedIds: string[]`, `onChange: (ids: string[]) => void`
- Used by Agent 2's RolePermissions component

## Agent 2 — Users + Roles

### Users Module (5 endpoints → 3 components + 1 page)
| Component | Purpose |
|-----------|---------|
| `UsersTable.tsx` | User list with role badge, status, last login. Filter: role select + search. Row actions: edit/disable/delete |
| `UserFormDialog.tsx` | Create/edit user dialog (RHF+Zod). Fields: email, display_name, role select, is_active toggle |
| `UsersPage.tsx` | Page container (useQuery + useMutation + providers) |
| `users/index.astro` | Astro route page |

### Roles Module (18 endpoints → 4 components + 2 pages)
| Component | Purpose |
|-----------|---------|
| `RolesTable.tsx` | Role list with built-in badge, description. Built-in roles non-deletable |
| `RoleFormDialog.tsx` | Create/edit role dialog (name, description) |
| `RolePermissions.tsx` | Permission editor with 2 tabs (API + Menu), uses PermissionTree. "Apply Template" button at top |
| `RolesPage.tsx` | List page container |
| `roles/index.astro` | Role list route |
| `roles/[id]/permissions.astro` | Permission editor route |

### Key Interactions
- RolePermissions loads: `rbacApi.listApis()` + `rolesApi.getApis(id)` + `rolesApi.getMenus(id)` in parallel
- "Apply Template" opens select dialog → one-click overwrites current permissions

## Agent 3 — Sites + Settings + API Keys

### Sites Module (8 endpoints → 4 components + 2 pages)
| Component | Purpose |
|-----------|---------|
| `SitesTable.tsx` | Site list (name, slug, domain, status badge, timezone). Row actions: edit/manage users/delete |
| `SiteFormDialog.tsx` | Create/edit site dialog (name, slug, domain, logo, description, timezone, is_active) |
| `SiteUsersDialog.tsx` | Site user management dialog (user list + role assignment + remove) |
| `SitesPage.tsx` | Page container |
| `sites/index.astro` | Site list route |
| `sites/[slug]/users.astro` | Site users route (or dialog-based) |

### Settings Module (2 endpoints → 2 components + 1 page)
| Component | Purpose |
|-----------|---------|
| `SettingsForm.tsx` | Settings form (site_name, logo, domain, timezone select, KV config items). RHF+Zod |
| `SettingsPage.tsx` | Page container |
| `settings/index.astro` | Route |

### API Keys Module (3 endpoints → 3 components + 1 page)
| Component | Purpose |
|-----------|---------|
| `ApiKeysTable.tsx` | Key list (name, prefix `sk-xxx...`, status, last_used, expires, rate_limit). Row action: revoke |
| `CreateApiKeyDialog.tsx` | Create dialog + **one-time full key display** with copy button (cannot view again after close) |
| `ApiKeysPage.tsx` | Page container |
| `api-keys/index.astro` | Route |

## Agent 4 — Comments + Audit

### Comments Module (7 endpoints → 3 components + 1 page)
| Component | Purpose |
|-----------|---------|
| `CommentsTable.tsx` | Comment list (author, excerpt, post, status badge, pin icon). Filter: status select + post_id + search. Batch: checkbox multi-select + floating action bar (approve/reject/spam) |
| `CommentDetailDialog.tsx` | Comment detail dialog (full content + reply tree + admin reply form) |
| `CommentsPage.tsx` | Page container |
| `comments/index.astro` | Route |

### Batch Operation UX
- Header checkbox selects all on current page
- Floating bottom bar shows selected count + action buttons
- Max 100 items per batch (API constraint)

### Audit Module (1 endpoint → 2 components + 1 page)
| Component | Purpose |
|-----------|---------|
| `AuditTable.tsx` | Read-only audit log list (actor, action type badge, resource type, resource ID, IP, timestamp). Filter: action type + resource type + date range (native `<input type="date">`) |
| `AuditPage.tsx` | Page container |
| `audit/index.astro` | Route |

## Agent 5 — Menus + Redirects

### Menus Module (9 endpoints → 4 components + 2 pages)
| Component | Purpose |
|-----------|---------|
| `MenusTable.tsx` | Menu list (name, slug, location badge, item count). Row actions: edit/manage items/delete |
| `MenuFormDialog.tsx` | Create/edit menu dialog (name, slug, location: header/footer/sidebar/custom, description) |
| `MenuItemsEditor.tsx` | **Core complex component**: 3-level tree with dnd-kit drag-drop reorder (reuses CategoryTree pattern). Add item dialog (5 types: custom/post/category/tag/page). Edit/delete per item |
| `MenusPage.tsx` | List page container |
| `menus/index.astro` | Menu list route |
| `menus/[id]/items.astro` | Menu items editor route |

### Redirects Module (7 endpoints → 4 components + 1 page)
| Component | Purpose |
|-----------|---------|
| `RedirectsTable.tsx` | Redirect list (source, target, status code badge 301/302, active toggle, hit count, last hit). Filter: status code + search. Batch delete with checkbox |
| `RedirectFormDialog.tsx` | Create/edit dialog (source_path validates `/` prefix, target_url, status_code select, is_active toggle) |
| `CsvImportDialog.tsx` | CSV import dialog (file select → preview first 10 rows → confirm → result stats: success/skipped/errors) |
| `RedirectsPage.tsx` | Page container (includes export button triggering download) |
| `redirects/index.astro` | Route |

## Shared Patterns (All Agents Follow)

### Component Pattern
- TanStack Table for all list views via shared `DataTable` component
- `usePagination()` + `useDebounce()` hooks for list pages
- `react-hook-form` + `zod` for all forms
- `Sonner` toast for success/error feedback
- `ConfirmDialog` for destructive operations
- `useTranslation()` with `system.*` namespace

### Page Container Pattern
```tsx
function ModulePageInner() {
  // State + queries + mutations + handlers
}
export function ModulePage() {
  return <QueryProvider><I18nProvider><ModulePageInner /></I18nProvider></QueryProvider>;
}
```

### Astro Page Pattern
```astro
---
import DashboardLayout from '@/layouts/DashboardLayout.astro';
import { ModulePage } from '@/components/system/ModulePage';
---
<DashboardLayout title="Module - Sky Flux CMS">
  <ModulePage client:load />
</DashboardLayout>
```

## File Structure

```
web/src/
├── lib/system-api.ts                          # Agent 1
├── components/
│   ├── shared/
│   │   ├── PermissionTree.tsx                 # Agent 1
│   │   └── __tests__/PermissionTree.test.tsx  # Agent 1
│   └── system/
│       ├── UsersTable.tsx                     # Agent 2
│       ├── UserFormDialog.tsx                 # Agent 2
│       ├── UsersPage.tsx                      # Agent 2
│       ├── RolesTable.tsx                     # Agent 2
│       ├── RoleFormDialog.tsx                 # Agent 2
│       ├── RolePermissions.tsx                # Agent 2
│       ├── RolesPage.tsx                      # Agent 2
│       ├── SitesTable.tsx                     # Agent 3
│       ├── SiteFormDialog.tsx                 # Agent 3
│       ├── SiteUsersDialog.tsx                # Agent 3
│       ├── SitesPage.tsx                      # Agent 3
│       ├── SettingsForm.tsx                   # Agent 3
│       ├── SettingsPage.tsx                   # Agent 3
│       ├── ApiKeysTable.tsx                   # Agent 3
│       ├── CreateApiKeyDialog.tsx             # Agent 3
│       ├── ApiKeysPage.tsx                    # Agent 3
│       ├── CommentsTable.tsx                  # Agent 4
│       ├── CommentDetailDialog.tsx            # Agent 4
│       ├── CommentsPage.tsx                   # Agent 4
│       ├── AuditTable.tsx                     # Agent 4
│       ├── AuditPage.tsx                      # Agent 4
│       ├── MenusTable.tsx                     # Agent 5
│       ├── MenuFormDialog.tsx                 # Agent 5
│       ├── MenuItemsEditor.tsx                # Agent 5
│       ├── MenusPage.tsx                      # Agent 5
│       ├── RedirectsTable.tsx                 # Agent 5
│       ├── RedirectFormDialog.tsx             # Agent 5
│       ├── CsvImportDialog.tsx                # Agent 5
│       ├── RedirectsPage.tsx                  # Agent 5
│       └── __tests__/
│           ├── Users.test.tsx                 # Agent 2
│           ├── Roles.test.tsx                 # Agent 2
│           ├── Sites.test.tsx                 # Agent 3
│           ├── Settings.test.tsx              # Agent 3
│           ├── ApiKeys.test.tsx               # Agent 3
│           ├── Comments.test.tsx              # Agent 4
│           ├── Audit.test.tsx                 # Agent 4
│           ├── Menus.test.tsx                 # Agent 5
│           └── Redirects.test.tsx             # Agent 5
├── pages/dashboard/
│   ├── users/index.astro                      # Agent 2
│   ├── roles/index.astro                      # Agent 2
│   ├── roles/[id]/permissions.astro           # Agent 2
│   ├── sites/index.astro                      # Agent 3
│   ├── sites/[slug]/users.astro               # Agent 3
│   ├── settings/index.astro                   # Agent 3
│   ├── api-keys/index.astro                   # Agent 3
│   ├── comments/index.astro                   # Agent 4
│   ├── audit/index.astro                      # Agent 4
│   ├── menus/index.astro                      # Agent 5
│   ├── menus/[id]/items.astro                 # Agent 5
│   └── redirects/index.astro                  # Agent 5
```

## Execution Order

1. **Phase 1**: Agent 1 (infra) runs first — creates system-api.ts, i18n keys, PermissionTree
2. **Phase 2**: Agents 2-5 run in parallel — each creates their components, tests, and pages
3. **Phase 3**: Integration testing — run full Vitest suite + `astro check`
4. **Phase 4**: Fix any TypeScript errors, commit

## Estimated Output
- ~29 React components
- ~12 Astro pages
- ~120 new i18n keys per locale
- ~200+ new tests
- 0 new npm dependencies (all using existing: dnd-kit, react-hook-form, zod, tanstack)
