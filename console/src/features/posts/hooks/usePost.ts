import { createQuery } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Post } from '../types/posts';

export function usePost(siteSlug: string, postId: string) {
  return createQuery({
    queryKey: ['post', siteSlug, postId],
    queryFn: async (): Promise<Post> => {
      const response = await apiClient.get<Post>(
        `/sites/${siteSlug}/posts/${postId}`
      );
      return response;
    },
    enabled: !!postId,
  });
}
