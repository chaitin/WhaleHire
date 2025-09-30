// 认证状态管理Hook - Cookie 认证版本
import React, { useState, useEffect, useContext, createContext, ReactNode, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { getUserInfo, setUserInfo, clearUserInfo } from '@/lib/api';
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

  const isLoading = authStatus === AuthStatus.LOADING;
  const isAuthenticated = authStatus === AuthStatus.AUTHENTICATED;

  // 刷新认证状态 - Cookie 认证版本
  const refreshAuth = useCallback(async () => {
    console.log('🔐 刷新认证状态...');
    const userInfo = getUserInfo();
    
    console.log('🔐 本地存储状态:', { hasUserInfo: !!userInfo });

    // 如果有本地用户信息，先设置为已认证状态
    if (userInfo) {
      console.log('🔐 使用本地存储的用户信息');
      setUser(userInfo);
      setAuthStatus(AuthStatus.AUTHENTICATED);
      
      // 在后台验证 Cookie 认证状态
      try {
        const fetchedUser = await getCurrentUser();
        console.log('🔐 Cookie 认证验证成功，更新用户信息:', fetchedUser);
        setUser(fetchedUser);
        setUserInfo(fetchedUser as unknown as Record<string, unknown>); // 更新本地存储
      } catch (error) {
        console.warn('🔐 Cookie 认证验证失败，清理认证信息:', error);
        clearUserInfo();
        setUser(null);
        setAuthStatus(AuthStatus.UNAUTHENTICATED);
      }
      return;
    }

    // 如果没有本地用户信息，直接验证服务器认证状态
    console.log('🔐 没有本地用户信息，检查服务器认证状态...');
    try {
      const fetchedUser = await getCurrentUser();
      console.log('🔐 服务器认证验证成功，获取用户信息:', fetchedUser);
      setUser(fetchedUser);
      setUserInfo(fetchedUser as unknown as Record<string, unknown>); // 保存到本地存储
      setAuthStatus(AuthStatus.AUTHENTICATED);
    } catch (error) {
      console.log('🔐 服务器认证验证失败，设置为未认证状态:', error);
      setUser(null);
      setAuthStatus(AuthStatus.UNAUTHENTICATED);
    }
  }, []);

  // 初始化认证状态
  useEffect(() => {
    let isMounted = true;

    const initAuth = async () => {
      console.log('🔐 开始初始化认证状态...');
      await refreshAuth();
      
      if (isMounted) {
        console.log('🔐 认证状态初始化完成');
      }
    };

    void initAuth();

    return () => {
      isMounted = false;
    };
  }, [refreshAuth]);

  // 登录
  const login = (userInfo: UserInfo) => {
    console.log('🔐 useAuth.login 被调用，用户信息:', userInfo);
    setUser(userInfo);
    setUserInfo(userInfo as unknown as Record<string, unknown>); // 保存到本地存储
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
      clearUserInfo();
      setUser(null);
      setAuthStatus(AuthStatus.UNAUTHENTICATED);
      navigate('/login');
      console.log('🔐 登出完成，已跳转到登录页');
    }
  }, [navigate]);

  // 更新用户信息
  const updateUser = (updatedUser: Partial<UserInfo>) => {
    if (user) {
      const newUser = { ...user, ...updatedUser };
      setUser(newUser);
      setUserInfo(newUser as unknown as Record<string, unknown>); // 更新本地存储
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