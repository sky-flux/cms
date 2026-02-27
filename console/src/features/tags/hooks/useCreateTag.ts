import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Tag, CreateTagRequest } from '../types/tags';

export function useCreateTag(siteSlug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: CreateTagRequest): Promise<Tag> => {
      const response = await apiClient.post<Tag>(
        `/sites/${siteSlug}/tags`,
        data
      );
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags', siteSlug] });
    },
  });
}
