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
  GetScreeningResultResp,
  GetResumeProgressResp,
} from '@/types/screening';

// 创建筛选任务
export const createScreeningTask = async (
  params: CreateScreeningTaskReq
): Promise<CreateScreeningTaskResp> => {
  return apiPost<CreateScreeningTaskResp>(
    '/v1/screening/tasks',
    params as unknown as Record<string, unknown>
  );
};

// 启动筛选任务
export const startScreeningTask = async (taskId: string): Promise<void> => {
  return apiPost<void>(`/v1/screening/tasks/${taskId}/start`, {});
};

// 获取任务进度
export const getTaskProgress = async (
  taskId: string
): Promise<GetTaskProgressResp> => {
  return apiGet<GetTaskProgressResp>(`/v1/screening/tasks/${taskId}/progress`);
};

// 获取筛选任务详情
export const getScreeningTask = async (
  taskId: string
): Promise<GetScreeningTaskResp> => {
  return apiGet<GetScreeningTaskResp>(`/v1/screening/tasks/${taskId}`);
};

// 获取筛选任务列表
export const listScreeningTasks = async (
  params: ListScreeningTasksParams = {}
): Promise<ListScreeningTasksResp> => {
  const queryParams = new URLSearchParams();

  if (params.page) queryParams.append('page', params.page.toString());
  if (params.size) queryParams.append('size', params.size.toString());
  if (params.next_token) queryParams.append('next_token', params.next_token);
  if (params.status && params.status !== 'all')
    queryParams.append('status', params.status);

  const raw = await apiGet<ListTasksAPIResponse>(
    `/v1/screening/tasks?${queryParams.toString()}`
  );

  const normalizeToSeconds = (value: unknown): number => {
    if (typeof value === 'number') return value;
    if (typeof value === 'string') {
      const ms = Date.parse(value);
      if (!isNaN(ms)) return Math.floor(ms / 1000);
    }
    return 0;
  };

  const items: TaskItemRaw[] = Array.isArray(raw?.items)
    ? raw.items!
    : Array.isArray(raw?.tasks)
      ? raw.tasks!
      : [];

  const tasks = items.map((item) => ({
    id: item?.id ?? item?.task_id ?? '',
    task_id: item?.task_id ?? item?.id ?? '',
    job_position_id: item?.job_position_id ?? '',
    job_position_name: item?.job_position_name ?? '',
    status: item?.status ?? 'pending',
    resume_count:
      item?.resume_count ?? item?.resume_total ?? item?.resumeProcessed ?? 0,
    created_by: item?.created_by ?? '',
    creator_name: item?.creator_name ?? item?.created_by ?? '',
    created_at: normalizeToSeconds(item?.created_at),
  }));

  const total =
    raw?.total ??
    raw?.page_info?.total_count ??
    (Array.isArray(items) ? items.length : 0);
  const next_token = raw?.next_token ?? raw?.page_info?.next_token;

  return { tasks, total, next_token } as ListScreeningTasksResp;
};

// 删除筛选任务
export const deleteScreeningTask = async (
  taskId: string
): Promise<DeleteScreeningTaskResp> => {
  return apiDelete<DeleteScreeningTaskResp>(`/v1/screening/tasks/${taskId}`);
};

// 获取筛选任务运行指标
export const getScreeningMetrics = async (
  taskId: string
): Promise<GetScreeningMetricsResp> => {
  return apiGet<GetScreeningMetricsResp>(
    `/v1/screening/tasks/${taskId}/metrics`
  );
};

// 获取单个简历的匹配报告详情
export const getScreeningResult = async (
  taskId: string,
  resumeId: string
): Promise<GetScreeningResultResp> => {
  return apiGet<GetScreeningResultResp>(
    `/v1/screening/tasks/${taskId}/results/${resumeId}`
  );
};

// 获取单个简历的处理进度
export const getResumeProgress = async (
  taskId: string,
  resumeId: string
): Promise<GetResumeProgressResp> => {
  return apiGet<GetResumeProgressResp>(
    `/v1/screening/tasks/${taskId}/resumes/${resumeId}/progress`
  );
};

// 定义筛选任务列表原始响应类型，替代 any
type TaskItemRaw = {
  id?: string;
  task_id?: string;
  job_position_id?: string;
  job_position_name?: string;
  status?: import('@/types/screening').ScreeningTaskStatus;
  resume_count?: number;
  resume_total?: number;
  resumeProcessed?: number;
  created_by?: string;
  creator_name?: string;
  created_at?: number | string;
};

type PageInfoRaw = {
  total_count?: number;
  next_token?: string;
};

type ListTasksAPIResponse = {
  items?: TaskItemRaw[];
  tasks?: TaskItemRaw[];
  total?: number;
  page_info?: PageInfoRaw;
  next_token?: string;
};
