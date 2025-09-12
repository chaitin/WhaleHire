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
  // General Agent 相关类型
  Message,
  Attachment,
  Conversation,
  CreateConversationRequest,
  GenerateRequest,
  GenerateResponse,
  StreamChunk,
  ListConversationsRequest,
  ListConversationsResponse,
  GetConversationHistoryRequest,
  DeleteConversationRequest,
  AddMessageToConversationRequest,
  PageInfo,
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

// 导出 General Agent 相关 API
export {
  generate,
  generateStream,
  createConversation,
  listConversations,
  getConversationHistory,
  deleteConversation,
  addMessageToConversation,
} from './general-agent';