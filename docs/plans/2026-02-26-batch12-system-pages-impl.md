# Batch 12: System Management Pages Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build all system management frontend pages (users, roles, sites, settings, API keys, audit, comments, menus, redirects) with full TDD coverage.

**Architecture:** Astro 5 SSR pages wrapping React 19 Islands via `client:load`. DashboardLayout provides sidebar+header shell. TanStack Query manages server state. All API calls routed through typed wrapper in `system-api.ts`. Components live under `components/system/`, sharing DataTable/ConfirmDialog/StatusBadge from `components/shared/`.

**Tech Stack:** Astro 5 + React 19 + @tanstack/react-table + @dnd-kit (menu reorder) + react-hook-form + zod + shadcn/ui + TanStack Query v5 + Vitest + RTL

---

## Agent Division

| Agent | Tasks | Dependencies |
|-------|-------|-------------|
| Agent 1 (infra) | Tasks 1–4 | None |
| Agent 2 (user-role) | Tasks 5–8 | Agent 1 done |
| Agent 3 (site-settings) | Tasks 9–13 | Agent 1 done |
| Agent 4 (comment-audit) | Tasks 14–16 | Agent 1 done |
| Agent 5 (menu-redirect) | Tasks 17–21 | Agent 1 done |

---

## Agent 1: Shared Infrastructure

### Task 1: Create system-api.ts with TypeScript interfaces and API wrappers

**Files:**
- Create: `web/src/lib/system-api.ts`

**Step 1: Create system-api.ts**

Create `web/src/lib/system-api.ts` with all interfaces and API wrappers. Follow the exact pattern from `content-api.ts`. Import `api`, `requestFormData`, and `RequestOptions` from `./api-client`. Reuse `PaginatedResponse`, `ApiResponse`, `PaginationMeta` from `./content-api` (re-export them).

```typescript
import { api, requestFormData, type RequestOptions } from './api-client';

// Re-export shared types from content-api
export { type PaginatedResponse, type ApiResponse, type PaginationMeta } from './content-api';
import type { PaginatedResponse, ApiResponse } from './content-api';

// --- Helpers ---
function buildQuery(params: Record<string, unknown> | object): string {
  const p = params as Record<string, unknown>;
  const entries = Object.entries(p).filter(([, v]) => v !== undefined && v !== null && v !== '');
  if (entries.length === 0) return '';
  return '?' + entries.map(([k, v]) => `${k}=${encodeURIComponent(String(v))}`).join('&');
}

// --- Users ---
export interface User {
  id: string;
  email: string;
  display_name: string;
  role: string;
  is_active: boolean;
  avatar_url: string | null;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface UserListParams {
  page: number;
  per_page: number;
  role?: string;
  q?: string;
}

export interface CreateUserDTO {
  email: string;
  password: string;
  display_name: string;
  role: string;
}

export interface UpdateUserDTO {
  display_name?: string;
  role?: string;
  is_active?: boolean;
}

export const usersApi = {
  list: (params: UserListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<User>>(`/api/v1/users${buildQuery(params)}`, opts),
  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<User>>(`/api/v1/users/${id}`, opts),
  create: (data: CreateUserDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<User>>('/api/v1/users', data, opts),
  update: (id: string, data: UpdateUserDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<User>>(`/api/v1/users/${id}`, data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/users/${id}`, opts),
};

// --- Roles ---
export interface Role {
  id: string;
  name: string;
  slug: string;
  description: string;
  built_in: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateRoleDTO {
  name: string;
  slug: string;
  description?: string;
}

export interface UpdateRoleDTO {
  name?: string;
  description?: string;
}

export interface ApiEndpoint {
  id: string;
  method: string;
  path: string;
  description: string;
}

export interface AdminMenu {
  id: string;
  name: string;
  path: string;
  icon: string;
  parent_id: string | null;
  sort_order: number;
  children: AdminMenu[];
}

export const rolesApi = {
  list: (opts?: RequestOptions) =>
    api.get<ApiResponse<Role[]>>('/api/v1/rbac/roles', opts),
  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Role>>(`/api/v1/rbac/roles/${id}`, opts),
  create: (data: CreateRoleDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<Role>>('/api/v1/rbac/roles', data, opts),
  update: (id: string, data: UpdateRoleDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<Role>>(`/api/v1/rbac/roles/${id}`, data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/rbac/roles/${id}`, opts),
  getApis: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<string[]>>(`/api/v1/rbac/roles/${id}/apis`, opts),
  setApis: (id: string, apiIds: string[], opts?: RequestOptions) =>
    api.put<{ success: boolean }>(`/api/v1/rbac/roles/${id}/apis`, { api_ids: apiIds }, opts),
  getMenus: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<string[]>>(`/api/v1/rbac/roles/${id}/menus`, opts),
  setMenus: (id: string, menuIds: string[], opts?: RequestOptions) =>
    api.put<{ success: boolean }>(`/api/v1/rbac/roles/${id}/menus`, { menu_ids: menuIds }, opts),
};

// --- Templates ---
export interface RoleTemplate {
  id: string;
  name: string;
  description: string;
  created_at: string;
}

export const templatesApi = {
  list: (opts?: RequestOptions) =>
    api.get<ApiResponse<RoleTemplate[]>>('/api/v1/rbac/templates', opts),
  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<RoleTemplate>>(`/api/v1/rbac/templates/${id}`, opts),
  create: (data: { name: string; description?: string }, opts?: RequestOptions) =>
    api.post<ApiResponse<RoleTemplate>>('/api/v1/rbac/templates', data, opts),
  update: (id: string, data: { name?: string; description?: string }, opts?: RequestOptions) =>
    api.put<ApiResponse<RoleTemplate>>(`/api/v1/rbac/templates/${id}`, data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/rbac/templates/${id}`, opts),
  apply: (roleId: string, templateId: string, opts?: RequestOptions) =>
    api.post<{ success: boolean }>(`/api/v1/rbac/roles/${roleId}/apply-template`, { template_id: templateId }, opts),
};

// --- RBAC Helpers ---
export const rbacApi = {
  listApis: (opts?: RequestOptions) =>
    api.get<ApiResponse<ApiEndpoint[]>>('/api/v1/rbac/apis', opts),
  getMyMenus: (opts?: RequestOptions) =>
    api.get<ApiResponse<AdminMenu[]>>('/api/v1/rbac/me/menus', opts),
  listAdminMenus: (opts?: RequestOptions) =>
    api.get<ApiResponse<AdminMenu[]>>('/api/v1/rbac/menus', opts),
};

// --- Sites ---
export interface Site {
  id: string;
  name: string;
  slug: string;
  domain: string | null;
  description: string | null;
  logo_url: string | null;
  default_locale: string;
  timezone: string;
  is_active: boolean;
  settings: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface SiteListParams {
  page: number;
  per_page: number;
  q?: string;
  is_active?: boolean;
}

export interface CreateSiteDTO {
  name: string;
  slug: string;
  domain?: string;
  description?: string;
  default_locale?: string;
  timezone?: string;
}

export interface UpdateSiteDTO {
  name?: string;
  domain?: string;
  description?: string;
  logo_url?: string;
  default_locale?: string;
  timezone?: string;
  is_active?: boolean;
}

export interface SiteUser {
  user: {
    id: string;
    email: string;
    display_name: string;
    avatar_url: string | null;
    is_active: boolean;
  };
  role: string;
  created_at: string;
}

export interface SiteUserListParams {
  page: number;
  per_page: number;
  role?: string;
  q?: string;
}

export const sitesApi = {
  list: (params: SiteListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<Site>>(`/api/v1/sites${buildQuery(params)}`, opts),
  get: (slug: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Site>>(`/api/v1/sites/${slug}`, opts),
  create: (data: CreateSiteDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<Site>>('/api/v1/sites', data, opts),
  update: (slug: string, data: UpdateSiteDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<Site>>(`/api/v1/sites/${slug}`, data, opts),
  delete: (slug: string, confirmSlug: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean; message: string }>(`/api/v1/sites/${slug}`, { ...opts, headers: { ...opts?.headers } }),
  deleteSite: (slug: string, confirmSlug: string, opts?: RequestOptions) =>
    api.post<{ success: boolean }>(`/api/v1/sites/${slug}/delete`, { confirm_slug: confirmSlug }, opts),
  listUsers: (slug: string, params: SiteUserListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<SiteUser>>(`/api/v1/sites/${slug}/users${buildQuery(params)}`, opts),
  assignRole: (slug: string, userId: string, role: string, opts?: RequestOptions) =>
    api.put<ApiResponse<{ user_id: string; site_slug: string; role: string }>>(`/api/v1/sites/${slug}/users/${userId}/role`, { role }, opts),
  removeRole: (slug: string, userId: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/sites/${slug}/users/${userId}/role`, opts),
};

// --- Settings ---
export interface SettingItem {
  key: string;
  value: string;
  description: string;
}

export const settingsApi = {
  get: (opts?: RequestOptions) =>
    api.get<ApiResponse<SettingItem[]>>('/api/v1/site/settings', opts),
  update: (key: string, value: string, opts?: RequestOptions) =>
    api.put<ApiResponse<SettingItem>>('/api/v1/site/settings', { key, value }, opts),
};

// --- API Keys ---
export interface ApiKey {
  id: string;
  name: string;
  key_prefix: string;
  is_active: boolean;
  last_used_at: string | null;
  expires_at: string | null;
  rate_limit: number;
  created_at: string;
}

export interface CreateApiKeyDTO {
  name: string;
  expires_at?: string | null;
  rate_limit?: number;
}

export interface CreateApiKeyResponse {
  id: string;
  name: string;
  key: string;
  key_prefix: string;
  expires_at: string | null;
  rate_limit: number;
  created_at: string;
}

export const apiKeysApi = {
  list: (opts?: RequestOptions) =>
    api.get<ApiResponse<ApiKey[]>>('/api/v1/site/api-keys', opts),
  create: (data: CreateApiKeyDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<CreateApiKeyResponse>>('/api/v1/site/api-keys', data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/site/api-keys/${id}`, opts),
};

// --- Audit ---
export interface AuditLog {
  id: string;
  actor: { id: string; display_name: string };
  action: string;
  resource_type: string;
  resource_id: string;
  resource_snapshot: Record<string, unknown> | null;
  ip_address: string;
  created_at: string;
}

export interface AuditListParams {
  page: number;
  per_page: number;
  actor_id?: string;
  action?: string;
  resource_type?: string;
  start_date?: string;
  end_date?: string;
}

export const auditApi = {
  list: (params: AuditListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<AuditLog>>(`/api/v1/site/audit-logs${buildQuery(params)}`, opts),
};

// --- Comments ---
export interface Comment {
  id: string;
  post: { id: string; title: string; slug: string };
  parent_id: string | null;
  user_id: string | null;
  author_name: string;
  author_email: string;
  author_url: string | null;
  author_ip: string;
  gravatar_url: string;
  content: string;
  status: string;
  is_pinned: boolean;
  replies?: Comment[];
  created_at: string;
  updated_at: string;
}

export interface CommentListParams {
  page: number;
  per_page: number;
  post_id?: string;
  status?: string;
  q?: string;
  sort?: string;
}

export const commentsApi = {
  list: (params: CommentListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<Comment>>(`/api/v1/site/comments${buildQuery(params)}`, opts),
  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Comment>>(`/api/v1/site/comments/${id}`, opts),
  updateStatus: (id: string, status: string, opts?: RequestOptions) =>
    api.put<ApiResponse<{ id: string; status: string }>>(`/api/v1/site/comments/${id}/status`, { status }, opts),
  togglePin: (id: string, isPinned: boolean, opts?: RequestOptions) =>
    api.put<ApiResponse<{ id: string; is_pinned: boolean }>>(`/api/v1/site/comments/${id}/pin`, { is_pinned: isPinned }, opts),
  reply: (id: string, content: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Comment>>(`/api/v1/site/comments/${id}/reply`, { content }, opts),
  batchStatus: (commentIds: string[], status: string, opts?: RequestOptions) =>
    api.put<ApiResponse<{ updated_count: number }>>('/api/v1/site/comments/batch-status', { comment_ids: commentIds, status }, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/site/comments/${id}`, opts),
};

// --- Menus (site-scoped navigation menus) ---
export interface SiteMenu {
  id: string;
  name: string;
  slug: string;
  location: string;
  description: string | null;
  item_count: number;
  created_at: string;
  updated_at: string;
}

export interface SiteMenuItem {
  id: string;
  parent_id: string | null;
  label: string;
  url: string | null;
  target: string;
  icon: string | null;
  css_class: string | null;
  type: string;
  reference_id: string | null;
  sort_order: number;
  is_active: boolean;
  is_broken: boolean;
  children: SiteMenuItem[];
}

export interface SiteMenuDetail extends SiteMenu {
  items: SiteMenuItem[];
}

export interface CreateSiteMenuDTO {
  name: string;
  slug: string;
  location?: string;
  description?: string;
}

export interface UpdateSiteMenuDTO {
  name?: string;
  slug?: string;
  location?: string;
  description?: string;
}

export interface CreateMenuItemDTO {
  parent_id?: string | null;
  label: string;
  url?: string | null;
  target?: string;
  icon?: string | null;
  css_class?: string | null;
  type: string;
  reference_id?: string | null;
  sort_order: number;
  is_active?: boolean;
}

export interface UpdateMenuItemDTO extends Partial<CreateMenuItemDTO> {}

export interface ReorderMenuItemDTO {
  id: string;
  parent_id: string | null;
  sort_order: number;
}

export const siteMenusApi = {
  list: (params?: { location?: string }, opts?: RequestOptions) =>
    api.get<ApiResponse<SiteMenu[]>>(`/api/v1/site/menus${params ? buildQuery(params) : ''}`, opts),
  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<SiteMenuDetail>>(`/api/v1/site/menus/${id}`, opts),
  create: (data: CreateSiteMenuDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<SiteMenu>>('/api/v1/site/menus', data, opts),
  update: (id: string, data: UpdateSiteMenuDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<SiteMenu>>(`/api/v1/site/menus/${id}`, data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/site/menus/${id}`, opts),
  addItem: (menuId: string, data: CreateMenuItemDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<SiteMenuItem>>(`/api/v1/site/menus/${menuId}/items`, data, opts),
  updateItem: (menuId: string, itemId: string, data: UpdateMenuItemDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<SiteMenuItem>>(`/api/v1/site/menus/${menuId}/items/${itemId}`, data, opts),
  deleteItem: (menuId: string, itemId: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/site/menus/${menuId}/items/${itemId}`, opts),
  reorderItems: (menuId: string, items: ReorderMenuItemDTO[], opts?: RequestOptions) =>
    api.put<{ success: boolean }>(`/api/v1/site/menus/${menuId}/items/reorder`, { items }, opts),
};

// --- Redirects ---
export interface Redirect {
  id: string;
  source_path: string;
  target_url: string;
  status_code: number;
  is_active: boolean;
  hit_count: number;
  last_hit_at: string | null;
  created_by: { id: string; display_name: string } | null;
  created_at: string;
  updated_at: string;
}

export interface RedirectListParams {
  page: number;
  per_page: number;
  q?: string;
  status_code?: number;
  is_active?: boolean;
  sort?: string;
}

export interface CreateRedirectDTO {
  source_path: string;
  target_url: string;
  status_code?: number;
  is_active?: boolean;
}

export interface UpdateRedirectDTO extends Partial<CreateRedirectDTO> {}

export interface CsvImportResult {
  imported: number;
  skipped: number;
  errors: { row: number; source_path: string; error: string }[];
}

export const redirectsApi = {
  list: (params: RedirectListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<Redirect>>(`/api/v1/site/redirects${buildQuery(params)}`, opts),
  create: (data: CreateRedirectDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<Redirect>>('/api/v1/site/redirects', data, opts),
  update: (id: string, data: UpdateRedirectDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<Redirect>>(`/api/v1/site/redirects/${id}`, data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/site/redirects/${id}`, opts),
  batchDelete: (ids: string[], opts?: RequestOptions) =>
    api.post<ApiResponse<{ deleted_count: number }>>('/api/v1/site/redirects/batch', { ids }, opts),
  import: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return requestFormData<ApiResponse<CsvImportResult>>('POST', '/api/v1/site/redirects/import', formData);
  },
  export: (opts?: RequestOptions) =>
    api.get<Blob>('/api/v1/site/redirects/export', opts),
};
```

**Step 2: Commit**

```bash
git add web/src/lib/system-api.ts
git commit -m "feat(web): add system-api.ts with all admin module API wrappers"
```

---

### Task 2: Add i18n keys for all system modules

**Files:**
- Modify: `web/src/i18n/locales/en.json`
- Modify: `web/src/i18n/locales/zh-CN.json`

**Step 1: Add `system` section to en.json**

Add a new top-level `"system"` key to `en.json` with all keys needed by Agents 2-5. Key naming follows the pattern `system.<module>.<field>`. Include these subsections:

```json
{
  "system": {
    "users": {
      "title": "Users",
      "newUser": "New User",
      "editUser": "Edit User",
      "email": "Email",
      "displayName": "Display Name",
      "role": "Role",
      "status": "Status",
      "active": "Active",
      "disabled": "Disabled",
      "lastLogin": "Last Login",
      "createdAt": "Created",
      "noUsersFound": "No users found",
      "deleteUserConfirm": "Delete user \"{{name}}\"? This action cannot be undone.",
      "disableUserConfirm": "Disable user \"{{name}}\"? They will lose access.",
      "searchPlaceholder": "Search users...",
      "filterByRole": "Filter by role",
      "password": "Password",
      "passwordHelp": "Minimum 8 characters. User will receive this via email."
    },
    "roles": {
      "title": "Roles",
      "newRole": "New Role",
      "editRole": "Edit Role",
      "roleName": "Role Name",
      "roleSlug": "Slug",
      "description": "Description",
      "builtIn": "Built-in",
      "custom": "Custom",
      "permissions": "Permissions",
      "apiPermissions": "API Permissions",
      "menuPermissions": "Menu Permissions",
      "applyTemplate": "Apply Template",
      "selectTemplate": "Select a template to apply",
      "applyTemplateConfirm": "Apply template \"{{name}}\"? This will overwrite current permissions.",
      "noRolesFound": "No roles found",
      "deleteRoleConfirm": "Delete role \"{{name}}\"?",
      "cannotDeleteBuiltIn": "Built-in roles cannot be deleted.",
      "selectAll": "Select All",
      "deselectAll": "Deselect All"
    },
    "sites": {
      "title": "Sites",
      "newSite": "New Site",
      "editSite": "Edit Site",
      "siteName": "Site Name",
      "slug": "Slug",
      "domain": "Domain",
      "description": "Description",
      "logo": "Logo URL",
      "locale": "Default Locale",
      "timezone": "Timezone",
      "status": "Status",
      "active": "Active",
      "inactive": "Inactive",
      "manageUsers": "Manage Users",
      "siteUsers": "Site Users",
      "assignUser": "Assign User",
      "removeUser": "Remove User",
      "removeUserConfirm": "Remove \"{{name}}\" from this site?",
      "deleteSiteConfirm": "Delete site \"{{name}}\"? This will permanently destroy all content. Type the slug to confirm:",
      "confirmSlug": "Confirm slug",
      "noSitesFound": "No sites found",
      "cannotDeleteLast": "Cannot delete the last site.",
      "searchPlaceholder": "Search sites..."
    },
    "settings": {
      "title": "Settings",
      "saveSettings": "Save Settings",
      "settingKey": "Setting",
      "settingValue": "Value",
      "settingDescription": "Description",
      "noSettings": "No settings configured"
    },
    "apiKeys": {
      "title": "API Keys",
      "newKey": "New API Key",
      "keyName": "Key Name",
      "keyPrefix": "Key Prefix",
      "status": "Status",
      "active": "Active",
      "revoked": "Revoked",
      "lastUsed": "Last Used",
      "expiresAt": "Expires",
      "rateLimit": "Rate Limit",
      "rateLimitHelp": "Requests per minute (default: 100)",
      "revokeKey": "Revoke",
      "revokeConfirm": "Revoke API key \"{{name}}\"? This action cannot be undone.",
      "keyCreated": "API Key Created",
      "keyCreatedDescription": "Copy this key now. You won't be able to see it again.",
      "copyKey": "Copy Key",
      "keyCopied": "Key copied to clipboard",
      "never": "Never",
      "noExpiry": "No expiry",
      "noKeysFound": "No API keys found"
    },
    "audit": {
      "title": "Audit Logs",
      "actor": "Actor",
      "action": "Action",
      "resourceType": "Resource Type",
      "resourceId": "Resource ID",
      "ipAddress": "IP Address",
      "timestamp": "Timestamp",
      "filterByAction": "Filter by action",
      "filterByResource": "Filter by resource type",
      "startDate": "Start Date",
      "endDate": "End Date",
      "noLogsFound": "No audit logs found",
      "actions": {
        "create": "Create",
        "update": "Update",
        "delete": "Delete",
        "login": "Login"
      },
      "resources": {
        "post": "Post",
        "user": "User",
        "setting": "Setting",
        "comment": "Comment",
        "media": "Media",
        "menu": "Menu",
        "redirect": "Redirect"
      }
    },
    "comments": {
      "title": "Comments",
      "author": "Author",
      "content": "Content",
      "post": "Post",
      "status": "Status",
      "pinned": "Pinned",
      "pending": "Pending",
      "approved": "Approved",
      "spam": "Spam",
      "trash": "Trash",
      "approve": "Approve",
      "reject": "Reject",
      "markSpam": "Mark as Spam",
      "moveToTrash": "Move to Trash",
      "pin": "Pin",
      "unpin": "Unpin",
      "reply": "Reply",
      "adminReply": "Admin Reply",
      "replyPlaceholder": "Write your reply...",
      "commentDetail": "Comment Detail",
      "replies": "Replies",
      "batchApprove": "Approve Selected",
      "batchReject": "Reject Selected",
      "batchSpam": "Mark as Spam",
      "deleteCommentConfirm": "Delete this comment and all replies?",
      "noCommentsFound": "No comments found",
      "filterByStatus": "Filter by status",
      "filterByPost": "Filter by post",
      "searchPlaceholder": "Search comments...",
      "selected": "{{count}} selected",
      "maxPinsReached": "Maximum 3 pinned comments per post"
    },
    "menus": {
      "title": "Menus",
      "newMenu": "New Menu",
      "editMenu": "Edit Menu",
      "menuName": "Menu Name",
      "slug": "Slug",
      "location": "Location",
      "description": "Description",
      "items": "Menu Items",
      "itemCount": "{{count}} items",
      "manageItems": "Manage Items",
      "addItem": "Add Item",
      "editItem": "Edit Item",
      "itemLabel": "Label",
      "itemUrl": "URL",
      "itemTarget": "Open in",
      "targetSelf": "Same window",
      "targetBlank": "New window",
      "itemType": "Type",
      "typeCustom": "Custom URL",
      "typePost": "Post",
      "typeCategory": "Category",
      "typeTag": "Tag",
      "typePage": "Page",
      "referenceId": "Reference",
      "itemIcon": "Icon",
      "itemCssClass": "CSS Class",
      "itemActive": "Active",
      "broken": "Broken Reference",
      "deleteMenuConfirm": "Delete menu \"{{name}}\" and all its items?",
      "deleteItemConfirm": "Delete this item and its children?",
      "noMenusFound": "No menus found",
      "noItems": "No items in this menu",
      "dragToReorder": "Drag to reorder",
      "locationHeader": "Header",
      "locationFooter": "Footer",
      "locationSidebar": "Sidebar",
      "locationCustom": "Custom"
    },
    "redirects": {
      "title": "Redirects",
      "newRedirect": "New Redirect",
      "editRedirect": "Edit Redirect",
      "sourcePath": "Source Path",
      "targetUrl": "Target URL",
      "statusCode": "Status Code",
      "active": "Active",
      "hitCount": "Hits",
      "lastHit": "Last Hit",
      "createdBy": "Created By",
      "permanent301": "301 Permanent",
      "temporary302": "302 Temporary",
      "deleteRedirectConfirm": "Delete this redirect?",
      "batchDeleteConfirm": "Delete {{count}} selected redirects?",
      "importCsv": "Import CSV",
      "exportCsv": "Export CSV",
      "csvFile": "CSV File",
      "csvPreview": "Preview",
      "csvImportResult": "Import complete: {{imported}} imported, {{skipped}} skipped, {{errors}} errors",
      "csvFormat": "CSV format: source_path, target_url, status_code (optional)",
      "sourcePathHelp": "Must start with /",
      "noRedirectsFound": "No redirects found",
      "searchPlaceholder": "Search redirects...",
      "filterByStatusCode": "Filter by status code",
      "selected": "{{count}} selected"
    }
  }
}
```

**Step 2: Add matching Chinese translations to zh-CN.json**

Add corresponding `"system"` section with Chinese translations. Follow the same structure. Key translations:
- users.title: "用户管理"
- roles.title: "角色管理"
- sites.title: "站点管理"
- settings.title: "系统设置"
- apiKeys.title: "API 密钥"
- audit.title: "审计日志"
- comments.title: "评论管理"
- menus.title: "菜单管理"
- redirects.title: "重定向管理"

**Step 3: Commit**

```bash
git add web/src/i18n/locales/en.json web/src/i18n/locales/zh-CN.json
git commit -m "feat(web): add system management i18n keys for all 9 modules"
```

---

### Task 3: Create PermissionTree shared component (TDD)

**Files:**
- Create: `web/src/components/shared/__tests__/PermissionTree.test.tsx`
- Create: `web/src/components/shared/PermissionTree.tsx`

**Step 1: Write PermissionTree tests**

Create `web/src/components/shared/__tests__/PermissionTree.test.tsx`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PermissionTree } from '../PermissionTree';
import type { TreeNode } from '../PermissionTree';

const mockItems: TreeNode[] = [
  {
    id: '1',
    label: 'Posts',
    children: [
      { id: '1-1', label: 'Create Post', children: [] },
      { id: '1-2', label: 'Edit Post', children: [] },
      { id: '1-3', label: 'Delete Post', children: [] },
    ],
  },
  {
    id: '2',
    label: 'Users',
    children: [
      { id: '2-1', label: 'Create User', children: [] },
      { id: '2-2', label: 'Edit User', children: [] },
    ],
  },
  { id: '3', label: 'Settings', children: [] },
];

describe('PermissionTree', () => {
  const defaultProps = {
    items: mockItems,
    checkedIds: [] as string[],
    onChange: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders all top-level items', () => {
    render(<PermissionTree {...defaultProps} />);
    expect(screen.getByText('Posts')).toBeInTheDocument();
    expect(screen.getByText('Users')).toBeInTheDocument();
    expect(screen.getByText('Settings')).toBeInTheDocument();
  });

  it('renders child items', () => {
    render(<PermissionTree {...defaultProps} />);
    expect(screen.getByText('Create Post')).toBeInTheDocument();
    expect(screen.getByText('Edit Post')).toBeInTheDocument();
  });

  it('checks items that are in checkedIds', () => {
    render(<PermissionTree {...defaultProps} checkedIds={['1-1', '3']} />);
    const checkboxes = screen.getAllByRole('checkbox');
    // 1-1 (Create Post) and 3 (Settings) should be checked
    const createPostCb = checkboxes.find(
      (cb) => cb.closest('label')?.textContent?.includes('Create Post'),
    );
    expect(createPostCb).toBeTruthy();
  });

  it('calls onChange when a leaf node is toggled', async () => {
    const user = userEvent.setup();
    render(<PermissionTree {...defaultProps} checkedIds={[]} />);
    // Click "Settings" checkbox (leaf node with id '3')
    const settingsCb = screen.getByText('Settings').closest('label')?.querySelector('button') ||
      screen.getByText('Settings').parentElement?.querySelector('[role="checkbox"]');
    if (settingsCb) {
      await user.click(settingsCb);
      expect(defaultProps.onChange).toHaveBeenCalledWith(['3']);
    }
  });

  it('selects all children when parent is checked', async () => {
    const user = userEvent.setup();
    render(<PermissionTree {...defaultProps} checkedIds={[]} />);
    const postsCb = screen.getByText('Posts').parentElement?.querySelector('[role="checkbox"]');
    if (postsCb) {
      await user.click(postsCb);
      expect(defaultProps.onChange).toHaveBeenCalledWith(
        expect.arrayContaining(['1', '1-1', '1-2', '1-3']),
      );
    }
  });

  it('deselects all children when parent is unchecked', async () => {
    const user = userEvent.setup();
    render(
      <PermissionTree {...defaultProps} checkedIds={['1', '1-1', '1-2', '1-3']} />,
    );
    const postsCb = screen.getByText('Posts').parentElement?.querySelector('[role="checkbox"]');
    if (postsCb) {
      await user.click(postsCb);
      expect(defaultProps.onChange).toHaveBeenCalledWith([]);
    }
  });

  it('renders empty message when no items', () => {
    render(<PermissionTree {...defaultProps} items={[]} />);
    expect(screen.getByText('No permissions available')).toBeInTheDocument();
  });
});
```

**Step 2: Run test — verify it fails**

```bash
cd web && bun run vitest run src/components/shared/__tests__/PermissionTree.test.tsx
```

Expected: FAIL — PermissionTree module not found.

**Step 3: Implement PermissionTree component**

Create `web/src/components/shared/PermissionTree.tsx`:

```tsx
import { useCallback, useMemo } from 'react';
import { Checkbox } from '@/components/ui/checkbox';

export interface TreeNode {
  id: string;
  label: string;
  children: TreeNode[];
}

interface PermissionTreeProps {
  items: TreeNode[];
  checkedIds: string[];
  onChange: (ids: string[]) => void;
}

function getAllDescendantIds(node: TreeNode): string[] {
  const ids: string[] = [node.id];
  for (const child of node.children) {
    ids.push(...getAllDescendantIds(child));
  }
  return ids;
}

function getCheckState(
  node: TreeNode,
  checkedSet: Set<string>,
): 'checked' | 'unchecked' | 'indeterminate' {
  if (node.children.length === 0) {
    return checkedSet.has(node.id) ? 'checked' : 'unchecked';
  }
  const childStates = node.children.map((c) => getCheckState(c, checkedSet));
  if (childStates.every((s) => s === 'checked') && checkedSet.has(node.id)) return 'checked';
  if (childStates.some((s) => s === 'checked' || s === 'indeterminate')) return 'indeterminate';
  return 'unchecked';
}

function TreeNodeRow({
  node,
  checkedSet,
  onToggle,
  depth = 0,
}: {
  node: TreeNode;
  checkedSet: Set<string>;
  onToggle: (node: TreeNode) => void;
  depth?: number;
}) {
  const state = getCheckState(node, checkedSet);

  return (
    <div>
      <label
        className="flex items-center gap-2 py-1 hover:bg-muted/50 rounded px-2 cursor-pointer"
        style={{ paddingLeft: `${depth * 24 + 8}px` }}
      >
        <Checkbox
          checked={state === 'checked' ? true : state === 'indeterminate' ? 'indeterminate' : false}
          onCheckedChange={() => onToggle(node)}
        />
        <span className="text-sm">{node.label}</span>
      </label>
      {node.children.map((child) => (
        <TreeNodeRow
          key={child.id}
          node={child}
          checkedSet={checkedSet}
          onToggle={onToggle}
          depth={depth + 1}
        />
      ))}
    </div>
  );
}

export function PermissionTree({ items, checkedIds, onChange }: PermissionTreeProps) {
  const checkedSet = useMemo(() => new Set(checkedIds), [checkedIds]);

  const handleToggle = useCallback(
    (node: TreeNode) => {
      const allIds = getAllDescendantIds(node);
      const currentState = getCheckState(node, checkedSet);
      const newSet = new Set(checkedIds);

      if (currentState === 'checked') {
        // Uncheck all descendants
        for (const id of allIds) newSet.delete(id);
      } else {
        // Check all descendants
        for (const id of allIds) newSet.add(id);
      }

      onChange(Array.from(newSet));
    },
    [checkedIds, checkedSet, onChange],
  );

  if (items.length === 0) {
    return <p className="text-sm text-muted-foreground py-4">No permissions available</p>;
  }

  return (
    <div className="space-y-0.5">
      {items.map((item) => (
        <TreeNodeRow
          key={item.id}
          node={item}
          checkedSet={checkedSet}
          onToggle={handleToggle}
        />
      ))}
    </div>
  );
}
```

**Step 4: Run test — verify it passes**

```bash
cd web && bun run vitest run src/components/shared/__tests__/PermissionTree.test.tsx
```

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/shared/PermissionTree.tsx web/src/components/shared/__tests__/PermissionTree.test.tsx
git commit -m "feat(web): add PermissionTree shared component with tri-state checkboxes"
```

---

### Task 4: Run full test suite to verify infrastructure

**Step 1: Run all tests**

```bash
cd web && bun run vitest run
```

Expected: All existing tests pass + new PermissionTree tests pass.

**Step 2: Run astro check**

```bash
cd web && bun run astro check
```

Expected: 0 errors.

---

## Agent 2: Users + Roles

### Task 5: Users module (TDD)

**Files:**
- Create: `web/src/components/system/__tests__/Users.test.tsx`
- Create: `web/src/components/system/UsersTable.tsx`
- Create: `web/src/components/system/UserFormDialog.tsx`
- Create: `web/src/components/system/UsersPage.tsx`
- Create: `web/src/pages/dashboard/users/index.astro`

**Step 1: Write Users tests**

Create `web/src/components/system/__tests__/Users.test.tsx`. Follow PostsTable test pattern. Test:
- UsersTable renders column headers (email, name, role, status, last login, actions)
- UsersTable renders user data rows
- UsersTable shows role badges
- UsersTable shows active/disabled status badges
- UsersTable filter bar has role select + search input + "New User" button
- UsersTable calls onSearch, onRoleFilter, onDelete, onEdit, onNewUser callbacks
- UsersTable shows empty state when no users
- UsersTable shows loading skeletons
- UserFormDialog renders create mode with empty fields
- UserFormDialog renders edit mode with pre-filled data
- UserFormDialog validates required fields (email, display_name, password for create)
- UserFormDialog calls onSubmit with form data

Mock `react-i18next` the same way as PostsTable tests. Mock `@/lib/system-api` for types only.

**Step 2: Run tests — verify fail**

```bash
cd web && bun run vitest run src/components/system/__tests__/Users.test.tsx
```

**Step 3: Implement UsersTable**

Create `web/src/components/system/UsersTable.tsx` following PostsTable pattern:
- Props: `users: User[]`, `pagination: PaginationMeta`, `loading: boolean`, `onPageChange`, `onRoleFilter`, `onSearch`, `onEdit`, `onDelete`, `onNewUser`
- Columns: email (linked), display_name, role (StatusBadge), status (active/disabled badge), last_login (formatted date), actions (dropdown: edit/delete)
- Filter bar: role Select + search Input + "New User" Button

**Step 4: Implement UserFormDialog**

Create `web/src/components/system/UserFormDialog.tsx`:
- Props: `open`, `onOpenChange`, `user?: User` (edit mode if provided), `roles: Role[]`, `onSubmit`
- Uses Dialog from shadcn/ui, react-hook-form + zod validation
- Create mode: email + display_name + password + role select
- Edit mode: display_name + role select + is_active toggle (no email/password change)

**Step 5: Implement UsersPage**

Create `web/src/components/system/UsersPage.tsx` following PostsListPage pattern:
- UsersPageInner: state + queries (usersApi.list, rolesApi.list for filter) + mutations (create, update, delete)
- Export UsersPage wrapping with QueryProvider + I18nProvider

**Step 6: Create Astro page**

Create `web/src/pages/dashboard/users/index.astro`:
```astro
---
import DashboardLayout from '@/layouts/DashboardLayout.astro';
import { UsersPage } from '@/components/system/UsersPage';
---
<DashboardLayout title="Users - Sky Flux CMS">
  <UsersPage client:load />
</DashboardLayout>
```

**Step 7: Run tests — verify pass**

```bash
cd web && bun run vitest run src/components/system/__tests__/Users.test.tsx
```

**Step 8: Commit**

```bash
git add web/src/components/system/UsersTable.tsx web/src/components/system/UserFormDialog.tsx web/src/components/system/UsersPage.tsx web/src/components/system/__tests__/Users.test.tsx web/src/pages/dashboard/users/index.astro
git commit -m "feat(web): add users management page with table, form dialog, and tests"
```

---

### Task 6: Roles list module (TDD)

**Files:**
- Create: `web/src/components/system/__tests__/Roles.test.tsx`
- Create: `web/src/components/system/RolesTable.tsx`
- Create: `web/src/components/system/RoleFormDialog.tsx`
- Create: `web/src/components/system/RolesPage.tsx`
- Create: `web/src/pages/dashboard/roles/index.astro`

**Step 1: Write Roles tests**

Test:
- RolesTable renders role name, slug, description, built-in badge
- RolesTable doesn't show delete action for built-in roles
- RolesTable shows "Permissions" link for each role
- RoleFormDialog validates name/slug required
- RoleFormDialog slug validation (lowercase, numbers, hyphens)

**Step 2: Implement RolesTable, RoleFormDialog, RolesPage**

RolesTable: columns = name, slug, description, built-in (badge), actions (edit/permissions/delete — delete hidden for built_in)
RoleFormDialog: name + slug + description fields
RolesPage: list query + create/update/delete mutations

**Step 3: Create Astro page, run tests, commit**

---

### Task 7: Role permissions editor (TDD)

**Files:**
- Create: `web/src/components/system/RolePermissions.tsx`
- Create: `web/src/pages/dashboard/roles/[id]/permissions.astro`
- Extend: `web/src/components/system/__tests__/Roles.test.tsx`

**Step 1: Add RolePermissions tests**

Test:
- RolePermissions renders two tabs: "API Permissions" and "Menu Permissions"
- RolePermissions loads API list and current role permissions
- RolePermissions shows "Apply Template" button
- RolePermissions calls setApis/setMenus on save
- "Apply Template" dialog lists available templates

**Step 2: Implement RolePermissions**

- Uses Tabs from shadcn/ui
- Tab 1: API Permissions — PermissionTree with rbacApi.listApis() as items, rolesApi.getApis(id) as checkedIds
- Tab 2: Menu Permissions — PermissionTree with rbacApi.listAdminMenus() as items, rolesApi.getMenus(id) as checkedIds
- "Apply Template" button opens Select dialog → calls templatesApi.apply(roleId, templateId)
- Save button calls rolesApi.setApis() and rolesApi.setMenus()
- Parallel load: useQueries for all 3 API calls

**Step 3: Create Astro page**

```astro
---
import DashboardLayout from '@/layouts/DashboardLayout.astro';
import { RolePermissions } from '@/components/system/RolePermissions';
const { id } = Astro.params;
---
<DashboardLayout title="Role Permissions - Sky Flux CMS">
  <RolePermissions roleId={id!} client:load />
</DashboardLayout>
```

**Step 4: Run tests, commit**

---

### Task 8: Agent 2 integration — run all tests

**Step 1: Run Agent 2 tests**

```bash
cd web && bun run vitest run src/components/system/__tests__/Users.test.tsx src/components/system/__tests__/Roles.test.tsx
```

**Step 2: Run astro check**

```bash
cd web && bun run astro check
```

**Step 3: Commit any fixes**

---

## Agent 3: Sites + Settings + API Keys

### Task 9: Sites module (TDD)

**Files:**
- Create: `web/src/components/system/__tests__/Sites.test.tsx`
- Create: `web/src/components/system/SitesTable.tsx`
- Create: `web/src/components/system/SiteFormDialog.tsx`
- Create: `web/src/components/system/SiteUsersDialog.tsx`
- Create: `web/src/components/system/SitesPage.tsx`
- Create: `web/src/pages/dashboard/sites/index.astro`

**Step 1: Write Sites tests**

Test:
- SitesTable renders site name, slug, domain, status, timezone columns
- SitesTable row actions: edit/manage users/delete
- SiteFormDialog validates name+slug required, slug regex `^[a-z0-9_]{3,50}$`
- SiteFormDialog create mode doesn't show is_active toggle
- SiteFormDialog edit mode shows all fields, slug disabled
- SiteUsersDialog renders user list with role badges
- SiteUsersDialog has "Add User" and role assignment select
- Delete site requires typing slug to confirm

**Step 2: Implement components**

SitesTable: name, slug, domain, status badge (active/inactive), timezone, actions
SiteFormDialog: name + slug + domain + description + default_locale select + timezone select + is_active toggle
SiteUsersDialog: opens as Dialog, shows site users list + add user form (user select + role select) + remove button
SitesPage: list query + CRUD mutations + user management mutations

Delete site: custom ConfirmDialog that requires typing the slug (compare input vs site.slug before enabling confirm button)

**Step 3: Create Astro page, run tests, commit**

---

### Task 10: Settings module (TDD)

**Files:**
- Create: `web/src/components/system/__tests__/Settings.test.tsx`
- Create: `web/src/components/system/SettingsForm.tsx`
- Create: `web/src/components/system/SettingsPage.tsx`
- Create: `web/src/pages/dashboard/settings/index.astro`

**Step 1: Write Settings tests**

Test:
- SettingsForm renders all config items as key-value pairs
- SettingsForm shows description for each setting
- SettingsForm validates value is not empty
- SettingsForm calls onSave with key+value on individual item save
- SettingsPage loads settings and handles update

**Step 2: Implement**

SettingsForm: renders list of SettingItem as card rows, each with key label + value Input + description text + Save button per row. Not a single form submit — each setting saves individually via settingsApi.update(key, value).

**Step 3: Create Astro page, run tests, commit**

---

### Task 11: API Keys module (TDD)

**Files:**
- Create: `web/src/components/system/__tests__/ApiKeys.test.tsx`
- Create: `web/src/components/system/ApiKeysTable.tsx`
- Create: `web/src/components/system/CreateApiKeyDialog.tsx`
- Create: `web/src/components/system/ApiKeysPage.tsx`
- Create: `web/src/pages/dashboard/api-keys/index.astro`

**Step 1: Write API Keys tests**

Test:
- ApiKeysTable renders name, key_prefix, status, last_used, expires, rate_limit columns
- ApiKeysTable shows "Revoke" action (not delete)
- ApiKeysTable shows "Never" for null last_used_at and "No expiry" for null expires_at
- CreateApiKeyDialog validates name required
- CreateApiKeyDialog shows full key after creation with copy button
- CreateApiKeyDialog disables close until user acknowledges (checkbox: "I've copied the key")

**Step 2: Implement**

ApiKeysTable: columns with formatted dates, rate_limit display
CreateApiKeyDialog: two-phase dialog:
  1. Form: name + expires_at (optional date) + rate_limit (number, default 100)
  2. After creation: shows full key in monospace font + Copy button + "I've saved this key" checkbox

Copy uses `navigator.clipboard.writeText()`.

**Step 3: Create Astro page, run tests, commit**

---

### Task 12: Sites users management page

**Files:**
- Create: `web/src/pages/dashboard/sites/[slug]/users.astro`

**Step 1: Create Astro page**

```astro
---
import DashboardLayout from '@/layouts/DashboardLayout.astro';
import { SitesPage } from '@/components/system/SitesPage';
const { slug } = Astro.params;
---
<DashboardLayout title="Site Users - Sky Flux CMS">
  <SitesPage initialSiteSlug={slug} client:load />
</DashboardLayout>
```

Note: The SiteUsersDialog is a dialog within SitesPage, so this page just opens SitesPage with the dialog pre-opened for the given slug. Alternatively, pass slug as prop to auto-open the dialog.

**Step 2: Commit**

---

### Task 13: Agent 3 integration — run all tests

```bash
cd web && bun run vitest run src/components/system/__tests__/Sites.test.tsx src/components/system/__tests__/Settings.test.tsx src/components/system/__tests__/ApiKeys.test.tsx
```

---

## Agent 4: Comments + Audit

### Task 14: Comments module (TDD)

**Files:**
- Create: `web/src/components/system/__tests__/Comments.test.tsx`
- Create: `web/src/components/system/CommentsTable.tsx`
- Create: `web/src/components/system/CommentDetailDialog.tsx`
- Create: `web/src/components/system/CommentsPage.tsx`
- Create: `web/src/pages/dashboard/comments/index.astro`

**Step 1: Write Comments tests**

Test:
- CommentsTable renders author, content excerpt (first 100 chars), post title, status badge, pin icon
- CommentsTable filter bar: status select (pending/approved/spam/trash) + search
- CommentsTable row actions: approve/reject/spam/pin/reply/delete
- CommentsTable checkbox selection + floating batch action bar
- CommentsTable batch bar shows selected count + approve/reject/spam buttons
- CommentDetailDialog renders full content + replies tree
- CommentDetailDialog admin reply form with submit
- CommentsPage handles batch status mutation

**Step 2: Implement CommentsTable**

Key features:
- Checkbox column using TanStack Table row selection
- Floating batch bar at bottom when selection > 0: `<div className="fixed bottom-0 ...">`
- Content truncated to 100 chars with ellipsis
- Pin icon: `Pin` from lucide-react (filled for pinned)
- Status-specific row actions: "Approve" only shown for non-approved, etc.

**Step 3: Implement CommentDetailDialog**

Dialog showing:
- Full comment content
- Author info (name, email, IP, gravatar)
- Reply tree (recursive rendering, max 3 levels)
- Admin reply textarea + submit button at bottom
- Uses commentsApi.get(id) to load details, commentsApi.reply(id, content) to submit

**Step 4: Implement CommentsPage**

- Filter state: status + search + post_id
- Batch selection state: `selectedIds: string[]`
- Mutations: updateStatus, togglePin, reply, batchStatus, delete
- Batch action handler: `commentsApi.batchStatus(selectedIds, status)` → invalidate + clear selection

**Step 5: Create Astro page, run tests, commit**

---

### Task 15: Audit module (TDD)

**Files:**
- Create: `web/src/components/system/__tests__/Audit.test.tsx`
- Create: `web/src/components/system/AuditTable.tsx`
- Create: `web/src/components/system/AuditPage.tsx`
- Create: `web/src/pages/dashboard/audit/index.astro`

**Step 1: Write Audit tests**

Test:
- AuditTable renders actor name, action badge, resource type, resource ID, IP, timestamp columns
- AuditTable filter bar: action type select + resource type select + date range inputs
- AuditTable is read-only (no row actions, no edit/delete)
- AuditTable shows empty state when no logs
- AuditPage handles filter state and pagination

**Step 2: Implement AuditTable**

Columns: actor (display_name), action (StatusBadge), resource_type (badge), resource_id (monospace), ip_address, created_at (formatted)
Filter bar: action Select (create/update/delete/login) + resource type Select (post/user/setting/comment/media/menu/redirect) + start_date `<input type="date">` + end_date `<input type="date">`
No row actions — purely read-only.

**Step 3: Implement AuditPage, create Astro page, run tests, commit**

---

### Task 16: Agent 4 integration — run all tests

```bash
cd web && bun run vitest run src/components/system/__tests__/Comments.test.tsx src/components/system/__tests__/Audit.test.tsx
```

---

## Agent 5: Menus + Redirects

### Task 17: Menus list module (TDD)

**Files:**
- Create: `web/src/components/system/__tests__/Menus.test.tsx`
- Create: `web/src/components/system/MenusTable.tsx`
- Create: `web/src/components/system/MenuFormDialog.tsx`
- Create: `web/src/components/system/MenusPage.tsx`
- Create: `web/src/pages/dashboard/menus/index.astro`

**Step 1: Write Menus tests**

Test:
- MenusTable renders name, slug, location badge, item count columns
- MenusTable row actions: edit/manage items/delete
- MenuFormDialog validates name+slug required
- MenuFormDialog slug regex validation
- MenuFormDialog location select (header/footer/sidebar/custom)

**Step 2: Implement MenusTable, MenuFormDialog, MenusPage**

MenusTable: name, slug, location (StatusBadge), item_count ("N items"), actions
MenuFormDialog: name + slug + location Select + description textarea
MenusPage: list query + create/update/delete mutations

**Step 3: Create Astro page, run tests, commit**

---

### Task 18: Menu items editor (TDD)

**Files:**
- Create: `web/src/components/system/MenuItemsEditor.tsx`
- Create: `web/src/pages/dashboard/menus/[id]/items.astro`
- Extend: `web/src/components/system/__tests__/Menus.test.tsx`

**Step 1: Add MenuItemsEditor tests**

Test:
- MenuItemsEditor renders menu name and items tree
- MenuItemsEditor shows "Add Item" button
- MenuItemsEditor renders item label, type badge, URL, active toggle
- MenuItemsEditor shows broken reference warning icon for is_broken items
- MenuItemsEditor drag handle is present on each item
- Add item dialog: type select changes visible fields (URL for custom, reference for others)
- MenuItemsEditor calls reorder API on drag-drop

**Step 2: Implement MenuItemsEditor**

Complex component structure:
- Loads menu detail: `siteMenusApi.get(menuId)` → renders tree
- Uses @dnd-kit/sortable for drag-drop reorder (same pattern as CategoryTree from Batch 11)
- Each item row: drag handle + label + type badge + url + active toggle + edit/delete buttons
- "Add Item" opens Dialog with: type Select (custom/post/category/tag/page) → conditional fields (URL for custom, reference search for others) + label + target + icon + css_class + sort_order
- On drag end: flatten tree to ReorderMenuItemDTO[], call siteMenusApi.reorderItems()
- 3-level nesting max — prevent dragging to 4th level

**Step 3: Create Astro page**

```astro
---
import DashboardLayout from '@/layouts/DashboardLayout.astro';
import { MenuItemsEditor } from '@/components/system/MenuItemsEditor';
const { id } = Astro.params;
---
<DashboardLayout title="Menu Items - Sky Flux CMS">
  <MenuItemsEditor menuId={id!} client:load />
</DashboardLayout>
```

**Step 4: Run tests, commit**

---

### Task 19: Redirects module (TDD)

**Files:**
- Create: `web/src/components/system/__tests__/Redirects.test.tsx`
- Create: `web/src/components/system/RedirectsTable.tsx`
- Create: `web/src/components/system/RedirectFormDialog.tsx`
- Create: `web/src/components/system/RedirectsPage.tsx`
- Create: `web/src/pages/dashboard/redirects/index.astro`

**Step 1: Write Redirects tests**

Test:
- RedirectsTable renders source_path, target_url, status_code badge, active toggle, hit_count, last_hit columns
- RedirectsTable filter: status code select (301/302) + search
- RedirectsTable checkbox selection + batch delete
- RedirectsTable shows "Import CSV" and "Export CSV" buttons
- RedirectFormDialog validates source_path starts with "/"
- RedirectFormDialog validates source_path doesn't contain "?"
- RedirectFormDialog status_code select (301/302)

**Step 2: Implement RedirectsTable**

Columns: source_path (monospace), target_url, status_code (badge: 301=blue, 302=yellow), is_active (Switch toggle — calls update inline), hit_count, last_hit_at (formatted), actions (edit/delete)
Filter bar: status_code Select + search Input + "New Redirect" Button + "Import CSV" Button + "Export CSV" Button
Checkbox column for batch delete — floating bar with "Delete Selected" button.

Export: `redirectsApi.export()` → create download link with `URL.createObjectURL()`.

**Step 3: Implement RedirectFormDialog**

Fields: source_path (Input, validate starts with `/`, no `?`), target_url (Input), status_code (Select: 301/302), is_active (Switch)

**Step 4: Implement RedirectsPage, create Astro page, run tests, commit**

---

### Task 20: CSV import dialog (TDD)

**Files:**
- Create: `web/src/components/system/CsvImportDialog.tsx`
- Extend: `web/src/components/system/__tests__/Redirects.test.tsx`

**Step 1: Add CsvImportDialog tests**

Test:
- CsvImportDialog renders file input
- CsvImportDialog shows preview of first 10 rows after file selection
- CsvImportDialog shows "Import" button after preview
- CsvImportDialog calls redirectsApi.import with file
- CsvImportDialog shows result stats (imported/skipped/errors)

**Step 2: Implement CsvImportDialog**

Two-phase dialog:
1. **Select phase**: File input (accept=".csv"), on file select → read first 10 lines with FileReader → parse CSV → show preview table
2. **Import phase**: "Import" button calls `redirectsApi.import(file)` → shows result: "N imported, N skipped, N errors" + error details table

```tsx
// CSV preview parsing (client-side, just for preview)
function parseCSVPreview(text: string, maxRows = 10): string[][] {
  const lines = text.split('\n').filter(Boolean);
  return lines.slice(0, maxRows + 1).map(line => line.split(',').map(s => s.trim()));
}
```

**Step 3: Run tests, commit**

---

### Task 21: Agent 5 integration — run all tests

```bash
cd web && bun run vitest run src/components/system/__tests__/Menus.test.tsx src/components/system/__tests__/Redirects.test.tsx
```

---

## Final Integration (After All Agents Complete)

### Task 22: Full test suite + astro check

**Step 1: Run all tests**

```bash
cd web && bun run vitest run
```

Expected: All tests pass (290 existing + ~200 new = ~490 total).

**Step 2: Run astro check**

```bash
cd web && bun run astro check
```

Expected: 0 errors. Fix any TypeScript errors discovered.

**Step 3: Final commit**

```bash
git add -A
git commit -m "feat(web): complete Batch 12 system management pages with 9 modules"
```

---

## Reference: Existing Patterns to Follow

### i18n mock pattern (all test files)
```typescript
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const label = key.split('.').pop() || key;
      return params
        ? label.replace(/\{\{(\w+)\}\}/g, (_: string, k: string) => String(params[k]))
        : label;
    },
  }),
}));
```

### Page container pattern
```tsx
function ModulePageInner() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { page, perPage, setPage, resetPage } = usePagination();
  // ... state, queries, mutations, handlers
  return <div className="p-6">...</div>;
}
export function ModulePage() {
  return <QueryProvider><I18nProvider><ModulePageInner /></I18nProvider></QueryProvider>;
}
```

### Astro page pattern
```astro
---
import DashboardLayout from '@/layouts/DashboardLayout.astro';
import { ModulePage } from '@/components/system/ModulePage';
---
<DashboardLayout title="Module - Sky Flux CMS">
  <ModulePage client:load />
</DashboardLayout>
```

### Date formatting helper
```typescript
function formatDate(dateStr: string | null): string {
  if (!dateStr) return '--';
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric',
  });
}
```
