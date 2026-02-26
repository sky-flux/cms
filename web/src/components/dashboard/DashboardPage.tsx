import { useQuery } from '@tanstack/react-query';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { QueryProvider } from '@/components/providers/QueryProvider';
import {
  fetchDashboardStats,
  dashboardKeys,
  formatBytes,
  type DashboardStats,
} from '@/lib/dashboard-api';

interface StatCardProps {
  label: string;
  value: string | number;
  color?: string;
}

function StatCard({ label, value, color }: StatCardProps) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="text-sm text-muted-foreground">{label}</div>
        <div className={`text-2xl font-bold ${color ?? ''}`}>{value}</div>
      </CardContent>
    </Card>
  );
}

function StatCardSkeleton() {
  return (
    <Card>
      <CardContent className="p-4 space-y-2">
        <Skeleton className="h-4 w-20" />
        <Skeleton className="h-8 w-16" />
      </CardContent>
    </Card>
  );
}

function renderStats(stats: DashboardStats): React.ReactNode {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Posts" value={stats.posts.total} />
        <StatCard label="Published" value={stats.posts.published} color="text-green-600" />
        <StatCard label="Drafts" value={stats.posts.draft} color="text-yellow-600" />
        <StatCard label="Scheduled" value={stats.posts.scheduled} color="text-blue-600" />
      </div>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Comments" value={stats.comments.total} />
        <StatCard label="Pending" value={stats.comments.pending} color="text-orange-600" />
        <StatCard label="Approved" value={stats.comments.approved} color="text-green-600" />
        <StatCard label="Spam" value={stats.comments.spam} color="text-red-600" />
      </div>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Users" value={stats.users.total} />
        <StatCard label="Active" value={stats.users.active} color="text-green-600" />
        <StatCard label="Inactive" value={stats.users.inactive} color="text-muted-foreground" />
        <StatCard label="Media Files" value={stats.media.total} />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <StatCard label="Storage Used" value={formatBytes(stats.media.storage_used)} />
      </div>
    </div>
  );
}

function DashboardPageInner() {
  const { data: stats, isLoading, error } = useQuery({
    queryKey: dashboardKeys.stats,
    queryFn: fetchDashboardStats,
  });

  if (error) {
    return (
      <div className="p-6 text-red-500">
        Failed to load dashboard statistics.
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      {isLoading ? (
        <div className="space-y-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <StatCardSkeleton key={i} />
            ))}
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <StatCardSkeleton key={i} />
            ))}
          </div>
        </div>
      ) : stats ? (
        renderStats(stats)
      ) : null}
    </div>
  );
}

export function DashboardPage() {
  return (
    <QueryProvider>
      <DashboardPageInner />
    </QueryProvider>
  );
}
