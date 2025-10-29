import { useEffect, useState, useRef } from 'react';
import {
  X,
  Upload,
  Link,
  ArrowRight,
  FileText,
  User,
  Mail,
  Phone,
  MapPin,
  GraduationCap,
  Briefcase,
  CheckCircle,
  XCircle,
  Clock,
  Loader2,
} from 'lucide-react';
import { Button } from '@/ui/button';
import { Label } from '@/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/ui/dialog';
import { useResumeUpload } from '@/hooks/useResume';
import { useResumePollingManager } from '@/hooks/useResumePollingManager';
import {
  Resume,
  ResumeParseProgress,
  ResumeDetail,
  ResumeStatus,
} from '@/types/resume';
import { formatDate } from '@/lib/utils';
import {
  getResumeProgress,
  getResumeDetail,
  getBatchUploadStatus,
} from '@/services/resume';
import { MultiSelect, Option } from '@/ui/multi-select';
import { listJobProfiles } from '@/services/job-profile';

interface UploadResumeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: (resume?: Resume) => void;
}

// 简历解析状态组件
function ResumeParsingStatus({
  resumeId,
  filename,
}: {
  resumeId: string;
  filename: string;
}) {
  const [parseProgress, setParseProgress] =
    useState<ResumeParseProgress | null>(null);
  const { startPolling, stopPolling } = useResumePollingManager();

  useEffect(() => {
    // 获取解析进度并启动轮询
    const initPolling = async () => {
      try {
        const progress = await getResumeProgress(resumeId);
        setParseProgress(progress);

        // 如果需要轮询，启动轮询
        if (
          progress.status === ResumeStatus.PENDING ||
          progress.status === ResumeStatus.PROCESSING
        ) {
          startPolling(resumeId, (newProgress) => {
            setParseProgress(newProgress);
          });
        }
      } catch (error) {
        console.error('获取解析进度失败:', error);
      }
    };

    initPolling();

    return () => {
      stopPolling();
    };
  }, [resumeId, startPolling, stopPolling]);

  const getStatusIcon = () => {
    if (!parseProgress) {
      return <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />;
    }

    switch (parseProgress.status) {
      case ResumeStatus.COMPLETED:
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case ResumeStatus.FAILED:
        return <XCircle className="w-4 h-4 text-red-500" />;
      case ResumeStatus.PROCESSING:
        return <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />;
      case ResumeStatus.PENDING:
        return <Clock className="w-4 h-4 text-gray-400" />;
      default:
        return <FileText className="w-4 h-4 text-gray-400" />;
    }
  };

  const getStatusText = () => {
    if (!parseProgress) {
      return '加载中...';
    }

    switch (parseProgress.status) {
      case ResumeStatus.COMPLETED:
        return '解析完成';
      case ResumeStatus.FAILED:
        return `解析失败${parseProgress.error_message ? ': ' + parseProgress.error_message : ''}`;
      case ResumeStatus.PROCESSING:
        return '解析中'; // 移除百分比，在进度条后单独显示
      case ResumeStatus.PENDING:
        return '等待解析';
      default:
        return '未知状态';
    }
  };

  const getStatusColor = () => {
    if (!parseProgress) return 'text-gray-600';

    switch (parseProgress.status) {
      case ResumeStatus.COMPLETED:
        return 'text-green-700';
      case ResumeStatus.FAILED:
        return 'text-red-700';
      case ResumeStatus.PROCESSING:
        return 'text-blue-700';
      case ResumeStatus.PENDING:
        return 'text-gray-600';
      default:
        return 'text-gray-600';
    }
  };

  return (
    <div className="bg-white rounded px-2.5 py-1.5 text-xs">
      {/* 第一行：图标和文件名 */}
      <div className="flex items-center gap-2">
        <div className="flex-shrink-0">{getStatusIcon()}</div>
        <p className="truncate text-gray-700 font-medium text-xs flex-1">
          {filename}
        </p>
      </div>

      {/* 第二行：状态文本、进度条和百分比 */}
      <div className="flex items-center gap-2 mt-1.5 ml-6">
        <p className={`text-[10px] ${getStatusColor()} flex-shrink-0`}>
          {getStatusText()}
        </p>
        {/* 解析中时显示进度条和百分比 */}
        {parseProgress && parseProgress.status === ResumeStatus.PROCESSING && (
          <>
            <div className="flex-1 max-w-[120px]">
              <div className="bg-gray-200 rounded-full h-1">
                <div
                  className="bg-blue-500 h-1 rounded-full transition-all duration-300"
                  style={{ width: `${parseProgress.progress}%` }}
                ></div>
              </div>
            </div>
            <span className="text-[10px] text-blue-700 font-medium flex-shrink-0">
              {parseProgress.progress}%
            </span>
          </>
        )}
      </div>
    </div>
  );
}

export function UploadResumeModal({
  open,
  onOpenChange,
  onSuccess,
}: UploadResumeModalProps) {
  const [position, setPosition] = useState('');
  const [uploadMethod, setUploadMethod] = useState<'local' | 'link' | null>(
    null
  );
  const [currentStep, setCurrentStep] = useState<
    'upload' | 'preview' | 'complete'
  >('upload');
  const [uploadedResume, setUploadedResume] = useState<Resume | null>(null);
  const [showContentTip, setShowContentTip] = useState(false);

  // 多个简历相关状态
  const [uploadedResumes, setUploadedResumes] = useState<ResumeDetail[]>([]);
  const [currentResumeIndex, setCurrentResumeIndex] = useState(0);

  // 批量上传状态轮询的定时器引用
  const batchStatusPollingTimerRef = useRef<NodeJS.Timeout | null>(null);

  // 岗位选择相关状态
  const [selectedJobIds, setSelectedJobIds] = useState<string[]>([]);
  const [jobOptions, setJobOptions] = useState<Option[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);

  // 选择的文件列表
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);

  const {
    uploadFile,
    uploading,
    uploadProgress,
    error,
    uploadStatus,
    stopPolling: stopUploadPolling,
    taskId,
  } = useResumeUpload();
  const { startPolling, stopPolling } = useResumePollingManager();

  // 获取岗位列表
  useEffect(() => {
    if (open) {
      fetchJobProfiles();
    }
  }, [open]);

  const fetchJobProfiles = async () => {
    setLoadingJobs(true);
    try {
      const response = await listJobProfiles({
        page: 1,
        page_size: 100, // 获取足够多的岗位供选择
      });

      // 只显示发布状态的岗位，过滤掉草稿
      const publishedJobs = response.items.filter(
        (job) => job.status === 'published'
      );

      // 转换为 MultiSelect 所需的选项格式
      const options: Option[] = publishedJobs.map((job) => ({
        value: job.id,
        label: job.name,
      }));
      setJobOptions(options);
    } catch (err) {
      console.error('获取岗位列表失败:', err);
    } finally {
      setLoadingJobs(false);
    }
  };

  useEffect(() => {
    if (!open) {
      // 关闭弹窗时停止所有轮询
      stopPolling();
      stopUploadPolling();

      // 清理批量状态轮询定时器
      if (batchStatusPollingTimerRef.current) {
        clearInterval(batchStatusPollingTimerRef.current);
        batchStatusPollingTimerRef.current = null;
      }

      // 只有在解析完成后关闭弹窗时才重置状态
      if (currentStep === 'complete') {
        // 关闭弹窗时通知父组件刷新列表
        console.log('📤 关闭弹窗，通知父组件刷新列表');
        if (uploadedResumes.length > 0) {
          onSuccess?.(uploadedResumes[0] as Resume);
        } else {
          onSuccess?.();
        }

        setPosition('');
        setUploadMethod(null);
        setCurrentStep('upload');
        setUploadedResume(null);
        setShowContentTip(false);
        setSelectedJobIds([]); // 重置岗位选择
        setSelectedFiles([]); // 重置文件选择
        setUploadedResumes([]); // 重置多简历列表
        setCurrentResumeIndex(0); // 重置简历索引
      }
    }
  }, [
    open,
    stopPolling,
    stopUploadPolling,
    currentStep,
    uploadedResumes,
    onSuccess,
  ]);

  // 监听批量上传状态变化
  useEffect(() => {
    if (uploadStatus && currentStep === 'preview') {
      console.log('📤 批量上传状态更新:', uploadStatus);

      // 上传完成，处理结果
      if (uploadStatus.status === 'completed') {
        console.log('✅ 批量上传完成');

        // 如果有成功上传的简历项，获取所有成功的简历详情
        if (uploadStatus.items && uploadStatus.items.length > 0) {
          const successItems = uploadStatus.items.filter(
            (item) => item.status === 'completed' && item.resume_id
          );

          if (successItems.length > 0) {
            console.log(`📄 开始获取 ${successItems.length} 个成功简历的详情`);

            // 获取所有成功简历的详情
            Promise.all(
              successItems.map((item) => getResumeDetail(item.resume_id!))
            )
              .then((details) => {
                console.log(
                  '📄 获取到所有简历详情:',
                  details.map((d) => ({
                    id: d.id,
                    name: d.name,
                    status: d.status,
                  }))
                );

                // 保存所有简历详情
                setUploadedResumes(details);
                setCurrentResumeIndex(0);

                // 保留第一个简历的兼容性
                if (details.length > 0) {
                  setUploadedResume(details[0] as Resume);

                  // 为第一个简历启动解析进度轮询
                  const firstDetail = details[0];
                  if (
                    firstDetail.status === 'pending' ||
                    firstDetail.status === 'processing'
                  ) {
                    console.log('启动第一个简历解析进度轮询:', firstDetail.id);
                    startPolling(firstDetail.id, async (newProgress) => {
                      // 解析完成后更新简历详情
                      if (newProgress.status === 'completed') {
                        try {
                          const updatedDetail = await getResumeDetail(
                            firstDetail.id
                          );
                          console.log('📄 解析完成后获取的简历详情:', {
                            id: updatedDetail.id,
                            name: updatedDetail.name,
                            status: updatedDetail.status,
                          });

                          // 更新列表中的对应简历
                          setUploadedResumes((prev) =>
                            prev.map((r) =>
                              r.id === updatedDetail.id ? updatedDetail : r
                            )
                          );

                          // 检查是否所有简历都解析完成
                          const allCompleted = details.every(
                            (d) =>
                              d.status === 'completed' ||
                              d.id === updatedDetail.id
                          );

                          if (allCompleted) {
                            setCurrentStep('complete');
                            // 传递第一个简历数据给父组件
                            onSuccess?.(updatedDetail as Resume);
                          }
                        } catch (error) {
                          console.error('获取简历详情失败:', error);
                        }
                      } else if (newProgress.status === 'failed') {
                        console.error('简历解析失败:', newProgress);
                        onSuccess?.(firstDetail as Resume);
                      }
                    });
                  } else if (firstDetail.status === 'completed') {
                    // 已经解析完成
                    setCurrentStep('complete');
                    onSuccess?.(firstDetail as Resume);
                  } else {
                    onSuccess?.(firstDetail as Resume);
                  }
                }
              })
              .catch((error) => {
                console.error('获取简历详情失败:', error);
                onSuccess?.();
              });
          } else {
            console.log('❌ 没有成功上传的简历项，通知刷新列表');
            onSuccess?.();
          }
        } else {
          console.log('❌ 批量上传完成但没有上传项，通知刷新列表');
          onSuccess?.();
        }
      } else if (uploadStatus.status === 'failed') {
        console.error('❌ 批量上传失败');
        onSuccess?.();
      }
    }
  }, [uploadStatus, currentStep, startPolling, onSuccess]);

  // 监听简历上传成功后的状态变化，手动启动轮询（保留兼容旧逻辑）
  useEffect(() => {
    if (uploadedResume && currentStep === 'preview') {
      // 如果简历状态需要轮询，则启动轮询
      if (
        uploadedResume.status === 'pending' ||
        uploadedResume.status === 'processing'
      ) {
        console.log('启动简历解析进度轮询:', uploadedResume.id);
        startPolling(uploadedResume.id, async (newProgress) => {
          // 解析完成后获取详细信息并进入完成步骤
          if (newProgress.status === 'completed') {
            try {
              const detail = await getResumeDetail(uploadedResume.id);
              // 更新简历列表中的数据
              setUploadedResumes((prev) =>
                prev.map((r) => (r.id === detail.id ? detail : r))
              );
              setCurrentStep('complete');
            } catch (error) {
              console.error('获取简历详情失败:', error);
            }
          } else if (newProgress.status === 'failed') {
            console.error('简历解析失败:', newProgress);
          }
        });
      }
    }
  }, [uploadedResume, currentStep, startPolling]);

  // 在 complete 步骤轮询批量上传状态，获取所有简历的最新数据
  useEffect(() => {
    // 清理函数
    const cleanup = () => {
      if (batchStatusPollingTimerRef.current) {
        console.log('🧹 清除批量上传状态轮询定时器');
        clearInterval(batchStatusPollingTimerRef.current);
        batchStatusPollingTimerRef.current = null;
      }
    };

    // 只在 complete 步骤且有 taskId 时启动轮询
    if (currentStep === 'complete' && taskId) {
      console.log('🔄 启动批量上传状态轮询，taskId:', taskId);

      let pollCount = 0;
      const maxPollCount = 60; // 最多轮询60次（6分钟）
      let hasNotifiedSuccess = false; // 标记是否已经通知过成功

      // 定义轮询函数
      const pollBatchStatus = async () => {
        // 防止无限轮询
        pollCount++;
        if (pollCount > maxPollCount) {
          console.log('⛔ 达到最大轮询次数(60次)，停止轮询');
          cleanup();
          return;
        }

        try {
          const batchStatus = await getBatchUploadStatus(taskId);
          console.log(`📊 批量上传状态轮询结果 (第${pollCount}次):`, {
            taskId,
            status: batchStatus.status,
            items: batchStatus.items?.length || 0,
          });

          // 获取所有成功上传的简历ID
          const successItems =
            batchStatus.items?.filter(
              (item) => item.status === 'completed' && item.resume_id
            ) || [];

          if (successItems.length > 0) {
            console.log(`📄 获取 ${successItems.length} 个简历的最新详情`);

            // 批量获取所有简历的最新详情
            const resumeDetails = await Promise.all(
              successItems.map((item) => getResumeDetail(item.resume_id!))
            );

            console.log(
              '📄 所有简历详情已更新:',
              resumeDetails.map((d) => ({
                id: d.id,
                name: d.name,
                status: d.status,
              }))
            );

            // 更新简历列表
            setUploadedResumes(resumeDetails);

            // 如果所有简历都解析完成，停止轮询并通知父组件
            const allCompleted = resumeDetails.every(
              (d) => d.status === 'completed' || d.status === 'failed'
            );

            if (allCompleted && !hasNotifiedSuccess) {
              console.log('✅ 所有简历解析完成，停止轮询并通知父组件刷新列表');
              hasNotifiedSuccess = true; // 标记已通知，防止重复调用
              cleanup();
              // 通知父组件刷新列表，传递第一个简历数据
              if (resumeDetails.length > 0 && onSuccess) {
                onSuccess(resumeDetails[0] as Resume);
              }
            }
          }
        } catch (error) {
          console.error('❌ 获取批量上传状态失败:', error);
        }
      };

      // 立即执行一次
      pollBatchStatus();

      // 每8秒轮询一次
      batchStatusPollingTimerRef.current = setInterval(pollBatchStatus, 8000);

      // 清理函数
      return cleanup;
    } else {
      // 如果不在 complete 步骤，确保清理定时器
      cleanup();
    }
    // 移除 onSuccess 从依赖项，避免因其引用变化导致重新创建轮询
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentStep, taskId]);

  // 拖拽状态
  const [isDragging, setIsDragging] = useState(false);

  // 选择文件
  const handleSelectFiles = () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.pdf,.doc,.docx';
    input.multiple = true; // 允许选择多个文件
    input.onchange = (e) => {
      const files = (e.target as HTMLInputElement).files;
      if (files && files.length > 0) {
        const fileArray = Array.from(files);
        console.log(
          `📁 选择了 ${fileArray.length} 个文件:`,
          fileArray.map((f) => f.name)
        );
        setSelectedFiles(fileArray);
      }
    };
    input.click();
  };

  // 处理拖拽进入
  const handleDragEnter = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  // 处理拖拽离开
  const handleDragLeave = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  };

  // 处理拖拽悬停
  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
  };

  // 处理文件放置
  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      // 过滤出支持的文件类型
      const supportedFiles = Array.from(files).filter((file) => {
        const ext = file.name.toLowerCase().split('.').pop();
        return ext === 'pdf' || ext === 'doc' || ext === 'docx';
      });

      if (supportedFiles.length > 0) {
        console.log(
          `📁 拖拽上传了 ${supportedFiles.length} 个文件:`,
          supportedFiles.map((f) => f.name)
        );
        setSelectedFiles(supportedFiles);
      } else {
        alert('请上传 PDF、DOC 或 DOCX 格式的文件');
      }
    }
  };

  // 上传文件
  const handleFileUpload = async () => {
    if (selectedFiles.length === 0) {
      alert('请先选择文件');
      return;
    }

    try {
      console.log(`📤 开始上传 ${selectedFiles.length} 个简历文件`);

      // 调用批量上传接口，支持单个或多个文件
      const response = await uploadFile(
        selectedFiles.length === 1 ? selectedFiles[0] : selectedFiles,
        selectedJobIds,
        position || undefined
      );

      console.log('📤 上传响应:', response);

      // 批量上传接口已经在 Hook 中自动开始轮询状态
      // 这里直接进入预览步骤，等待上传完成
      setCurrentStep('preview');
    } catch (err) {
      // 错误已在 hook 中处理，这里可以根据需要添加额外逻辑
      console.error('上传简历失败:', err);
    }
  };

  const handleNext = () => {
    if (currentStep === 'upload' && uploadMethod === 'local' && !uploading) {
      // 验证是否选择了岗位
      if (selectedJobIds.length === 0) {
        alert('请先选择岗位');
        return;
      }
      // 验证是否选择了文件
      if (selectedFiles.length === 0) {
        alert('请先选择要上传的文件');
        return;
      }
      handleFileUpload();
    } else if (currentStep === 'preview') {
      setShowContentTip(true);
      setTimeout(() => {
        onOpenChange(false);
      }, 2000);
    }
  };

  const getStepNumber = (step: string) => {
    switch (step) {
      case 'upload':
        return 1;
      case 'preview':
        return 2;
      case 'complete':
        return 3;
      default:
        return 1;
    }
  };

  const isStepCompleted = (step: string) => {
    const stepNum = getStepNumber(step);
    const currentStepNum = getStepNumber(currentStep);
    return (
      stepNum < currentStepNum ||
      (stepNum === currentStepNum && currentStep === 'complete')
    );
  };

  const isStepActive = (step: string) => {
    return step === currentStep;
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] p-0 bg-white rounded-2xl flex flex-col">
        {/* 弹窗头部 */}
        <DialogHeader className="px-6 py-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <DialogTitle className="text-xl font-semibold text-gray-900">
              上传简历
            </DialogTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => onOpenChange(false)}
              className="h-7 w-7 rounded-full hover:bg-gray-100"
            >
              <X className="h-4 w-4 text-gray-500" />
            </Button>
          </div>
        </DialogHeader>

        {/* 步骤指示器 */}
        <div className="px-6 py-6 border-b border-gray-200 flex-shrink-0">
          <div className="flex items-center justify-center max-w-2xl mx-auto">
            {/* 步骤1 - 上传文件 */}
            <div className="flex items-center">
              <div
                className={`flex items-center justify-center w-10 h-10 rounded-full text-sm font-medium ${
                  isStepCompleted('upload')
                    ? 'text-white'
                    : isStepActive('upload')
                      ? 'text-white'
                      : 'bg-gray-200 text-gray-600'
                }`}
                style={
                  isStepCompleted('upload') || isStepActive('upload')
                    ? {
                        background:
                          'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                      }
                    : undefined
                }
              >
                {isStepCompleted('upload') ? (
                  <CheckCircle className="h-5 w-5" />
                ) : (
                  '1'
                )}
              </div>
              <div className="ml-3 hidden sm:block">
                <div
                  className={`text-sm font-medium ${
                    isStepCompleted('upload')
                      ? 'text-[#7bb8ff]'
                      : isStepActive('upload')
                        ? 'text-[#7bb8ff]'
                        : 'text-gray-600'
                  }`}
                >
                  上传文件
                </div>
              </div>
            </div>

            {/* 连接线1 */}
            <div className="flex-1 mx-4 h-1 bg-gray-200 rounded max-w-24">
              <div
                className={`h-full rounded transition-all duration-300 ${
                  isStepCompleted('upload') ? 'w-full' : 'w-0'
                }`}
                style={{
                  background:
                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }}
              ></div>
            </div>

            {/* 步骤2 - 预览内容 */}
            <div className="flex items-center">
              <div
                className={`flex items-center justify-center w-10 h-10 rounded-full text-sm font-medium ${
                  isStepCompleted('preview')
                    ? 'text-white'
                    : isStepActive('preview')
                      ? 'text-white'
                      : 'bg-gray-200 text-gray-600'
                }`}
                style={
                  isStepCompleted('preview') || isStepActive('preview')
                    ? {
                        background:
                          'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                      }
                    : undefined
                }
              >
                {isStepCompleted('preview') ? (
                  <CheckCircle className="h-5 w-5" />
                ) : (
                  '2'
                )}
              </div>
              <div className="ml-3 hidden sm:block">
                <div
                  className={`text-sm font-medium ${
                    isStepCompleted('preview')
                      ? 'text-[#7bb8ff]'
                      : isStepActive('preview')
                        ? 'text-[#7bb8ff]'
                        : 'text-gray-600'
                  }`}
                >
                  预览内容
                </div>
              </div>
            </div>

            {/* 连接线2 */}
            <div className="flex-1 mx-4 h-1 bg-gray-200 rounded max-w-24">
              <div
                className={`h-full rounded transition-all duration-300 ${
                  isStepCompleted('preview') ? 'w-full' : 'w-0'
                }`}
                style={{
                  background:
                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }}
              ></div>
            </div>

            {/* 步骤3 - 完成 */}
            <div className="flex items-center">
              <div
                className={`flex items-center justify-center w-10 h-10 rounded-full text-sm font-medium ${
                  isStepCompleted('complete')
                    ? 'text-white'
                    : isStepActive('complete')
                      ? 'text-white'
                      : 'bg-gray-200 text-gray-600'
                }`}
                style={
                  isStepCompleted('complete') || isStepActive('complete')
                    ? {
                        background:
                          'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                      }
                    : undefined
                }
              >
                {isStepCompleted('complete') ? (
                  <CheckCircle className="h-5 w-5" />
                ) : (
                  '3'
                )}
              </div>
              <div className="ml-3 hidden sm:block">
                <div
                  className={`text-sm font-medium ${
                    isStepCompleted('complete')
                      ? 'text-[#7bb8ff]'
                      : isStepActive('complete')
                        ? 'text-[#7bb8ff]'
                        : 'text-gray-600'
                  }`}
                >
                  完成
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* 表单内容 */}
        <div className="flex-1 overflow-y-auto px-6 py-8 space-y-8">
          {currentStep === 'upload' && (
            <>
              {/* 岗位选择区块 */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <Label className="text-sm font-medium text-gray-700">
                    选择岗位 <span className="text-red-500">*</span>
                  </Label>
                </div>

                <MultiSelect
                  options={jobOptions}
                  selected={selectedJobIds}
                  onChange={setSelectedJobIds}
                  placeholder={loadingJobs ? '加载岗位中...' : '请选择岗位'}
                  multiple={true}
                  searchPlaceholder="搜索岗位名称..."
                  disabled={loadingJobs}
                  selectCountLabel="岗位"
                />

                {/* 已选择岗位展示 */}
                {selectedJobIds.length > 0 && (
                  <div className="bg-gray-50 rounded-lg px-3 py-2">
                    <div className="text-sm text-gray-700">
                      {selectedJobIds.length === 1 ? (
                        <span>
                          已选择：
                          <span className="font-medium text-gray-900">
                            {jobOptions.find(
                              (job) => job.value === selectedJobIds[0]
                            )?.label || '未知岗位'}
                          </span>
                        </span>
                      ) : (
                        <span>
                          已选择：
                          <span className="font-medium text-gray-900">
                            {jobOptions.find(
                              (job) => job.value === selectedJobIds[0]
                            )?.label || '未知岗位'}
                          </span>
                          等
                          <span className="relative font-medium text-[#7bb8ff] cursor-pointer group">
                            {selectedJobIds.length}
                            {/* Hover 提示框 */}
                            <div className="absolute left-0 bottom-full mb-2 hidden group-hover:block z-50 w-max max-w-xs">
                              <div className="bg-gray-900 text-white text-xs rounded-lg px-3 py-2 shadow-lg">
                                <div className="font-medium mb-1.5">
                                  已选择的岗位：
                                </div>
                                <div className="space-y-1 max-h-48 overflow-y-auto">
                                  {selectedJobIds.map((jobId, index) => (
                                    <div key={jobId} className="text-gray-200">
                                      {index + 1}.{' '}
                                      {jobOptions.find(
                                        (job) => job.value === jobId
                                      )?.label || '未知岗位'}
                                    </div>
                                  ))}
                                </div>
                                {/* 小三角 */}
                                <div className="absolute left-4 top-full w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-900"></div>
                              </div>
                            </div>
                          </span>
                          个岗位
                        </span>
                      )}
                    </div>
                  </div>
                )}
              </div>

              {/* 上传方式选择 */}
              <div className="space-y-4">
                <Label className="text-sm font-medium text-gray-700">
                  选择上传方式 <span className="text-red-500">*</span>
                </Label>

                <div className="grid grid-cols-2 gap-3">
                  {/* 本地上传 */}
                  <div
                    className={`relative p-3 border-2 rounded-lg cursor-pointer transition-all ${
                      uploadMethod === 'local'
                        ? 'border-[#7bb8ff]'
                        : 'border-gray-300 hover:border-gray-400 hover:bg-gray-50'
                    }
                    style={
                      uploadMethod === 'local'
                        ? { background: 'linear-gradient(135deg, rgba(123, 184, 255, 0.1) 0%, rgba(63, 54, 99, 0.1) 100%)' }
                        : undefined
                    }`}
                    onClick={() => setUploadMethod('local')}
                  >
                    <div className="text-center">
                      <div
                        className="mx-auto w-8 h-8 rounded-full flex items-center justify-center mb-2"
                        style={{
                          background:
                            'linear-gradient(135deg, rgba(123, 184, 255, 0.2) 0%, rgba(63, 54, 99, 0.2) 100%)',
                        }}
                      >
                        <Upload className="h-4 w-4 text-[#7bb8ff]" />
                      </div>
                      <h4 className="text-sm font-medium text-gray-900 mb-1">
                        本地上传
                      </h4>
                      <p className="text-xs text-gray-600 leading-relaxed">
                        从您的设备上传简历文件
                      </p>
                    </div>
                    {uploadMethod === 'local' && (
                      <div
                        className="absolute top-2 right-2 w-4 h-4 rounded-full flex items-center justify-center"
                        style={{
                          background:
                            'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                        }}
                      >
                        <div className="w-1.5 h-1.5 bg-white rounded-full"></div>
                      </div>
                    )}
                  </div>

                  {/* 链接导入 - 置灰不可选 */}
                  <div className="relative p-3 border-2 border-gray-200 rounded-lg opacity-50 cursor-not-allowed">
                    <div className="text-center">
                      <div
                        className="mx-auto w-8 h-8 rounded-full flex items-center justify-center mb-2"
                        style={{
                          background:
                            'linear-gradient(135deg, rgba(123, 184, 255, 0.2) 0%, rgba(63, 54, 99, 0.2) 100%)',
                        }}
                      >
                        <Link className="h-4 w-4 text-[#7bb8ff]" />
                      </div>
                      <h4 className="text-sm font-medium text-gray-900 mb-1">
                        链接导入
                      </h4>
                      <p className="text-xs text-gray-600 leading-relaxed">
                        通过简历链接导入
                      </p>
                    </div>
                  </div>
                </div>

                {/* 本地上传 - 文件选择和列表 */}
                {uploadMethod === 'local' && (
                  <div className="space-y-3">
                    {/* 拖拽上传区域 */}
                    <div
                      className={`relative border-2 border-dashed rounded-xl p-8 transition-all cursor-pointer ${
                        isDragging
                          ? 'border-[#7bb8ff]'
                          : 'border-gray-300 hover:border-[#7bb8ff] hover:bg-gray-50'
                      }
                      style={
                        isDragging
                          ? { background: 'linear-gradient(135deg, rgba(123, 184, 255, 0.1) 0%, rgba(63, 54, 99, 0.1) 100%)' }
                          : undefined
                      } ${uploading ? 'opacity-50 cursor-not-allowed' : ''}`}
                      onClick={() => !uploading && handleSelectFiles()}
                      onDragEnter={handleDragEnter}
                      onDragOver={handleDragOver}
                      onDragLeave={handleDragLeave}
                      onDrop={handleDrop}
                    >
                      <div className="text-center">
                        <div
                          className="mx-auto w-16 h-16 rounded-full flex items-center justify-center mb-4"
                          style={{
                            background:
                              'linear-gradient(135deg, rgba(123, 184, 255, 0.2) 0%, rgba(63, 54, 99, 0.2) 100%)',
                          }}
                        >
                          <Upload
                            className={`h-8 w-8 ${isDragging ? 'text-[#7bb8ff] animate-bounce' : 'text-[#7bb8ff]'}`}
                          />
                        </div>
                        <h4 className="text-base font-medium text-gray-900 mb-2">
                          {isDragging
                            ? '释放文件以上传'
                            : selectedFiles.length > 0
                              ? '点击或拖拽文件到此处重新选择'
                              : '点击或拖拽文件到此处'}
                        </h4>
                        <p className="text-sm text-gray-600 leading-relaxed">
                          支持 PDF、DOC、DOCX 格式，单个文件最大 100MB
                        </p>
                        <p className="text-xs text-gray-500 mt-2">
                          可同时选择多个文件批量上传
                        </p>
                      </div>
                    </div>

                    {/* 已选择的文件列表 */}
                    {selectedFiles.length > 0 && (
                      <div className="bg-gray-50 rounded-lg p-4 space-y-3">
                        <div className="flex items-center justify-between">
                          <div className="text-sm font-medium text-gray-700">
                            已选择 {selectedFiles.length} 个文件
                          </div>
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedFiles([]);
                            }}
                            className="text-xs text-[#7bb8ff] hover:text-[#6aa8ee] font-medium"
                          >
                            清空列表
                          </button>
                        </div>
                        <div className="space-y-2 max-h-40 overflow-y-auto">
                          {selectedFiles.map((file, index) => (
                            <div
                              key={index}
                              className="flex items-center justify-between text-xs bg-white rounded-lg px-3 py-2.5 shadow-sm"
                            >
                              <div className="flex items-center gap-2 flex-1 min-w-0">
                                <FileText className="w-4 h-4 text-[#7bb8ff] flex-shrink-0" />
                                <span className="truncate text-gray-700 font-medium">
                                  {file.name}
                                </span>
                              </div>
                              <div className="flex items-center gap-3 ml-2 flex-shrink-0">
                                <span className="text-gray-500">
                                  {(file.size / 1024 / 1024).toFixed(2)} MB
                                </span>
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    setSelectedFiles(
                                      selectedFiles.filter(
                                        (_, i) => i !== index
                                      )
                                    );
                                  }}
                                  className="text-gray-400 hover:text-red-500 transition-colors"
                                  title="移除此文件"
                                >
                                  <X className="w-4 h-4" />
                                </button>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* 上传进度和错误信息 */}
                    {uploading && (
                      <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                        <div className="flex items-center justify-between text-sm text-blue-700 mb-2">
                          <span className="font-medium">上传中...</span>
                          <span className="font-semibold">
                            {uploadProgress}%
                          </span>
                        </div>
                        <div className="w-full bg-blue-100 rounded-full h-2">
                          <div
                            className="h-2 rounded-full transition-all duration-300"
                            style={{
                              width: `${uploadProgress}%`,
                              background:
                                'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                            }}
                          ></div>
                        </div>
                      </div>
                    )}

                    {error && !uploading && (
                      <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                        <div className="flex items-start gap-2">
                          <XCircle className="w-4 h-4 text-red-500 mt-0.5 flex-shrink-0" />
                          <div className="text-sm text-red-700">{error}</div>
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </>
          )}

          {currentStep === 'preview' && (
            <>
              {/* 上传进度展示 */}
              <div className="space-y-4">
                {/* 批量上传状态展示 - 整合进度条和文件详情 */}
                {uploadStatus && (
                  <div className="bg-gradient-to-r from-blue-50 to-blue-100 rounded-lg p-3 border border-blue-100">
                    {/* 进度条 */}
                    <div className="mb-3">
                      <div className="flex items-center justify-between mb-1.5">
                        <span className="text-xs font-medium text-gray-700">
                          上传进度
                        </span>
                        <span className="text-xs font-medium text-[#7bb8ff]">
                          {uploadProgress}%
                        </span>
                      </div>
                      <div className="w-full bg-white rounded-full h-2 shadow-inner">
                        <div
                          className="bg-gradient-to-r from-[#7bb8ff] to-[#5aa3e6] h-2 rounded-full transition-all duration-500 ease-out relative overflow-hidden"
                          style={{ width: `${uploadProgress}%` }}
                        >
                          <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 to-transparent animate-shimmer"></div>
                        </div>
                      </div>
                    </div>

                    {/* 统计数字 */}
                    <div className="grid grid-cols-3 gap-2 mb-3">
                      <div className="bg-white rounded px-2 py-1.5 text-center">
                        <p className="text-[10px] text-gray-500 mb-0.5">总数</p>
                        <p className="text-sm font-semibold text-gray-900">
                          {uploadStatus.total_count}
                        </p>
                      </div>
                      <div className="bg-white rounded px-2 py-1.5 text-center">
                        <p className="text-[10px] text-gray-500 mb-0.5">成功</p>
                        <p className="text-sm font-semibold text-green-600">
                          {uploadStatus.success_count}
                        </p>
                      </div>
                      <div className="bg-white rounded px-2 py-1.5 text-center">
                        <p className="text-[10px] text-gray-500 mb-0.5">失败</p>
                        <p className="text-sm font-semibold text-red-600">
                          {uploadStatus.failed_count}
                        </p>
                      </div>
                    </div>

                    {/* 文件上传详情列表 */}
                    {uploadStatus.items && uploadStatus.items.length > 0 && (
                      <div className="space-y-1.5 max-h-48 overflow-y-auto custom-scrollbar">
                        {uploadStatus.items.map((item, index) => (
                          <div
                            key={item.item_id || index}
                            className="flex items-center justify-between bg-white rounded px-2.5 py-2 text-xs"
                          >
                            <div className="flex items-center gap-2 flex-1 min-w-0">
                              <div className="flex-shrink-0">
                                {item.status === 'completed' ? (
                                  <CheckCircle className="w-3.5 h-3.5 text-green-500" />
                                ) : item.status === 'failed' ? (
                                  <XCircle className="w-3.5 h-3.5 text-red-500" />
                                ) : item.status === 'processing' ? (
                                  <div className="w-3.5 h-3.5 border-2 border-[#7bb8ff] border-t-transparent rounded-full animate-spin"></div>
                                ) : (
                                  <Clock className="w-3.5 h-3.5 text-gray-400" />
                                )}
                              </div>
                              <div className="flex-1 min-w-0">
                                <p className="truncate text-gray-700 font-medium">
                                  {item.filename}
                                </p>
                                {item.error_message && (
                                  <p className="text-[10px] text-red-600 mt-0.5 truncate">
                                    {item.error_message}
                                  </p>
                                )}
                              </div>
                            </div>
                            <div className="flex-shrink-0 ml-2">
                              {item.status === 'completed' ? (
                                <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-green-100 text-green-700">
                                  完成
                                </span>
                              ) : item.status === 'failed' ? (
                                <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-100 text-red-700">
                                  失败
                                </span>
                              ) : item.status === 'processing' ? (
                                <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-blue-100 text-blue-700">
                                  处理中
                                </span>
                              ) : (
                                <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-gray-100 text-gray-600">
                                  等待
                                </span>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    )}

                    {/* 简历解析进度 - 整合在上传进度框体内 */}
                    {uploadStatus.status === 'completed' &&
                      uploadStatus.items &&
                      uploadStatus.items.length > 0 &&
                      uploadStatus.items.some(
                        (item) => item.status === 'completed' && item.resume_id
                      ) && (
                        <div className="mt-3 pt-3 border-t border-blue-200">
                          <h4 className="text-xs font-medium text-gray-700 mb-2 flex items-center gap-1.5">
                            <FileText className="w-3.5 h-3.5 text-[#7bb8ff]" />
                            简历解析进度
                          </h4>
                          <div className="space-y-1.5 max-h-40 overflow-y-auto custom-scrollbar">
                            {uploadStatus.items
                              .filter(
                                (item) =>
                                  item.status === 'completed' && item.resume_id
                              )
                              .map((item) => (
                                <ResumeParsingStatus
                                  key={item.resume_id}
                                  resumeId={item.resume_id!}
                                  filename={item.filename}
                                />
                              ))}
                          </div>
                        </div>
                      )}
                  </div>
                )}

                {/* 如果有成功的简历，展示简历信息 */}
                {uploadedResumes.length > 0 && (
                  <div>
                    <div className="text-center mb-2">
                      <div className="flex items-center justify-between">
                        <div className="flex-1"></div>
                        <div className="flex-1 text-center">
                          <h3 className="text-sm font-semibold text-gray-900 mb-1">
                            简历信息预览
                          </h3>
                          <p className="text-xs text-gray-600">
                            以下是解析到的简历信息
                            {uploadedResumes.length > 1 && (
                              <span className="text-[#7bb8ff] ml-1">
                                ({currentResumeIndex + 1}/
                                {uploadedResumes.length})
                              </span>
                            )}
                          </p>
                        </div>
                        <div className="flex-1"></div>
                      </div>
                    </div>

                    {/* 简历列表容器 - 支持上下滚动查看多个简历 */}
                    <div className="space-y-2">
                      {/* 导航指示器 */}
                      {uploadedResumes.length > 1 && (
                        <div className="flex items-center justify-center gap-2 mb-2">
                          {uploadedResumes.map((_, index) => (
                            <button
                              key={index}
                              onClick={() => setCurrentResumeIndex(index)}
                              className={`h-1.5 rounded-full transition-all ${
                                index === currentResumeIndex
                                  ? 'w-8'
                                  : 'w-1.5 bg-gray-300 hover:bg-gray-400'
                              }
                              style={
                                index === currentResumeIndex
                                  ? { background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)' }
                                  : undefined
                              }`}
                              title={`查看第 ${index + 1} 份简历`}
                            />
                          ))}
                        </div>
                      )}

                      {/* 简历卡片容器 - 可滚动 */}
                      <div className="max-h-[400px] overflow-y-auto custom-scrollbar">
                        {uploadedResumes.map((resume, index) => (
                          <div
                            key={resume.id}
                            className={`mb-3 ${index !== uploadedResumes.length - 1 ? 'border-b border-gray-300 pb-3' : ''}`}
                          >
                            {/* 简历信息卡片 */}
                            <div className="bg-gray-50 rounded-lg p-3 space-y-3">
                              {/* 简历序号标识 */}
                              {uploadedResumes.length > 1 && (
                                <div className="mb-2 flex items-center justify-between">
                                  <span className="text-xs font-medium text-[#7bb8ff]">
                                    简历 {index + 1} / {uploadedResumes.length}
                                  </span>
                                  {resume.name && (
                                    <span className="text-xs font-medium text-gray-700">
                                      {resume.name}
                                    </span>
                                  )}
                                </div>
                              )}

                              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                                {/* 基本信息 */}
                                <div className="space-y-2">
                                  <div className="flex items-center space-x-2">
                                    <User className="h-3.5 w-3.5 text-gray-500" />
                                    <span className="text-xs text-gray-600">
                                      姓名：
                                    </span>
                                    <span
                                      className={`text-xs font-medium ${resume.name ? 'text-gray-900' : 'text-gray-400'}`}
                                    >
                                      {resume.name || '无数据'}
                                    </span>
                                  </div>
                                  <div className="flex items-center space-x-2">
                                    <Mail className="h-3.5 w-3.5 text-gray-500" />
                                    <span className="text-xs text-gray-600">
                                      邮箱：
                                    </span>
                                    <span
                                      className={`text-xs font-medium truncate ${resume.email ? 'text-gray-900' : 'text-gray-400'}`}
                                    >
                                      {resume.email || '无数据'}
                                    </span>
                                  </div>
                                  <div className="flex items-center space-x-2">
                                    <Phone className="h-3.5 w-3.5 text-gray-500" />
                                    <span className="text-xs text-gray-600">
                                      电话：
                                    </span>
                                    <span
                                      className={`text-xs font-medium ${resume.phone ? 'text-gray-900' : 'text-gray-400'}`}
                                    >
                                      {resume.phone || '无数据'}
                                    </span>
                                  </div>
                                </div>

                                {/* 详细信息 */}
                                <div className="space-y-2">
                                  <div className="flex items-center space-x-2">
                                    <MapPin className="h-3.5 w-3.5 text-gray-500" />
                                    <span className="text-xs text-gray-600">
                                      城市：
                                    </span>
                                    <span
                                      className={`text-xs font-medium ${resume.current_city ? 'text-gray-900' : 'text-gray-400'}`}
                                    >
                                      {resume.current_city || '无数据'}
                                    </span>
                                  </div>
                                  <div className="flex items-center space-x-2">
                                    <GraduationCap className="h-3.5 w-3.5 text-gray-500" />
                                    <span className="text-xs text-gray-600">
                                      学历：
                                    </span>
                                    <span
                                      className={`text-xs font-medium ${resume.highest_education ? 'text-gray-900' : 'text-gray-400'}`}
                                    >
                                      {resume.highest_education || '无数据'}
                                    </span>
                                  </div>
                                  <div className="flex items-center space-x-2">
                                    <Briefcase className="h-3.5 w-3.5 text-gray-500" />
                                    <span className="text-xs text-gray-600">
                                      经验：
                                    </span>
                                    <span
                                      className={`text-xs font-medium ${resume.years_experience ? 'text-gray-900' : 'text-gray-400'}`}
                                    >
                                      {resume.years_experience
                                        ? `${resume.years_experience}年`
                                        : '无数据'}
                                    </span>
                                  </div>
                                </div>
                              </div>

                              {/* 状态信息 */}
                              <div className="pt-2 border-t border-gray-200">
                                <div className="flex items-center justify-between text-xs">
                                  <div className="flex items-center space-x-1.5">
                                    <span className="text-gray-600">
                                      解析状态：
                                    </span>
                                    <span
                                      className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium ${
                                        resume.status === 'completed'
                                          ? 'text-[#7bb8ff]'
                                          : resume.status === 'processing'
                                            ? 'bg-yellow-100 text-yellow-800'
                                            : resume.status === 'pending'
                                              ? 'bg-blue-100 text-blue-800'
                                              : 'bg-red-100 text-red-800'
                                      }`}
                                    >
                                      {resume.status === 'completed'
                                        ? '解析完成'
                                        : resume.status === 'processing'
                                          ? '解析中'
                                          : resume.status === 'pending'
                                            ? '待解析'
                                            : '解析失败'}
                                    </span>
                                  </div>
                                  <div className="text-[10px] text-gray-500">
                                    {formatDate(resume.created_at, 'datetime')}
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>

                    {/* 内容提示 */}
                    {showContentTip && (
                      <div
                        className="border border-[#7bb8ff]/30 rounded-lg p-3 mt-3"
                        style={{
                          background:
                            'linear-gradient(135deg, rgba(123, 184, 255, 0.1) 0%, rgba(63, 54, 99, 0.1) 100%)',
                        }}
                      >
                        <div className="flex items-center space-x-2">
                          <CheckCircle className="h-4 w-4 text-[#7bb8ff]" />
                          <span className="text-xs font-medium text-[#7bb8ff]">
                            简历上传成功！
                          </span>
                        </div>
                        <p className="text-xs text-[#7bb8ff] mt-1">
                          简历正在后台解析中，解析完成后您可以在简历列表中查看详细信息。
                        </p>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </>
          )}

          {currentStep === 'complete' && uploadedResumes.length > 0 && (
            <>
              {/* 完成步骤内容 */}
              <div className="space-y-4">
                {/* 解析完成的简历信息展示 */}
                <div className="bg-gradient-to-r from-[#7bb8ff]/10 to-blue-50 rounded-lg p-4">
                  <div className="text-center mb-3">
                    <h4 className="text-sm font-semibold text-gray-900 mb-1">
                      解析结果
                    </h4>
                    <p className="text-xs text-gray-600">
                      以下是从简历中提取的关键信息
                      {uploadedResumes.length > 1 && (
                        <span className="text-[#7bb8ff] ml-1">
                          (共 {uploadedResumes.length} 份简历)
                        </span>
                      )}
                    </p>
                  </div>

                  {/* 导航指示器 */}
                  {uploadedResumes.length > 1 && (
                    <div className="flex items-center justify-center gap-2 mb-3">
                      {uploadedResumes.map((_, index) => (
                        <button
                          key={index}
                          onClick={() => setCurrentResumeIndex(index)}
                          className={`h-1.5 rounded-full transition-all ${
                            index === currentResumeIndex
                              ? 'w-8'
                              : 'w-1.5 bg-gray-300 hover:bg-gray-400'
                          }`}
                          style={
                            index === currentResumeIndex
                              ? {
                                  background:
                                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                                }
                              : undefined
                          }
                          title={`查看第 ${index + 1} 份简历`}
                        />
                      ))}
                    </div>
                  )}

                  {/* 简历列表容器 - 可滚动 */}
                  <div className="max-h-[500px] overflow-y-auto custom-scrollbar">
                    {uploadedResumes.map((resume, index) => {
                      const detail = resume; // 使用已经获取的详情
                      return (
                        <div
                          key={resume.id}
                          className={`${index !== uploadedResumes.length - 1 ? 'border-b border-gray-300 pb-4 mb-4' : ''}`}
                        >
                          {/* 简历序号标识 */}
                          {uploadedResumes.length > 1 && (
                            <div className="mb-3 flex items-center justify-between">
                              <span className="text-xs font-medium text-[#7bb8ff]">
                                简历 {index + 1} / {uploadedResumes.length}
                              </span>
                              {detail.name && (
                                <span className="text-xs font-semibold text-gray-900">
                                  {detail.name}
                                </span>
                              )}
                            </div>
                          )}

                          {/* 基本信息和职业信息 - 左右布局 */}
                          <div className="flex gap-4 mb-3">
                            {/* 基本信息 - 左侧 */}
                            <div className="flex-1 bg-white rounded-lg p-3">
                              <h5 className="text-xs font-medium text-gray-900 border-b border-gray-200 pb-1.5 mb-2">
                                基本信息
                              </h5>
                              <div className="space-y-1.5">
                                <div className="flex items-center justify-between">
                                  <span className="text-xs text-gray-600">
                                    姓名
                                  </span>
                                  <span
                                    className={`text-xs font-medium ${detail.name ? 'text-gray-900' : 'text-gray-400'}`}
                                  >
                                    {detail.name || '无数据'}
                                  </span>
                                </div>
                                <div className="flex items-center justify-between">
                                  <span className="text-xs text-gray-600">
                                    邮箱
                                  </span>
                                  <span
                                    className={`text-xs font-medium truncate ml-2 ${detail.email ? 'text-gray-900' : 'text-gray-400'}`}
                                  >
                                    {detail.email || '无数据'}
                                  </span>
                                </div>
                                <div className="flex items-center justify-between">
                                  <span className="text-xs text-gray-600">
                                    电话
                                  </span>
                                  <span
                                    className={`text-xs font-medium ${detail.phone ? 'text-gray-900' : 'text-gray-400'}`}
                                  >
                                    {detail.phone || '无数据'}
                                  </span>
                                </div>
                                <div className="flex items-center justify-between">
                                  <span className="text-xs text-gray-600">
                                    城市
                                  </span>
                                  <span
                                    className={`text-xs font-medium ${detail.current_city ? 'text-gray-900' : 'text-gray-400'}`}
                                  >
                                    {detail.current_city || '无数据'}
                                  </span>
                                </div>
                              </div>
                            </div>

                            {/* 职业信息 - 右侧 */}
                            <div className="flex-1 bg-white rounded-lg p-3">
                              <h5 className="text-xs font-medium text-gray-900 border-b border-gray-200 pb-1.5 mb-2">
                                职业信息
                              </h5>
                              <div className="space-y-1.5">
                                <div className="flex items-center justify-between">
                                  <span className="text-xs text-gray-600">
                                    学历
                                  </span>
                                  <span
                                    className={`text-xs font-medium ${detail.highest_education ? 'text-gray-900' : 'text-gray-400'}`}
                                  >
                                    {detail.highest_education || '无数据'}
                                  </span>
                                </div>
                                <div className="flex items-center justify-between">
                                  <span className="text-xs text-gray-600">
                                    工作经验
                                  </span>
                                  <span
                                    className={`text-xs font-medium ${detail.years_experience ? 'text-gray-900' : 'text-gray-400'}`}
                                  >
                                    {detail.years_experience
                                      ? `${detail.years_experience}年`
                                      : '无数据'}
                                  </span>
                                </div>
                                <div className="flex items-center justify-between">
                                  <span className="text-xs text-gray-600">
                                    解析时间
                                  </span>
                                  <span
                                    className={`text-xs font-medium ${detail.parsed_at ? 'text-gray-900' : 'text-gray-400'}`}
                                  >
                                    {detail.parsed_at
                                      ? formatDate(detail.parsed_at)
                                      : '无数据'}
                                  </span>
                                </div>
                                <div className="flex items-center justify-between">
                                  <span className="text-xs text-gray-600">
                                    解析状态
                                  </span>
                                  <span
                                    className={`inline-flex items-center px-1.5 py-0.5 rounded-full text-[10px] font-medium ${
                                      detail.status === 'completed'
                                        ? 'text-[#7bb8ff]'
                                        : detail.status === 'processing'
                                          ? 'bg-yellow-100 text-yellow-800'
                                          : detail.status === 'pending'
                                            ? 'bg-blue-100 text-blue-800'
                                            : 'bg-red-100 text-red-800'
                                    }`}
                                  >
                                    {detail.status === 'completed'
                                      ? '解析完成'
                                      : detail.status === 'processing'
                                        ? '解析中'
                                        : detail.status === 'pending'
                                          ? '待解析'
                                          : '解析失败'}
                                  </span>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>

                {/* 操作提示 */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                  <div className="flex items-start space-x-2">
                    <CheckCircle className="h-4 w-4 text-blue-600 mt-0.5" />
                    <div>
                      <h5 className="text-xs font-medium text-blue-900">
                        简历已成功保存
                      </h5>
                      <p className="text-xs text-blue-700 mt-0.5">
                        您可以在简历管理页面查看完整的简历信息，或继续上传更多简历。
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </>
          )}
        </div>

        {/* 底部按钮 */}
        {currentStep === 'upload' && (
          <div className="flex-shrink-0 px-6 py-4 border-t border-gray-200 bg-gray-50">
            <div className="flex flex-col gap-2">
              {/* 验证提示 */}
              {(selectedJobIds.length === 0 || uploadMethod !== 'local') && (
                <div className="text-xs text-amber-600 bg-amber-50 px-3 py-2 rounded-lg">
                  {selectedJobIds.length === 0 && '⚠️ 请先选择岗位'}
                  {selectedJobIds.length === 0 &&
                    uploadMethod !== 'local' &&
                    '，并'}
                  {uploadMethod !== 'local' && '⚠️ 请选择上传方式'}
                </div>
              )}

              <div className="flex justify-end">
                <Button
                  onClick={handleNext}
                  disabled={
                    uploadMethod !== 'local' ||
                    uploading ||
                    selectedJobIds.length === 0
                  }
                  className="hover:bg-[#5aa3e6] disabled:bg-gray-300 disabled:cursor-not-allowed text-white px-6 py-2.5 rounded-lg flex items-center gap-2 font-medium transition-colors"
                  style={{
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }}
                >
                  {uploading ? '上传中...' : '下一步'}
                  <ArrowRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
