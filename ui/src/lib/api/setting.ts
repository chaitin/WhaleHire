

import type { 
  ApiResult,
  Setting,
  UpdateSettingRequest,
} from './types';


/**
 * 获取系统设置
 */
export async function getSetting(): Promise<ApiResult<Setting>> {
  try {
    const response = await fetch('/api/v1/admin/setting', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    const result = await response.json();
    
    if (response.ok) {
      return { success: true, data: result.data };
    } else {
      return { success: false, message: result.message || '获取系统设置失败' };
    }
  } catch (error) {
    console.error('获取系统设置请求失败:', error);
    return { success: false, message: '网络错误，请稍后重试' };
  }
}

/**
 * 更新系统设置
 */
export async function updateSetting(settingData: UpdateSettingRequest): Promise<ApiResult<Setting>> {
  try {
    const response = await fetch('/api/v1/admin/setting', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(settingData),
    });
    
    const result = await response.json();
    
    if (response.ok) {
      return { success: true, data: result.data };
    } else {
      return { success: false, message: result.message || '更新系统设置失败' };
    }
  } catch (error) {
    console.error('更新系统设置请求失败:', error);
    return { success: false, message: '网络错误，请稍后重试' };
  }
}
