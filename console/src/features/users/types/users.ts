export interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  role: string;
  status: 'active' | 'inactive';
  createdAt: string;
  updatedAt: string;
}

export interface CreateUserRequest {
  email: string;
  name: string;
  password: string;
  role?: string;
}

export interface UpdateUserRequest {
  email?: string;
  name?: string;
  password?: string;
  role?: string;
}
