import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Site, CreateSiteRequest } from '../types/sites';

export function useCreateSite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: CreateSiteRequest): Promise<Site> => {
      const response = await apiClient.post<Site>('/sites', data);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sites'] });
    },
  });
}
