/**
 * 操作日志类型定义
 */

// 操作类型枚举
export type OperationType =
  | 'create' // 创建操作
  | 'update' // 更新操作
  | 'delete' // 删除操作
  | 'system_setting' // 系统设置
  | 'login' // 登录操作
  | 'export'; // 导出操作

// 操作状态枚举
export type OperationStatus =
  | 'success' // 成功
  | 'warning' // 警告
  | 'failed'; // 失败

// 操作日志接口
export interface AuditLog {
  id: string; // 日志ID
  operator_name?: string; // 操作人姓名
  operator_id?: string; // 操作人ID
  operation_type: OperationType; // 操作类型
  resource_name?: string; // 资源名称（操作内容描述）
  resource_type?: string; // 资源类型
  resource_id?: string; // 资源ID
  ip?: string; // IP地址
  status?: string; // 状态
  error_message?: string; // 错误信息
  request_method?: string; // 请求方法
  request_path?: string; // 请求路径
  business_data?: Record<string, unknown>; // 业务数据
  created_at: number; // 创建时间（Unix时间戳）
}

// 操作日志列表请求参数
export interface ListAuditLogsRequest {
  page: number; // 页码
  size: number; // 每页数量
  operator_name?: string; // 操作人筛选
  operation_type?: OperationType; // 操作类型筛选
  operation_status?: OperationStatus; // 操作状态筛选
  start_time?: number; // 开始时间（Unix时间戳）
  end_time?: number; // 结束时间（Unix时间戳）
}

// 操作日志列表响应
export interface ListAuditLogsResponse {
  audit_logs: AuditLog[]; // 日志列表
  total_count: number; // 总数量
  has_next_page?: boolean; // 是否有下一页
  next_token?: string; // 下一页标记
}

// 获取操作日志详情请求参数
export interface GetAuditLogRequest {
  id: string; // 日志ID
}

// 获取操作日志详情响应
export type GetAuditLogResponse = AuditLog;

// 操作类型显示名称映射
export const OperationTypeLabels: Record<OperationType, string> = {
  create: '创建操作',
  update: '更新操作',
  delete: '删除操作',
  system_setting: '系统设置',
  login: '登录操作',
  export: '导出操作',
};

// 操作类型颜色映射
export const OperationTypeColors: Record<
  OperationType,
  { bg: string; text: string; icon: string }
> = {
  create: { bg: '#F0FFF4', text: '#00B42A', icon: '#00B42A' },
  update: { bg: '#FFF7E8', text: '#FF7D00', icon: '#FF7D00' },
  delete: { bg: '#FFF1F0', text: '#F53F3F', icon: '#F53F3F' },
  system_setting: { bg: '#E8F3FF', text: '#165DFF', icon: '#165DFF' },
  login: { bg: '#E8F3FF', text: '#165DFF', icon: '#165DFF' },
  export: { bg: '#E8F3FF', text: '#165DFF', icon: '#165DFF' },
};

// 操作状态显示名称映射
export const OperationStatusLabels: Record<OperationStatus, string> = {
  success: '成功',
  warning: '警告',
  failed: '失败',
};

// 操作状态颜色映射
export const OperationStatusColors: Record<
  OperationStatus,
  { bg: string; text: string; icon: string }
> = {
  success: { bg: '#F0FFF4', text: '#00B42A', icon: '#00B42A' },
  warning: { bg: '#FFF7E8', text: '#FF7D00', icon: '#FF7D00' },
  failed: { bg: '#FFF1F0', text: '#F53F3F', icon: '#F53F3F' },
};
