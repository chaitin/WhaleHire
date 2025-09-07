// API 通用类型定义

export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

// ==================== 用户相关类型 ====================

export interface UserProfile {
  id: string;
  username: string;
  email: string;
  avatar_url?: string;
  status: string;
  created_at: number;
  last_active_at: number;
  is_deleted: boolean;
}

// ==================== 认证相关类型 ====================

// 登录请求参数
export interface LoginRequest {
  username: string;
  password: string;
  source?: string;
}

// 注册请求参数
export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

// 登录响应数据
export interface LoginResponse {
  user_id?: string;
  admin_id?: string;
  redirect_url?: string;
  token?: string;
}

// 注册响应数据
export interface RegisterResponse {
  user_id?: string;
}

// ==================== 通用响应类型 ====================

export interface ApiResult<T = unknown> {
  success: boolean;
  data?: T;
  message?: string;
}