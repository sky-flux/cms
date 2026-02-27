import { createQuery } from '@tanstack/react-query';
import { apiClient, type ListResponse } from '../../shared';
import type { User } from '../types/users';

export interface ListUsersParams {
  page?: number;
  pageSize?: number;
  search?: string;
  role?: string;
  status?: string;
}

export function useUsers(params: ListUsersParams = {}) {
  return createQuery({
    queryKey: ['users', params],
    queryFn: async () => {
      const response = await apiClient.get<ListResponse<User>>(
        '/users',
        { params }
      );
      return response;
    },
  });
}
