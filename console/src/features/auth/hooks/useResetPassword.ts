import { createMutation } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { ResetPasswordRequest } from '../types/auth';

export function useResetPassword() {
  return createMutation({
    mutationFn: async (data: ResetPasswordRequest) => {
      await apiClient.post('/auth/reset-password', data);
    },
  });
}
