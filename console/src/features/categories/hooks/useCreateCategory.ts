import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Category, CreateCategoryRequest } from '../types/categories';

export function useCreateCategory(siteSlug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: CreateCategoryRequest): Promise<Category> => {
      const response = await apiClient.post<Category>(
        `/sites/${siteSlug}/categories`,
        data
      );
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories', siteSlug] });
    },
  });
}
