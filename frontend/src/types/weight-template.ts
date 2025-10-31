// 智能匹配权重模板类型定义

import { DimensionWeights } from './screening';

// 权重模板响应
export interface WeightTemplate {
  id: string; // 模板ID
  name: string; // 模板名称
  description?: string; // 模板描述
  weights: DimensionWeights; // 权重配置
  created_by: string; // 创建者ID
  created_at: string; // 创建时间
  updated_at: string; // 更新时间
}

// 创建权重模板请求
export interface CreateWeightTemplateReq {
  name: string; // 模板名称（必填，最大长度100）
  description?: string; // 模板描述（可选，最大长度500）
  weights: DimensionWeights; // 权重配置（必填）
}

// 更新权重模板请求
export interface UpdateWeightTemplateReq {
  name: string; // 模板名称（必填，最大长度100）
  description?: string; // 模板描述（可选，最大长度500）
  weights: DimensionWeights; // 权重配置（必填）
}

// 权重模板列表响应
export interface ListWeightTemplatesResp {
  items: WeightTemplate[]; // 模板列表
  page_info?: {
    total_count: number; // 总数
    next_token?: string; // 下一页标识
  };
}

// 权重模板查询参数
export interface WeightTemplateQueryParams {
  page?: number; // 分页，默认1
  size?: number; // 每页数量，默认10
  next_token?: string; // 下一页标识
  name?: string; // 模板名称（模糊匹配）
}
