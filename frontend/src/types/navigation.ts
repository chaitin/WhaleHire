import { LucideIcon } from 'lucide-react';

export interface NavigationItem {
  id: string;
  label: string;
  icon: LucideIcon;
  path: string;
  disabled?: boolean;
  children?: NavigationItem[];
}

export interface NavigationGroup {
  id: string;
  label: string;
  items: NavigationItem[];
}

export interface SidebarProps {
  isCollapsed?: boolean;
  onToggle?: () => void;
}
