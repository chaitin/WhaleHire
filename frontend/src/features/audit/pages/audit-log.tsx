/**
 * 操作日志页面组件
 * 严格按照Figma设计图实现界面布局、色彩方案、字体样式、间距和尺寸比例
 */

import { useState, useEffect, useCallback } from 'react';
import {
  Search,
  ChevronLeft,
  ChevronRight,
  Loader2,
  AlertCircle,
} from 'lucide-react';
import { Button } from '@/ui/button';
import { cn } from '@/lib/utils';
import { listAuditLogs } from '@/services/audit-log';
import type { AuditLog, OperationType } from '@/types/audit-log';
import { OperationTypeLabels } from '@/types/audit-log';

/**
 * 操作日志页面组件
 */
export default function AuditLogPage() {
  // 筛选条件状态
  const [operatorName, setOperatorName] = useState('');
  const [operationType, setOperationType] = useState<OperationType | ''>('');
  const [searchTerm, setSearchTerm] = useState(''); // 搜索框输入值

  // 日志列表状态
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 分页状态
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(6); // 根据Figma设计每页显示6条
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);

  // 获取操作日志列表
  const fetchLogs = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      console.log(
        `📋 获取操作日志列表 - 页码: ${currentPage}, 每页: ${pageSize}`
      );

      const params = {
        page: currentPage,
        size: pageSize,
        ...(operatorName && { operator_name: operatorName }),
        ...(operationType && { operation_type: operationType }),
      };

      const response = await listAuditLogs(params);

      // 检查数据结构
      const auditLogs = response?.audit_logs || [];
      const totalCount = response?.total_count || 0;
      const newTotalPages = Math.ceil(totalCount / pageSize);

      console.log(
        `📋 获取到 ${auditLogs.length} 条日志，总计: ${totalCount}，总页数: ${newTotalPages}`
      );

      // 调试：打印第一条日志的详细信息
      if (auditLogs.length > 0) {
        console.log('📋 第一条日志的详细数据:', auditLogs[0]);
        console.log('📋 resource_name:', auditLogs[0].resource_name);
        console.log('📋 所有字段:', Object.keys(auditLogs[0]));
      }

      setLogs(auditLogs);
      setTotalCount(totalCount);
      setTotalPages(newTotalPages);
    } catch (err) {
      setError('获取操作日志失败，请重试');
      console.error('获取操作日志失败:', err);
      setLogs([]);
      setTotalCount(0);
      setTotalPages(0);
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, operatorName, operationType]);

  // 页面加载时获取数据
  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  // 查询处理
  const handleSearch = () => {
    setOperatorName(searchTerm);
    setCurrentPage(1);
  };

  // 页面切换
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  // 生成页码数组
  const generatePageNumbers = () => {
    const pages = [];
    const safeTotalPages = totalPages || 1;
    const safeCurrentPage = currentPage || 1;

    const start = Math.max(1, safeCurrentPage - 2);
    const end = Math.min(safeTotalPages, safeCurrentPage + 2);

    for (let i = start; i <= end; i++) {
      pages.push(i);
    }

    return pages;
  };

  // 资源类型中文映射
  const resourceTypeLabels: Record<string, string> = {
    user: '用户',
    admin: '管理员',
    role: '角色',
    department: '部门',
    job_position: '职位',
    resume: '简历',
    screening: '筛选任务',
    setting: '系统设置',
    attachment: '附件',
    conversation: '对话',
    message: '消息',
  };

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

  // 获取操作内容描述
  const getOperationContent = (log: AuditLog): string => {
    // 优先使用 resource_name
    if (log.resource_name) {
      return log.resource_name;
    }

    // 如果没有 resource_name，尝试从 business_data 中提取
    if (log.business_data) {
      // 尝试获取一些常见的描述字段
      const description =
        log.business_data.description ||
        log.business_data.name ||
        log.business_data.title;
      if (description && typeof description === 'string') {
        return description;
      }
    }

    // 如果都没有，根据操作类型和资源类型生成描述
    if (log.resource_type) {
      const typeLabel =
        OperationTypeLabels[log.operation_type] || log.operation_type;
      const resourceLabel =
        resourceTypeLabels[log.resource_type] || log.resource_type;
      return `${typeLabel}${resourceLabel}`;
    }

    // 如果有请求路径，显示路径
    if (log.request_path) {
      return log.request_path;
    }

    // 最后的fallback
    return '-';
  };

  return (
    <div className="flex h-full flex-col gap-6 px-6 pb-6 pt-6 page-content">
      {/* 页面标题卡片 */}
      <div className="flex items-center justify-between rounded-xl border border-[#E5E7EB] bg-white px-6 py-5 shadow-sm transition-all duration-200 hover:shadow-md">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">操作日志</h1>
          <p className="text-sm text-muted-foreground">
            查看和管理系统中的所有操作日志记录
          </p>
        </div>
      </div>

      {/* 日志筛选卡片 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          {/* 左侧筛选器 */}
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <select
              value={operationType}
              onChange={(e) => {
                setOperationType(e.target.value as OperationType | '');
                setCurrentPage(1);
              }}
              className="h-11 w-full sm:w-40 rounded-lg border border-[#D1D5DB] bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="">全部类型</option>
              {Object.entries(OperationTypeLabels).map(([key, label]) => (
                <option key={key} value={key}>
                  {label}
                </option>
              ))}
            </select>
          </div>

          {/* 右侧搜索区域 */}
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <div className="relative sm:w-64">
              <Search className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                onKeyPress={(e) => {
                  if (e.key === 'Enter') {
                    handleSearch();
                  }
                }}
                placeholder="搜索操作人..."
                className="h-11 w-full rounded-lg border border-[#D1D5DB] bg-white pl-11 pr-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
              />
            </div>
            <Button
              type="button"
              className="h-11 rounded-lg px-6"
              onClick={handleSearch}
              disabled={loading}
            >
              搜索
            </Button>
          </div>
        </div>
      </div>

      {/* 操作日志列表卡片 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white shadow-sm overflow-hidden flex-1 flex flex-col">
        {/* 卡片头部 */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
          <h3 className="text-lg font-semibold text-gray-900">操作日志列表</h3>
        </div>

        {/* 表格内容 */}
        <div className="flex-1 overflow-auto">
          {/* 错误提示 */}
          {error && (
            <div className="flex items-center justify-center gap-2 py-8">
              <AlertCircle className="w-5 h-5 text-red-500" />
              <p className="text-sm text-red-500">{error}</p>
              <button
                onClick={fetchLogs}
                className="text-sm text-blue-500 hover:text-blue-600 ml-2 underline"
              >
                重试
              </button>
            </div>
          )}

          {/* 加载中 */}
          {loading && (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
              <span className="ml-2 text-sm text-gray-500">加载中...</span>
            </div>
          )}

          {/* 表格 */}
          {!loading && !error && (
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    序号
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    操作人
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    操作类型
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    操作内容
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    操作时间
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {logs.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-6 py-16">
                      <div className="flex flex-col items-center justify-center gap-2">
                        <svg
                          className="h-16 w-16 text-gray-300"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={1}
                            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                          />
                        </svg>
                        <div className="text-sm text-gray-500">
                          暂无简历数据
                        </div>
                        <div className="text-xs text-gray-400">
                          请尝试调整搜索条件或上传新的简历
                        </div>
                      </div>
                    </td>
                  </tr>
                ) : (
                  logs.map((log, index) => (
                    <tr key={log.id} className="hover:bg-gray-50">
                      {/* 序号 */}
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {(currentPage - 1) * pageSize + index + 1}
                      </td>

                      {/* 操作人 */}
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          {log.operator_name || '-'}
                        </div>
                      </td>

                      {/* 操作类型 */}
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm text-gray-500">
                          {OperationTypeLabels[log.operation_type] ||
                            log.operation_type ||
                            '未知操作'}
                        </span>
                      </td>

                      {/* 操作内容 */}
                      <td className="px-6 py-4 max-w-md">
                        <div className="text-sm text-gray-500 truncate">
                          {getOperationContent(log)}
                        </div>
                      </td>

                      {/* 操作时间 */}
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {formatTime(log.created_at)}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          )}
        </div>

        {/* 分页区域 */}
        {!loading && !error && logs.length > 0 && (
          <div className="border-t border-[#E5E7EB] px-6 py-6">
            <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
              <div className="flex items-center gap-4">
                <div className="text-sm text-[#6B7280]">
                  显示
                  <span className="text-[#6B7280]">
                    {(currentPage - 1) * pageSize + 1}
                  </span>{' '}
                  到{' '}
                  <span className="text-[#6B7280]">
                    {Math.min(currentPage * pageSize, totalCount)}
                  </span>{' '}
                  条，共 <span className="text-[#6B7280]">{totalCount}</span>{' '}
                  条结果
                </div>
              </div>

              {/* 分页按钮 */}
              {totalPages > 1 && (
                <div className="flex flex-wrap items-center gap-1">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange(currentPage - 1)}
                    disabled={currentPage === 1}
                    className={cn(
                      'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                      currentPage === 1 && 'opacity-50'
                    )}
                  >
                    <ChevronLeft className="h-4 w-4 text-[#6B7280]" />
                  </Button>

                  {generatePageNumbers().map((page) => (
                    <Button
                      key={page}
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(page)}
                      className={cn(
                        'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0 text-sm font-normal text-[#374151]',
                        page === currentPage &&
                          'border-[#7bb8ff] bg-[#7bb8ff] text-white'
                      )}
                    >
                      {page}
                    </Button>
                  ))}

                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange(currentPage + 1)}
                    disabled={currentPage >= totalPages}
                    className={cn(
                      'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                      currentPage >= totalPages && 'opacity-50'
                    )}
                  >
                    <ChevronRight className="h-4 w-4 text-[#6B7280]" />
                  </Button>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
