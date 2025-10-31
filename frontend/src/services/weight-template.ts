// 智能匹配权重模板API服务
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api';
import type {
  WeightTemplate,
  CreateWeightTemplateReq,
  UpdateWeightTemplateReq,
  ListWeightTemplatesResp,
  WeightTemplateQueryParams,
} from '@/types/weight-template';

// 创建权重模板
export const createWeightTemplate = async (
  data: CreateWeightTemplateReq
): Promise<WeightTemplate> => {
  return await apiPost<WeightTemplate>(
    '/v1/screening/weights/templates',
    data as unknown as Record<string, unknown>
  );
};

// 获取权重模板列表
export const listWeightTemplates = async (
  params?: WeightTemplateQueryParams
): Promise<ListWeightTemplatesResp> => {
  const queryParams = new URLSearchParams();

  if (params?.page) {
    queryParams.append('page', params.page.toString());
  }
  if (params?.size) {
    queryParams.append('size', params.size.toString());
  }
  if (params?.next_token) {
    queryParams.append('next_token', params.next_token);
  }
  if (params?.name) {
    queryParams.append('name', params.name);
  }

  const url = `/v1/screening/weights/templates${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
  return await apiGet<ListWeightTemplatesResp>(url);
};

// 获取权重模板详情
export const getWeightTemplate = async (
  id: string
): Promise<WeightTemplate> => {
  return await apiGet<WeightTemplate>(`/v1/screening/weights/templates/${id}`);
};

// 更新权重模板
export const updateWeightTemplate = async (
  id: string,
  data: UpdateWeightTemplateReq
): Promise<WeightTemplate> => {
  return await apiPut<WeightTemplate>(
    `/v1/screening/weights/templates/${id}`,
    data as unknown as Record<string, unknown>
  );
};

// 删除权重模板
export const deleteWeightTemplate = async (id: string): Promise<void> => {
  return await apiDelete<void>(`/v1/screening/weights/templates/${id}`);
};
