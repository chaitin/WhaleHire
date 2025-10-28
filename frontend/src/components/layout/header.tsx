import { Menu, Bell, ChevronDown, LogOut } from 'lucide-react';
import { Button } from '@/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/ui/dropdown-menu';
import { useAuth } from '@/hooks/useAuth';

interface HeaderProps {
  onToggleSidebar?: () => void;
}

export function Header({ onToggleSidebar }: HeaderProps) {
  const { user, logout } = useAuth();

  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('退出登录失败:', error);
    }
  };

  return (
    <header className="flex h-16 items-center justify-between border-b bg-background px-6">
      <div className="flex flex-1 items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          className="lg:hidden"
          onClick={onToggleSidebar}
          aria-label="展开侧边栏"
        >
          <Menu className="h-5 w-5" />
        </Button>
      </div>

      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          className="header-notification-btn disabled"
          aria-label="查看通知"
          disabled
        >
          <Bell className="h-4 w-4" />
          <span className="header-notification-badge" />
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <div className="header-profile cursor-pointer hover:bg-accent/50 transition-colors">
              <Avatar
                className="h-8 w-8 border-2"
                style={{ borderColor: '#7bb8ff' }}
              >
                {user?.avatar_url && (
                  <AvatarImage src={user.avatar_url} alt="用户头像" />
                )}
                <AvatarFallback
                  className="text-primary-foreground"
                  style={{
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }}
                >
                  {user?.username?.charAt(0)?.toUpperCase() || 'U'}
                </AvatarFallback>
              </Avatar>
              <div className="header-profile-info">
                <p className="header-profile-name">
                  {user?.username || '用户'}
                </p>
                <p className="header-profile-email">{user?.email || ''}</p>
              </div>
              <ChevronDown className="header-profile-chevron" />
            </div>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuItem
              onClick={handleLogout}
              className="text-destructive focus:text-destructive"
            >
              <LogOut className="mr-2 h-4 w-4" />
              退出登录
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
