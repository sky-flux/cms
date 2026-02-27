import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Category, UpdateCategoryRequest } from '../types/categories';

export function useUpdateCategory(siteSlug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateCategoryRequest }): Promise<Category> => {
      const response = await apiClient.put<Category>(
        `/sites/${siteSlug}/categories/${id}`,
        data
      );
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories', siteSlug] });
      queryClient.invalidateQueries({ queryKey: ['category', siteSlug] });
    },
  });
}
