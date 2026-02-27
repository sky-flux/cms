import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Site, UpdateSiteRequest } from '../types/sites';

export function useUpdateSite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateSiteRequest }): Promise<Site> => {
      const response = await apiClient.patch<Site>(`/sites/${id}`, data);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sites'] });
    },
  });
}
