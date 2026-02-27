export interface Post {
  id: string;
  title: string;
  slug: string;
  content: string;
  excerpt?: string;
  status: 'draft' | 'published' | 'scheduled' | 'private';
  authorId: string;
  siteId: string;
  publishedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreatePostRequest {
  title: string;
  slug: string;
  content?: string;
  excerpt?: string;
  status?: 'draft' | 'published';
  categoryId?: string;
  tagIds?: string[];
}

export interface UpdatePostRequest extends Partial<CreatePostRequest> {
  publishedAt?: string;
}

export interface ListParams {
  page?: number;
  pageSize?: number;
  status?: string;
  categoryId?: string;
  tagIds?: string[];
  authorId?: string;
  search?: string;
}
