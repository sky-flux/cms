import { createQuery } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { MeResponse } from '../types/auth';

export function useMe() {
  return createQuery({
    queryKey: ['me'],
    queryFn: async (): Promise<MeResponse> => {
      const response = await apiClient.get<MeResponse>('/auth/me');
      return response;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}
