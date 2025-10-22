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
  creator_name?: string; // 创建者名称
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
  creator_name?: string;
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

// 匹配等级
export type MatchLevel = 'excellent' | 'good' | 'fair' | 'poor';

// 基本信息匹配详情
export interface BasicMatchDetail {
  score: number;
  sub_scores: Record<string, number>;
  evidence: string[];
  notes: string;
}

// 技能匹配详情
export interface SkillMatchDetail {
  score: number;
  matched_skills: MatchedSkill[];
  missing_skills: JobSkill[];
  extra_skills: string[];
  llm_analysis?: SkillLLMAnalysis;
}

// 匹配的技能
export interface MatchedSkill {
  job_skill_id: string;
  resume_skill_id?: string;
  match_type: string;
  llm_score: number;
  proficiency_gap: number;
  score: number;
  llm_analysis?: SkillItemAnalysis;
}

// 职位技能
export interface JobSkill {
  id: string;
  name: string;
  required_level?: string;
  priority?: string;
}

// 技能大模型分析
export interface SkillLLMAnalysis {
  overall_match: number;
  technical_fit: number;
  learning_curve: string;
  strength_areas: string[];
  gap_areas: string[];
  recommendations: string[];
  analysis_detail: string;
}

// 技能项分析
export interface SkillItemAnalysis {
  match_level: string;
  match_percentage: number;
  proficiency_gap: string;
  transferability: string;
  learning_effort: string;
  match_reason: string;
}

// 职责匹配详情
export interface ResponsibilityMatchDetail {
  score: number;
  matched_responsibilities: MatchedResponsibility[];
  unmatched_responsibilities: JobResponsibility[];
  relevant_experiences: string[];
}

// 匹配的职责
export interface MatchedResponsibility {
  job_responsibility_id: string;
  resume_experience_id?: string;
  llm_analysis?: LLMMatchAnalysis;
  match_score: number;
  match_reason: string;
}

// 职位职责
export interface JobResponsibility {
  id: string;
  description: string;
  priority?: string;
}

// 大模型匹配分析
export interface LLMMatchAnalysis {
  match_level: string;
  match_percentage: number;
  strength_points: string[];
  weak_points: string[];
  recommended_actions: string[];
  analysis_detail: string;
}

// 工作经验匹配详情
export interface ExperienceMatchDetail {
  score: number;
  years_match?: YearsMatchInfo;
  position_matches?: PositionMatchInfo[];
  industry_matches?: IndustryMatchInfo[];
}

// 年限匹配信息
export interface YearsMatchInfo {
  required_years: number;
  actual_years: number;
  score: number;
  gap: number;
}

// 职位匹配信息
export interface PositionMatchInfo {
  resume_experience_id: string;
  position: string;
  relevance: number;
  score: number;
}

// 行业匹配信息
export interface IndustryMatchInfo {
  resume_experience_id: string;
  company: string;
  industry: string;
  relevance: number;
  score: number;
}

// 教育背景匹配详情
export interface EducationMatchDetail {
  score: number;
  degree_match?: DegreeMatchInfo;
  major_matches?: MajorMatchInfo[];
  school_matches?: SchoolMatchInfo[];
}

// 学历匹配信息
export interface DegreeMatchInfo {
  required_degree: string;
  actual_degree: string;
  score: number;
  meets: boolean;
}

// 专业匹配信息
export interface MajorMatchInfo {
  resume_education_id: string;
  major: string;
  relevance: number;
  score: number;
}

// 学校匹配信息
export interface SchoolMatchInfo {
  resume_education_id: string;
  school: string;
  reputation: number;
  score: number;
}

// 行业背景匹配详情
export interface IndustryMatchDetail {
  score: number;
  industry_matches?: IndustryMatchInfo[];
  company_matches?: CompanyMatchInfo[];
  overall_analysis?: string; // 整体分析文本
}

// 公司匹配信息
export interface CompanyMatchInfo {
  resume_experience_id: string;
  company: string;
  target_company: string;
  score: number;
  is_exact: boolean;
}

// 筛选结果详情
export interface ScreeningResult {
  task_id: string;
  job_position_id: string;
  resume_id: string;
  overall_score: number;
  match_level: MatchLevel;
  dimension_scores?: Record<string, number>;
  basic_detail?: BasicMatchDetail;
  education_detail?: EducationMatchDetail;
  experience_detail?: ExperienceMatchDetail;
  industry_detail?: IndustryMatchDetail;
  responsibility_detail?: ResponsibilityMatchDetail;
  skill_detail?: SkillMatchDetail;
  recommendations?: string[];
  trace_id?: string;
  runtime_metadata?: Record<string, unknown>;
  sub_agent_versions?: Record<string, unknown>;
  matched_at: string;
  created_at: string;
  updated_at: string;
}

// 获取筛选结果响应
export interface GetScreeningResultResp {
  result: ScreeningResult;
}

// 获取筛选任务进度响应
export interface GetResumeProgressResp {
  task_id: string; // 任务ID
  resume_id: string; // 简历ID
  status: ScreeningTaskStatus; // 状态：pending/processing/completed/failed
  progress_percent?: number; // 完成百分比 (0-100)
  stage?: string; // 当前阶段描述
  error_message?: string; // 错误信息
  updated_at?: string; // 最近更新时间
}
