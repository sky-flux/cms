import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Post, UpdatePostRequest } from '../types/posts';

export function useUpdatePost(siteSlug: string, postId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: UpdatePostRequest): Promise<Post> => {
      const response = await apiClient.put<Post>(
        `/sites/${siteSlug}/posts/${postId}`,
        data
      );
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts', siteSlug] });
      queryClient.invalidateQueries({ queryKey: ['post', siteSlug, postId] });
    },
  });
}
