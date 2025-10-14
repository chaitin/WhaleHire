import { useState, useEffect } from 'react';
import { User, ChevronLeft, ChevronRight, FileText, Download } from 'lucide-react';
import {
  Sheet,
  SheetContent,
} from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { MatchingTaskDetail } from '@/types/matching';
import { getMockMatchingTaskDetail } from '@/data/mockMatchingData';

interface MatchingDetailDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  taskId: string | null;
}

export function MatchingDetailDrawer({
  open,
  onOpenChange,
  taskId,
}: MatchingDetailDrawerProps) {
  const [taskDetail, setTaskDetail] = useState<MatchingTaskDetail | null>(null);
  const [currentPage, setCurrentPage] = useState(1);

  // 加载任务详情
  useEffect(() => {
    if (taskId && open) {
      const detail = getMockMatchingTaskDetail(taskId, currentPage, 3);
      setTaskDetail(detail);
    }
  }, [taskId, open, currentPage]);

  // 重置页码当打开新任务时
  useEffect(() => {
    if (open && taskId) {
      setCurrentPage(1);
    }
  }, [open, taskId]);

  if (!taskDetail) return null;

  // 获取状态样式
  const getStatusStyle = (status: string) => {
    switch (status) {
      case 'completed':
        return {
          bg: 'bg-green-50',
          text: 'text-green-700',
          label: '已完成',
        };
      case 'in_progress':
        return {
          bg: 'bg-orange-50',
          text: 'text-orange-700',
          label: '进行中',
        };
      case 'failed':
        return {
          bg: 'bg-red-50',
          text: 'text-red-700',
          label: '失败',
        };
      default:
        return {
          bg: 'bg-gray-50',
          text: 'text-gray-700',
          label: '未知',
        };
    }
  };

  const statusStyle = getStatusStyle(taskDetail.status);

  // 处理分页
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  // 渲染分页按钮
  const renderPaginationButtons = () => {
    const pages = [];
    const totalPages = taskDetail.resultsPagination.totalPages;
    const maxVisiblePages = 5;

    let startPage = Math.max(1, currentPage - 2);
    let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);

    if (endPage - startPage < maxVisiblePages - 1) {
      startPage = Math.max(1, endPage - maxVisiblePages + 1);
    }

    for (let i = startPage; i <= endPage; i++) {
      pages.push(
        <Button
          key={i}
          variant="outline"
          size="sm"
          onClick={() => handlePageChange(i)}
          className={cn(
            'h-[30px] w-[33px] rounded-md border border-[#C9CDD4] bg-white p-0 text-sm text-[#374151] hover:border-[#10B981] hover:text-[#10B981]',
            i === currentPage && 'border-[#10B981] bg-[#10B981] text-white hover:bg-[#10B981] hover:text-white'
          )}
        >
          {i}
        </Button>
      );
    }

    return pages;
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-[960px] p-0 overflow-y-auto">
        {/* 头部 */}
        <div className="border-b border-[#E5E6EB] px-6 py-4">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <div className="flex items-center gap-3 mb-2">
                <h3 className="text-lg font-medium text-[#1D2129]">
                  高级前端工程师匹配任务 #{taskDetail.taskId}
                </h3>
                <span
                  className={cn(
                    'inline-flex items-center rounded-full px-3 py-1 text-xs font-medium',
                    statusStyle.bg,
                    statusStyle.text
                  )}
                >
                  {statusStyle.label}
                </span>
              </div>
              <p className="text-sm text-[#4E5969]">
                创建时间: {new Date(taskDetail.createdAt * 1000).toLocaleString('zh-CN', {
                  year: 'numeric',
                  month: '2-digit',
                  day: '2-digit',
                  hour: '2-digit',
                  minute: '2-digit',
                  second: '2-digit',
                  hour12: false,
                }).replace(/\//g, '-')}
              </p>
            </div>
          </div>
        </div>

        {/* 匹配结果 */}
        <div className="border-b border-[#E5E6EB] px-6 py-6">
          <h4 className="text-lg font-medium text-[#1D2129] mb-6">匹配结果</h4>

          {/* 表格 */}
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-[#F9FAFB]">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase tracking-wider text-[#4E5969]">
                    简历信息
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase tracking-wider text-[#4E5969]">
                    岗位信息
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase tracking-wider text-[#4E5969]">
                    <div className="flex items-center gap-1">
                      <span>匹配分数</span>
                      <svg
                        width="13"
                        height="16"
                        viewBox="0 0 13 16"
                        fill="none"
                        className="text-[#4E5969]"
                      >
                        <path
                          d="M6.5 0L12.1292 8H0.870835L6.5 0Z"
                          fill="currentColor"
                          opacity="0.3"
                        />
                        <path
                          d="M6.5 16L0.870835 8H12.1292L6.5 16Z"
                          fill="currentColor"
                        />
                      </svg>
                    </div>
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-medium uppercase tracking-wider text-[#4E5969]">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#E5E6EB] bg-white">
                {taskDetail.results.map((result) => (
                  <tr key={result.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[#D1FAE5]">
                          <User className="h-5 w-5 text-[#10B981]" />
                        </div>
                        <div>
                          <div className="text-sm font-medium text-[#1D2129]">
                            {result.resume.name}
                          </div>
                          <div className="text-sm text-[#4E5969]">
                            {result.resume.age}岁 · {result.resume.education} · {result.resume.experience}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div>
                        <div className="text-sm font-medium text-[#1D2129] mb-1">
                          {result.job.title}
                        </div>
                        <div className="text-sm text-[#4E5969]">
                          {result.job.department} · ID: {result.job.jobId}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <FileText className="h-4 w-4 text-[#10B981]" />
                        <span className="text-sm font-medium text-[#1D2129]">
                          {result.matchScore}分
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <div className="flex items-center justify-end gap-4">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-[#10B981] hover:text-[#10B981]/80 hover:bg-[#D1FAE5]/50"
                          onClick={() => console.log('查看报告', result.id)}
                          title="查看报告"
                        >
                          <FileText className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-[#4E5969] hover:text-[#4E5969]/80 hover:bg-gray-100"
                          onClick={() => console.log('下载', result.id)}
                          title="下载"
                        >
                          <Download className="h-4 w-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* 分页 */}
          <div className="mt-6 flex items-center justify-between">
            <div className="text-sm text-[#1D2129]">
              显示第{' '}
              <span className="font-normal text-[#1D2129]">
                {(currentPage - 1) * 3 + 1}
              </span>{' '}
              到{' '}
              <span className="font-normal text-[#1D2129]">
                {Math.min(currentPage * 3, taskDetail.resultsPagination.total)}
              </span>{' '}
              条，共{' '}
              <span className="font-normal text-[#1D2129]">
                {taskDetail.resultsPagination.total}
              </span>{' '}
              条结果
            </div>

            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="sm"
                onClick={() => handlePageChange(currentPage - 1)}
                disabled={currentPage === 1}
                className={cn(
                  'h-[30px] w-[34px] rounded-md border border-[#C9CDD4] bg-white p-0 hover:border-[#10B981] hover:text-[#10B981]',
                  currentPage === 1 && 'opacity-50 cursor-not-allowed'
                )}
              >
                <ChevronLeft className="h-4 w-4 text-[#6B7280]" />
              </Button>

              {renderPaginationButtons()}

              <Button
                variant="outline"
                size="sm"
                onClick={() => handlePageChange(currentPage + 1)}
                disabled={currentPage >= taskDetail.resultsPagination.totalPages}
                className={cn(
                  'h-[30px] w-[34px] rounded-md border border-[#C9CDD4] bg-white p-0 hover:border-[#10B981] hover:text-[#10B981]',
                  currentPage >= taskDetail.resultsPagination.totalPages &&
                    'opacity-50 cursor-not-allowed'
                )}
              >
                <ChevronRight className="h-4 w-4 text-[#6B7280]" />
              </Button>
            </div>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}

