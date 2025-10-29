import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api';
import type {
  JobProfileDetail,
  CreateJobProfileReq,
  UpdateJobProfileReq,
  ListJobProfilesResp,
  SearchJobProfilesResp,
  JobProfileQueryParams,
  JobSkillMeta,
  SkillMetaQueryParams,
  ListSkillMetaResp,
} from '@/types/job-profile';

// 获取岗位画像列表
export const listJobProfiles = async (
  params?: JobProfileQueryParams
): Promise<ListJobProfilesResp> => {
  const queryParams = new URLSearchParams();

  if (params?.page) {
    queryParams.append('page', params.page.toString());
  }
  if (params?.page_size) {
    queryParams.append('page_size', params.page_size.toString());
  }
  if (params?.keyword) {
    queryParams.append('keyword', params.keyword);
  }
  if (params?.department_id) {
    queryParams.append('department_id', params.department_id);
  }
  if (params?.skill_ids && params.skill_ids.length > 0) {
    params.skill_ids.forEach((id) => queryParams.append('skill_ids', id));
  }
  if (params?.locations && params.locations.length > 0) {
    params.locations.forEach((location) =>
      queryParams.append('locations', location)
    );
  }
  if (params?.salary_min) {
    queryParams.append('salary_min', params.salary_min.toString());
  }
  if (params?.salary_max) {
    queryParams.append('salary_max', params.salary_max.toString());
  }
  if (params?.next_token) {
    queryParams.append('next_token', params.next_token);
  }

  const url = `/v1/job-profiles${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
  return await apiGet<ListJobProfilesResp>(url);
};

// 创建岗位画像
export const createJobProfile = async (
  data: CreateJobProfileReq
): Promise<JobProfileDetail> => {
  return await apiPost<JobProfileDetail>(
    '/v1/job-profiles',
    data as unknown as Record<string, unknown>
  );
};

// 解析岗位描述返回类型
interface ParseJobProfileResult {
  name?: string;
  location?: string;
  work_type?: string;
  salary_min?: number;
  salary_max?: number;
  education_requirements?: Array<{ education_type: string }>;
  experience_requirements?: Array<{
    experience_type: string;
    min_years?: number;
    ideal_years?: number;
  }>;
  industry_requirements?: Array<{ industry?: string; company_name?: string }>;
  skills?: Array<{
    skill_id?: string;
    skill_name?: string;
    name?: string;
    type?: string;
  }>;
  responsibilities?: Array<
    { responsibility?: string; description?: string } | string
  >;
}

// 解析岗位描述
export const parseJobProfile = async (
  description: string
): Promise<ParseJobProfileResult> => {
  return await apiPost<ParseJobProfileResult>('/v1/job-profiles/parse', {
    description,
  });
};

// 优化提示词返回类型
interface PolishPromptResult {
  polished_prompt: string;
  responsibility_tips?: string[];
  requirement_tips?: string[];
  bonus_tips?: string[];
}

// 优化提示词
export const polishPrompt = async (
  idea: string
): Promise<PolishPromptResult> => {
  return await apiPost<PolishPromptResult>('/v1/job-profiles/polish-prompt', {
    idea,
  });
};

// 根据提示词生成岗位画像
export const generateByPrompt = async (
  prompt: string
): Promise<JobProfileDetail> => {
  // apiPost已经返回data.data,所以这里的response就是后端返回的data内容
  const data = await apiPost<{
    profile: JobProfileDetail;
    description_markdown?: string;
  }>('/v1/job-profiles/generate-by-prompt', {
    prompt,
  });

  // 打印完整数据以便调试
  console.log('=== generateByPrompt 返回数据 ===', data);
  console.log('data.profile:', data.profile);
  console.log('data.description_markdown:', data.description_markdown);

  // 后端返回 { profile: {...}, description_markdown: "..." }
  // 需要将description_markdown合并到profile中
  console.log('使用 data.profile 并合并 description_markdown');
  return {
    ...data.profile,
    description_markdown:
      data.description_markdown || data.profile.description_markdown,
  };
};

// 搜索岗位画像
export const searchJobProfiles = async (
  params?: JobProfileQueryParams
): Promise<SearchJobProfilesResp> => {
  const queryParams = new URLSearchParams();

  if (params?.page) {
    queryParams.append('page', params.page.toString());
  }
  if (params?.page_size) {
    queryParams.append('page_size', params.page_size.toString());
  }
  if (params?.keyword) {
    queryParams.append('keyword', params.keyword);
  }
  if (params?.department_id) {
    queryParams.append('department_id', params.department_id);
  }
  if (params?.skill_ids && params.skill_ids.length > 0) {
    params.skill_ids.forEach((id) => queryParams.append('skill_ids', id));
  }
  if (params?.locations && params.locations.length > 0) {
    params.locations.forEach((location) =>
      queryParams.append('locations', location)
    );
  }
  if (params?.salary_min) {
    queryParams.append('salary_min', params.salary_min.toString());
  }
  if (params?.salary_max) {
    queryParams.append('salary_max', params.salary_max.toString());
  }
  if (params?.next_token) {
    queryParams.append('next_token', params.next_token);
  }

  const url = `/v1/job-profiles/search${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
  return await apiGet<SearchJobProfilesResp>(url);
};

// 获取岗位画像详情
export const getJobProfile = async (id: string): Promise<JobProfileDetail> => {
  return await apiGet<JobProfileDetail>(`/v1/job-profiles/${id}`);
};

// 更新岗位画像
export const updateJobProfile = async (
  id: string,
  data: UpdateJobProfileReq
): Promise<JobProfileDetail> => {
  // 确保请求数据中包含id
  const requestData = { ...data, id };
  return await apiPut<JobProfileDetail>(`/v1/job-profiles/${id}`, requestData);
};

// 删除岗位画像
export const deleteJobProfile = async (id: string): Promise<void> => {
  return await apiDelete<void>(`/v1/job-profiles/${id}`);
};

// 获取技能元数据列表
export const listJobSkillMeta = async (
  params?: SkillMetaQueryParams
): Promise<ListSkillMetaResp> => {
  const queryParams = new URLSearchParams();

  if (params?.page) {
    queryParams.append('page', params.page.toString());
  }
  if (params?.size) {
    queryParams.append('size', params.size.toString());
  }
  if (params?.keyword) {
    queryParams.append('keyword', params.keyword);
  }
  if (params?.next_token) {
    queryParams.append('next_token', params.next_token);
  }

  const url = `/v1/job-skills/meta${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
  return await apiGet<ListSkillMetaResp>(url);
};

// 获取技能元数据列表（向后兼容的简化版本）
export const listJobSkillMetaSimple = async (): Promise<{
  skills: JobSkillMeta[];
}> => {
  const response = await listJobSkillMeta();
  return { skills: response.items };
};

// 创建技能元数据
export const createJobSkillMeta = async (data: {
  name: string;
}): Promise<JobSkillMeta> => {
  return await apiPost<JobSkillMeta>('/v1/job-skills/meta', data);
};

// 更新技能元数据
export const updateJobSkillMeta = async (
  id: string,
  data: { name: string }
): Promise<JobSkillMeta> => {
  return await apiPut<JobSkillMeta>(`/v1/job-skills/meta/${id}`, data);
};

// 删除技能元数据
export const deleteJobSkillMeta = async (id: string): Promise<void> => {
  return await apiDelete<void>(`/v1/job-skills/meta/${id}`);
};
