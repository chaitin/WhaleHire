import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api';

/**
 * 简历邮箱配置接口类型定义（匹配后端返回字段）
 */
export interface ResumeMailboxSetting {
  id: string;
  name: string; // 任务名称
  email_address: string; // 邮箱地址
  protocol: string; // 邮箱协议（imap/pop3）
  host: string; // 邮箱服务器地址
  port: number; // 端口
  use_ssl: boolean; // 是否使用SSL
  folder?: string; // IMAP文件夹
  auth_type: string; // 认证类型
  encrypted_credential: Record<string, string>; // 加密凭证
  uploader_id: string; // 上传人ID
  job_profile_ids?: string[]; // 岗位画像ID列表
  sync_interval_minutes?: number | null; // 同步频率（分钟）
  status: 'enabled' | 'disabled'; // 状态
  last_synced_at?: string | null; // 最后同步时间（ISO字符串）
  last_error?: string; // 最后错误信息
  retry_count: number; // 重试次数
  created_at: string; // 创建时间（ISO字符串）
  updated_at: string; // 更新时间（ISO字符串）
  // 关联数据（通过后端join或额外查询获取）
  uploader_name?: string; // 上传人姓名
  job_profile_names?: string[]; // 岗位名称列表
  total_resumes?: number; // 已同步简历总数
}

/**
 * 创建简历邮箱配置的请求参数（按照后端实际字段）
 */
export interface CreateResumeMailboxSettingReq {
  name: string; // 任务名称
  email_address: string; // 邮箱地址
  auth_type: string; // 认证类型，默认 "password"
  encrypted_credential: Record<string, string>; // 加密凭证对象，如 {"password":"授权码"}
  protocol: string; // 协议类型（IMAP等）
  host: string; // 服务器地址
  port: number; // 端口
  use_ssl: boolean; // 是否使用SSL，默认 true
  sync_interval_minutes: number; // 同步频率（分钟）
  job_profile_ids: string[]; // 岗位画像ID数组
  uploader_id: string; // 上传者ID（必需）
  status: 'enabled' | 'disabled'; // 状态：enabled 或 disabled
}

/**
 * 更新简历邮箱配置的请求参数（按照后端实际字段）
 */
export interface UpdateResumeMailboxSettingReq {
  name?: string; // 任务名称
  email_address?: string; // 邮箱地址
  auth_type?: string; // 认证类型
  encrypted_credential?: Record<string, string>; // 加密凭证对象，如 {"password":"授权码"}
  protocol?: string; // 协议类型
  host?: string; // 服务器地址
  port?: number; // 端口
  use_ssl?: boolean; // 是否使用SSL
  sync_interval_minutes?: number; // 同步频率（分钟）
  job_profile_ids?: string[]; // 岗位画像ID数组
  status?: 'enabled' | 'disabled'; // 状态：enabled 或 disabled
}

/**
 * 列表响应
 */
export interface ListResumeMailboxSettingsResponse {
  items: ResumeMailboxSetting[];
  total_count: number;
  has_next_page: boolean;
  next_token?: string;
}

/**
 * 获取简历邮箱配置列表
 */
export const getResumeMailboxSettings = async (): Promise<
  ResumeMailboxSetting[]
> => {
  const response = await apiGet<ListResumeMailboxSettingsResponse>(
    '/v1/resume-mailbox-settings'
  );
  return response.items || [];
};

/**
 * 获取单个简历邮箱配置详情
 */
export const getResumeMailboxSetting = async (
  id: string
): Promise<ResumeMailboxSetting> => {
  return apiGet<ResumeMailboxSetting>(`/v1/resume-mailbox-settings/${id}`);
};

/**
 * 创建简历邮箱配置
 */
export const createResumeMailboxSetting = async (
  data: CreateResumeMailboxSettingReq
): Promise<ResumeMailboxSetting> => {
  return apiPost<ResumeMailboxSetting>(
    '/v1/resume-mailbox-settings',
    data as unknown as Record<string, unknown>
  );
};

/**
 * 更新简历邮箱配置
 */
export const updateResumeMailboxSetting = async (
  id: string,
  data: UpdateResumeMailboxSettingReq
): Promise<ResumeMailboxSetting> => {
  return apiPut<ResumeMailboxSetting>(
    `/v1/resume-mailbox-settings/${id}`,
    data as unknown as Record<string, unknown>
  );
};

/**
 * 删除简历邮箱配置
 */
export const deleteResumeMailboxSetting = async (id: string): Promise<void> => {
  return apiDelete<void>(`/v1/resume-mailbox-settings/${id}`);
};

/**
 * 手动触发同步
 */
export const syncResumeMailboxNow = async (id: string): Promise<void> => {
  return apiPost<void>(`/v1/resume-mailbox-settings/${id}/sync-now`, {});
};

/**
 * 邮箱统计单项数据
 */
export interface ResumeMailboxStatistic {
  id: string;
  mailbox_id: string;
  date: string; // 日期
  synced_emails: number; // 同步邮件数
  parsed_resumes: number; // 解析简历数
  failed_resumes: number; // 失败简历数
  skipped_attachments: number; // 跳过附件数
  last_sync_duration_ms: number; // 最后同步时长（毫秒）
  created_at: string;
  updated_at: string;
}

/**
 * 邮箱统计列表响应
 */
export interface ListResumeMailboxStatisticsResponse {
  items: ResumeMailboxStatistic[]; // 统计列表
  total_count: number; // 总数
  has_next_page: boolean; // 是否有下一页
  next_token?: string; // 下一页token
}

/**
 * 获取邮箱统计列表
 */
export const getMailboxStatistics = async (
  mailboxId?: string,
  dateFrom?: string,
  dateTo?: string,
  page?: number,
  pageSize?: number
): Promise<ListResumeMailboxStatisticsResponse> => {
  const queryParams = new URLSearchParams();

  if (mailboxId) queryParams.append('mailbox_id', mailboxId);
  if (dateFrom) queryParams.append('date_from', dateFrom);
  if (dateTo) queryParams.append('date_to', dateTo);
  if (page) queryParams.append('page', page.toString());
  if (pageSize) queryParams.append('page_size', pageSize.toString());

  const queryString = queryParams.toString();
  const url = queryString
    ? `/v1/resume-mailbox-statistics?${queryString}`
    : '/v1/resume-mailbox-statistics';

  return apiGet<ListResumeMailboxStatisticsResponse>(url);
};
