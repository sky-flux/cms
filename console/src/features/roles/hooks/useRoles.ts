import { createQuery } from '@tanstack/react-query';
import { apiClient, type ListResponse } from '../../shared';
import type { Role } from '../types/roles';

export interface ListRolesParams {
  page?: number;
  pageSize?: number;
}

export function useRoles(params: ListRolesParams = {}) {
  return createQuery({
    queryKey: ['roles', params],
    queryFn: async () => {
      const response = await apiClient.get<ListResponse<Role>>(
        '/rbac/roles',
        { params }
      );
      return response;
    },
  });
}
