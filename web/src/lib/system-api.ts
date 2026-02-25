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
    api.get<PaginatedResponse<User>>(`/v1/users${buildQuery(params)}`, opts),
  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<User>>(`/v1/users/${id}`, opts),
  create: (data: CreateUserDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<User>>('/v1/users', data, opts),
  update: (id: string, data: UpdateUserDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<User>>(`/v1/users/${id}`, data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/users/${id}`, opts),
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
    api.get<ApiResponse<Role[]>>('/v1/rbac/roles', opts),
  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Role>>(`/v1/rbac/roles/${id}`, opts),
  create: (data: CreateRoleDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<Role>>('/v1/rbac/roles', data, opts),
  update: (id: string, data: UpdateRoleDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<Role>>(`/v1/rbac/roles/${id}`, data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/rbac/roles/${id}`, opts),
  getApis: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<string[]>>(`/v1/rbac/roles/${id}/apis`, opts),
  setApis: (id: string, apiIds: string[], opts?: RequestOptions) =>
    api.put<{ success: boolean }>(`/v1/rbac/roles/${id}/apis`, { api_ids: apiIds }, opts),
  getMenus: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<string[]>>(`/v1/rbac/roles/${id}/menus`, opts),
  setMenus: (id: string, menuIds: string[], opts?: RequestOptions) =>
    api.put<{ success: boolean }>(`/v1/rbac/roles/${id}/menus`, { menu_ids: menuIds }, opts),
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
    api.get<ApiResponse<RoleTemplate[]>>('/v1/rbac/templates', opts),
  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<RoleTemplate>>(`/v1/rbac/templates/${id}`, opts),
  create: (data: { name: string; description?: string }, opts?: RequestOptions) =>
    api.post<ApiResponse<RoleTemplate>>('/v1/rbac/templates', data, opts),
  update: (id: string, data: { name?: string; description?: string }, opts?: RequestOptions) =>
    api.put<ApiResponse<RoleTemplate>>(`/v1/rbac/templates/${id}`, data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/rbac/templates/${id}`, opts),
  apply: (roleId: string, templateId: string, opts?: RequestOptions) =>
    api.post<{ success: boolean }>(`/v1/rbac/roles/${roleId}/apply-template`, { template_id: templateId }, opts),
};

// --- RBAC Helpers ---
export const rbacApi = {
  listApis: (opts?: RequestOptions) =>
    api.get<ApiResponse<ApiEndpoint[]>>('/v1/rbac/apis', opts),
  getMyMenus: (opts?: RequestOptions) =>
    api.get<ApiResponse<AdminMenu[]>>('/v1/rbac/me/menus', opts),
  listAdminMenus: (opts?: RequestOptions) =>
    api.get<ApiResponse<AdminMenu[]>>('/v1/rbac/menus', opts),
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
    api.get<PaginatedResponse<Site>>(`/v1/sites${buildQuery(params)}`, opts),
  get: (slug: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Site>>(`/v1/sites/${slug}`, opts),
  create: (data: CreateSiteDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<Site>>('/v1/sites', data, opts),
  update: (slug: string, data: UpdateSiteDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<Site>>(`/v1/sites/${slug}`, data, opts),
  deleteSite: (slug: string, confirmSlug: string, opts?: RequestOptions) =>
    api.post<{ success: boolean }>(`/v1/sites/${slug}/delete`, { confirm_slug: confirmSlug }, opts),
  listUsers: (slug: string, params: SiteUserListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<SiteUser>>(`/v1/sites/${slug}/users${buildQuery(params)}`, opts),
  assignRole: (slug: string, userId: string, role: string, opts?: RequestOptions) =>
    api.put<ApiResponse<{ user_id: string; site_slug: string; role: string }>>(`/v1/sites/${slug}/users/${userId}/role`, { role }, opts),
  removeRole: (slug: string, userId: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/sites/${slug}/users/${userId}/role`, opts),
};

// --- Settings ---
export interface SettingItem {
  key: string;
  value: string;
  description: string;
}

export const settingsApi = {
  get: (opts?: RequestOptions) =>
    api.get<ApiResponse<SettingItem[]>>('/v1/site/settings', opts),
  update: (key: string, value: string, opts?: RequestOptions) =>
    api.put<ApiResponse<SettingItem>>('/v1/site/settings', { key, value }, opts),
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
    api.get<ApiResponse<ApiKey[]>>('/v1/site/api-keys', opts),
  create: (data: CreateApiKeyDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<CreateApiKeyResponse>>('/v1/site/api-keys', data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/site/api-keys/${id}`, opts),
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
    api.get<PaginatedResponse<AuditLog>>(`/v1/site/audit-logs${buildQuery(params)}`, opts),
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
    api.get<PaginatedResponse<Comment>>(`/v1/site/comments${buildQuery(params)}`, opts),
  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Comment>>(`/v1/site/comments/${id}`, opts),
  updateStatus: (id: string, status: string, opts?: RequestOptions) =>
    api.put<ApiResponse<{ id: string; status: string }>>(`/v1/site/comments/${id}/status`, { status }, opts),
  togglePin: (id: string, isPinned: boolean, opts?: RequestOptions) =>
    api.put<ApiResponse<{ id: string; is_pinned: boolean }>>(`/v1/site/comments/${id}/pin`, { is_pinned: isPinned }, opts),
  reply: (id: string, content: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Comment>>(`/v1/site/comments/${id}/reply`, { content }, opts),
  batchStatus: (commentIds: string[], status: string, opts?: RequestOptions) =>
    api.put<ApiResponse<{ updated_count: number }>>('/v1/site/comments/batch-status', { comment_ids: commentIds, status }, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/site/comments/${id}`, opts),
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
    api.get<ApiResponse<SiteMenu[]>>(`/v1/site/menus${params ? buildQuery(params) : ''}`, opts),
  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<SiteMenuDetail>>(`/v1/site/menus/${id}`, opts),
  create: (data: CreateSiteMenuDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<SiteMenu>>('/v1/site/menus', data, opts),
  update: (id: string, data: UpdateSiteMenuDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<SiteMenu>>(`/v1/site/menus/${id}`, data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/site/menus/${id}`, opts),
  addItem: (menuId: string, data: CreateMenuItemDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<SiteMenuItem>>(`/v1/site/menus/${menuId}/items`, data, opts),
  updateItem: (menuId: string, itemId: string, data: UpdateMenuItemDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<SiteMenuItem>>(`/v1/site/menus/${menuId}/items/${itemId}`, data, opts),
  deleteItem: (menuId: string, itemId: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/site/menus/${menuId}/items/${itemId}`, opts),
  reorderItems: (menuId: string, items: ReorderMenuItemDTO[], opts?: RequestOptions) =>
    api.put<{ success: boolean }>(`/v1/site/menus/${menuId}/items/reorder`, { items }, opts),
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
    api.get<PaginatedResponse<Redirect>>(`/v1/site/redirects${buildQuery(params)}`, opts),
  create: (data: CreateRedirectDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<Redirect>>('/v1/site/redirects', data, opts),
  update: (id: string, data: UpdateRedirectDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<Redirect>>(`/v1/site/redirects/${id}`, data, opts),
  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/site/redirects/${id}`, opts),
  batchDelete: (ids: string[], opts?: RequestOptions) =>
    api.post<ApiResponse<{ deleted_count: number }>>('/v1/site/redirects/batch', { ids }, opts),
  import: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return requestFormData<ApiResponse<CsvImportResult>>('POST', '/v1/site/redirects/import', formData);
  },
  export: (opts?: RequestOptions) =>
    api.get<Blob>('/v1/site/redirects/export', opts),
};
