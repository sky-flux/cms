import { api } from './api-client';

export interface LoginResponse {
  success: boolean;
  data: LoginSuccessData | Login2FAData;
}

export interface LoginSuccessData {
  user: { id: string; email: string; display_name: string };
  access_token: string;
  token_type: string;
  expires_in: number;
}

export interface Login2FAData {
  temp_token: string;
  token_type: string;
  expires_in: number;
  requires: 'totp';
}

export interface SetupInstallPayload {
  site_name: string;
  site_slug: string;
  site_url: string;
  admin_email: string;
  admin_password: string;
  admin_display_name: string;
  locale?: string;
}

export function isLogin2FA(data: LoginSuccessData | Login2FAData): data is Login2FAData {
  return 'requires' in data && data.requires === 'totp';
}

export const authApi = {
  login: (email: string, password: string) =>
    api.post<LoginResponse>('/v1/auth/login', { email, password }),

  validate2FA: (code: string, tempToken: string) =>
    api.post<LoginResponse>('/v1/auth/2fa/validate', { code }, {
      headers: { Authorization: `Bearer ${tempToken}` },
    }),

  forgotPassword: (email: string) =>
    api.post('/v1/auth/forgot-password', { email }),

  resetPassword: (token: string, newPassword: string) =>
    api.post('/v1/auth/reset-password', { token, new_password: newPassword }),

  setupCheck: () =>
    api.post<{ success: boolean; data: { installed: boolean } }>('/v1/setup/check'),

  setupInstall: (payload: SetupInstallPayload) =>
    api.post('/v1/setup/initialize', payload),
};
