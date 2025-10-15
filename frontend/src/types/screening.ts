// 筛选任务相关类型定义

// 筛选任务状态
export type ScreeningTaskStatus =
  | 'pending'
  | 'in_progress'
  | 'completed'
  | 'failed';

// 维度权重配置
export interface DimensionWeights {
  skill?: number; // 技能权重
  responsibility?: number; // 职责权重
  experience?: number; // 经验权重
  education?: number; // 教育权重
  industry?: number; // 行业权重
  basic?: number; // 基本信息权重
}

// 创建筛选任务请求
export interface CreateScreeningTaskReq {
  job_position_id: string; // 职位ID
  resume_ids: string[]; // 简历ID列表
  dimension_weights?: DimensionWeights; // 维度权重配置
  notes?: string; // 任务备注
  created_by?: string; // 创建者ID
  agent_version?: string; // 代理版本
  llm_config?: Record<string, unknown>; // LLM配置
}

// 创建筛选任务响应
export interface CreateScreeningTaskResp {
  task_id: string; // 任务ID
}

// 任务进度响应
export interface GetTaskProgressResp {
  task_id: string; // 任务ID
  status: ScreeningTaskStatus; // 任务状态
  progress_percent: number; // 完成百分比 (0-100)
  resume_total: number; // 简历总数
  resume_processed: number; // 已处理简历数量
  resume_succeeded: number; // 处理成功简历数量
  resume_failed: number; // 处理失败简历数量
  started_at?: string; // 开始时间
  estimated_finish?: string; // 预计完成时间
}

// 筛选任务简历信息
export interface ScreeningTaskResume {
  resume_id: string; // 简历ID
  resume_name: string; // 简历名称
  status: ScreeningTaskStatus; // 处理状态
  score?: number; // 匹配分数
  match_reason?: string; // 匹配原因
  created_at: number; // 创建时间
  updated_at: number; // 更新时间
}

// 筛选任务运行指标
export interface ScreeningRunMetric {
  task_id: string; // 任务ID
  total_duration: number; // 总耗时(秒)
  avg_resume_duration: number; // 单份简历平均耗时(秒)
  success_rate: number; // 成功率
  llm_calls: number; // LLM调用次数
  llm_tokens: number; // LLM Token消耗
}

// 筛选任务基本信息
export interface ScreeningTask {
  id: string; // 任务ID
  task_id: string; // 任务编号
  job_position_id: string; // 职位ID
  job_position_name: string; // 职位名称
  status: ScreeningTaskStatus; // 任务状态
  dimension_weights: DimensionWeights; // 维度权重
  notes?: string; // 备注
  created_by: string; // 创建者
  created_at: number; // 创建时间
  updated_at: number; // 更新时间
  started_at?: number; // 开始时间
  finished_at?: number; // 完成时间
}

// 获取筛选任务详情响应
export interface GetScreeningTaskResp {
  task: ScreeningTask; // 任务基本信息
  resumes: ScreeningTaskResume[]; // 简历列表
  metrics?: ScreeningRunMetric; // 运行指标
}

// 筛选任务列表项
export interface ScreeningTaskItem {
  id: string;
  task_id: string;
  job_position_id: string;
  job_position_name: string;
  status: ScreeningTaskStatus;
  resume_count: number;
  created_by: string;
  created_at: number;
}

// 筛选任务列表响应
export interface ListScreeningTasksResp {
  tasks: ScreeningTaskItem[];
  total: number;
  next_token?: string;
}

// 筛选任务列表参数
export interface ListScreeningTasksParams {
  page?: number;
  size?: number;
  next_token?: string;
  status?: string;
}

// 删除筛选任务响应
export interface DeleteScreeningTaskResp {
  success: boolean;
}

// 获取筛选任务指标响应
export interface GetScreeningMetricsResp {
  metrics: ScreeningRunMetric;
}
