import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api';

/**
 * 钉钉通知配置
 */
export interface DingTalkConfig {
  webhook_url: string; // 钉钉Webhook URL
  secret: string; // 加签密钥
  token?: string; // token值（从URL中提取access_token）
  keywords?: string; // 关键词（默认不传）
  max_retry?: number; // 最大重试次数（默认不传）
  timeout?: number; // 超时时间（默认不传）
}

/**
 * 通知配置接口类型定义
 */
export interface NotificationSetting {
  id: string;
  description: string; // 通知名称
  channel: 'dingtalk' | 'email'; // 通知方式
  dingtalk_config?: DingTalkConfig; // 钉钉配置
  enabled: boolean; // 是否启用
  created_at: string;
  updated_at: string;
}

/**
 * 创建通知配置的请求参数
 */
export interface CreateNotificationSettingReq {
  description: string; // 通知名称
  channel: 'dingtalk' | 'email'; // 通知方式
  dingtalk_config?: DingTalkConfig; // 钉钉配置
  enabled: boolean; // 是否启用
  [key: string]: unknown; // 索引签名
}

/**
 * 更新通知配置的请求参数
 */
export interface UpdateNotificationSettingReq {
  description?: string; // 通知名称
  channel?: 'dingtalk' | 'email'; // 通知方式
  dingtalk_config?: DingTalkConfig; // 钉钉配置
  enabled?: boolean; // 是否启用
  [key: string]: unknown; // 索引签名
}

/**
 * 获取通知配置列表
 */
export const getNotificationSettings = async (): Promise<
  NotificationSetting[]
> => {
  const response = await apiGet<{
    settings?: NotificationSetting[];
    items?: NotificationSetting[];
  }>('/v1/notification-settings');
  // 后端返回的是 settings 字段，兼容 items 字段
  return response.settings || response.items || [];
};

/**
 * 创建通知配置
 */
export const createNotificationSetting = async (
  data: CreateNotificationSettingReq
): Promise<NotificationSetting> => {
  return apiPost<NotificationSetting>('/v1/notification-settings', data);
};

/**
 * 更新通知配置
 */
export const updateNotificationSetting = async (
  id: string,
  data: UpdateNotificationSettingReq
): Promise<NotificationSetting> => {
  return apiPut<NotificationSetting>(`/v1/notification-settings/${id}`, data);
};

/**
 * 删除通知配置
 */
export const deleteNotificationSetting = async (id: string): Promise<void> => {
  return apiDelete<void>(`/v1/notification-settings/${id}`);
};
