// 简历相关类型定义 - 根据swagger.json定义

// 简历状态枚举
export enum ResumeStatus {
  PENDING = 'pending',      // 已上传，等待解析
  PROCESSING = 'processing', // 正在解析中
  COMPLETED = 'completed',   // 解析完成
  FAILED = 'failed',        // 解析失败
  ARCHIVED = 'archived'     // 已归档
}

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
  user_id: string;
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

// 简历日志
export interface ResumeLog {
  id: string;
  resume_id: string;
  action: string;
  message: string;
  created_at: number;
  updated_at: number;
}

// 简历详细信息
export interface ResumeDetail extends Resume {
  experiences: ResumeExperience[];
  educations: ResumeEducation[];
  skills: ResumeSkill[];
  logs: ResumeLog[];
}

export type ResumeStatusFilter = 'all' | ResumeStatus;

export interface ResumeFilters {
  position?: string;
  status: ResumeStatusFilter;
  keywords: string;
}

export interface PaginationInfo {
  current: number;
  total: number;
  pageSize: number;
  hasNextPage: boolean;
  nextToken?: string;
}