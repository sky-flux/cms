export interface NavItem {
  label: string;
  href: string;
  icon: string;
  children?: NavItem[];
}

export interface NavSection {
  title?: string;
  items: NavItem[];
}

export const adminNavSections: NavSection[] = [
  {
    items: [
      { label: 'nav.dashboard', href: '/dashboard', icon: 'LayoutDashboard' },
    ],
  },
  {
    title: 'nav.content',
    items: [
      { label: 'nav.posts', href: '/dashboard/posts', icon: 'FileText' },
      { label: 'nav.categories', href: '/dashboard/categories', icon: 'FolderTree' },
      { label: 'nav.tags', href: '/dashboard/tags', icon: 'Tags' },
      { label: 'nav.media', href: '/dashboard/media', icon: 'Image' },
      { label: 'nav.comments', href: '/dashboard/comments', icon: 'MessageSquare' },
      { label: 'nav.menus', href: '/dashboard/menus', icon: 'Menu' },
      { label: 'nav.redirects', href: '/dashboard/redirects', icon: 'ArrowLeftRight' },
    ],
  },
  {
    title: 'nav.system',
    items: [
      { label: 'nav.users', href: '/dashboard/users', icon: 'Users' },
      { label: 'nav.roles', href: '/dashboard/roles', icon: 'Shield' },
      { label: 'nav.apiKeys', href: '/dashboard/api-keys', icon: 'Key' },
      { label: 'nav.audit', href: '/dashboard/audit', icon: 'ClipboardList' },
      { label: 'nav.sites', href: '/dashboard/sites', icon: 'Globe' },
      { label: 'nav.settings', href: '/dashboard/settings', icon: 'Settings' },
    ],
  },
];

export const allNavItems: NavItem[] = adminNavSections.flatMap((s) => s.items);
