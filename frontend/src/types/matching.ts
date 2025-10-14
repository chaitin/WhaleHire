/**
 * 匹配任务状态
 */
export type MatchingTaskStatus = 'completed' | 'in_progress' | 'failed';

/**
 * 匹配任务接口
 */
export interface MatchingTask {
  id: string;
  taskId: string; // 匹配任务ID，如 MT20230512001
  jobPositions: string[]; // 匹配岗位列表
  resumeCount: number; // 匹配简历数
  status: MatchingTaskStatus; // 任务状态
  creator: string; // 创建任务人
  createdAt: number; // 创建时间（Unix时间戳，秒）
}

/**
 * 匹配任务统计数据
 */
export interface MatchingStats {
  total: number; // 匹配总数
  inProgress: number; // 匹配中
  completed: number; // 匹配已完成
}

/**
 * 匹配任务筛选条件
 */
export interface MatchingFilters {
  position?: string; // 岗位筛选
  status?: string; // 状态筛选
  keywords?: string; // 关键词搜索
}

/**
 * 分页信息
 */
export interface Pagination {
  current: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

/**
 * 匹配结果-简历信息
 */
export interface MatchResultResume {
  id: string;
  name: string; // 候选人姓名
  age: number; // 年龄
  education: string; // 学历
  experience: string; // 工作经验
  avatar?: string; // 头像URL
}

/**
 * 匹配结果-岗位信息
 */
export interface MatchResultJob {
  id: string;
  title: string; // 岗位标题
  department: string; // 部门
  jobId: string; // 岗位ID
}

/**
 * 匹配结果项
 */
export interface MatchResult {
  id: string;
  resume: MatchResultResume;
  job: MatchResultJob;
  matchScore: number; // 匹配分数 (0-100)
}

/**
 * 匹配任务详情
 */
export interface MatchingTaskDetail extends MatchingTask {
  results: MatchResult[]; // 匹配结果列表
  resultsPagination: Pagination; // 结果分页信息
}
