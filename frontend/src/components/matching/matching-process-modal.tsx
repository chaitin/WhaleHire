import { useState, useEffect, useRef } from 'react';
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
import { getTaskProgress, cancelScreeningTask } from '@/services/screening';
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
  const [displayProgress, setDisplayProgress] = useState<number>(0);
  const taskStartAtRef = useRef<number | null>(null);

  // 新增：判定任务是否已完成（兼容部分失败但总体完成）
  const isFinished = (data: GetTaskProgressResp | null) => {
    if (!data) return false;
    const completed = data.status === 'completed';
    // 只有在明确完成状态时才判定为已完成，不依赖 progress_percent
    // 因为后端可能一开始就返回 100%，导致进度条不展示
    const total = data.resume_total ?? 0;
    const processed = data.resume_processed ?? 0;
    const allProcessed = total > 0 && processed >= total;
    return completed || allProcessed;
  };

  // 重置进度条状态（当弹窗打开时）
  useEffect(() => {
    if (open) {
      console.log('🔄 弹窗打开，重置进度条状态');
      setDisplayProgress(0);
      setProgressData(null);
      setPreviousStatus(null);
      taskStartAtRef.current = null;
    }
  }, [open]);

  // 获取任务进度
  useEffect(() => {
    if (!open || !taskId) return;

    // 立即获取一次
    const fetchProgress = async () => {
      try {
        const data = await getTaskProgress(taskId);
        console.log('📊 获取任务进度:', {
          status: data.status,
          progress_percent: data.progress_percent,
          resume_total: data.resume_total,
          resume_processed: data.resume_processed,
          isFinished: isFinished(data),
        });
        setProgressData(data);

        // 当任务判定已完成时自动跳转到结果
        if (isFinished(data)) {
          if (previousStatus !== 'completed') {
            console.log('✅ 任务完成，1秒后跳转到结果页');
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

    // 每10秒轮询一次进度
    const interval = setInterval(fetchProgress, 10000);

    return () => clearInterval(interval);
  }, [open, taskId, onComplete, previousStatus]);

  // 进度条缓慢增长动画：在轮询获取到新的progress_percent后，逐步逼近目标值
  useEffect(() => {
    if (!open) return;

    const status = progressData?.status;

    // 初始化任务开始时间（用于估算进度），优先使用后端提供的 started_at
    if (
      (status === 'in_progress' || String(status) === 'running') &&
      !taskStartAtRef.current
    ) {
      const startedAt = progressData?.started_at
        ? new Date(progressData.started_at).getTime()
        : Date.now();
      taskStartAtRef.current = startedAt;
    }

    // 完成后重置开始时间
    if (isFinished(progressData)) {
      taskStartAtRef.current = null;
    }
  }, [open, progressData, progressData?.status, progressData?.started_at]);

  // 进度条缓慢增长动画：在轮询获取到新的progress_percent后，逐步逼近目标值
  useEffect(() => {
    if (!open) return;

    const status = progressData?.status;
    const realProgress = Math.min(100, progressData?.progress_percent ?? 0);

    // 完成状态直接显示为100%
    if (isFinished(progressData)) {
      console.log('🎯 任务已完成，设置进度条为100%');
      setDisplayProgress(100);
      return;
    }

    console.log('🔄 进度条动画更新:', {
      status,
      realProgress,
      taskStarted: !!taskStartAtRef.current,
    });

    // 基于简历总数和开始时间估算目标进度（前端动画，不依赖后端）
    const totalResumes = Math.max(
      1,
      progressData?.resume_total || selectedResumeCount || 1
    );
    // 统一使用较慢的增长速度，确保所有情况下都有平滑动画
    const basePerResumeMs = 3000; // 每份简历预计3秒（统一速率）
    const expectedTotalMs = Math.max(60000, totalResumes * basePerResumeMs); // 保底至少60秒
    const startedAt = taskStartAtRef.current;
    const elapsedMs = startedAt ? Date.now() - startedAt : 0;

    const isActive = status === 'in_progress' || String(status) === 'running';
    const estimatedPercent = isActive
      ? Math.min(85, Math.max(0, (elapsedMs / expectedTotalMs) * 100)) // 基于时间的估算进度，上限85%
      : 0;

    const step = () => {
      setDisplayProgress((prev) => {
        // 在任务开始或运行中，缓慢增长
        const baseline = isActive
          ? Math.min(85, Math.max(prev, 1, estimatedPercent)) // 基线下限1%，上限85%
          : prev;

        // 目标值：优先使用估算的平滑进度，只有当后端进度超过估算进度较多时才缓慢追赶
        // 但绝不允许直接跳跃到后端进度
        let target = baseline;
        if (realProgress > baseline + 5) {
          // 如果后端进度超过估算进度5%以上，缓慢追赶，但不超过当前进度+2%
          target = Math.min(baseline + 2, realProgress);
        }

        // 确保目标不回退
        target = Math.max(prev, target);

        const delta = target - prev;
        if (Math.abs(delta) <= 0.05) return target; // 接近目标则停止
        // 更平滑的增长：每次只增加很小的量
        const inc = Math.max(0.015, delta / 80);
        return Math.min(prev + inc, target);
      });
    };

    // 立即执行一次，然后每800ms缓慢增长
    step();
    const intervalId = setInterval(step, 800);

    return () => {
      if (intervalId) clearInterval(intervalId);
    };
  }, [
    open,
    progressData,
    progressData?.progress_percent,
    progressData?.status,
    progressData?.resume_total,
    selectedResumeCount,
  ]);

  const handlePrevious = () => {
    setShowConfirmDialog(true);
  };

  const handleConfirmPrevious = async () => {
    setShowConfirmDialog(false);

    // 如果任务ID存在，调用取消任务接口
    if (taskId) {
      try {
        await cancelScreeningTask(taskId);
        console.log('任务已取消:', taskId);
      } catch (error) {
        console.error('取消任务失败:', error);
        // 即使取消失败，也继续返回上一步
      }
    }

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

  // 获取状态标签 - 按照用户要求的中文显示（优先使用完成判定）
  const getStatusLabel = (
    status: string,
    data?: GetTaskProgressResp | null
  ) => {
    if (data && isFinished(data)) return '匹配完成';
    switch (status) {
      case 'pending':
        return '创建完成';
      case 'in_progress':
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

  // 格式化时间（通用）
  const formatTime = (iso?: string) => {
    if (!iso) return '计算中...';
    try {
      const date = new Date(iso);
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

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent
          className="max-w-[880px] max-h-[90vh] p-0"
          onInteractOutside={(e) => e.preventDefault()}
        >
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
                      <Loader2
                        className="h-16 w-16 animate-spin"
                        style={{ color: '#7bb8ff' }}
                      />
                    </div>
                  </div>

                  {/* 匹配进度 */}
                  <div className="rounded-xl bg-white p-6 shadow-sm">
                    <div className="mb-3 flex items-center justify-between">
                      <span className="text-sm font-medium text-gray-900">
                        匹配进度
                      </span>
                      <span className="text-xl font-bold text-green-600">
                        {displayProgress.toFixed(1)}%
                      </span>
                    </div>
                    <div className="h-2.5 w-full rounded-full bg-gray-200">
                      <div
                        className="h-2.5 rounded-full bg-gradient-to-r from-orange-500 via-blue-500 to-green-500 transition-all duration-300"
                        style={{
                          width: `${Math.min(100, displayProgress)}%`,
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
                      {/* 任务开始时间 */}
                      <div className="flex items-start gap-3">
                        <div className="rounded-lg bg-indigo-50 p-2">
                          <Clock className="h-5 w-5 text-indigo-600" />
                        </div>
                        <div>
                          <p className="text-xs text-gray-500 mb-1">
                            任务开始时间
                          </p>
                          <p className="text-sm font-semibold text-gray-900">
                            {formatTime(progressData?.started_at)}
                          </p>
                        </div>
                      </div>

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

                      {/* 已处理 / 简历总数 */}
                      <div className="flex items-start gap-3">
                        <div className="rounded-lg bg-green-50 p-2">
                          <FileCheck className="h-5 w-5 text-green-600" />
                        </div>
                        <div>
                          <p className="text-xs text-gray-500 mb-1">
                            已处理 / 简历总数
                          </p>
                          <p className="text-sm font-semibold text-gray-900">
                            {progressData?.resume_processed || 0} /{' '}
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
                            {getStatusLabel(
                              progressData?.status || 'pending',
                              progressData
                            )}
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
            <AlertDialogTitle>确认返回上一步</AlertDialogTitle>
            <AlertDialogDescription>
              返回上一步将取消当前正在进行的匹配任务，已处理的数据将不会保存。确定要继续吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirmPrevious}>
              确定
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
