import { useState, type ReactNode } from 'react';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { adminNavSections } from './nav-items';

interface DashboardShellProps {
  currentPath: string;
  children?: ReactNode;
}

export function DashboardShell({ currentPath, children }: DashboardShellProps) {
  const [collapsed, setCollapsed] = useState(false);

  // Placeholder user until auth store integration
  const user = {
    displayName: 'Admin',
    email: 'admin@example.com',
  };

  return (
    <div className="flex h-screen w-full">
      <Sidebar
        sections={adminNavSections}
        currentPath={currentPath}
        collapsed={collapsed}
        onToggle={() => setCollapsed((v) => !v)}
      />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header
          user={user}
          onToggleSidebar={() => setCollapsed((v) => !v)}
          onLogout={() => {
            window.location.href = '/login';
          }}
        />
        <main className="flex-1 overflow-auto">{children}</main>
      </div>
    </div>
  );
}
