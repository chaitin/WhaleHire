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

// ==================== 用户更新相关类型 ====================

// 更新用户信息请求参数
export interface UpdateUserRequest {
  id: string;
  password?: string;
  status?: 'active' | 'locked' | 'inactive';
}

// 更新用户资料请求参数
export interface ProfileUpdateRequest {
  username?: string;
  password?: string;
  old_password?: string;
  avatar?: string;
}

// ==================== General Agent 相关类型 ====================

// 消息类型
export interface Message {
  id: string;
  conversation_id: string;
  role: 'user' | 'assistant' | 'system' | 'agent';
  content?: string;
  agent_name?: string;
  type: 'text' | 'image' | 'audio' | 'video' | 'file';
  media_url?: string;
  metadata?: Record<string, unknown>;
  attachments?: Attachment[];
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

// 附件类型
export interface Attachment {
  id: string;
  message_id: string;
  file_name: string;
  file_size: number;
  file_type: string;
  file_url: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

// 对话类型
export interface Conversation {
  id: string;
  user_id: string;
  title: string;
  metadata?: Record<string, unknown>;
  status: string;
  messages: Message[];
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

// 创建对话请求
export interface CreateConversationRequest {
  title: string;
}

// 生成请求
export interface GenerateRequest {
  prompt: string;
  history?: Message[];
  conversation_id?: string;
}

// 生成响应
export interface GenerateResponse {
  answer: string;
}

// 流式数据块
export interface StreamChunk {
  content: string;
  done: boolean;
}

// 流式元数据
export interface StreamMetadata {
  version: string;
  conversation_id: string;
}

// 对话列表请求
export interface ListConversationsRequest {
  page?: number;
  size?: number;
  search?: string;
}

// 分页信息
export interface PageInfo {
  page: number;
  size: number;
  total: number;
  total_pages: number;
}

// 对话列表响应
export interface ListConversationsResponse {
  conversations: Conversation[];
  page_info: PageInfo;
}

// 获取对话历史请求
export interface GetConversationHistoryRequest {
  conversation_id: string;
}

// 删除对话请求
export interface DeleteConversationRequest {
  conversation_id: string;
}

// 添加消息到对话请求
export interface AddMessageToConversationRequest {
  conversation_id: string;
  message: Message;
}

// ==================== 通用响应类型 ====================

export interface ApiResult<T = unknown> {
  success: boolean;
  data?: T;
  message?: string;
}