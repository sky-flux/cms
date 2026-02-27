import { createMutation } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { ForgotPasswordRequest } from '../types/auth';

export function useForgotPassword() {
  return createMutation({
    mutationFn: async (data: ForgotPasswordRequest) => {
      await apiClient.post('/auth/forgot-password', data);
    },
  });
}
