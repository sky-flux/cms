import { createQuery } from '@tanstack/react-query';
import { apiClient, type ListResponse } from '../../shared';
import type { Site } from '../types/sites';

export interface ListSitesParams {
  page?: number;
  pageSize?: number;
  search?: string;
  status?: string;
}

export function useSites(params: ListSitesParams = {}) {
  return createQuery({
    queryKey: ['sites', params],
    queryFn: async () => {
      const response = await apiClient.get<ListResponse<Site>>(
        '/sites',
        { params }
      );
      return response;
    },
  });
}
