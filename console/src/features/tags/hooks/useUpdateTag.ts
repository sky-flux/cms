import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Tag, UpdateTagRequest } from '../types/tags';

export function useUpdateTag(siteSlug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateTagRequest }): Promise<Tag> => {
      const response = await apiClient.put<Tag>(
        `/sites/${siteSlug}/tags/${id}`,
        data
      );
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags', siteSlug] });
    },
  });
}
