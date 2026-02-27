import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Role, UpdateRoleRequest } from '../types/roles';

export function useUpdateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateRoleRequest }): Promise<Role> => {
      const response = await apiClient.put<Role>(`/rbac/roles/${id}`, data);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
    },
  });
}
