import { CheckCircle, AlertCircle, Loader2, Clock } from 'lucide-react';
import { ResumeParseProgress, ResumeStatus } from '@/types/resume';
import { cn } from '@/lib/utils';

interface ResumeParseProgressProps {
  progress: ResumeParseProgress;
  className?: string;
  showDetails?: boolean;
}

export function ResumeParseProgressComponent({
  progress,
  className,
  showDetails = true,
}: ResumeParseProgressProps) {
  const getStatusIcon = (status: ResumeStatus) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-4 w-4 text-gray-500" />;
      case 'processing':
        return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />;
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'failed':
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Clock className="h-4 w-4 text-gray-500" />;
    }
  };

  const getStatusColor = (status: ResumeStatus) => {
    switch (status) {
      case 'pending':
        return 'text-gray-600';
      case 'processing':
        return 'text-blue-600';
      case 'completed':
        return 'text-green-600';
      case 'failed':
        return 'text-red-600';
      default:
        return 'text-gray-600';
    }
  };

  const getProgressBarColor = (status: ResumeStatus) => {
    switch (status) {
      case 'pending':
        return 'bg-gray-200';
      case 'processing':
        return 'bg-blue-500';
      case 'completed':
        return 'bg-green-500';
      case 'failed':
        return 'bg-red-500';
      default:
        return 'bg-gray-200';
    }
  };

  const getStatusText = (status: ResumeStatus) => {
    switch (status) {
      case 'pending':
        return '等待解析';
      case 'processing':
        return '解析中';
      case 'completed':
        return '解析完成';
      case 'failed':
        return '解析失败';
      default:
        return '未知状态';
    }
  };

  return (
    <div className={cn('space-y-3', className)}>
      {/* 状态标题行 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          {getStatusIcon(progress.status)}
          <span
            className={cn(
              'text-sm font-medium',
              getStatusColor(progress.status)
            )}
          >
            {getStatusText(progress.status)}
          </span>
        </div>
        <span className="text-sm text-gray-500">{progress.progress}%</span>
      </div>

      {/* 进度条 */}
      <div className="w-full bg-gray-200 rounded-full h-2">
        <div
          className={cn(
            'h-2 rounded-full transition-all duration-300 ease-in-out',
            getProgressBarColor(progress.status)
          )}
          style={{ width: `${progress.progress}%` }}
        />
      </div>

      {/* 详细信息 */}
      {showDetails && (
        <div className="space-y-2">
          {/* 进度消息 */}
          {progress.message && (
            <p className="text-sm text-gray-600">{progress.message}</p>
          )}

          {/* 错误消息 */}
          {progress.error_message && progress.status === 'failed' && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
              <p className="text-sm text-red-700">
                <span className="font-medium">错误详情：</span>
                {progress.error_message}
              </p>
            </div>
          )}

          {/* 时间信息 */}
          <div className="flex justify-between text-xs text-gray-500">
            <span>
              开始时间：{new Date(progress.started_at).toLocaleString()}
            </span>
            {progress.completed_at && (
              <span>
                完成时间：{new Date(progress.completed_at).toLocaleString()}
              </span>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

// 简化版进度条，用于列表显示
interface SimpleProgressProps {
  status: ResumeStatus;
  progress?: number;
  className?: string;
}

export function SimpleResumeProgress({
  status,
  progress = 0,
  className,
}: SimpleProgressProps) {
  const getStatusIcon = (status: ResumeStatus) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-3 w-3 text-gray-500" />;
      case 'processing':
        return <Loader2 className="h-3 w-3 text-blue-500 animate-spin" />;
      case 'completed':
        return <CheckCircle className="h-3 w-3 text-green-500" />;
      case 'failed':
        return <AlertCircle className="h-3 w-3 text-red-500" />;
      default:
        return <Clock className="h-3 w-3 text-gray-500" />;
    }
  };

  const getStatusText = (status: ResumeStatus) => {
    switch (status) {
      case 'pending':
        return '等待解析';
      case 'processing':
        return '解析中';
      case 'completed':
        return '已完成';
      case 'failed':
        return '解析失败';
      default:
        return '未知';
    }
  };

  const getProgressBarColor = (status: ResumeStatus) => {
    switch (status) {
      case 'pending':
        return 'bg-gray-300';
      case 'processing':
        return 'bg-blue-500';
      case 'completed':
        return 'bg-green-500';
      case 'failed':
        return 'bg-red-500';
      default:
        return 'bg-gray-300';
    }
  };

  return (
    <div className={cn('flex items-center space-x-2', className)}>
      {getStatusIcon(status)}
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between mb-1">
          <span className="text-xs text-gray-600 truncate">
            {getStatusText(status)}
          </span>
          {(status === 'processing' || status === 'completed') && (
            <span className="text-xs text-gray-500 ml-2">{progress}%</span>
          )}
        </div>
        {(status === 'processing' || status === 'completed') && (
          <div className="w-full bg-gray-200 rounded-full h-1">
            <div
              className={cn(
                'h-1 rounded-full transition-all duration-300',
                getProgressBarColor(status)
              )}
              style={{ width: `${progress}%` }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
