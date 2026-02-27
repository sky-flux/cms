import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Post, CreatePostRequest } from '../types/posts';

export function useCreatePost(siteSlug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: CreatePostRequest): Promise<Post> => {
      const response = await apiClient.post<Post>(
        `/sites/${siteSlug}/posts`,
        data
      );
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts', siteSlug] });
    },
  });
}
