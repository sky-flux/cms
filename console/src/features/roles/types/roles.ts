export interface Permission {
  id: string;
  name: string;
  description?: string;
  category: string;
}

export interface Role {
  id: string;
  name: string;
  slug: string;
  description?: string;
  permissions: Permission[];
  isSystem: boolean;
  status: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateRoleRequest {
  name: string;
  slug: string;
  description?: string;
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
  status?: number;
}
