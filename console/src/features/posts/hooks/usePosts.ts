import { createQuery } from '@tanstack/react-query';
import { apiClient, type ListResponse } from '../../shared';
import type { Post, ListParams } from '../types/posts';

export function usePosts(siteSlug: string, params: ListParams = {}) {
  return createQuery({
    queryKey: ['posts', siteSlug, params],
    queryFn: async () => {
      const response = await apiClient.get<ListResponse<Post>>(
        `/sites/${siteSlug}/posts`,
        { params }
      );
      return response;
    },
  });
}
