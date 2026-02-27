import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useLogout() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async () => {
      await apiClient.post('/auth/logout');
    },
    onSuccess: () => {
      queryClient.clear();
    },
  });
}
