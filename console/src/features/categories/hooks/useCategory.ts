import { createQuery } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Category } from '../types/categories';

export function useCategory(siteSlug: string, categoryId: string) {
  return createQuery({
    queryKey: ['category', siteSlug, categoryId],
    queryFn: async () => {
      const response = await apiClient.get<Category>(
        `/sites/${siteSlug}/categories/${categoryId}`
      );
      return response;
    },
    enabled: !!categoryId,
  });
}
