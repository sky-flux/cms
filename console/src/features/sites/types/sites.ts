export interface Site {
  id: string;
  name: string;
  slug: string;
  domain?: string;
  status: 'active' | 'inactive';
  createdAt: string;
  updatedAt: string;
}

export interface CreateSiteRequest {
  name: string;
  slug: string;
  domain?: string;
}

export interface UpdateSiteRequest {
  name?: string;
  domain?: string;
  status?: 'active' | 'inactive';
}
