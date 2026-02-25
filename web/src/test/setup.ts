import '@testing-library/jest-dom/vitest';

// Polyfill ResizeObserver for Radix UI components (Switch, Select, etc.)
if (typeof globalThis.ResizeObserver === 'undefined') {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as unknown as typeof globalThis.ResizeObserver;
}
