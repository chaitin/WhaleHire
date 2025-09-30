import { useCallback, useState } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './sidebar';
import { Header } from './header';
import { ProtectedRoute } from '@/components/auth/protected-route';
import { cn } from '@/lib/utils';

export function MainLayout() {
  // 移动端侧边栏开关状态
  const [isMobileSidebarOpen, setIsMobileSidebarOpen] = useState(false);

  // 切换侧边栏显示状态
  const handleToggleSidebar = useCallback(() => {
    setIsMobileSidebarOpen((prev) => !prev);
  }, []);

  // 关闭侧边栏
  const handleCloseSidebar = useCallback(() => {
    setIsMobileSidebarOpen(false);
  }, []);

  return (
    <ProtectedRoute>
      <div className="flex min-h-screen bg-background">
        {/* 桌面端侧边栏 */}
        <Sidebar className="hidden lg:flex" />

        {/* 移动端遮罩 */}
        {isMobileSidebarOpen && (
          <div
            className="fixed inset-0 z-40 bg-black/40 lg:hidden"
            onClick={handleCloseSidebar}
            aria-hidden="true"
          />
        )}

        {/* 移动端侧边栏 */}
        <Sidebar
          className={cn(
            'fixed inset-y-0 left-0 z-50 w-64 border-r bg-card shadow-lg transition-transform duration-200 lg:hidden',
            isMobileSidebarOpen ? 'translate-x-0' : '-translate-x-full'
          )}
          isMobile
          onClose={handleCloseSidebar}
        />

        {/* Main Content */}
        <div className="flex flex-1 flex-col overflow-hidden">
          {/* 顶部栏 */}
          <Header onToggleSidebar={handleToggleSidebar} />

          {/* 页面内容 */}
          <main className="flex-1 overflow-auto bg-[#F9FAFB] min-h-0">
            <div className="min-h-full">
              <Outlet />
            </div>
          </main>
        </div>

      </div>
    </ProtectedRoute>
  );
}
