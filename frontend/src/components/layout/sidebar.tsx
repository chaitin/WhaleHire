import { Link, useLocation } from 'react-router-dom';
import {
  FileText,
  Calendar,
  BarChart3,
  Settings,
  UserCircle,
  Rocket,
  X,
  Crown,
  UserCheck,
  Sparkles,
  GitBranch,
  FileCode,
  MessageSquare,
  HelpCircle,
  Users,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { NavigationGroup } from '@/types/navigation';
// 导入package.json以获取版本号
import packageJson from '../../../package.json';

// 菜单分组配置
const navigationGroups: NavigationGroup[] = [
  {
    id: 'function-menu',
    label: '主菜单',
    items: [
      {
        id: 'dashboard',
        label: '仪表盘',
        icon: BarChart3,
        path: '/dashboard',
        disabled: true,
      },
      {
        id: 'job-profile',
        label: '岗位画像',
        icon: UserCheck,
        path: '/job-profile',
        disabled: false,
      },
      {
        id: 'resume-management',
        label: '简历管理',
        icon: FileText,
        path: '/resume-management',
        disabled: false,
      },
      {
        id: 'intelligent-matching',
        label: '智能匹配',
        icon: Sparkles,
        path: '/intelligent-matching',
        disabled: false,
      },
      {
        id: 'interview-schedule',
        label: '面试安排',
        icon: Calendar,
        path: '/interview-schedule',
        disabled: true,
      },
      {
        id: 'talent-pool',
        label: '人才库',
        icon: Users,
        path: '/talent-pool',
        disabled: true,
      },
      {
        id: 'intelligent-orchestration',
        label: '智能编排',
        icon: GitBranch,
        path: '/intelligent-orchestration',
        disabled: true,
      },
    ],
  },
  {
    id: 'system-config',
    label: '系统配置',
    items: [
      {
        id: 'platform-config',
        label: '基础功能配置',
        icon: Settings,
        path: '/platform-config',
        disabled: false,
      },
      {
        id: 'basic-config',
        label: '基本信息',
        icon: Settings,
        path: '/basic-config',
        disabled: true,
      },
      {
        id: 'user-management',
        label: '用户管理',
        icon: UserCircle,
        path: '/user-management',
        disabled: true,
      },
      {
        id: 'operation-log',
        label: '操作日志',
        icon: FileCode,
        path: '/operation-log',
        disabled: true,
      },
    ],
  },
  {
    id: 'help-center',
    label: '帮助中心',
    items: [
      {
        id: 'document-center',
        label: '文档中心',
        icon: HelpCircle,
        path: '/document-center',
        disabled: true,
      },
      {
        id: 'discussion-area',
        label: '讨论专区',
        icon: MessageSquare,
        path: '/discussion-area',
        disabled: true,
      },
    ],
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
          {/* 添加顶部间距，让菜单向下移动更多距离 */}
          <div className="pt-12">
            {navigationGroups.map((group, groupIndex) => (
              <div
                key={group.id}
                className={cn('px-3', groupIndex > 0 && 'mt-6')}
              >
                {/* 主菜单标题 - 不可点击，浅灰色 */}
                <div className="mb-3 px-3 text-xs font-semibold text-gray-400 uppercase tracking-wider">
                  {group.label}
                </div>

                {/* 子菜单项 */}
                <nav className="sidebar-nav space-y-1">
                  {group.items.map((item) => {
                    const Icon = item.icon;

                    if (item.disabled) {
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
              </div>
            ))}
          </div>
        </div>

        {/* Upgrade Section - 固定在左下角 */}
        <div className="border-t px-4 pt-4 pb-4">
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
