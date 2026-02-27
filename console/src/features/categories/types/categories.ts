export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  parentId?: string;
  siteId: string;
  order: number;
  createdAt: string;
  updatedAt: string;
  children?: Category[];
}

export interface CreateCategoryRequest {
  name: string;
  slug: string;
  description?: string;
  parentId?: string;
}

export interface UpdateCategoryRequest extends Partial<CreateCategoryRequest> {
  order?: number;
}

export interface ListCategoryParams {
  page?: number;
  pageSize?: number;
  parentId?: string;
  search?: string;
}
