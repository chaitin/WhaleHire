import { useState, useEffect } from 'react';
import { Eye, Download, Trash2, ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  type Resume,
  type PaginationInfo,
  ResumeStatus,
  type ResumeParseProgress,
} from '@/types/resume';
import { formatDate, formatPhone, cn } from '@/lib/utils';

// 状态标签映射
const statusLabels: Record<string, string> = {
  [ResumeStatus.PENDING]: '待解析',
  [ResumeStatus.PROCESSING]: '解析中',
  [ResumeStatus.COMPLETED]: '待筛选', // 解析成功翻译为待筛选
  [ResumeStatus.FAILED]: '解析失败',
  [ResumeStatus.ARCHIVED]: '已归档',
};
import { downloadResumeFile } from '@/services/resume';
import { SimpleResumeProgress } from '@/components/ui/resume-parse-progress';
import { useResumePollingManager } from '@/hooks/useResumePollingManager';

interface ResumeTableProps {
  resumes: Resume[];
  pagination: PaginationInfo;
  onPageChange: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  onEdit: (resume: Resume) => void;
  onDelete?: (id: string) => Promise<void>;
  onPreview?: (resume: Resume) => void;
  isLoading?: boolean;
  className?: string;
}

export function ResumeTable({
  resumes,
  pagination,
  onPageChange,
  onPageSizeChange,
  onEdit: _onEdit,
  onDelete,
  onPreview,
  isLoading = false,
  className,
}: ResumeTableProps) {
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deletingResumeId, setDeletingResumeId] = useState<string | null>(null);
  const [deletingResumeName, setDeletingResumeName] = useState<string>('');
  const [isDeleting, setIsDeleting] = useState(false);
  const [downloadingResumeId, setDownloadingResumeId] = useState<string | null>(
    null
  );

  const handleDeleteClick = (resume: Resume) => {
    setDeletingResumeId(resume.id);
    setDeletingResumeName(resume.name);
    setDeleteConfirmOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!onDelete || !deletingResumeId || isDeleting) {
      return;
    }

    setIsDeleting(true);
    try {
      await onDelete(deletingResumeId);
    } catch (error) {
      console.error('删除简历失败:', error);
    } finally {
      setIsDeleting(false);
      setDeletingResumeId(null);
      setDeletingResumeName('');
    }
  };

  const handleDownloadClick = async (resume: Resume) => {
    setDownloadingResumeId(resume.id);
    try {
      console.log('🔽 简历列表开始下载简历:', {
        resumeId: resume.id,
        name: resume.name,
        hasFileUrl: !!resume.resume_file_url,
      });

      await downloadResumeFile(resume);
      console.log('✅ 简历列表下载完成');
    } catch (error) {
      console.error('❌ 简历列表下载简历失败:', error);

      // 显示用户友好的错误信息
      let errorMessage = '下载失败，请稍后重试';
      if (error instanceof Error) {
        errorMessage = error.message;
      }

      // 这里可以添加 toast 通知或其他用户提示
      // 暂时使用 console.error，后续可以集成通知组件
      console.error('用户错误提示:', errorMessage);
    } finally {
      setDownloadingResumeId(null);
    }
  };

  const generatePageNumbers = () => {
    const pages = [];
    const { current, total, pageSize } = pagination;

    // 计算总页数
    const totalPages = Math.ceil(total / pageSize);

    // 显示当前页前后2页
    const start = Math.max(1, current - 2);
    const end = Math.min(totalPages, current + 2);

    for (let i = start; i <= end; i++) {
      pages.push(i);
    }

    return pages;
  };

  return (
    <div className={cn('', className)}>
      {/* 表格标题 */}
      <div className="border-b border-[#E5E7EB] px-6 py-4">
        <h3 className="text-lg font-semibold text-gray-900">
          简历列表 ({pagination.total})
        </h3>
      </div>

      {/* 表格内容 */}
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                简历信息
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                上传时间
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                上传人
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                状态
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                操作
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {resumes && resumes.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-16">
                    <div className="flex flex-col items-center justify-center space-y-3">
                      <div className="text-gray-400">
                        <svg
                          className="h-12 w-12 mx-auto"
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
                      </div>
                      <div className="text-sm text-gray-500">暂无简历数据</div>
                      <div className="text-xs text-gray-400">
                        请尝试调整搜索条件或上传新的简历
                      </div>
                    </div>
                  </td>
                </tr>
            ) : (
              resumes?.map((resume) => (
                  <tr key={resume.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-4">
                        <Avatar className="h-10 w-10">
                          <AvatarFallback className="bg-primary/10 text-primary">
                            {resume.name.charAt(0)}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="text-sm font-medium text-gray-900">
                            {resume.name}
                          </div>
                          <div className="text-sm text-gray-500">
                            {formatPhone(resume.phone)}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatDate(resume.created_at, 'datetime')}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {resume.uploader_name || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getStatusBadge(resume.status, resume.id)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          className={cn(
                            'h-8 w-8 p-0 text-gray-400 hover:text-gray-600',
                            resume.status !== ResumeStatus.COMPLETED &&
                              'opacity-50 cursor-not-allowed'
                          )}
                          onClick={() => onPreview?.(resume)}
                          disabled={
                            !onPreview ||
                            resume.status !== ResumeStatus.COMPLETED ||
                            isLoading
                          }
                          title="预览简历"
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className={cn(
                            'h-8 w-8 p-0 text-gray-400 hover:text-gray-600',
                            (downloadingResumeId === resume.id || isLoading) &&
                              'opacity-50 cursor-not-allowed'
                          )}
                          onClick={() => handleDownloadClick(resume)}
                          disabled={
                            downloadingResumeId === resume.id || isLoading
                          }
                          title="下载简历"
                        >
                          <Download className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-gray-400 hover:text-red-600"
                          onClick={() => handleDeleteClick(resume)}
                          disabled={!onDelete || isLoading || isDeleting}
                          title="删除简历"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))
            )}
          </tbody>
        </table>
      </div>

      {/* 分页区域 */}
      <div className="border-t border-[#E5E7EB] px-6 py-6">
        <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex items-center gap-4">
          <div className="text-sm text-[#6B7280]">
            显示
            <span className="text-[#6B7280]">
              {resumes?.length > 0
                ? (pagination.current - 1) * pagination.pageSize + 1
                : 0}
            </span>{' '}
            到{' '}
            <span className="text-[#6B7280]">
              {resumes?.length > 0
                ? Math.min(
                    pagination.current * pagination.pageSize,
                    pagination.total
                  )
                : 0}
            </span>{' '}
            条，共 <span className="text-[#6B7280]">{pagination.total}</span>{' '}
            条结果
          </div>

          {/* 每页条数选择器 */}
          {onPageSizeChange && (
            <div className="flex items-center gap-2">
              <span className="text-sm text-[#6B7280]">每页显示</span>
              <Select
                value={pagination.pageSize.toString()}
                onValueChange={(value) => onPageSizeChange(parseInt(value))}
              >
                <SelectTrigger className="w-20 h-8">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="10">10</SelectItem>
                  <SelectItem value="20">20</SelectItem>
                  <SelectItem value="50">50</SelectItem>
                </SelectContent>
              </Select>
              <span className="text-sm text-[#6B7280]">条</span>
            </div>
          )}
        </div>

        {/* 分页按钮 - 只在有数据时显示 */}
        {pagination.total > 0 && (
          <div className="flex flex-wrap items-center gap-1">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(pagination.current - 1)}
              disabled={pagination.current === 1}
              className={cn(
                'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                pagination.current === 1 && 'opacity-50'
              )}
            >
              <ChevronLeft className="h-4 w-4 text-[#6B7280]" />
            </Button>

            {generatePageNumbers().map((page) => (
              <Button
                key={page}
                variant="outline"
                size="sm"
                onClick={() => onPageChange(page)}
                className={cn(
                  'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0 text-sm font-normal text-[#374151]',
                  page === pagination.current &&
                    'border-[#10B981] bg-[#10B981] text-white'
                )}
              >
                {page}
              </Button>
            ))}

            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(pagination.current + 1)}
              disabled={
                pagination.current >=
                Math.ceil(pagination.total / pagination.pageSize)
              }
              className={cn(
                'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                pagination.current >=
                  Math.ceil(pagination.total / pagination.pageSize) &&
                  'opacity-50'
              )}
            >
              <ChevronRight className="h-4 w-4 text-[#6B7280]" />
            </Button>
          </div>
        )}
        </div>
      </div>

      {/* 删除确认对话框 */}
      <ConfirmDialog
        open={deleteConfirmOpen}
        onOpenChange={setDeleteConfirmOpen}
        title="确认删除"
        description={
          <span>
            确定要删除 <strong>{deletingResumeName}</strong>{' '}
            的简历吗？此操作无法撤销。
          </span>
        }
        confirmText="删除"
        cancelText="取消"
        onConfirm={handleDeleteConfirm}
        variant="destructive"
        loading={isDeleting}
      />
    </div>
  );

  function getStatusBadge(status: Resume['status'], resumeId?: string) {
    const statusConfig: Record<
      ResumeStatus | 'interview' | 'rejected' | 'hired',
      string
    > = {
      [ResumeStatus.PENDING]: 'status-badge pending',
      [ResumeStatus.PROCESSING]: 'status-badge processing',
      [ResumeStatus.COMPLETED]: 'status-badge pending', // 解析成功（待筛选）使用待处理样式
      [ResumeStatus.FAILED]: 'status-badge rejected',
      [ResumeStatus.ARCHIVED]: 'status-badge archived',
      interview: 'status-badge approved',
      rejected: 'status-badge rejected',
      hired: 'status-badge used',
    };

    // 对于处理中的简历，显示进度组件
    if (status === ResumeStatus.PROCESSING && resumeId) {
      return <ResumeProgressCell resumeId={resumeId} />;
    }

    const label = statusLabels[status as keyof typeof statusLabels] || status;

    return (
      <span className={statusConfig[status] ?? 'status-badge pending'}>
        {label}
      </span>
    );
  }

  // 简历进度单元格组件
  function ResumeProgressCell({ resumeId }: { resumeId: string }) {
    const [progress, setProgress] = useState<ResumeParseProgress | null>(null);
    const { startPolling, stopPolling } = useResumePollingManager();

    useEffect(() => {
      // 启动轮询
      startPolling(resumeId, (newProgress) => {
        setProgress(newProgress);
      });

      // 清理函数
      return () => {
        stopPolling();
      };
    }, [resumeId, startPolling, stopPolling]);

    if (!progress) {
      return <span className="status-badge processing">处理中</span>;
    }

    return (
      <SimpleResumeProgress
        status={progress.status}
        progress={progress.progress}
        className="min-w-[120px]"
      />
    );
  }
}
