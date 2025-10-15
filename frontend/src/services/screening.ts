// 筛选任务相关API服务
import { apiGet, apiPost, apiDelete } from '@/lib/api';
import {
  CreateScreeningTaskReq,
  CreateScreeningTaskResp,
  GetTaskProgressResp,
  GetScreeningTaskResp,
  ListScreeningTasksResp,
  ListScreeningTasksParams,
  DeleteScreeningTaskResp,
  GetScreeningMetricsResp,
} from '@/types/screening';

/**
 * 创建筛选任务
 * POST /api/v1/screening/tasks
 */
export const createScreeningTask = async (
  params: CreateScreeningTaskReq
): Promise<CreateScreeningTaskResp> => {
  return apiPost<CreateScreeningTaskResp>(
    '/v1/screening/tasks',
    params as unknown as Record<string, unknown>
  );
};

/**
 * 启动筛选任务
 * POST /api/v1/screening/tasks/{id}/start
 */
export const startScreeningTask = async (taskId: string): Promise<void> => {
  return apiPost<void>(`/v1/screening/tasks/${taskId}/start`, {});
};

/**
 * 获取任务进度
 * GET /api/v1/screening/tasks/{id}/progress
 */
export const getTaskProgress = async (
  taskId: string
): Promise<GetTaskProgressResp> => {
  return apiGet<GetTaskProgressResp>(`/v1/screening/tasks/${taskId}/progress`);
};

/**
 * 获取筛选任务详情
 * GET /api/v1/screening/tasks/{id}
 */
export const getScreeningTask = async (
  taskId: string
): Promise<GetScreeningTaskResp> => {
  return apiGet<GetScreeningTaskResp>(`/v1/screening/tasks/${taskId}`);
};

/**
 * 获取筛选任务列表
 * GET /api/v1/screening/tasks
 */
export const listScreeningTasks = async (
  params: ListScreeningTasksParams = {}
): Promise<ListScreeningTasksResp> => {
  const queryParams = new URLSearchParams();

  if (params.page) queryParams.append('page', params.page.toString());
  if (params.size) queryParams.append('size', params.size.toString());
  if (params.next_token) queryParams.append('next_token', params.next_token);
  if (params.status && params.status !== 'all')
    queryParams.append('status', params.status);

  return apiGet<ListScreeningTasksResp>(
    `/v1/screening/tasks?${queryParams.toString()}`
  );
};

/**
 * 删除筛选任务
 * DELETE /api/v1/screening/tasks/{id}
 */
export const deleteScreeningTask = async (
  taskId: string
): Promise<DeleteScreeningTaskResp> => {
  return apiDelete<DeleteScreeningTaskResp>(`/v1/screening/tasks/${taskId}`);
};

/**
 * 获取筛选任务运行指标
 * GET /api/v1/screening/tasks/{id}/metrics
 */
export const getScreeningMetrics = async (
  taskId: string
): Promise<GetScreeningMetricsResp> => {
  return apiGet<GetScreeningMetricsResp>(
    `/v1/screening/tasks/${taskId}/metrics`
  );
};
