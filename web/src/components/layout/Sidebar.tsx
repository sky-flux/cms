import {
  LayoutDashboard,
  FileText,
  FolderTree,
  Tags,
  Image,
  MessageSquare,
  Menu,
  ArrowLeftRight,
  Users,
  Shield,
  Settings,
  Key,
  ClipboardList,
  Globe,
  PanelLeft,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import type { NavSection } from './nav-items';

const iconMap: Record<string, LucideIcon> = {
  LayoutDashboard,
  FileText,
  FolderTree,
  Tags,
  Image,
  MessageSquare,
  Menu,
  ArrowLeftRight,
  Users,
  Shield,
  Settings,
  Key,
  ClipboardList,
  Globe,
};

interface SidebarProps {
  sections: NavSection[];
  currentPath: string;
  collapsed?: boolean;
  onToggle?: () => void;
}

export function Sidebar({ sections, currentPath, collapsed = false, onToggle }: SidebarProps) {
  const { t } = useTranslation();

  return (
    <nav
      data-collapsed={collapsed}
      className={cn(
        'flex h-full flex-col border-r bg-background transition-[width] duration-200',
        collapsed ? 'w-16' : 'w-60',
      )}
    >
      <div className="flex h-14 items-center justify-between border-b px-3">
        {!collapsed && (
          <span className="text-sm font-semibold tracking-tight">Sky Flux CMS</span>
        )}
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={onToggle}
          aria-label="Toggle sidebar"
          className={cn(collapsed && 'mx-auto')}
        >
          <PanelLeft className="size-4" />
        </Button>
      </div>

      <ScrollArea className="flex-1 py-2">
        <TooltipProvider delayDuration={0}>
          {sections.map((section, sIdx) => (
            <div key={sIdx}>
              {sIdx > 0 && <Separator className="my-2" />}
              {section.title && !collapsed && (
                <p className="mb-1 px-4 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t(section.title)}
                </p>
              )}
              <ul className="space-y-0.5 px-2">
                {section.items.map((item) => {
                  const Icon = iconMap[item.icon];
                  const isActive =
                    item.href === '/dashboard'
                      ? currentPath === '/dashboard'
                      : currentPath.startsWith(item.href);

                  const link = (
                    <a
                      href={item.href}
                      data-active={isActive}
                      className={cn(
                        'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                        'hover:bg-accent hover:text-accent-foreground',
                        isActive && 'bg-accent text-accent-foreground',
                        collapsed && 'justify-center px-2',
                      )}
                    >
                      {Icon && <Icon className="size-4 shrink-0" />}
                      {!collapsed && <span>{t(item.label)}</span>}
                      {collapsed && <span className="sr-only">{t(item.label)}</span>}
                    </a>
                  );

                  return (
                    <li key={item.href}>
                      {collapsed ? (
                        <Tooltip>
                          <TooltipTrigger asChild>{link}</TooltipTrigger>
                          <TooltipContent side="right">{t(item.label)}</TooltipContent>
                        </Tooltip>
                      ) : (
                        link
                      )}
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </TooltipProvider>
      </ScrollArea>
    </nav>
  );
}
