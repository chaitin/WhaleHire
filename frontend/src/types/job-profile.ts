// 岗位画像相关类型定义

// 技能类型枚举
export type SkillType = 'required' | 'bonus';

// 工作性质选项
export const WORK_TYPE_OPTIONS = [
  { value: 'full_time', label: '全职' },
  { value: 'part_time', label: '兼职' },
  { value: 'internship', label: '实习' },
  { value: 'outsourcing', label: '外包' },
];

// 学历要求选项
export const EDUCATION_TYPE_OPTIONS = [
  { value: 'unlimited', label: '不限' },
  { value: 'junior_college', label: '大专' },
  { value: 'bachelor', label: '本科' },
  { value: 'master', label: '硕士' },
  { value: 'doctor', label: '博士' },
];

// 工作经验选项
export const EXPERIENCE_TYPE_OPTIONS = [
  { value: 'unlimited', label: '不限' },
  { value: 'fresh_graduate', label: '应届毕业生' },
  { value: 'under_one_year', label: '1年以下' },
  { value: 'one_to_three_years', label: '1-3年' },
  { value: 'three_to_five_years', label: '3-5年' },
  { value: 'five_to_ten_years', label: '5-10年' },
  { value: 'over_ten_years', label: '10年以上' },
];

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
  education_type: string;
  weight?: number;
}

// 教育要求输入
export interface JobEducationRequirementInput {
  id?: string;
  education_type: string; // 学历要求: unlimited/junior_college/bachelor/master/doctor
}

// 工作经验要求
export interface JobExperienceRequirement {
  id: string;
  job_id: string;
  experience_type: string;
  min_years?: number;
  ideal_years?: number;
  weight?: number;
}

// 工作经验要求输入
export interface JobExperienceRequirementInput {
  id?: string;
  experience_type: string; // 工作经验: unlimited/fresh_graduate/under_one_year/one_to_three_years/three_to_five_years/five_to_ten_years/over_ten_years
  min_years?: number;
  ideal_years?: number;
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
  skill_name?: string; // AI生成时可能返回的字段
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
  description_markdown?: string; // AI生成的markdown格式描述
  department: string;
  department_id: string;
  work_type?: string; // 工作性质: full_time/part_time/internship/outsourcing
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
  work_type?: string; // 工作性质: full_time/part_time/internship/outsourcing
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
  work_type?: string; // 工作性质: full_time/part_time/internship/outsourcing
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
