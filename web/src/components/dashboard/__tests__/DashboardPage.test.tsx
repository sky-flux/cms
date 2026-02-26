import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClientProvider } from '@tanstack/react-query';
import { createQueryClient } from '@/lib/query-client';
import { DashboardPage } from '../DashboardPage';

// Mock the dashboard-api module
vi.mock('@/lib/dashboard-api', () => ({
  fetchDashboardStats: vi.fn().mockResolvedValue({
    posts: { total: 100, published: 80, draft: 15, scheduled: 5 },
    users: { total: 10, active: 9, inactive: 1 },
    comments: { total: 50, pending: 5, approved: 40, spam: 5 },
    media: { total: 200, storage_used: 2684354560 },
  }),
  dashboardKeys: { stats: ['dashboard', 'stats'] },
  formatBytes: (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  },
}));

function renderWithProviders(ui: React.ReactNode) {
  const qc = createQueryClient();
  return render(
    <QueryClientProvider client={qc}>{ui}</QueryClientProvider>,
  );
}

describe('DashboardPage', () => {
  it('should render stat cards after loading', async () => {
    renderWithProviders(<DashboardPage />);

    await waitFor(() => {
      expect(screen.getByText('100')).toBeInTheDocument();
    });

    // Posts section
    expect(screen.getByText('80')).toBeInTheDocument(); // published
    expect(screen.getByText('15')).toBeInTheDocument(); // drafts
    // '5' appears 3 times (scheduled + pending + spam), use getAllByText
    expect(screen.getAllByText('5')).toHaveLength(3);

    // Comments
    expect(screen.getByText('50')).toBeInTheDocument();
    expect(screen.getByText('40')).toBeInTheDocument(); // approved

    // Media + Users
    expect(screen.getByText('200')).toBeInTheDocument();
    expect(screen.getByText('2.5 GB')).toBeInTheDocument();
    expect(screen.getByText('10')).toBeInTheDocument(); // total users
  });

  it('should show loading skeleton initially', () => {
    renderWithProviders(<DashboardPage />);
    // Skeleton cards should be present before data loads
    const skeletons = document.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });
});
