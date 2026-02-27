import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useDeletePost(siteSlug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (postId: string): Promise<void> => {
      await apiClient.delete(`/sites/${siteSlug}/posts/${postId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts', siteSlug] });
    },
  });
}
