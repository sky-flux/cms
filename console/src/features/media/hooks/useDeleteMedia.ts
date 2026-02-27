import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useDeleteMedia(siteSlug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (mediaId: string): Promise<void> => {
      await apiClient.delete(`/sites/${siteSlug}/media/${mediaId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['media', siteSlug] });
    },
  });
}
