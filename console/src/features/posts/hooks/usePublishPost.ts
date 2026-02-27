import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Post } from '../types/posts';

export type PostStatus = 'draft' | 'published' | 'scheduled' | 'private';

export interface UpdatePostStatusRequest {
  status: PostStatus;
  publishedAt?: string;
}

export function usePublishPost(siteSlug: string, postId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: UpdatePostStatusRequest): Promise<Post> => {
      const response = await apiClient.patch<Post>(
        `/sites/${siteSlug}/posts/${postId}/status`,
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
