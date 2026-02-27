import { createQuery } from '@tanstack/react-query';
import { apiClient, type ListResponse } from '../../shared';
import type { Tag, ListTagParams } from '../types/tags';

export function useTags(siteSlug: string, params: ListTagParams = {}) {
  return createQuery({
    queryKey: ['tags', siteSlug, params],
    queryFn: async () => {
      const response = await apiClient.get<ListResponse<Tag>>(
        `/sites/${siteSlug}/tags`,
        { params }
      );
      return response;
    },
  });
}
