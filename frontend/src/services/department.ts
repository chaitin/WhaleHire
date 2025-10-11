// 部门管理相关API服务
import { apiGet, apiPost, apiPut, apiDelete, debugLog } from '@/lib/api';

// 分页信息接口 - 根据swagger定义
export interface PageInfo {
  page: number;
  size: number;
  total: number;
  has_next_page: boolean;
}

// 部门数据接口 - 根据swagger domain.Department定义
export interface Department {
  id: string;
  name: string;
  description?: string;
  parent_id?: string;
  created_at: number;
  updated_at: number;
}

// 创建部门请求参数 - 根据swagger domain.CreateDepartmentReq定义
export interface CreateDepartmentRequest {
  name: string;
  description?: string;
  parent_id?: string;
}

// 更新部门请求参数 - 根据swagger domain.UpdateDepartmentReq定义
export interface UpdateDepartmentRequest {
  name?: string;
  description?: string;
  parent_id?: string;
}

// 部门列表响应 - 根据swagger domain.ListDepartmentResp定义
export interface ListDepartmentResponse {
  items: Department[];
  page_info: PageInfo;
}

// 获取部门列表查询参数
export interface ListDepartmentParams {
  page?: number;
  size?: number;
  keyword?: string;
  parent_id?: string;
}

/**
 * 获取部门列表
 * GET /api/v1/departments
 */
export const listDepartments = async (
  params?: ListDepartmentParams
): Promise<ListDepartmentResponse> => {
  debugLog('📋 获取部门列表...');
  try {
    const response = await apiGet<ListDepartmentResponse>(
      '/v1/departments',
      params as Record<string, unknown>
    );
    debugLog('📋 部门列表获取成功:', response);
    return response;
  } catch (error) {
    console.error('📋 获取部门列表失败:', error);
    throw error;
  }
};

/**
 * 创建部门
 * POST /api/v1/departments
 */
export const createDepartment = async (
  data: CreateDepartmentRequest
): Promise<Department> => {
  debugLog('📋 创建部门...');
  try {
    const response = await apiPost<Department>(
      '/v1/departments',
      data as unknown as Record<string, unknown>
    );
    debugLog('📋 部门创建成功:', response);
    return response;
  } catch (error) {
    console.error('📋 创建部门失败:', error);
    throw error;
  }
};

/**
 * 获取部门详情
 * GET /api/v1/departments/{id}
 */
export const getDepartment = async (id: string): Promise<Department> => {
  debugLog('📋 获取部门详情...');
  try {
    const response = await apiGet<Department>(`/v1/departments/${id}`);
    debugLog('📋 部门详情获取成功:', response);
    return response;
  } catch (error) {
    console.error('📋 获取部门详情失败:', error);
    throw error;
  }
};

/**
 * 更新部门
 * PUT /api/v1/departments/{id}
 */
export const updateDepartment = async (
  id: string,
  data: UpdateDepartmentRequest
): Promise<Department> => {
  debugLog('📋 更新部门...');
  try {
    const response = await apiPut<Department>(
      `/v1/departments/${id}`,
      data as Record<string, unknown>
    );
    debugLog('📋 部门更新成功:', response);
    return response;
  } catch (error) {
    console.error('📋 更新部门失败:', error);
    throw error;
  }
};

/**
 * 删除部门
 * DELETE /api/v1/departments/{id}
 */
export const deleteDepartment = async (id: string): Promise<void> => {
  debugLog('📋 删除部门...');
  try {
    await apiDelete(`/v1/departments/${id}`);
    debugLog('📋 部门删除成功');
  } catch (error) {
    console.error('📋 删除部门失败:', error);
    throw error;
  }
};
