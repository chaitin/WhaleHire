import { Link, useLocation } from 'react-router-dom';
import {
  FileText,
  Briefcase,
  Users,
  Calendar,
  BarChart3,
  Settings,
  UserCircle,
  Shield,
  Rocket,
  X,
  Crown,
  UserCheck,
  Sparkles,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { NavigationItem } from '@/types/navigation';
// 导入package.json以获取版本号
import packageJson from '../../../package.json';

const navigationItems: NavigationItem[] = [
  {
    id: 'dashboard',
    label: '仪表盘',
    icon: BarChart3,
    path: '/dashboard',
  },
  {
    id: 'job-profile',
    label: '岗位画像',
    icon: UserCheck,
    path: '/job-profile',
  },
  {
    id: 'resume-management',
    label: '简历管理',
    icon: FileText,
    path: '/resume-management',
  },
  {
    id: 'position-management',
    label: '岗位管理',
    icon: Briefcase,
    path: '/position-management',
  },
  {
    id: 'intelligent-matching',
    label: '智能匹配',
    icon: Sparkles,
    path: '/intelligent-matching',
  },
  {
    id: 'candidate',
    label: '候选人',
    icon: Users,
    path: '/candidate',
  },
  {
    id: 'interview-schedule',
    label: '面试安排',
    icon: Calendar,
    path: '/interview-schedule',
  },
  {
    id: 'data-analysis',
    label: '数据分析',
    icon: BarChart3,
    path: '/data-analysis',
  },
  {
    id: 'platform-config',
    label: '平台配置',
    icon: Settings,
    path: '/platform-config',
  },
];

const systemItems: NavigationItem[] = [
  {
    id: 'system-settings',
    label: '系统设置',
    icon: Settings,
    path: '/system-settings',
  },
  {
    id: 'user-management',
    label: '用户管理',
    icon: UserCircle,
    path: '/user-management',
  },
  {
    id: 'permission-management',
    label: '权限管理',
    icon: Shield,
    path: '/permission-management',
  },
];

interface SidebarProps {
  className?: string;
  isMobile?: boolean;
  onClose?: () => void;
}

export function Sidebar({
  className,
  isMobile = false,
  onClose,
}: SidebarProps) {
  const location = useLocation();

  const isActive = (path: string) => {
    return location.pathname === path;
  };

  // 获取版本号：优先从环境变量获取，否则从package.json获取
  const getVersion = () => {
    const envVersion = import.meta.env.VITE_APP_VERSION;
    return envVersion || packageJson.version;
  };

  return (
    <div
      className={cn(
        'flex h-full w-64 flex-col border-r bg-card',
        isMobile && 'pt-2',
        className
      )}
      style={{ height: '150vh' }}
      role="navigation"
      aria-label="主导航"
    >
      {/* Logo */}
      <div className="flex h-16 items-center justify-between gap-2 border-b px-6">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
            <Rocket className="h-4 w-4 text-primary-foreground" />
          </div>
          <div className="flex items-center gap-2">
            <span className="text-lg font-semibold">WhaleHire</span>
            <span className="inline-flex items-center gap-1 rounded-md bg-gray-50 px-1.5 py-0.5 text-xs font-medium text-gray-600 border border-gray-200">
              <Crown className="h-2.5 w-2.5" />
              免费版
            </span>
          </div>
        </div>
        {isMobile ? (
          <button
            type="button"
            className="rounded-md p-2 text-muted-foreground transition hover:bg-accent hover:text-accent-foreground"
            onClick={onClose}
            aria-label="关闭侧边栏"
          >
            <X className="h-4 w-4" />
          </button>
        ) : null}
      </div>

      {/* Navigation */}
      <div
        className="flex flex-1 flex-col overflow-hidden"
        style={{
          height: 'calc(100vh - 64px)',
          justifyContent: 'space-between',
        }}
      >
        <div className="scrollbar-thin scrollbar-thumb-muted-foreground/20 scrollbar-track-transparent -mr-1 flex-1 overflow-y-auto pr-1">
          {/* 添加顶部间距，让菜单向下移动 */}
          <div className="pt-8">
            <nav className="sidebar-nav px-3">
              {navigationItems.map((item) => {
                const Icon = item.icon;
                const isDisabled =
                  item.id !== 'resume-management' &&
                  item.id !== 'job-profile' &&
                  item.id !== 'platform-config' &&
                  item.id !== 'intelligent-matching';

                if (isDisabled) {
                  return (
                    <div
                      key={item.id}
                      className={cn(
                        'sidebar-nav-item disabled',
                        'cursor-not-allowed opacity-50'
                      )}
                    >
                      <Icon className="h-4 w-4" />
                      <span>{item.label}</span>
                    </div>
                  );
                }

                return (
                  <Link
                    key={item.id}
                    to={item.path}
                    className={cn(
                      'sidebar-nav-item',
                      isActive(item.path) && 'active'
                    )}
                    onClick={isMobile ? onClose : undefined}
                  >
                    <Icon className="h-4 w-4" />
                    <span>{item.label}</span>
                  </Link>
                );
              })}
            </nav>

            {/* System Settings */}
            <nav className="sidebar-nav px-3 mt-3">
              {systemItems.map((item) => {
                const Icon = item.icon;
                return (
                  <div
                    key={item.id}
                    className={cn(
                      'sidebar-nav-item disabled',
                      'cursor-not-allowed opacity-50'
                    )}
                  >
                    <Icon className="h-4 w-4" />
                    <span>{item.label}</span>
                  </div>
                );
              })}
            </nav>
          </div>
        </div>

        {/* Upgrade Section - positioned to align with 10th row */}
        <div
          className="border-t px-4 pt-4 pb-4"
          style={{ marginTop: 'auto', position: 'relative', top: '-120px' }}
        >
          <div className="rounded-lg bg-gradient-to-r from-green-50 to-green-100 border border-green-200 p-4 opacity-60 cursor-not-allowed">
            <div className="flex items-center gap-2 mb-2">
              <div className="flex h-6 w-6 items-center justify-center rounded-full bg-gradient-to-r from-green-400 to-green-500">
                <Rocket className="h-3 w-3 text-white" />
              </div>
              <span className="text-sm font-medium text-gray-700">
                升级专业版
              </span>
            </div>
            <p className="text-xs text-gray-500 mb-3">
              解锁更多高级功能，提升工作效率
            </p>
            <button
              disabled
              className="w-full rounded-md bg-gradient-to-r from-green-400 to-green-500 px-3 py-2 text-xs font-medium text-white cursor-not-allowed opacity-70"
            >
              立即升级
            </button>
            {/* 版本号显示 */}
            <div className="mt-3 text-center">
              <span className="text-xs text-gray-500">
                版本号：{getVersion()}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
