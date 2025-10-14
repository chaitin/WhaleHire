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
import { Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';

interface MatchingProcessModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onPrevious?: () => void;
  onComplete?: () => void;
  selectedJobCount: number;
  selectedResumeCount: number;
}

interface WeightProgress {
  name: string;
  emoji: string;
  status: 'completed' | 'in_progress' | 'pending';
}

interface LogEntry {
  time: string;
  message: string;
}

export function MatchingProcessModal({
  open,
  onOpenChange,
  onPrevious,
  onComplete,
  selectedJobCount,
  selectedResumeCount,
}: MatchingProcessModalProps) {
  const [progress, setProgress] = useState(0);
  const [processedCount, setProcessedCount] = useState(0);
  const [currentStage, setCurrentStage] = useState('基本信息权重');
  const [estimatedTime, setEstimatedTime] = useState('计算中...');
  const [logs, setLogs] = useState<LogEntry[]>([
    {
      time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
      message: '系统初始化完成',
    },
    {
      time: new Date(Date.now() + 1000).toLocaleTimeString('zh-CN', {
        hour12: false,
      }),
      message: '等待开始匹配任务...',
    },
    {
      time: new Date(Date.now() + 2000).toLocaleTimeString('zh-CN', {
        hour12: false,
      }),
      message: '提示：点击"开始匹配"按钮开始处理',
    },
  ]);

  const [weights, setWeights] = useState<WeightProgress[]>([
    { name: '基本信息权重', emoji: '✓', status: 'completed' },
    { name: '工作职责权重', emoji: '⏳', status: 'in_progress' },
    { name: '工作技能权重', emoji: '⏸', status: 'pending' },
    { name: '教育经历权重', emoji: '⏸', status: 'pending' },
    { name: '工作经历权重', emoji: '⏸', status: 'pending' },
    { name: '行业背景权重', emoji: '⏸', status: 'pending' },
    { name: '其他自定义权重', emoji: '⏸', status: 'pending' },
  ]);

  const logContainerRef = useRef<HTMLDivElement>(null);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);

  // 自动滚动日志到底部
  useEffect(() => {
    if (logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs]);

  // 添加日志
  const addLog = (message: string) => {
    const time = new Date().toLocaleTimeString('zh-CN', { hour12: false });
    setLogs((prev) => [...prev, { time, message }]);
  };

  // 模拟匹配进度
  useEffect(() => {
    if (!open) return;

    const interval = setInterval(() => {
      setProgress((prev) => {
        if (prev >= 100) {
          clearInterval(interval);
          addLog('✓ 所有匹配任务已完成');
          // 延迟1秒后打开结果页面
          setTimeout(() => {
            onOpenChange(false);
            onComplete?.();
          }, 1000);
          return 100;
        }

        const newProgress = prev + 2;
        const newProcessed = Math.floor(
          (newProgress / 100) * selectedResumeCount
        );

        if (newProcessed > processedCount) {
          setProcessedCount(newProcessed);
          if (newProcessed % 10 === 0) {
            addLog(`正在处理简历 ${newProcessed}/${selectedResumeCount}`);
          }
        }

        // 更新预计剩余时间
        const remaining = 100 - newProgress;
        const estimatedSeconds = Math.ceil((remaining / 2) * 0.5); // 假设每2%需要0.5秒
        setEstimatedTime(`约 ${estimatedSeconds} 秒`);

        // 更新权重进度
        const currentWeightIndex = Math.floor(
          (newProgress / 100) * weights.length
        );
        setWeights((prev) =>
          prev.map((w, i) => {
            if (i < currentWeightIndex) {
              return { ...w, status: 'completed' as const, emoji: '✓' };
            } else if (i === currentWeightIndex) {
              return { ...w, status: 'in_progress' as const, emoji: '⏳' };
            }
            return w;
          })
        );

        if (currentWeightIndex < weights.length) {
          setCurrentStage(weights[currentWeightIndex].name);
        }

        return newProgress;
      });
    }, 500);

    return () => clearInterval(interval);
  }, [
    open,
    processedCount,
    selectedResumeCount,
    weights,
    onComplete,
    onOpenChange,
  ]);

  // 开始匹配
  useEffect(() => {
    if (open && progress === 0) {
      setTimeout(() => {
        addLog('开始匹配任务');
        addLog(`共需匹配 ${selectedResumeCount} 份简历`);
        addLog(`匹配 ${selectedJobCount} 个岗位`);
      }, 100);
    }
  }, [open, progress, selectedResumeCount, selectedJobCount]);

  const handlePrevious = () => {
    setShowConfirmDialog(true);
  };

  const handleConfirmPrevious = () => {
    setShowConfirmDialog(false);
    onPrevious?.();
  };

  const getWeightStatusStyle = (status: WeightProgress['status']) => {
    switch (status) {
      case 'completed':
        return {
          bg: 'bg-green-50',
          border: 'border-l-4 border-green-500',
          textColor: 'text-gray-900',
          statusColor: 'text-green-600',
          statusLabel: '已完成',
          iconBg: 'bg-green-500',
        };
      case 'in_progress':
        return {
          bg: 'bg-yellow-50',
          border: 'border-l-4 border-yellow-500',
          textColor: 'text-gray-900',
          statusColor: 'text-yellow-600',
          statusLabel: '进行中',
          iconBg: 'bg-yellow-500',
        };
      case 'pending':
        return {
          bg: 'bg-gray-50',
          border: 'border-l-4 border-gray-300',
          textColor: 'text-gray-500',
          statusColor: 'text-gray-500',
          statusLabel: '待处理',
          iconBg: 'bg-gray-300',
        };
    }
  };

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-[880px] max-h-[90vh] overflow-hidden p-0">
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
                        {progress}%
                      </span>
                    </div>
                    <div className="h-2.5 w-full rounded-full bg-gray-200">
                      <div
                        className="h-2.5 rounded-full bg-gradient-to-r from-orange-500 via-blue-500 to-green-500 transition-all duration-300"
                        style={{ width: `${progress}%` }}
                      />
                    </div>
                  </div>

                  {/* 当前处理状态 */}
                  <div className="grid grid-cols-2 gap-4 rounded-xl bg-white p-6 shadow-sm">
                    <div>
                      <p className="mb-1 text-xs text-gray-500">当前处理状态</p>
                      <p className="text-base font-semibold text-gray-900">
                        正在分析简历 {processedCount}/{selectedResumeCount}
                      </p>
                    </div>
                    <div>
                      <p className="mb-1 text-xs text-gray-500">预计剩余时间</p>
                      <p className="text-base font-medium text-gray-900">
                        {estimatedTime}
                      </p>
                    </div>
                  </div>

                  {/* 权重处理进度 */}
                  <div className="rounded-xl bg-white p-6 shadow-sm">
                    <h4 className="mb-4 text-base font-semibold text-gray-900">
                      权重处理进度
                    </h4>
                    <div className="mb-4 rounded-lg bg-blue-50 px-4 py-3 text-sm text-blue-900 border-l-4 border-blue-500">
                      当前阶段：正在根据【{currentStage}】进行匹配...
                    </div>
                    <div className="space-y-3">
                      {weights.map((weight, index) => {
                        const style = getWeightStatusStyle(weight.status);
                        return (
                          <div
                            key={index}
                            className={cn(
                              'flex items-center gap-3 rounded-lg p-3',
                              style.bg,
                              style.border
                            )}
                          >
                            <div
                              className={cn(
                                'flex h-8 w-8 items-center justify-center rounded-full',
                                style.iconBg,
                                'bg-opacity-60'
                              )}
                            >
                              <span className="text-sm">{weight.emoji}</span>
                            </div>
                            <div className="flex-1">
                              <p
                                className={cn(
                                  'text-sm font-medium',
                                  style.textColor
                                )}
                              >
                                {weight.name}
                              </p>
                            </div>
                            <div>
                              <span
                                className={cn(
                                  'text-xs font-medium',
                                  style.statusColor
                                )}
                              >
                                {style.statusLabel}
                              </span>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* 实时处理日志 */}
                  <div className="rounded-xl bg-white p-6 shadow-sm">
                    <h4 className="mb-4 text-base font-semibold text-gray-900">
                      实时处理日志
                    </h4>
                    <div
                      ref={logContainerRef}
                      className="h-[200px] overflow-y-auto rounded-lg bg-gray-900 p-4 font-mono text-xs"
                    >
                      <div className="space-y-1">
                        {logs.map((log, index) => (
                          <div key={index} className="flex gap-2">
                            <span className="text-gray-400">[{log.time}]</span>
                            <span className="text-gray-300">{log.message}</span>
                          </div>
                        ))}
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
