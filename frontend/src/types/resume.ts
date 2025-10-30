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
  job_ids?: string[]; // 关联的岗位ID列表
  job_names?: string[]; // 关联的岗位名称列表
  job_positions?: JobApplication[]; // 关联的岗位信息
  source?: 'manual' | 'email'; // 来源：manual(手动) 或 email(邮件)
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
  university_type?: 'ordinary' | '211' | '985'; // 大学类型：普通学校、211工程、985工程（已废弃，使用university_types）
  university_types?: Array<
    'ordinary' | '211' | '985' | 'double_first_class' | 'qs_top100'
  >; // 大学类型列表
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

// 项目经历 - 根据swagger文档domain.ResumeProject定义
export interface ResumeProject {
  id: string;
  resume_id: string;
  name: string; // 项目名称
  project_name?: string; // 兼容旧字段
  description: string; // 项目描述
  start_date: string; // 开始日期
  end_date: string; // 结束日期
  achievements?: string; // 项目成就
  company?: string; // 所属公司
  project_type?: string; // 项目类型
  project_url?: string; // 项目链接
  responsibilities?: string; // 项目职责
  role?: string; // 担任角色
  technologies?: string; // 技术栈
  tech_stack?: string; // 兼容旧字段
  created_at: number;
  updated_at: number;
}

// 简历岗位申请关联信息 - 根据backend domain.JobApplication定义
export interface JobApplication {
  id: string;
  resume_id: string;
  job_position_id: string;
  job_title: string; // 岗位标题
  job_department: string; // 岗位部门
  status: string; // 申请状态
  source: string; // 申请来源
  notes: string; // 备注信息
  applied_at?: string; // 申请时间
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
  job_position_id?: string; // 单个岗位ID筛选（后端只支持单个）
  job_position_ids?: string[]; // 保留兼容性，但实际使用job_position_id
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
  job_ids?: string[]; // 关联的岗位ID列表（废弃，保留兼容）
  job_position_ids?: string[]; // 关联的岗位ID列表（新字段）
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
  university_type?: 'ordinary' | '211' | '985'; // 大学类型：普通学校、211工程、985工程
}

// 更新技能参数
export interface UpdateResumeSkill {
  id?: string; // 更新时必填
  action: 'create' | 'update' | 'delete'; // 操作类型
  skill_name?: string;
  level?: string;
  description?: string;
}

// 更新项目经历参数 - 根据swagger文档扩展
export interface UpdateResumeProject {
  id?: string; // 更新时必填
  action: 'create' | 'update' | 'delete'; // 操作类型
  name?: string; // 项目名称
  project_name?: string; // 兼容旧字段
  description?: string; // 项目描述
  start_date?: string; // 开始日期
  end_date?: string; // 结束日期
  achievements?: string; // 项目成就
  company?: string; // 所属公司
  project_type?: string; // 项目类型
  project_url?: string; // 项目链接
  responsibilities?: string; // 项目职责
  role?: string; // 担任角色
  technologies?: string; // 技术栈
  tech_stack?: string; // 兼容旧字段
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
