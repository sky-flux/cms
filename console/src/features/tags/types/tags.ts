export interface Tag {
  id: string;
  name: string;
  slug: string;
  siteId: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTagRequest {
  name: string;
  slug: string;
}

export interface UpdateTagRequest extends Partial<CreateTagRequest> {}

export interface ListTagParams {
  page?: number;
  pageSize?: number;
  search?: string;
}
