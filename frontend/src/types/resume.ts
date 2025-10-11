// 简历相关类型定义 - 根据swagger.json定义

// ==================== 枚举类型 ====================

// 简历状态枚举
export enum ResumeStatus {
  PENDING = 'pending', // 已上传，等待解析
  PROCESSING = 'processing', // 正在解析中
  COMPLETED = 'completed', // 解析完成
  FAILED = 'failed', // 解析失败
  ARCHIVED = 'archived', // 已归档
}

// ==================== 实体类型 ====================

// 简历基础信息
export interface Resume {
  id: string;
  name: string;
  email: string;
  phone: string;
  gender?: string;
  birthday?: string;
  current_city?: string;
  highest_education?: string;
  years_experience?: number;
  status: ResumeStatus;
  resume_file_url?: string;
  parsed_at?: string;
  error_message?: string;
  uploader_id: string;
  uploader_name?: string; // 上传人姓名
  created_at: number;
  updated_at: number;
}

// 工作经历
export interface ResumeExperience {
  id: string;
  resume_id: string;
  company: string;
  position: string;
  title: string;
  start_date: string;
  end_date: string;
  description: string;
  created_at: number;
  updated_at: number;
}

// 教育背景
export interface ResumeEducation {
  id: string;
  resume_id: string;
  school: string;
  major: string;
  degree: string;
  start_date: string;
  end_date: string;
  created_at: number;
  updated_at: number;
}

// 技能信息
export interface ResumeSkill {
  id: string;
  resume_id: string;
  skill_name: string;
  level: string;
  description: string;
  created_at: number;
  updated_at: number;
}

// 项目经历
export interface ResumeProject {
  id: string;
  resume_id: string;
  project_name: string;
  description: string;
  tech_stack: string;
  start_date: string;
  end_date: string;
  created_at: number;
  updated_at: number;
}

// 简历日志
export interface ResumeLog {
  id: string;
  resume_id: string;
  action: string;
  message: string;
  created_at: number;
  updated_at: number;
}

// 简历详细信息（包含关联数据）
export interface ResumeDetail extends Resume {
  experiences: ResumeExperience[];
  educations: ResumeEducation[];
  skills: ResumeSkill[];
  projects: ResumeProject[];
  logs: ResumeLog[];
}

// ==================== API 请求/响应类型 ====================

// 简历列表查询参数 - 根据swagger定义
export interface ResumeListParams {
  page?: number;
  size?: number;
  next_token?: string;
  position?: string;
  status?: string;
  keywords?: string;
}

// 简历搜索参数
export interface ResumeSearchParams extends ResumeListParams {
  keywords: string;
}

// 简历列表响应 - 根据swagger ListResumeResp定义
export interface ResumeListResponse {
  resumes: Resume[];
  total_count: number;
  has_next_page: boolean;
  next_token?: string;
}

// 简历搜索响应 - 根据swagger SearchResumeResp定义
export interface ResumeSearchResponse {
  resumes: Resume[];
  total_count: number;
  has_next_page: boolean;
  next_token?: string;
}

// 简历更新参数 - 根据swagger UpdateResumeReq定义
export interface ResumeUpdateParams {
  name?: string;
  email?: string;
  phone?: string;
  gender?: string;
  birthday?: string;
  current_city?: string;
  highest_education?: string;
  years_experience?: number;
  experiences?: UpdateResumeExperience[];
  educations?: UpdateResumeEducation[];
  skills?: UpdateResumeSkill[];
  projects?: UpdateResumeProject[];
}

// 更新工作经历参数
export interface UpdateResumeExperience {
  id?: string; // 更新时必填
  action: 'create' | 'update' | 'delete'; // 操作类型
  company?: string;
  position?: string;
  title?: string;
  start_date?: string;
  end_date?: string;
  description?: string;
}

// 更新教育背景参数
export interface UpdateResumeEducation {
  id?: string; // 更新时必填
  action: 'create' | 'update' | 'delete'; // 操作类型
  school?: string;
  major?: string;
  degree?: string;
  start_date?: string;
  end_date?: string;
}

// 更新技能参数
export interface UpdateResumeSkill {
  id?: string; // 更新时必填
  action: 'create' | 'update' | 'delete'; // 操作类型
  skill_name?: string;
  level?: string;
  description?: string;
}

// 更新项目经历参数
export interface UpdateResumeProject {
  id?: string; // 更新时必填
  action: 'create' | 'update' | 'delete'; // 操作类型
  project_name?: string;
  description?: string;
  tech_stack?: string;
  start_date?: string;
  end_date?: string;
}

// 简历解析进度 - 根据swagger ResumeParseProgress定义
export interface ResumeParseProgress {
  resume_id: string;
  status: ResumeStatus;
  progress: number; // 0-100
  message: string;
  error_message?: string;
  started_at: string;
  completed_at?: string;
}

// ==================== 业务逻辑类型 ====================

// 简历状态筛选类型
export type ResumeStatusFilter = 'all' | ResumeStatus;

// 简历筛选条件
export interface ResumeFilters {
  position?: string;
  status: ResumeStatusFilter;
  keywords: string;
}

// 分页信息
export interface PaginationInfo {
  current: number;
  total: number;
  pageSize: number;
  hasNextPage: boolean;
  nextToken?: string;
}
