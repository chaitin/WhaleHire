import { useEffect, useState } from 'react';
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
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useResumeUpload } from '@/hooks/useResume';
import { useResumePollingManager } from '@/hooks/useResumePollingManager';
import { Resume, ResumeParseProgress, ResumeDetail } from '@/types/resume';
import { formatDate } from '@/lib/utils';
import { ResumeParseProgressComponent } from '@/components/ui/resume-parse-progress';
import { getResumeProgress, getResumeDetail } from '@/services/resume';
import { MultiSelect, Option } from '@/components/ui/multi-select';
import { listJobProfiles } from '@/services/job-profile';

interface UploadResumeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: (resume?: Resume) => void;
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
  const [progress, setProgress] = useState<ResumeParseProgress | null>(null);
  const [resumeDetail, setResumeDetail] = useState<ResumeDetail | null>(null);

  // 岗位选择相关状态
  const [selectedJobIds, setSelectedJobIds] = useState<string[]>([]);
  const [jobOptions, setJobOptions] = useState<Option[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);

  const { uploadFile, uploading, uploadProgress, error } = useResumeUpload();
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
      
      // 转换为 MultiSelect 所需的选项格式
      const options: Option[] = response.items.map((job) => ({
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
      // 关闭弹窗时停止轮询
      stopPolling();
      // 只有在解析完成后关闭弹窗时才重置状态
      if (currentStep === 'complete') {
        setPosition('');
        setUploadMethod(null);
        setCurrentStep('upload');
        setUploadedResume(null);
        setShowContentTip(false);
        setSelectedJobIds([]); // 重置岗位选择
      }
    }
  }, [open, stopPolling, currentStep]);

  // 监听简历上传成功后的状态变化，手动启动轮询
  useEffect(() => {
    if (uploadedResume && currentStep === 'preview') {
      // 如果简历状态需要轮询，则启动轮询
      if (
        uploadedResume.status === 'pending' ||
        uploadedResume.status === 'processing'
      ) {
        console.log('启动简历解析进度轮询:', uploadedResume.id);
        startPolling(uploadedResume.id, async (newProgress) => {
          setProgress(newProgress);

          // 解析完成后获取详细信息并进入完成步骤
          if (newProgress.status === 'completed') {
            try {
              const detail = await getResumeDetail(uploadedResume.id);
              setResumeDetail(detail);
              setCurrentStep('complete');
            } catch (error) {
              console.error('获取简历详情失败:', error);
            }
          } else if (newProgress.status === 'failed') {
            console.error('简历解析失败:', newProgress);
          }
        });
      } else {
        // 获取初始进度状态
        getResumeProgress(uploadedResume.id)
          .then(setProgress)
          .catch(console.error);
      }
    }
  }, [uploadedResume, currentStep, startPolling]);

  const handleFileUpload = () => {
    // 创建文件输入元素
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.pdf,.doc,.docx';
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (file) {
        try {
          // 将选中的岗位ID作为数组传递给上传函数
          const resume = await uploadFile(file, selectedJobIds, position || undefined);
          setUploadedResume(resume);
          setCurrentStep('preview');
          onSuccess?.(resume);
        } catch (err) {
          // 错误已在 hook 中处理，这里可以根据需要添加额外逻辑
          console.error('上传简历失败:', err);
        }
      }
    };
    input.click();
  };

  const handleNext = () => {
    if (currentStep === 'upload' && uploadMethod === 'local' && !uploading) {
      // 验证是否选择了岗位
      if (selectedJobIds.length === 0) {
        // 这里可以添加提示
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
                    ? 'bg-green-500 text-white'
                    : isStepActive('upload')
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-200 text-gray-600'
                }`}
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
                      ? 'text-green-600'
                      : isStepActive('upload')
                        ? 'text-blue-600'
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
                className={`h-full bg-green-500 rounded transition-all duration-300 ${
                  isStepCompleted('upload') ? 'w-full' : 'w-0'
                }`}
              ></div>
            </div>

            {/* 步骤2 - 预览内容 */}
            <div className="flex items-center">
              <div
                className={`flex items-center justify-center w-10 h-10 rounded-full text-sm font-medium ${
                  isStepCompleted('preview')
                    ? 'bg-green-500 text-white'
                    : isStepActive('preview')
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-200 text-gray-600'
                }`}
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
                      ? 'text-green-600'
                      : isStepActive('preview')
                        ? 'text-blue-600'
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
                className={`h-full bg-green-500 rounded transition-all duration-300 ${
                  isStepCompleted('preview') ? 'w-full' : 'w-0'
                }`}
              ></div>
            </div>

            {/* 步骤3 - 完成 */}
            <div className="flex items-center">
              <div
                className={`flex items-center justify-center w-10 h-10 rounded-full text-sm font-medium ${
                  isStepCompleted('complete')
                    ? 'bg-green-500 text-white'
                    : isStepActive('complete')
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-200 text-gray-600'
                }`}
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
                      ? 'text-green-600'
                      : isStepActive('complete')
                        ? 'text-blue-600'
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
              </div>

              {/* 上传方式选择 */}
              <div className="space-y-4">
                <Label className="text-sm font-medium text-gray-700">
                  选择上传方式 <span className="text-red-500">*</span>
                </Label>

                <div className="grid grid-cols-2 gap-5">
                  {/* 本地上传 */}
                  <div
                    className={`relative p-6 border-2 border-dashed rounded-xl cursor-pointer transition-all ${
                      uploadMethod === 'local'
                        ? 'border-green-500 bg-green-50'
                        : 'border-gray-300 hover:border-gray-400 hover:bg-gray-50'
                    }`}
                    onClick={() => setUploadMethod('local')}
                  >
                    <div className="text-center">
                      <div className="mx-auto w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mb-4">
                        <Upload className="h-5 w-5 text-blue-600" />
                      </div>
                      <h4 className="text-base font-medium text-gray-900 mb-2">
                        本地上传
                      </h4>
                      <p className="text-sm text-gray-600 leading-relaxed">
                        从您的设备上传简历文件
                      </p>
                    </div>
                    {uploadMethod === 'local' && (
                      <div className="absolute top-3 right-3 w-5 h-5 bg-green-500 rounded-full flex items-center justify-center">
                        <div className="w-2 h-2 bg-white rounded-full"></div>
                      </div>
                    )}
                  </div>

                  {/* 链接导入 - 置灰不可选 */}
                  <div className="relative p-6 border-2 border-dashed border-gray-200 rounded-xl opacity-50 cursor-not-allowed">
                    <div className="text-center">
                      <div className="mx-auto w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mb-4">
                        <Link className="h-5 w-5 text-green-600" />
                      </div>
                      <h4 className="text-base font-medium text-gray-900 mb-2">
                        链接导入
                      </h4>
                      <p className="text-sm text-gray-600 leading-relaxed">
                        通过简历链接导入
                      </p>
                    </div>
                  </div>
                </div>

                {/* 本地上传说明文字 */}
                {uploadMethod === 'local' && (
                  <div className="text-sm text-gray-600 bg-blue-50 p-3 rounded-lg">
                    默认支持 .pdf、.doc、.docx 格式，文件大小 100M
                    {uploading && (
                      <div className="mt-2 flex items-center justify-between text-xs text-gray-500">
                        <span>上传中...</span>
                        <span>{uploadProgress}%</span>
                      </div>
                    )}
                    {error && !uploading && (
                      <div className="mt-2 text-xs text-red-500">{error}</div>
                    )}
                  </div>
                )}
              </div>
            </>
          )}

          {currentStep === 'preview' && uploadedResume && (
            <>
              {/* 简历预览内容 */}
              <div className="space-y-6">
                <div className="text-center">
                  <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-4">
                    <FileText className="h-8 w-8 text-green-600" />
                  </div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-2">
                    简历上传成功！
                  </h3>
                  <p className="text-sm text-gray-600">
                    以下是解析到的简历信息
                  </p>
                </div>

                {/* 简历信息卡片 */}
                <div className="bg-gray-50 rounded-lg p-6 space-y-4">
                  {/* 解析进度显示 */}
                  {progress && (
                    <div className="mb-6">
                      <ResumeParseProgressComponent
                        progress={progress}
                        showDetails={true}
                        className="bg-white rounded-lg p-4 border border-gray-200"
                      />
                    </div>
                  )}

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {/* 基本信息 */}
                    <div className="space-y-3">
                      <div className="flex items-center space-x-3">
                        <User className="h-4 w-4 text-gray-500" />
                        <span className="text-sm text-gray-600">姓名：</span>
                        <span className="text-sm font-medium text-gray-900">
                          {uploadedResume.name || '待解析'}
                        </span>
                      </div>
                      <div className="flex items-center space-x-3">
                        <Mail className="h-4 w-4 text-gray-500" />
                        <span className="text-sm text-gray-600">邮箱：</span>
                        <span className="text-sm font-medium text-gray-900">
                          {uploadedResume.email || '待解析'}
                        </span>
                      </div>
                      <div className="flex items-center space-x-3">
                        <Phone className="h-4 w-4 text-gray-500" />
                        <span className="text-sm text-gray-600">电话：</span>
                        <span className="text-sm font-medium text-gray-900">
                          {uploadedResume.phone || '待解析'}
                        </span>
                      </div>
                    </div>

                    {/* 详细信息 */}
                    <div className="space-y-3">
                      <div className="flex items-center space-x-3">
                        <MapPin className="h-4 w-4 text-gray-500" />
                        <span className="text-sm text-gray-600">城市：</span>
                        <span className="text-sm font-medium text-gray-900">
                          {uploadedResume.current_city || '待解析'}
                        </span>
                      </div>
                      <div className="flex items-center space-x-3">
                        <GraduationCap className="h-4 w-4 text-gray-500" />
                        <span className="text-sm text-gray-600">学历：</span>
                        <span className="text-sm font-medium text-gray-900">
                          {uploadedResume.highest_education || '待解析'}
                        </span>
                      </div>
                      <div className="flex items-center space-x-3">
                        <Briefcase className="h-4 w-4 text-gray-500" />
                        <span className="text-sm text-gray-600">经验：</span>
                        <span className="text-sm font-medium text-gray-900">
                          {uploadedResume.years_experience
                            ? `${uploadedResume.years_experience}年`
                            : '待解析'}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* 状态信息 */}
                  <div className="pt-4 border-t border-gray-200">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <span className="text-sm text-gray-600">
                          解析状态：
                        </span>
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            uploadedResume.status === 'completed'
                              ? 'bg-green-100 text-green-800'
                              : uploadedResume.status === 'processing'
                                ? 'bg-yellow-100 text-yellow-800'
                                : uploadedResume.status === 'pending'
                                  ? 'bg-blue-100 text-blue-800'
                                  : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {uploadedResume.status === 'completed'
                            ? '解析完成'
                            : uploadedResume.status === 'processing'
                              ? '解析中'
                              : uploadedResume.status === 'pending'
                                ? '待解析'
                                : '解析失败'}
                        </span>
                      </div>
                      <div className="text-xs text-gray-500">
                        上传时间：
                        {formatDate(uploadedResume.created_at, 'datetime')}
                      </div>
                    </div>
                  </div>
                </div>

                {/* PDF预览区域（如果有文件URL） */}
                {uploadedResume.resume_file_url && (
                  <div className="bg-white border border-gray-200 rounded-lg p-4">
                    <div className="text-sm font-medium text-gray-700 mb-3">
                      简历文件预览
                    </div>
                    <div className="bg-gray-100 rounded-lg p-8 text-center">
                      <FileText className="h-12 w-12 text-gray-400 mx-auto mb-3" />
                      <p className="text-sm text-gray-600 mb-2">
                        PDF文件已上传
                      </p>
                      <p className="text-xs text-gray-500">
                        文件正在后台解析中，请稍候...
                      </p>
                    </div>
                  </div>
                )}

                {/* 内容提示 */}
                {showContentTip && (
                  <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                    <div className="flex items-center space-x-2">
                      <CheckCircle className="h-5 w-5 text-green-600" />
                      <span className="text-sm font-medium text-green-800">
                        简历上传成功！
                      </span>
                    </div>
                    <p className="text-sm text-green-700 mt-1">
                      简历正在后台解析中，解析完成后您可以在简历列表中查看详细信息。
                    </p>
                  </div>
                )}
              </div>
            </>
          )}

          {currentStep === 'complete' && uploadedResume && (
            <>
              {/* 完成步骤内容 */}
              <div className="space-y-6">
                <div className="text-center">
                  <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-4">
                    <CheckCircle className="h-8 w-8 text-green-600" />
                  </div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-2">
                    简历解析完成！
                  </h3>
                  <p className="text-sm text-gray-600">
                    简历信息已成功解析并保存
                  </p>
                </div>

                {/* 解析完成的简历信息展示 */}
                <div className="bg-gradient-to-r from-green-50 to-blue-50 rounded-lg p-6 space-y-4">
                  <div className="text-center mb-4">
                    <h4 className="text-base font-semibold text-gray-900 mb-2">
                      解析结果
                    </h4>
                    <p className="text-sm text-gray-600">
                      以下是从简历中提取的关键信息
                    </p>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {/* 基本信息 */}
                    <div className="bg-white rounded-lg p-4 space-y-3">
                      <h5 className="text-sm font-medium text-gray-900 border-b border-gray-200 pb-2">
                        基本信息
                      </h5>
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-gray-600">姓名</span>
                          <span className="text-sm font-medium text-gray-900">
                            {resumeDetail?.name ||
                              uploadedResume.name ||
                              '未解析'}
                          </span>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-gray-600">邮箱</span>
                          <span className="text-sm font-medium text-gray-900">
                            {resumeDetail?.email ||
                              uploadedResume.email ||
                              '未解析'}
                          </span>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-gray-600">电话</span>
                          <span className="text-sm font-medium text-gray-900">
                            {resumeDetail?.phone ||
                              uploadedResume.phone ||
                              '未解析'}
                          </span>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-gray-600">城市</span>
                          <span className="text-sm font-medium text-gray-900">
                            {resumeDetail?.current_city ||
                              uploadedResume.current_city ||
                              '未解析'}
                          </span>
                        </div>
                      </div>
                    </div>

                    {/* 职业信息 */}
                    <div className="bg-white rounded-lg p-4 space-y-3">
                      <h5 className="text-sm font-medium text-gray-900 border-b border-gray-200 pb-2">
                        职业信息
                      </h5>
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-gray-600">学历</span>
                          <span className="text-sm font-medium text-gray-900">
                            {resumeDetail?.highest_education ||
                              uploadedResume.highest_education ||
                              '未解析'}
                          </span>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-gray-600">
                            工作经验
                          </span>
                          <span className="text-sm font-medium text-gray-900">
                            {resumeDetail?.years_experience ||
                            uploadedResume.years_experience
                              ? `${resumeDetail?.years_experience || uploadedResume.years_experience}年`
                              : '未解析'}
                          </span>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-gray-600">
                            解析时间
                          </span>
                          <span className="text-sm font-medium text-gray-900">
                            {resumeDetail?.parsed_at || uploadedResume.parsed_at
                              ? formatDate(
                                  resumeDetail?.parsed_at ||
                                    uploadedResume.parsed_at
                                )
                              : '未解析'}
                          </span>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-gray-600">
                            解析状态
                          </span>
                          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                            解析完成
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* 技能和经验概览 - 使用resumeDetail数据 */}
                  {resumeDetail && (
                    <div className="bg-white rounded-lg p-4 space-y-3">
                      <h5 className="text-sm font-medium text-gray-900 border-b border-gray-200 pb-2">
                        技能与经验
                      </h5>
                      {resumeDetail.skills &&
                        resumeDetail.skills.length > 0 && (
                          <div>
                            <span className="text-sm text-gray-600">
                              技能：
                            </span>
                            <div className="flex flex-wrap gap-1 mt-1">
                              {resumeDetail.skills
                                .slice(0, 8)
                                .map((skill, index) => (
                                  <span
                                    key={index}
                                    className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                                  >
                                    {skill.skill_name}
                                  </span>
                                ))}
                              {resumeDetail.skills.length > 8 && (
                                <span className="text-xs text-gray-500">
                                  +{resumeDetail.skills.length - 8}个技能
                                </span>
                              )}
                            </div>
                          </div>
                        )}
                      {resumeDetail.experiences &&
                        resumeDetail.experiences.length > 0 && (
                          <div>
                            <span className="text-sm text-gray-600">
                              工作经历：
                            </span>
                            <p className="text-sm text-gray-900 mt-1">
                              共{resumeDetail.experiences.length}段工作经历
                            </p>
                          </div>
                        )}
                      {resumeDetail.educations &&
                        resumeDetail.educations.length > 0 && (
                          <div>
                            <span className="text-sm text-gray-600">
                              教育经历：
                            </span>
                            <p className="text-sm text-gray-900 mt-1">
                              共{resumeDetail.educations.length}段教育经历
                            </p>
                          </div>
                        )}
                    </div>
                  )}
                </div>

                {/* 操作提示 */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <div className="flex items-start space-x-3">
                    <CheckCircle className="h-5 w-5 text-blue-600 mt-0.5" />
                    <div>
                      <h5 className="text-sm font-medium text-blue-900">
                        简历已成功保存
                      </h5>
                      <p className="text-sm text-blue-700 mt-1">
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
                  {selectedJobIds.length === 0 && uploadMethod !== 'local' && '，并'}
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
                  className="bg-green-600 hover:bg-green-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-white px-6 py-2.5 rounded-lg flex items-center gap-2 font-medium transition-colors"
                >
                  {uploading ? '上传中...' : '上传文件'}
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
