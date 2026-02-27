import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useDeleteTag(siteSlug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (tagId: string): Promise<void> => {
      await apiClient.delete(
        `/sites/${siteSlug}/tags/${tagId}`
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags', siteSlug] });
    },
  });
}
