// 受保护的路由组件
import { ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';

interface ProtectedRouteProps {
  children: ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, authStatus } = useAuth();
  const location = useLocation();

  console.log('🛡️ ProtectedRoute 状态:', { isAuthenticated, isLoading, authStatus, pathname: location.pathname });

  // 正在加载认证状态
  if (isLoading) {
    console.log('🛡️ 正在验证认证状态，显示加载界面');
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="flex flex-col items-center space-y-4">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
          <p className="text-sm text-muted-foreground">正在验证登录状态...</p>
        </div>
      </div>
    );
  }

  // 未认证，重定向到登录页面
  if (!isAuthenticated) {
    console.log('🛡️ 用户未认证，重定向到登录页面');
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // 已认证，渲染子组件
  console.log('🛡️ 用户已认证，渲染受保护的内容');
  return <>{children}</>;
}

