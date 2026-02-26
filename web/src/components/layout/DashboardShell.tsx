import { useState, useEffect, type ReactNode } from 'react';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { adminNavSections } from './nav-items';
import { api, ApiError } from '@/lib/api-client';
import { useAuthStore } from '@/stores/auth-store';
import { useUIStore } from '@/stores/ui-store';
import { I18nProvider } from '@/components/providers/I18nProvider';

interface DashboardShellProps {
  currentPath: string;
  children?: ReactNode;
}

interface Site {
  id: string;
  name: string;
  slug: string;
}

interface AuthUser {
  id: string;
  email: string;
  display_name: string;
  sites?: Site[];
}

export function DashboardShell({ currentPath, children }: DashboardShellProps) {
  const [collapsed, setCollapsed] = useState(false);
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);
  const clearAuth = useAuthStore((s) => s.clearAuth);
  const setSiteSlug = useUIStore((s) => s.setSiteSlug);

  useEffect(() => {
    // Fetch user info on mount
    const fetchUser = async () => {
      try {
        const resp = await api.get<{ success: boolean; data: AuthUser }>('/v1/auth/me');
        if (resp.success && resp.data) {
          setUser(resp.data);
          // Set the first site as the active site
          if (resp.data.sites && resp.data.sites.length > 0) {
            setSiteSlug(resp.data.sites[0].slug);
          }
        }
      } catch (err) {
        // If unauthorized, redirect to login
        if (err instanceof ApiError && err.status === 401) {
          window.location.href = '/login';
          return;
        }
        console.error('Failed to fetch user:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, [setSiteSlug]);

  const handleLogout = async () => {
    try {
      await api.post('/v1/auth/logout');
    } catch {
      // Ignore errors, proceed with redirect anyway
    }
    clearAuth();
    window.location.href = '/login';
  };

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <I18nProvider>
      <div className="flex h-screen w-full">
        <Sidebar
          sections={adminNavSections}
          currentPath={currentPath}
          collapsed={collapsed}
          onToggle={() => setCollapsed((v) => !v)}
        />
        <div className="flex flex-1 flex-col overflow-hidden">
          <Header
            user={user ? { displayName: user.display_name, email: user.email } : undefined}
            onToggleSidebar={() => setCollapsed((v) => !v)}
            onLogout={handleLogout}
          />
          <main className="flex-1 overflow-auto">{children}</main>
        </div>
      </div>
    </I18nProvider>
  );
}
