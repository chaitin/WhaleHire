/**
 * 操作日志详情弹窗组件
 */

import { X, Circle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import type { AuditLog, OperationType } from '@/types/audit-log';
import { OperationTypeLabels, OperationTypeColors } from '@/types/audit-log';

interface AuditLogDetailModalProps {
  open: boolean;
  onClose: () => void;
  log: AuditLog | null;
}

/**
 * 操作日志详情弹窗组件
 */
export function AuditLogDetailModal({
  open,
  onClose,
  log,
}: AuditLogDetailModalProps) {
  if (!open || !log) return null;

  // 格式化时间
  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp * 1000);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  // 获取操作类型图标
  const getTypeIcon = (_type: OperationType) => {
    return <Circle className="w-3 h-3 fill-current" />;
  };

  // 复制日志信息
  const handleCopyLog = () => {
    const logInfo = `
日志ID: ${log.id}
操作人: ${log.operator_name}
操作类型: ${OperationTypeLabels[log.operation_type]}
操作时间: ${formatTime(log.created_at)}
IP地址: ${log.ip || '-'}
资源名称: ${log.resource_name || '-'}
资源类型: ${log.resource_type || '-'}
请求路径: ${log.request_path || '-'}

${log.business_data ? `业务数据:\n${JSON.stringify(log.business_data, null, 2)}` : ''}
    `.trim();

    navigator.clipboard.writeText(logInfo).then(() => {
      alert('日志信息已复制到剪贴板');
    });
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-lg shadow-xl w-[672px] max-h-[90vh] overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* 弹窗头部 */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <h3 className="text-lg font-semibold text-[#1D2129]">操作日志详情</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* 弹窗内容 */}
        <div className="px-6 py-4 overflow-y-auto max-h-[calc(90vh-140px)]">
          {/* 基本信息网格 */}
          <div className="grid grid-cols-2 gap-4 mb-6">
            {/* 左侧列 */}
            <div className="space-y-4">
              {/* 日志ID */}
              <div>
                <h4 className="text-sm font-medium text-[#6B7280] mb-1.5">
                  日志ID
                </h4>
                <p className="text-base font-medium text-[#1D2129]">{log.id}</p>
              </div>

              {/* 操作人 */}
              <div>
                <h4 className="text-sm font-medium text-[#6B7280] mb-1.5">
                  操作人
                </h4>
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-400 to-blue-600 flex items-center justify-center text-white font-medium text-sm">
                    {log.operator_name?.charAt(0) || '-'}
                  </div>
                  <div>
                    <div className="text-sm font-medium text-[#1D2129]">
                      {log.operator_name || '-'}
                    </div>
                  </div>
                </div>
              </div>

              {/* 操作类型 */}
              <div>
                <h4 className="text-sm font-medium text-[#6B7280] mb-1.5">
                  操作类型
                </h4>
                <span
                  className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm"
                  style={{
                    backgroundColor: OperationTypeColors[log.operation_type].bg,
                    color: OperationTypeColors[log.operation_type].text,
                  }}
                >
                  <span
                    style={{
                      color: OperationTypeColors[log.operation_type].icon,
                    }}
                  >
                    {getTypeIcon(log.operation_type)}
                  </span>
                  {OperationTypeLabels[log.operation_type]}
                </span>
              </div>

              {/* 资源类型 */}
              {log.resource_type && (
                <div>
                  <h4 className="text-sm font-medium text-[#6B7280] mb-1.5">
                    资源类型
                  </h4>
                  <p className="text-base text-[#1D2129]">
                    {log.resource_type}
                  </p>
                </div>
              )}
            </div>

            {/* 右侧列 */}
            <div className="space-y-4">
              {/* 操作时间 */}
              <div>
                <h4 className="text-sm font-medium text-[#6B7280] mb-1.5">
                  操作时间
                </h4>
                <p className="text-base text-[#1D2129]">
                  {formatTime(log.created_at)}
                </p>
              </div>

              {/* IP地址 */}
              {log.ip && (
                <div>
                  <h4 className="text-sm font-medium text-[#6B7280] mb-1.5">
                    IP地址
                  </h4>
                  <p className="text-base text-[#1D2129]">{log.ip}</p>
                </div>
              )}

              {/* 资源名称 */}
              {log.resource_name && (
                <div>
                  <h4 className="text-sm font-medium text-[#6B7280] mb-1.5">
                    资源名称
                  </h4>
                  <p className="text-base text-[#1D2129]">
                    {log.resource_name}
                  </p>
                </div>
              )}

              {/* 请求路径 */}
              {log.request_path && (
                <div>
                  <h4 className="text-sm font-medium text-[#6B7280] mb-1.5">
                    请求路径
                  </h4>
                  <p className="text-base text-[#1D2129] text-xs font-mono break-all">
                    {log.request_path}
                  </p>
                </div>
              )}
            </div>
          </div>

          {/* 状态信息 */}
          {log.status && (
            <div className="mb-6">
              <h4 className="text-sm font-medium text-[#6B7280] mb-2">状态</h4>
              <div className="bg-[#F9FAFB] border border-[#E5E7EB] rounded-md p-4">
                <p className="text-base leading-relaxed text-[#1D2129]">
                  {log.status}
                </p>
              </div>
            </div>
          )}

          {/* 错误信息 */}
          {log.error_message && (
            <div className="mb-6">
              <h4 className="text-sm font-medium text-[#6B7280] mb-2">
                错误信息
              </h4>
              <div className="bg-red-50 border border-red-200 rounded-md p-4">
                <p className="text-base leading-relaxed text-red-700">
                  {log.error_message}
                </p>
              </div>
            </div>
          )}

          {/* 业务数据 */}
          {log.business_data && Object.keys(log.business_data).length > 0 && (
            <div>
              <h4 className="text-sm font-medium text-[#6B7280] mb-2">
                业务数据
              </h4>
              <div className="bg-gray-900 rounded-lg p-4 overflow-x-auto">
                <pre className="text-sm text-green-400 font-mono">
                  {JSON.stringify(log.business_data, null, 2)}
                </pre>
              </div>
            </div>
          )}
        </div>

        {/* 弹窗底部 */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100">
          <Button variant="outline" className="px-6" onClick={onClose}>
            关闭
          </Button>
          <Button
            className="px-6 bg-[#165DFF] hover:bg-[#165DFF]/90"
            onClick={handleCopyLog}
          >
            复制日志信息
          </Button>
        </div>
      </div>
    </div>
  );
}
