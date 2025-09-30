import { LucideIcon } from 'lucide-react';

export interface NavigationItem {
  id: string;
  label: string;
  icon: LucideIcon;
  path: string;
  children?: NavigationItem[];
}

export interface SidebarProps {
  isCollapsed?: boolean;
  onToggle?: () => void;
}
