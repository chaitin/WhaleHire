/**
 * 操作日志API服务
 */

import { apiGet, apiDelete } from '@/lib/api';
import type {
  ListAuditLogsRequest,
  ListAuditLogsResponse,
  GetAuditLogRequest,
  GetAuditLogResponse,
} from '@/types/audit-log';

/**
 * 获取操作日志列表
 */
export const listAuditLogs = async (
  params: ListAuditLogsRequest
): Promise<ListAuditLogsResponse> => {
  return apiGet<ListAuditLogsResponse>(
    '/v1/audit/logs',
    params as unknown as Record<string, unknown>
  );
};

/**
 * 获取操作日志详情
 */
export const getAuditLog = async (
  params: GetAuditLogRequest
): Promise<GetAuditLogResponse> => {
  return apiGet<GetAuditLogResponse>(`/v1/audit/logs/${params.id}`);
};

/**
 * 删除操作日志
 */
export const deleteAuditLog = async (id: string): Promise<void> => {
  return apiDelete(`/v1/audit/logs/${id}`);
};
