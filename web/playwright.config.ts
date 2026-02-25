import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  expect: { timeout: 5_000 },
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: 1,
  reporter: process.env.CI ? 'github' : 'html',
  use: {
    baseURL: 'http://localhost:4321',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    locale: 'en-US',
  },
  projects: [
    { name: 'setup', testMatch: 'setup.spec.ts' },
    { name: 'auth', testMatch: 'auth.spec.ts', dependencies: ['setup'] },
    { name: 'content', testMatch: 'content.spec.ts', dependencies: ['setup'] },
    { name: 'media', testMatch: 'media.spec.ts', dependencies: ['setup'] },
    { name: 'system', testMatch: 'system.spec.ts', dependencies: ['setup'] },
    { name: 'rbac', testMatch: 'rbac.spec.ts', dependencies: ['setup'] },
    { name: 'multisite', testMatch: 'multisite.spec.ts', dependencies: ['setup'] },
  ],
  webServer: {
    command: 'bun run dev',
    url: 'http://localhost:4321',
    timeout: 30_000,
    reuseExistingServer: !process.env.CI,
  },
});
