// 认证相关 API

import type { 
  LoginRequest, 
  RegisterRequest, 
  LoginResponse, 
  RegisterResponse, 
  ApiResult,
  OAuthSignUpOrInRequest,
  OAuthURLResponse,
  OAuthCallbackRequest,
  OAuthCallbackResponse
} from './types';

/**
 * 用户登录
 */
export async function userLogin(loginData: LoginRequest): Promise<ApiResult<LoginResponse>> {
  try {
    const response = await fetch('/api/v1/user/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ...loginData,
        source: loginData.source || 'browser',
      }),
    });
    
    const result = await response.json();
    
    if (response.ok) {
      return { success: true, data: result };
    } else {
      return { success: false, message: result.message || '登录失败，请检查用户名和密码' };
    }
  } catch (error) {
    console.error('用户登录请求失败:', error);
    return { success: false, message: '网络错误，请稍后重试' };
  }
}

/**
 * 管理员登录
 */
export async function adminLogin(loginData: LoginRequest): Promise<ApiResult<LoginResponse>> {
  try {
    const response = await fetch('/api/v1/admin/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        ...loginData,
        source: loginData.source || 'browser',
      }),
    });
    
    const result = await response.json();
    
    if (response.ok) {
      return { success: true, data: result };
    } else {
      return { success: false, message: result.message || '登录失败，请检查管理员账号和密码' };
    }
  } catch (error) {
    console.error('管理员登录请求失败:', error);
    return { success: false, message: '网络错误，请稍后重试' };
  }
}

/**
 * 用户注册
 */
export async function userRegister(registerData: RegisterRequest): Promise<ApiResult<RegisterResponse>> {
  try {
    const response = await fetch('/api/v1/user/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(registerData),
    });
    
    const result = await response.json();
    
    if (response.ok) {
      return { success: true, data: result };
    } else {
      return { success: false, message: result.message || '注册失败，请检查输入信息' };
    }
  } catch (error) {
    console.error('用户注册请求失败:', error);
    return { success: false, message: '网络错误，请稍后重试' };
  }
}


/**
 * OAuth登录或注册
 */
export async function oauthSignUpOrIn(oauthData: OAuthSignUpOrInRequest): Promise<ApiResult<OAuthURLResponse>> {
  try {
    const queryParams = new URLSearchParams();
    
    if (oauthData.source) queryParams.append('source', oauthData.source);
    queryParams.append('platform', oauthData.platform);
    if (oauthData.session_id) queryParams.append('session_id', oauthData.session_id);
    if (oauthData.redirect_url) queryParams.append('redirect_url', oauthData.redirect_url);
    if (oauthData.inviate_code) queryParams.append('inviate_code', oauthData.inviate_code);
    console.log('OAuth登录，查询参数:', queryParams.toString());
    
    const response = await fetch(`/api/v1/user/oauth/signup-or-in?${queryParams.toString()}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    const result = await response.json();
    
    if (response.ok) {
      return { success: true, data: result.data };
    } else {
      return { success: false, message: result.message || 'OAuth登录失败' };
    }
  } catch (error) {
    console.error('OAuth登录请求失败:', error);
    return { success: false, message: '网络错误，请稍后重试' };
  }
}

/**
 * OAuth回调处理
 */
export async function oauthCallback(callbackData: OAuthCallbackRequest): Promise<ApiResult<OAuthCallbackResponse>> {
  try {
    const queryParams = new URLSearchParams();
    queryParams.append('state', callbackData.state);
    queryParams.append('code', callbackData.code);
    
    const response = await fetch(`/api/v1/user/oauth/callback?${queryParams.toString()}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    const result = await response.json();
    
    if (response.ok) {
      return { success: true, data: result.data };
    } else {
      return { success: false, message: result.message || 'OAuth回调处理失败' };
    }
  } catch (error) {
    console.error('OAuth回调请求失败:', error);
    return { success: false, message: '网络错误，请稍后重试' };
  }
}