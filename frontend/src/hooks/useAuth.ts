// 认证状态管理Hook - 纯Cookie认证版本
import React, { useState, useEffect, useContext, createContext, ReactNode, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { UserInfo, logout as logoutApi, getCurrentUser } from '@/services/auth';

// 认证状态枚举
enum AuthStatus {
  LOADING = 'loading',
  AUTHENTICATED = 'authenticated',
  UNAUTHENTICATED = 'unauthenticated',
}

// 认证上下文类型
interface AuthContextType {
  user: UserInfo | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  authStatus: AuthStatus;
  login: (user: UserInfo) => void;
  logout: () => Promise<void>;
  updateUser: (user: Partial<UserInfo>) => void;
  refreshAuth: () => Promise<void>;
}

// 创建认证上下文
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// 认证提供者组件
interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [authStatus, setAuthStatus] = useState<AuthStatus>(AuthStatus.LOADING);
  const navigate = useNavigate();
  
  // 添加标记，避免在OAuth登录后重复调用refreshAuth
  const [skipInitialRefresh, setSkipInitialRefresh] = useState(false);

  const isLoading = authStatus === AuthStatus.LOADING;
  const isAuthenticated = authStatus === AuthStatus.AUTHENTICATED;

  // 刷新认证状态 - 纯Cookie认证版本
  const refreshAuth = useCallback(async () => {
    console.log('🔐 刷新认证状态...');
    
    try {
      const fetchedUser = await getCurrentUser();
      console.log('🔐 Cookie认证验证成功，获取用户信息:', fetchedUser);
      setUser(fetchedUser);
      setAuthStatus(AuthStatus.AUTHENTICATED);
    } catch (error) {
      console.log('🔐 Cookie认证验证失败，设置为未认证状态:', error);
      setUser(null);
      setAuthStatus(AuthStatus.UNAUTHENTICATED);
    }
  }, []);

  // 初始化认证状态
  useEffect(() => {
    let isMounted = true;

    const initAuth = async () => {
      console.log('🔐 开始初始化认证状态...');
      
      // 如果已经通过OAuth登录设置了用户信息，跳过初始化认证检查
      if (skipInitialRefresh) {
        console.log('🔐 跳过初始化认证检查，用户已通过OAuth登录');
        return;
      }
      
      // 如果当前在OAuth回调页面，跳过初始化认证检查，等待OAuth流程完成
      if (window.location.pathname === '/oauth-callback') {
        console.log('🔐 当前在OAuth回调页面，跳过初始化认证检查');
        return;
      }
      
      await refreshAuth();
      
      if (isMounted) {
        console.log('🔐 认证状态初始化完成');
      }
    };

    void initAuth();

    return () => {
      isMounted = false;
    };
  }, [refreshAuth, skipInitialRefresh]);

  // 登录
  const login = (userInfo: UserInfo) => {
    console.log('🔐 useAuth.login 被调用，用户信息:', userInfo);
    
    // 先设置标记，防止后续的useEffect触发refreshAuth
    setSkipInitialRefresh(true);
    
    setUser(userInfo);
    setAuthStatus(AuthStatus.AUTHENTICATED);
    console.log('🔐 useAuth 状态已更新，isAuthenticated 将变为:', true);
  };

  // 退出登录
  const logout = useCallback(async () => {
    try {
      console.log('🔐 执行登出操作...');
      await logoutApi(); // 调用后端登出接口，清除服务器端 session/cookie
    } catch (error) {
      console.error('🔐 退出登录失败:', error);
    } finally {
      // 清除本地状态
      setUser(null);
      setAuthStatus(AuthStatus.UNAUTHENTICATED);
      setSkipInitialRefresh(false); // 重置标记，下次登录时需要重新检查认证状态
      navigate('/login');
      console.log('🔐 登出完成，已跳转到登录页');
    }
  }, [navigate]);

  // 更新用户信息
  const updateUser = (updatedUser: Partial<UserInfo>) => {
    if (user) {
      const newUser = { ...user, ...updatedUser };
      setUser(newUser);
      console.log('🔐 用户信息已更新:', newUser);
    }
  };

  const value: AuthContextType = {
    user,
    isAuthenticated,
    isLoading,
    authStatus,
    login,
    logout,
    updateUser,
    refreshAuth,
  };

  return React.createElement(
    AuthContext.Provider,
    { value },
    children
  );
}

// 使用认证Hook
export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// 路由守卫Hook
export function useAuthGuard() {
  const { isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      console.log('🔐 用户未认证，跳转到登录页');
      navigate('/login', { replace: true });
    }
  }, [isAuthenticated, isLoading, navigate]);

  return { isAuthenticated, isLoading };
}