import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useDeleteCategory(siteSlug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (categoryId: string): Promise<void> => {
      await apiClient.delete(
        `/sites/${siteSlug}/categories/${categoryId}`
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories', siteSlug] });
    },
  });
}
