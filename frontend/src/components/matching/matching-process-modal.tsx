import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { Loader2, Clock, FileCheck, CheckCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { getTaskProgress } from '@/services/screening';
import { GetTaskProgressResp } from '@/types/screening';

interface MatchingProcessModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onPrevious?: () => void;
  onComplete?: () => void;
  selectedJobCount: number;
  selectedResumeCount: number;
  taskId: string | null;
}

export function MatchingProcessModal({
  open,
  onOpenChange,
  onPrevious,
  onComplete,
  selectedResumeCount,
  taskId,
}: MatchingProcessModalProps) {
  const [progressData, setProgressData] = useState<GetTaskProgressResp | null>(
    null
  );
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [previousStatus, setPreviousStatus] = useState<string | null>(null);

  // 获取任务进度
  useEffect(() => {
    if (!open || !taskId) return;

    // 立即获取一次
    const fetchProgress = async () => {
      try {
        const data = await getTaskProgress(taskId);
        setProgressData(data);

        // 检查状态变化，当状态变为 completed 时自动跳转
        if (data.status === 'completed') {
          // 如果之前不是 completed 状态，现在变成了 completed，延迟1秒后跳转
          if (previousStatus !== 'completed') {
            setTimeout(() => {
              onComplete?.();
            }, 1000);
          }
        }

        // 更新上一次的状态
        setPreviousStatus(data.status);
      } catch (error) {
        console.error('获取任务进度失败:', error);
      }
    };

    fetchProgress();

    // 每2秒轮询一次进度
    const interval = setInterval(fetchProgress, 2000);

    return () => clearInterval(interval);
  }, [open, taskId, onComplete, previousStatus]);

  const handlePrevious = () => {
    setShowConfirmDialog(true);
  };

  const handleConfirmPrevious = () => {
    setShowConfirmDialog(false);
    onPrevious?.();
  };

  // 格式化预计完成时间
  const formatEstimatedTime = (estimatedFinish?: string) => {
    if (!estimatedFinish) return '计算中...';
    try {
      const date = new Date(estimatedFinish);
      return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    } catch {
      return '计算中...';
    }
  };

  // 获取状态标签 - 按照用户要求的中文显示
  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'pending':
        return '创建完成';
      case 'running':
        return '匹配中';
      case 'completed':
        return '匹配完成';
      case 'failed':
        return '匹配失败';
      default:
        return '未知';
    }
  };

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-[880px] max-h-[90vh] p-0">
          {/* 头部 */}
          <DialogHeader className="border-b border-gray-200 px-6 py-5">
            <div className="flex items-center justify-between">
              <DialogTitle className="text-lg font-semibold text-gray-900">
                创建新匹配任务
              </DialogTitle>
              <button
                onClick={() => onOpenChange(false)}
                className="rounded-lg bg-gray-100 p-2 text-gray-600 hover:bg-gray-200"
              >
                ×
              </button>
            </div>
          </DialogHeader>

          {/* 滚动内容区 */}
          <div
            className="overflow-y-auto px-6 py-6"
            style={{ maxHeight: 'calc(90vh - 180px)' }}
          >
            <div className="space-y-6">
              {/* 流程步骤指示器 - 横排展示 */}
              <div>
                <h3 className="mb-4 text-base font-semibold text-gray-900">
                  匹配任务创建流程
                </h3>
                <div className="relative">
                  {/* 连接线 */}
                  <div className="absolute left-[10%] right-[10%] top-7 h-0.5 bg-gray-200" />

                  {/* 步骤项 - 横排布局 */}
                  <div className="flex items-start justify-between">
                    {[
                      { name: '选择岗位', emoji: '💼', active: false },
                      { name: '选择简历', emoji: '👤', active: false },
                      { name: '权重配置', emoji: '⚖️', active: false },
                      { name: '匹配处理', emoji: '⚙️', active: true },
                      { name: '匹配结果', emoji: '📊', active: false },
                    ].map((step, index) => (
                      <div
                        key={index}
                        className="relative flex flex-col items-center"
                        style={{ width: '18%' }}
                      >
                        <div
                          className={cn(
                            'relative z-10 flex h-14 w-14 items-center justify-center rounded-full',
                            step.active
                              ? 'bg-green-100 shadow-[0_0_0_4px_rgba(16,185,129,0.1)]'
                              : 'bg-gray-200'
                          )}
                        >
                          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-white/60">
                            <span className="text-lg">{step.emoji}</span>
                          </div>
                        </div>
                        <div className="mt-3 text-center">
                          <p
                            className={cn(
                              'text-xs font-medium whitespace-nowrap',
                              step.active ? 'text-green-600' : 'text-gray-600'
                            )}
                          >
                            {step.name}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* 匹配中状态 */}
              <div className="rounded-lg bg-gray-50 p-6">
                <div className="space-y-6">
                  {/* 标题和动画 */}
                  <div>
                    <h3 className="mb-2 text-center text-xl font-semibold text-gray-900">
                      智能匹配中
                    </h3>
                    <p className="mb-4 text-center text-sm text-gray-600">
                      系统正在进行AI智能匹配，请稍候...
                    </p>
                    <div className="flex justify-center">
                      <Loader2 className="h-16 w-16 animate-spin text-green-500" />
                    </div>
                  </div>

                  {/* 匹配进度 */}
                  <div className="rounded-xl bg-white p-6 shadow-sm">
                    <div className="mb-3 flex items-center justify-between">
                      <span className="text-sm font-medium text-gray-900">
                        匹配进度
                      </span>
                      <span className="text-xl font-bold text-green-600">
                        {(progressData?.progress_percent || 0).toFixed(1)}%
                      </span>
                    </div>
                    <div className="h-2.5 w-full rounded-full bg-gray-200">
                      <div
                        className="h-2.5 rounded-full bg-gradient-to-r from-orange-500 via-blue-500 to-green-500 transition-all duration-300"
                        style={{
                          width: `${progressData?.progress_percent || 0}%`,
                        }}
                      />
                    </div>
                  </div>

                  {/* 任务信息详情 */}
                  <div className="rounded-xl bg-white p-6 shadow-sm">
                    <h4 className="mb-4 text-base font-semibold text-gray-900">
                      任务详情
                    </h4>
                    <div className="grid grid-cols-2 gap-6">
                      {/* 预计完成时间 */}
                      <div className="flex items-start gap-3">
                        <div className="rounded-lg bg-blue-50 p-2">
                          <Clock className="h-5 w-5 text-blue-600" />
                        </div>
                        <div>
                          <p className="text-xs text-gray-500 mb-1">
                            预计完成时间
                          </p>
                          <p className="text-sm font-semibold text-gray-900">
                            {formatEstimatedTime(
                              progressData?.estimated_finish
                            )}
                          </p>
                        </div>
                      </div>

                      {/* 已处理简历数量 */}
                      <div className="flex items-start gap-3">
                        <div className="rounded-lg bg-green-50 p-2">
                          <FileCheck className="h-5 w-5 text-green-600" />
                        </div>
                        <div>
                          <p className="text-xs text-gray-500 mb-1">
                            已处理简历
                          </p>
                          <p className="text-sm font-semibold text-gray-900">
                            {progressData?.resume_processed || 0} /{' '}
                            {progressData?.resume_total || selectedResumeCount}
                          </p>
                        </div>
                      </div>

                      {/* 简历总数 */}
                      <div className="flex items-start gap-3">
                        <div className="rounded-lg bg-purple-50 p-2">
                          <FileCheck className="h-5 w-5 text-purple-600" />
                        </div>
                        <div>
                          <p className="text-xs text-gray-500 mb-1">简历总数</p>
                          <p className="text-sm font-semibold text-gray-900">
                            {progressData?.resume_total || selectedResumeCount}
                          </p>
                        </div>
                      </div>

                      {/* 任务状态 */}
                      <div className="flex items-start gap-3">
                        <div className="rounded-lg bg-orange-50 p-2">
                          <CheckCircle className="h-5 w-5 text-orange-600" />
                        </div>
                        <div>
                          <p className="text-xs text-gray-500 mb-1">任务状态</p>
                          <p className="text-sm font-semibold text-gray-900">
                            {getStatusLabel(progressData?.status || 'pending')}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* 底部按钮 */}
          <DialogFooter className="border-t border-gray-200 px-6 py-4">
            <div className="flex w-full items-center justify-end gap-3">
              <Button variant="outline" onClick={handlePrevious}>
                上一步
              </Button>
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                取消
              </Button>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 确认对话框 */}
      <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认操作</AlertDialogTitle>
            <AlertDialogDescription>
              确认要进行此操作吗？操作后将不可撤回
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirmPrevious}>
              确认
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
