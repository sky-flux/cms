import { api, requestFormData, type RequestOptions } from './api-client';

// --- Types ---

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface PaginatedResponse<T> {
  success: boolean;
  data: T[];
  pagination: PaginationMeta;
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
}

// Posts
export interface PostSummary {
  id: string;
  title: string;
  slug: string;
  status: string;
  author: { id: string; display_name: string };
  cover_image: { id: string; url: string; thumbnail_urls?: { sm: string; md: string } } | null;
  categories: { id: string; name: string; slug: string }[];
  tags: { id: string; name: string }[];
  view_count: number;
  published_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Post extends PostSummary {
  content: string;
  content_json: unknown;
  excerpt: string;
  seo: { meta_title: string; meta_description: string; og_image_url: string } | null;
  extra_fields: Record<string, unknown> | null;
  scheduled_at: string | null;
  version: number;
}

export interface PostListParams {
  page: number;
  per_page: number;
  status?: string;
  q?: string;
  category_id?: string;
  tag_id?: string;
  author_id?: string;
  sort?: string;
  include_deleted?: boolean;
}

export interface CreatePostDTO {
  title: string;
  slug?: string;
  content?: string;
  content_json?: unknown;
  excerpt?: string;
  status?: string;
  scheduled_at?: string;
  cover_image_id?: string;
  category_ids?: string[];
  primary_category_id?: string;
  tag_ids?: string[];
  meta_title?: string;
  meta_description?: string;
  og_image_url?: string;
}

export interface UpdatePostDTO extends Partial<CreatePostDTO> {
  version: number;
}

export interface Revision {
  id: string;
  version: number;
  editor: { id: string; display_name: string };
  diff_summary: string;
  created_at: string;
}

// Categories
export interface CategoryNode {
  id: string;
  name: string;
  slug: string;
  path: string;
  description?: string;
  parent_id: string | null;
  post_count: number;
  sort_order: number;
  children: CategoryNode[];
  created_at?: string;
  updated_at?: string;
}

export interface CreateCategoryDTO {
  name: string;
  slug?: string;
  parent_id?: string | null;
  description?: string;
  sort_order?: number;
}

export interface UpdateCategoryDTO extends Partial<CreateCategoryDTO> {}

export interface ReorderItem {
  id: string;
  sort_order: number;
}

// Tags
export interface Tag {
  id: string;
  name: string;
  slug: string;
  post_count: number;
  created_at: string;
  updated_at?: string;
}

export interface TagListParams {
  page: number;
  per_page: number;
  q?: string;
  sort?: string;
}

export interface CreateTagDTO {
  name: string;
  slug?: string;
}

export interface UpdateTagDTO extends Partial<CreateTagDTO> {}

// Media
export interface MediaFile {
  id: string;
  file_name: string;
  original_name: string;
  mime_type: string;
  media_type: string;
  file_size: number;
  public_url: string;
  webp_url?: string;
  thumbnail_urls?: { sm: string; md: string };
  reference_count: number;
  created_at: string;
  updated_at: string;
}

export interface MediaFileDetail extends MediaFile {
  width: number;
  height: number;
  alt_text: string;
  title: string;
  referencing_posts: { id: string; title: string }[];
}

export interface MediaListParams {
  page: number;
  per_page: number;
  q?: string;
  media_type?: string;
}

export interface UpdateMediaDTO {
  alt_text?: string;
  title?: string;
}

export interface BatchDeleteResult {
  deleted_count: number;
  skipped: { id: string; reason: string; reference_count: number }[];
}

// --- Helpers ---

function buildQuery(params: Record<string, unknown> | object): string {
  params = params as Record<string, unknown>;
  const entries = Object.entries(params).filter(([, v]) => v !== undefined && v !== null);
  if (entries.length === 0) return '';
  return '?' + entries.map(([k, v]) => `${k}=${encodeURIComponent(String(v))}`).join('&');
}

// --- API Wrappers ---

export const postsApi = {
  list: (params: PostListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<PostSummary>>(
      `/v1/site/posts${buildQuery(params)}`,
      opts,
    ),

  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Post>>(`/v1/site/posts/${id}`, opts),

  create: (data: CreatePostDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>('/v1/site/posts', data, opts),

  update: (id: string, data: UpdatePostDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<Post>>(`/v1/site/posts/${id}`, data, opts),

  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/site/posts/${id}`, opts),

  publish: (id: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>(`/v1/site/posts/${id}/publish`, undefined, opts),

  unpublish: (id: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>(`/v1/site/posts/${id}/unpublish`, undefined, opts),

  revertToDraft: (id: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>(`/v1/site/posts/${id}/revert-to-draft`, undefined, opts),

  restore: (id: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>(`/v1/site/posts/${id}/restore`, undefined, opts),

  getRevisions: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Revision[]>>(`/v1/site/posts/${id}/revisions`, opts),

  rollback: (id: string, revisionId: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>(
      `/v1/site/posts/${id}/revisions/${revisionId}/rollback`,
      undefined,
      opts,
    ),
};

export const categoriesApi = {
  tree: (opts?: RequestOptions) =>
    api.get<ApiResponse<CategoryNode[]>>('/v1/site/categories', opts),

  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<CategoryNode>>(`/v1/site/categories/${id}`, opts),

  create: (data: CreateCategoryDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<CategoryNode>>('/v1/site/categories', data, opts),

  update: (id: string, data: UpdateCategoryDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<CategoryNode>>(`/v1/site/categories/${id}`, data, opts),

  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/site/categories/${id}`, opts),

  reorder: (orders: ReorderItem[], opts?: RequestOptions) =>
    api.put<{ success: boolean }>('/v1/site/categories/reorder', { orders }, opts),
};

export const tagsApi = {
  list: (params: TagListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<Tag>>(
      `/v1/site/tags${buildQuery(params)}`,
      opts,
    ),

  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Tag>>(`/v1/site/tags/${id}`, opts),

  create: (data: CreateTagDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<Tag>>('/v1/site/tags', data, opts),

  update: (id: string, data: UpdateTagDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<Tag>>(`/v1/site/tags/${id}`, data, opts),

  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/v1/site/tags/${id}`, opts),

  suggest: (q: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Tag[]>>(`/v1/site/tags/suggest?q=${encodeURIComponent(q)}`, opts),
};

export const mediaApi = {
  list: (params: MediaListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<MediaFile>>(
      `/v1/site/media${buildQuery(params)}`,
      opts,
    ),

  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<MediaFileDetail>>(`/v1/site/media/${id}`, opts),

  upload: (file: File, altText?: string) => {
    const formData = new FormData();
    formData.append('file', file);
    if (altText) formData.append('alt_text', altText);
    return requestFormData<ApiResponse<MediaFile>>('POST', '/v1/site/media', formData);
  },

  updateMeta: (id: string, data: UpdateMediaDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<MediaFile>>(`/v1/site/media/${id}`, data, opts),

  delete: (id: string, force?: boolean, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(
      `/v1/site/media/${id}${force ? '?force=true' : ''}`,
      opts,
    ),

  batchDelete: (ids: string[], force?: boolean, opts?: RequestOptions) =>
    api.post<ApiResponse<BatchDeleteResult>>(
      `/v1/site/media/batch-delete${force ? '?force=true' : ''}`,
      { ids },
      opts,
    ),
};
