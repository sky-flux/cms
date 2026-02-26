import { api } from './api-client';

export interface DashboardStats {
  posts: {
    total: number;
    published: number;
    draft: number;
    scheduled: number;
  };
  users: {
    total: number;
    active: number;
    inactive: number;
  };
  comments: {
    total: number;
    pending: number;
    approved: number;
    spam: number;
  };
  media: {
    total: number;
    storage_used: number;
  };
}

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export async function fetchDashboardStats(): Promise<DashboardStats> {
  const res = await api.get<ApiResponse<DashboardStats>>('/v1/site/dashboard/stats');
  return res.data;
}

export const dashboardKeys = {
  stats: ['dashboard', 'stats'] as const,
};

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}
