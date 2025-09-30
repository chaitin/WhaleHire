// 认证相关API服务 - Cookie 认证版本
import { apiPost, apiGet, setUserInfo, clearUserInfo, debugLog } from '@/lib/api';

// 登录请求参数 - 根据swagger定义
export interface LoginRequest {
  username: string;
  password: string;
  source: 'browser' | 'plugin'; // 登录来源
  session_id?: string; // 会话ID，插件登录时必填
}

// 前端登录表单参数（包含rememberMe）
export interface LoginFormData {
  username: string;
  password: string;
  rememberMe?: boolean;
}

// 用户信息类型 - 根据swagger User定义
export interface UserInfo {
  id: string;
  username: string;
  email: string;
  avatar_url?: string;
  status: 'active' | 'inactive' | 'locked';
  created_at: number;
  last_active_at: number;
  is_deleted: boolean;
}

// 登录响应数据 - 根据swagger LoginResp定义
export interface LoginResponse {
  user: UserInfo;
  redirect_url?: string;
}

// 登录
export const login = async (credentials: LoginFormData): Promise<LoginResponse> => {
  // 构建符合后端API的登录请求参数
  const loginData: LoginRequest = {
    username: credentials.username,
    password: credentials.password,
    source: 'browser', // 默认浏览器登录
  };

  try {
    debugLog('🔐 发送登录请求到后端...');
    const response = await apiPost<LoginResponse>('/v1/user/login', loginData as unknown as Record<string, unknown>);

    // 保存用户信息到 sessionStorage（Cookie 认证不需要区分 rememberMe）
    setUserInfo(response.user as unknown as Record<string, unknown>);
    debugLog('🔐 登录成功，用户信息已保存');

    return response;
  } catch (error) {
    // 登录失败时直接抛出错误，不使用mock数据
    console.error('🔐 登录失败:', error);
    throw error;
  }
};

// 退出登录
export const logout = async (): Promise<void> => {
  try {
    debugLog('🔐 调用后端登出接口...');
    // 调用后端退出登录接口，服务器会清除 session/cookie
    await apiPost('/v1/user/logout');
    debugLog('🔐 后端登出成功');
  } catch (error) {
    console.warn('🔐 后端登出失败:', error);
  } finally {
    // 清除本地存储的用户信息
    clearUserInfo();
    debugLog('🔐 本地用户信息已清除');
  }
};

// 获取当前用户信息
export const getCurrentUser = async (): Promise<UserInfo> => {
  debugLog('🔐 获取当前用户信息...');
  try {
    const response = await apiGet<UserInfo>('/v1/user/profile');
    // 更新本地存储的用户信息
    setUserInfo(response as unknown as Record<string, unknown>);
    debugLog('🔐 用户信息获取成功:', response);
    return response;
  } catch (error) {
    console.error('🔐 获取用户信息失败:', error);
    throw error;
  }
};

// OAuth 认证地址响应
export interface OAuthUrlResponse {
  url: string;
}

// 获取OAuth认证地址
export const getOAuthUrl = async (): Promise<string> => {
  debugLog('🔐 获取OAuth认证地址...');
  debugLog('🔐 使用固定OAuth回调地址:', "https://hire.chaitin.net//resume-management");
  
  try {
    // 传递固定的回调地址给后端
    const response = await apiGet<OAuthUrlResponse>('/v1/user/oauth/signup-or-in', {
      platform: 'custom',
      source: 'browser',
      redirect_url: "https://hire.chaitin.net//resume-management"
    });
    debugLog('🔐 OAuth认证地址获取成功:', response.url);
    return response.url;
  } catch (error) {
    console.error('🔐 获取OAuth认证地址失败:', error);
    throw error;
  }
};

// OAuth回调处理
export const handleOAuthCallback = async (code: string, state: string): Promise<LoginResponse> => {
  debugLog('🔐 处理OAuth回调...');
  debugLog('🔐 OAuth回调地址:', "https://hire.chaitin.net//resume-management");
  
  try {
    const response = await apiGet<LoginResponse>('/v1/user/oauth/callback', {
      code,
      state,
    });
    
    // 保存用户信息到 sessionStorage
    // setUserInfo(response.user as unknown as Record<string, unknown>);
    debugLog('🔐 OAuth登录成功，用户信息已保存');
    
    return response;
  } catch (error) {
    console.error('🔐 OAuth回调处理失败:', error);
    throw error;
  }
};

// 更新用户资料
export interface UpdateProfileRequest {
  username?: string;
  avatar?: string;
  old_password?: string;
  password?: string;
}

export const updateProfile = async (data: UpdateProfileRequest): Promise<UserInfo> => {
  debugLog('🔐 更新用户资料...');
  try {
    const response = await apiPost<UserInfo>('/v1/user/profile', data as unknown as Record<string, unknown>);
    // 更新本地存储的用户信息
    setUserInfo(response as unknown as Record<string, unknown>);
    debugLog('🔐 用户资料更新成功:', response);
    return response;
  } catch (error) {
    console.error('🔐 更新用户资料失败:', error);
    throw error;
  }
};