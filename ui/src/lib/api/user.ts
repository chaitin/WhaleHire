// 用户信息相关 API

import type { UserProfile, ApiResponse, UpdateUserRequest } from './types';

/**
 * 获取用户信息
 */
export async function getUserProfile(): Promise<UserProfile | null> {
  try {
    const response = await fetch('/api/v1/user/profile', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include', // 包含cookies用于身份验证
    });

    if (!response.ok) {
      console.error('获取用户信息失败:', response.status, response.statusText);
      return null;
    }

    const result: ApiResponse<UserProfile> = await response.json();
    
    if (result.code === 0 && result.data) {
      return result.data;
    } else {
      console.error('API返回错误:', result.message);
      return null;
    }
  } catch (error) {
    console.error('获取用户信息请求失败:', error);
    return null;
  }
}

/**
 * 获取用户头像URL，如果没有则返回默认头像
 */
export function getAvatarUrl(avatarUrl?: string): string {
  return avatarUrl || '/avatar.svg';
}

/**
 * 获取用户显示名称
 */
export function getDisplayName(user: UserProfile): string {
  return user.username || user.email || 'User';
}

/**
 * 用户注销登录
 */
export async function logout(): Promise<boolean> {
  try {
    const response = await fetch('/api/v1/user/logout', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include'
    });
    
    return response.ok;
  } catch (error) {
    console.error('注销请求失败:', error);
    return false;
  }
}

/**
 * 更新用户信息
 */
export async function updateUser(params: UpdateUserRequest): Promise<UserProfile | null> {
  try {
    const response = await fetch('/api/v1/user/update', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(params)
    });

    if (!response.ok) {
      console.error('更新用户信息失败:', response.status, response.statusText);
      return null;
    }

    const result: ApiResponse<UserProfile> = await response.json();
    
    if (result.code === 0 && result.data) {
      return result.data;
    } else {
      console.error('API返回错误:', result.message);
      return null;
    }
  } catch (error) {
    console.error('更新用户信息请求失败:', error);
    return null;
  }
}