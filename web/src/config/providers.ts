/**
 * Provider configuration
 *
 * Environment-aware configuration for React Query and theme providers.
 */

export interface ProviderConfig {
  queryClient: {
    logger: boolean;
    retries: number;
    defaultOptions?: {
      queries?: {
        staleTime: number;
      };
    };
  };
  theme: {
    defaultTheme: 'system' | 'light' | 'dark';
  };
}

const devConfig: ProviderConfig = {
  queryClient: {
    logger: true,
    retries: 0,
    defaultOptions: {
      queries: {
        staleTime: 60 * 1000, // 1 minute
      },
    },
  },
  theme: {
    defaultTheme: 'system',
  },
};

const prodConfig: ProviderConfig = {
  queryClient: {
    logger: false,
    retries: 3,
    defaultOptions: {
      queries: {
        staleTime: 5 * 60 * 1000, // 5 minutes
      },
    },
  },
  theme: {
    defaultTheme: 'system',
  },
};

/**
 * Get provider configuration based on environment
 * @param isDev - Whether the environment is development
 */
export const getProviderConfig = (isDev = import.meta.env.DEV): ProviderConfig =>
  isDev ? devConfig : prodConfig;

/**
 * Default provider configuration using current environment
 */
export const providerConfig: ProviderConfig = getProviderConfig();
