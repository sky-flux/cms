import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { LoginRequest, LoginResponse } from '../types/auth';

export function useLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: LoginRequest): Promise<LoginResponse> => {
      const response = await apiClient.post<LoginResponse>('/auth/login', data);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me'] });
    },
  });
}
