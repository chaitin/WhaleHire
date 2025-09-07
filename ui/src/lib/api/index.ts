// API 统一导出入口

// 导出类型定义
export type {
  ApiResponse,
  UserProfile,
  LoginRequest,
  RegisterRequest,
  LoginResponse,
  RegisterResponse,
  ApiResult,
  ProfileUpdateRequest,
} from './types';

// 导出认证相关 API
export {
  userLogin,
  adminLogin,
  userRegister,
} from './auth';

// 导出用户相关 API
export {
  getUserProfile,
  getAvatarUrl,
  getDisplayName,
  logout,
  updateProfile,
} from './user';