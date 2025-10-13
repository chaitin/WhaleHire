// 岗位画像相关类型定义

// 技能类型枚举
export type SkillType = 'required' | 'bonus';

// 基础岗位画像信息
export interface JobProfile {
  id: string;
  name: string;
  description?: string;
  department: string;
  department_id: string;
  location?: string;
  salary_min?: number;
  salary_max?: number;
  status?: string;
  created_at: number;
  created_by?: string;
  creator_name?: string;
  updated_at: number;
}

// 教育要求
export interface JobEducationRequirement {
  id: string;
  job_id: string;
  min_degree: string;
  weight?: number;
}

// 教育要求输入
export interface JobEducationRequirementInput {
  id?: string;
  min_degree: string;
  weight?: number;
}

// 工作经验要求
export interface JobExperienceRequirement {
  id: string;
  job_id: string;
  min_years?: number;
  ideal_years?: number;
  weight?: number;
}

// 工作经验要求输入
export interface JobExperienceRequirementInput {
  id?: string;
  min_years?: number;
  ideal_years?: number;
  weight?: number;
}

// 行业要求
export interface JobIndustryRequirement {
  id: string;
  job_id: string;
  industry: string;
  company_name?: string;
  weight?: number;
}

// 行业要求输入
export interface JobIndustryRequirementInput {
  id?: string;
  industry: string;
  company_name?: string;
  weight?: number;
}

// 岗位职责
export interface JobResponsibility {
  id: string;
  job_id: string;
  responsibility: string;
  sort_order?: number;
}

// 岗位职责输入
export interface JobResponsibilityInput {
  id?: string;
  responsibility: string;
  sort_order?: number;
}

// 岗位技能
export interface JobSkill {
  id: string;
  job_id: string;
  skill_id: string;
  skill: string;
  type: SkillType;
  weight?: number;
}

// 岗位技能输入
export interface JobSkillInput {
  id?: string;
  skill_id: string;
  type: SkillType;
  weight?: number;
}

// 技能元数据
export interface JobSkillMeta {
  id: string;
  name: string;
  created_at: number;
  updated_at: number;
}

// 详细岗位画像信息
export interface JobProfileDetail {
  id: string;
  name: string;
  description?: string;
  department: string;
  department_id: string;
  location?: string;
  salary_min?: number;
  salary_max?: number;
  status?: string;
  created_at: number;
  created_by?: string;
  creator_name?: string;
  updated_at: number;
  education_requirements?: JobEducationRequirement[];
  experience_requirements?: JobExperienceRequirement[];
  industry_requirements?: JobIndustryRequirement[];
  responsibilities?: JobResponsibility[];
  skills?: JobSkill[];
}

// 创建岗位画像请求
export interface CreateJobProfileReq {
  name: string;
  department_id: string;
  description?: string;
  location?: string;
  salary_min?: number;
  salary_max?: number;
  status?: string;
  created_by?: string;
  education_requirements?: JobEducationRequirementInput[];
  experience_requirements?: JobExperienceRequirementInput[];
  industry_requirements?: JobIndustryRequirementInput[];
  responsibilities?: JobResponsibilityInput[];
  skills?: JobSkillInput[];
}

// 更新岗位画像请求
export interface UpdateJobProfileReq {
  id: string;
  name?: string;
  department_id?: string;
  description?: string;
  location?: string;
  salary_min?: number;
  salary_max?: number;
  status?: string;
  created_by?: string;
  education_requirements?: JobEducationRequirementInput[];
  experience_requirements?: JobExperienceRequirementInput[];
  industry_requirements?: JobIndustryRequirementInput[];
  responsibilities?: JobResponsibilityInput[];
  skills?: JobSkillInput[];
}

// 分页信息
export interface PageInfo {
  has_next_page: boolean;
  next_token?: string;
  total_count: number;
}

// 岗位画像列表响应
export interface ListJobProfilesResp {
  items: JobProfile[];
  page_info: PageInfo;
}

// 搜索岗位画像响应
export interface SearchJobProfilesResp {
  items: JobProfile[];
  page_info: PageInfo;
}

// 岗位画像查询参数
export interface JobProfileQueryParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  department_id?: string;
  skill_ids?: string[];
  locations?: string[];
  salary_min?: number;
  salary_max?: number;
  next_token?: string;
}

// 技能元数据查询参数
export interface SkillMetaQueryParams {
  page?: number;
  size?: number;
  keyword?: string;
  next_token?: string;
}

// 技能元数据列表响应
export interface ListSkillMetaResp {
  items: JobSkillMeta[];
  page_info: PageInfo;
}
