import { createMutation } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { VerifyTOTPRequest, LoginResponse } from '../types/auth';

export function useVerifyTOTP() {
  return createMutation({
    mutationFn: async (data: VerifyTOTPRequest): Promise<LoginResponse> => {
      const response = await apiClient.post<LoginResponse>('/auth/verify-totp', data);
      return response;
    },
  });
}
