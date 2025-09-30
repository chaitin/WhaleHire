// API 基础配置和工具函数

// 环境配置接口
interface AppConfig {
  apiBaseUrl: string;
  appTitle: string;
  appVersion: string;
  appEnv: string;
  maxFileSize: number;
  allowedFileTypes: string;
  enableMock: boolean;
  debug: boolean;
  enableDevtools: boolean;
}

// 检测当前运行环境
const detectRuntime = () => {
  // 检测是否为开发环境（vite dev）
  const isDev = import.meta.env.DEV;
  
  // 检测是否为预览模式（vite preview）
  // 预览模式下 import.meta.env.PROD 为 true，但端口通常是 4173 或 4174
  const isPreview = import.meta.env.PROD && 
    typeof window !== 'undefined' && 
    (window.location.port === '4173' || window.location.port === '4174');
  
  return {
    isDev,
    isPreview,
    isProd: import.meta.env.PROD && !isPreview
  };
};

// 获取环境变量配置
const getEnvConfig = (): AppConfig => {
  const env = import.meta.env;
  const runtime = detectRuntime();
  
  // API 基础 URL 逻辑：
  // 1. 开发模式：使用 /api（通过 vite 代理）
  // 2. 预览模式：使用完整的生产环境 URL
  // 3. 生产环境：使用完整的生产环境 URL
  let apiBaseUrl: string;
  
  if (runtime.isDev) {
    // 开发模式使用相对路径，通过 vite 代理
    apiBaseUrl = '/api';
  } else {
    // 预览模式和生产环境都使用完整 URL
    apiBaseUrl = env.VITE_API_BASE_URL || 'https://hire.chaitin.net/api';
  }
  
  return {
    apiBaseUrl,
    appTitle: env.VITE_APP_TITLE || 'WhaleHire 智能招聘系统',
    appVersion: env.VITE_APP_VERSION || '1.0.0',
    appEnv: env.VITE_APP_ENV || 'development',
    maxFileSize: parseInt(env.VITE_MAX_FILE_SIZE) || 10485760, // 10MB
    allowedFileTypes: env.VITE_ALLOWED_FILE_TYPES || '.pdf,.doc,.docx',
    enableMock: env.VITE_ENABLE_MOCK === 'true',
    debug: env.VITE_DEBUG === 'true',
    enableDevtools: env.VITE_ENABLE_DEVTOOLS === 'true',
  };
};

// 导出配置对象
export const appConfig = getEnvConfig();

// 向后兼容的 API_BASE_URL 导出
export const API_BASE_URL = appConfig.apiBaseUrl;

// 调试信息：在开发环境下输出当前配置
if (appConfig.debug || import.meta.env.DEV) {
  const runtime = detectRuntime();
  console.log('[API Config] 运行环境检测:', {
    isDev: runtime.isDev,
    isPreview: runtime.isPreview,
    isProd: runtime.isProd,
    currentPort: typeof window !== 'undefined' ? window.location.port : 'N/A',
    apiBaseUrl: appConfig.apiBaseUrl,
    mode: import.meta.env.MODE,
    prod: import.meta.env.PROD,
    dev: import.meta.env.DEV
  });
}

// 环境检查工具函数
export const isDevelopment = () => appConfig.appEnv === 'development';
export const isProduction = () => appConfig.appEnv === 'production';

// 调试日志工具
export const debugLog = (...args: unknown[]) => {
  if (appConfig.debug) {
    console.log('[DEBUG]', ...args);
  }
};

// API 响应类型
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  success: boolean;
}

// 分页响应类型
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// 错误类型
export class ApiError<T = unknown> extends Error {
  constructor(
    public code: number,
    public message: string,
    public data?: T
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// 基础请求函数
export async function apiRequest<T = unknown>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const headers = new Headers(options.headers as HeadersInit | undefined);
  const method = options.method?.toUpperCase() ?? 'GET';

  // 设置默认 Content-Type（除非是 FormData）
  if (
    method !== 'GET' &&
    method !== 'HEAD' &&
    !headers.has('Content-Type') &&
    !(options.body instanceof FormData)
  ) {
    headers.set('Content-Type', 'application/json');
  }

  // 配置请求，确保包含 cookies
  const config: RequestInit = {
    ...options,
    headers,
    credentials: 'include', // 重要：确保 cookies 被包含在请求中
  };

  try {
    debugLog(`🌐 API Request: ${method} ${url}`);
    const response = await fetch(url, config);
    
    // 处理HTTP错误状态
    if (!response.ok) {
      if (response.status === 401) {
        // 未授权，Cookie已失效
        debugLog('🔐 401 Unauthorized - Cookie已失效');
        
        // 只有在不是登录页面时才跳转到登录页
        if (!window.location.pathname.includes('/login')) {
          debugLog('🔐 Redirecting to login page');
          window.location.href = '/login';
        }
        throw new ApiError(401, '登录已过期，请重新登录');
      }
      
      const errorData = await response.json().catch(() => ({}));
      throw new ApiError(
        response.status,
        errorData.message || `HTTP Error: ${response.status}`,
        errorData
      );
    }

    const data: ApiResponse<T> = await response.json();
    debugLog(`✅ API Response: ${method} ${url} - ${response.status}`);
    
    return data.data;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
    
    // 网络错误或其他错误
    console.error(`❌ API Error: ${method} ${url}`, error);
    throw new ApiError(0, '网络请求失败，请检查网络连接', error);
  }
}

// GET 请求
export const apiGet = <T = unknown>(endpoint: string, params?: Record<string, unknown>): Promise<T> => {
  let finalEndpoint = endpoint;
  
  if (params) {
    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        queryParams.append(key, String(value));
      }
    });
    
    const queryString = queryParams.toString();
    if (queryString) {
      finalEndpoint += (endpoint.includes('?') ? '&' : '?') + queryString;
    }
  }
  
  return apiRequest<T>(finalEndpoint);
};

// POST 请求
export function apiPost<T = unknown>(endpoint: string, data?: Record<string, unknown>, options?: RequestInit): Promise<T>;
export function apiPost<T = unknown>(endpoint: string, data: FormData, options?: RequestInit): Promise<T>;
export function apiPost<T = unknown>(endpoint: string, data?: FormData | Record<string, unknown>, options: RequestInit = {}): Promise<T> {
  const isFormData = data instanceof FormData;
  const headers = isFormData
    ? options.headers
    : {
        'Content-Type': 'application/json',
        ...options.headers,
      };

  return apiRequest<T>(endpoint, {
    method: 'POST',
    body: isFormData ? data : data ? JSON.stringify(data) : undefined,
    ...options,
    headers,
  });
}

// PUT 请求
export const apiPut = <T = unknown>(endpoint: string, data?: Record<string, unknown>): Promise<T> => {
  return apiRequest<T>(endpoint, {
    method: 'PUT',
    body: data ? JSON.stringify(data) : undefined,
  });
};

// DELETE 请求
export const apiDelete = <T = unknown>(endpoint: string): Promise<T> => {
  return apiRequest<T>(endpoint, {
    method: 'DELETE',
  });
};

// 文件上传请求
export const apiUpload = <T = unknown>(endpoint: string, formData: FormData): Promise<T> => {
  return apiRequest<T>(endpoint, {
    method: 'POST',
    body: formData,
    // 不设置 Content-Type，让浏览器自动设置 multipart/form-data
  });
};