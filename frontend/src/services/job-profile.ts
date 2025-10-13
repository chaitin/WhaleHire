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
