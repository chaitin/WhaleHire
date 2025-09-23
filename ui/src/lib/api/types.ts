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

// ==================== 系统设置相关类型 ====================

// 钉钉OAuth配置
export interface DingtalkOAuth {
  enable: boolean;
  client_id: string;
  client_secret: string;
}

// 自定义OAuth配置
export interface CustomOAuth {
  enable: boolean;
  client_id: string;
  client_secret: string;
  authorize_url: string;
  access_token_url: string;
  userinfo_url: string;
  scopes: string[];
  id_field: string;
  name_field: string;
  avatar_field: string;
  email_field: string;
}

// 系统设置
export interface Setting {
  enable_sso: boolean;
  force_two_factor_auth: boolean;
  disable_password_login: boolean;
  enable_auto_login: boolean;
  dingtalk_oauth: DingtalkOAuth;
  custom_oauth: CustomOAuth;
  base_url?: string;
  created_at: number;
  updated_at: number;
}

// 更新系统设置请求参数
export interface UpdateSettingRequest {
  enable_sso?: boolean;
  force_two_factor_auth?: boolean;
  disable_password_login?: boolean;
  enable_auto_login?: boolean;
  dingtalk_oauth?: {
    enable?: boolean;
    client_id?: string;
    client_secret?: string;
  };
  custom_oauth?: {
    enable?: boolean;
    client_id?: string;
    client_secret?: string;
    authorize_url?: string;
    access_token_url?: string;
    userinfo_url?: string;
    scopes?: string[];
    id_field?: string;
    name_field?: string;
    avatar_field?: string;
    email_field?: string;
  };
  base_url?: string;
}

// ==================== OAuth相关类型 ====================

// OAuth登录或注册请求参数
export interface OAuthSignUpOrInRequest {
  source?: 'plugin' | 'browser';
  platform: 'dingtalk' | 'custom';
  session_id?: string;
  redirect_url?: string;
  inviate_code?: string;
}

// OAuth URL响应
export interface OAuthURLResponse {
  url: string;
}

// OAuth回调请求参数
export interface OAuthCallbackRequest {
  state: string;
  code: string;
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
  agent_name: string;
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
  agent_name: string;
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