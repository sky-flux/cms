import { createQuery } from '@tanstack/react-query';
import { apiClient, type ListResponse } from '../../shared';
import type { Category, ListCategoryParams } from '../types/categories';

export function useCategories(siteSlug: string, params: ListCategoryParams = {}) {
  return createQuery({
    queryKey: ['categories', siteSlug, params],
    queryFn: async () => {
      const response = await apiClient.get<ListResponse<Category>>(
        `/sites/${siteSlug}/categories`,
        { params }
      );
      return response;
    },
  });
}
