import type { ApiResult } from './types'

// 管理员资料类型定义
export interface AdminProfile {
  id: string
  username: string
  created_at: number
  last_active_at: number
  status: 'active' | 'inactive'
  role: Role
}

// 管理员资料更新请求类型
export interface AdminProfileUpdateRequest {
  username: string
  password: string
  old_password: string
}

// 角色类型定义
export interface Role {
  id: number
  name: string
  description: string
}

/**
 * 获取管理员资料
 */
export async function getAdminProfile(): Promise<AdminProfile | null> {
  try {
    const response = await fetch('/api/v1/admin/profile', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (response.ok) {
      const result = await response.json()
      return result.data
    } else {
      console.error('获取管理员资料失败:', response.status)
      return null
    }
  } catch (error) {
    console.error('获取管理员资料请求失败:', error)
    return null
  }
}

/**
 * 更新管理员资料
 */
export async function updateAdminProfile(updateData: AdminProfileUpdateRequest): Promise<AdminProfile | null> {
  try {
    const response = await fetch('/api/v1/admin/profile', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(updateData),
    })

    if (response.ok) {
      const result = await response.json()
      return result.data
    } else {
      console.error('更新管理员资料失败:', response.status)
      return null
    }
  } catch (error) {
    console.error('更新管理员资料请求失败:', error)
    return null
  }
}

/**
 * 管理员退出登录
 */
export async function adminLogout(): Promise<ApiResult<void>> {
  try {
    const response = await fetch('/api/v1/admin/logout', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (response.ok) {
      return { success: true }
    } else {
      const result = await response.json()
      return { success: false, message: result.message || '退出登录失败' }
    }
  } catch (error) {
    console.error('管理员退出登录请求失败:', error)
    return { success: false, message: '网络错误，请稍后重试' }
  }
}

/**
 * 获取角色列表
 */
export async function getRoleList(): Promise<Role[] | null> {
  try {
    const response = await fetch('/api/v1/admin/role', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (response.ok) {
      const result = await response.json()
      return result.data
    } else {
      console.error('获取角色列表失败:', response.status)
      return null
    }
  } catch (error) {
    console.error('获取角色列表请求失败:', error)
    return null
  }
}

