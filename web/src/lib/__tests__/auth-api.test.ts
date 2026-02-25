import { describe, it, expect, vi, beforeEach } from 'vitest';
import { api } from '@/lib/api-client';

vi.mock('@/lib/api-client', () => ({
  api: {
    post: vi.fn(),
    get: vi.fn(),
  },
}));

import { authApi } from '@/lib/auth-api';

describe('authApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('login', () => {
    it('calls POST /v1/auth/login with email and password', async () => {
      const mockResp = {
        success: true,
        data: { user: { id: '1', email: 'a@b.com', display_name: 'A' }, access_token: 'tok' },
      };
      vi.mocked(api.post).mockResolvedValue(mockResp);

      const result = await authApi.login('a@b.com', 'password123');
      expect(api.post).toHaveBeenCalledWith('/v1/auth/login', {
        email: 'a@b.com',
        password: 'password123',
      });
      expect(result).toEqual(mockResp);
    });
  });

  describe('validate2FA', () => {
    it('calls POST /v1/auth/2fa/validate with code and temp token header', async () => {
      vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

      await authApi.validate2FA('123456', 'temp-tok');
      expect(api.post).toHaveBeenCalledWith(
        '/v1/auth/2fa/validate',
        { code: '123456' },
        { headers: { Authorization: 'Bearer temp-tok' } },
      );
    });
  });

  describe('forgotPassword', () => {
    it('calls POST /v1/auth/forgot-password with email', async () => {
      vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

      await authApi.forgotPassword('a@b.com');
      expect(api.post).toHaveBeenCalledWith('/v1/auth/forgot-password', {
        email: 'a@b.com',
      });
    });
  });

  describe('resetPassword', () => {
    it('calls POST /v1/auth/reset-password with token and new password', async () => {
      vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

      await authApi.resetPassword('reset-tok', 'newpass123');
      expect(api.post).toHaveBeenCalledWith('/v1/auth/reset-password', {
        token: 'reset-tok',
        new_password: 'newpass123',
      });
    });
  });

  describe('setupCheck', () => {
    it('calls POST /v1/setup/check', async () => {
      vi.mocked(api.post).mockResolvedValue({ success: true, data: { installed: false } });

      const result = await authApi.setupCheck();
      expect(api.post).toHaveBeenCalledWith('/v1/setup/check');
      expect(result).toEqual({ success: true, data: { installed: false } });
    });
  });

  describe('setupInstall', () => {
    it('calls POST /v1/setup/initialize with full payload', async () => {
      const payload = {
        site_name: 'My Site',
        site_slug: 'my-site',
        site_url: 'https://example.com',
        admin_email: 'a@b.com',
        admin_password: 'pass1234',
        admin_display_name: 'Admin',
        locale: 'en',
      };
      vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

      await authApi.setupInstall(payload);
      expect(api.post).toHaveBeenCalledWith('/v1/setup/initialize', payload);
    });
  });
});
