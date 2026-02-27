import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useDeleteSite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (siteId: string): Promise<void> => {
      await apiClient.delete(`/sites/${siteId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sites'] });
    },
  });
}
