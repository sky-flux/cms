import { describe, it, expect } from 'vitest';
import { getProviderConfig } from '../providers';

describe('providers.config', () => {
  describe('development mode', () => {
    it('should use development config when DEV mode', () => {
      // Act
      const providerConfig = getProviderConfig(true);

      // Assert
      expect(providerConfig.queryClient).toBeDefined();
      expect(providerConfig.queryClient.logger).toBe(true);
      expect(providerConfig.queryClient.retries).toBe(0);
      expect(providerConfig.queryClient.defaultOptions?.queries?.staleTime).toBe(60 * 1000); // 1 minute
      expect(providerConfig.theme.defaultTheme).toBe('system');
    });
  });

  describe('production mode', () => {
    it('should use production config when PROD mode', () => {
      // Act
      const providerConfig = getProviderConfig(false);

      // Assert
      expect(providerConfig.queryClient).toBeDefined();
      expect(providerConfig.queryClient.logger).toBe(false);
      expect(providerConfig.queryClient.retries).toBe(3);
      expect(providerConfig.queryClient.defaultOptions?.queries?.staleTime).toBe(5 * 60 * 1000); // 5 minutes
      expect(providerConfig.theme.defaultTheme).toBe('system');
    });
  });
});
