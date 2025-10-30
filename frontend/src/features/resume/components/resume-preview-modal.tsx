import { useState, useEffect } from 'react';
import {
  X,
  // Download, // 已注释 - 下载按钮已隐藏
  // Edit, // 已注释 - 编辑按钮已隐藏
  Mail,
  Phone,
  MapPin,
  Eye,
  Sparkles,
  User,
  Calendar,
  Save,
  DollarSign,
  Clock,
  GraduationCap,
  FileText,
  Users,
  Briefcase,
  FolderGit2,
  Award,
  Info,
  Edit2,
} from 'lucide-react';
import { Dialog, DialogContent, DialogTitle } from '@/ui/dialog';
import { Button } from '@/ui/button';
import { Input } from '@/ui/input';
import { Textarea } from '@/ui/textarea';
import { ResumeDetail } from '@/types/resume';
import {
  getResumeDetail,
  /* downloadResumeFile, */ updateResume,
} from '@/services/resume';
import { toast } from '@/ui/toast';
import { MultiSelect, Option } from '@/ui/multi-select';
import { listJobProfiles } from '@/services/job-profile';

interface ResumePreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  resumeId: string;
  onEdit: (resumeId: string) => void;
  onNavigate?: (direction: 'prev' | 'next') => void;
  canNavigatePrev?: boolean;
  canNavigateNext?: boolean;
}

export function ResumePreviewModal({
  isOpen,
  onClose,
  resumeId,
}: ResumePreviewModalProps) {
  const [resumeDetail, setResumeDetail] = useState<ResumeDetail | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  // const [isDownloading, setIsDownloading] = useState(false); // 已注释 - 下载功能已隐藏
  const [isEditMode, setIsEditMode] = useState(false);
  const [editedData, setEditedData] = useState<ResumeDetail | null>(null);

  // 岗位编辑相关状态
  const [isEditingJob, setIsEditingJob] = useState(false);
  const [selectedJobIds, setSelectedJobIds] = useState<string[]>([]);
  const [jobOptions, setJobOptions] = useState<Option[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [savingJobs, setSavingJobs] = useState(false);

  // 获取简历详情
  useEffect(() => {
    const fetchResumeDetail = async () => {
      setIsLoading(true);
      try {
        const detail = await getResumeDetail(resumeId);
        setResumeDetail(detail);
        // 初始化已选岗位
        if (detail.job_positions && detail.job_positions.length > 0) {
          setSelectedJobIds(
            detail.job_positions.map((jp) => jp.job_position_id)
          );
        } else {
          setSelectedJobIds([]);
        }
      } catch (error) {
        console.error('获取简历详情失败:', error);
        toast.error('获取简历详情失败');
      } finally {
        setIsLoading(false);
      }
    };

    if (isOpen && resumeId) {
      fetchResumeDetail();
    }
  }, [isOpen, resumeId]);

  // 获取岗位列表
  useEffect(() => {
    const fetchJobProfiles = async () => {
      if (!isOpen) return;

      setLoadingJobs(true);
      try {
        const response = await listJobProfiles({
          page: 1,
          page_size: 100,
        });

        // 只显示发布状态的岗位
        const publishedJobs = response.items.filter(
          (job) => job.status === 'published'
        );

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

    fetchJobProfiles();
  }, [isOpen]);

  // 下载简历 - 已注释
  // const handleDownload = async () => {
  //   if (!resumeDetail?.resume_file_url) {
  //     toast.error('简历文件不存在');
  //     return;
  //   }

  //   setIsDownloading(true);
  //   try {
  //     await downloadResumeFile(resumeDetail, resumeDetail.name);
  //     toast.success('简历下载成功');
  //   } catch (error) {
  //     console.error('下载简历失败:', error);
  //     toast.error('下载简历失败');
  //   } finally {
  //     setIsDownloading(false);
  //   }
  // };

  // 查看原始简历
  const handleViewOriginal = () => {
    if (resumeDetail?.resume_file_url) {
      window.open(resumeDetail.resume_file_url, '_blank');
    } else {
      toast.error('原始简历文件不存在');
    }
  };

  // AI分析(占位功能)
  const handleAIAnalysis = () => {
    toast.info('AI分析功能开发中');
  };

  // 进入编辑模式 - 已注释
  // const handleEdit = () => {
  //   setIsEditMode(true);
  //   setEditedData(resumeDetail);
  // };

  // 取消编辑
  const handleCancelEdit = () => {
    setIsEditMode(false);
    setEditedData(null);
  };

  // 保存编辑
  const handleSaveEdit = () => {
    // TODO: 调用API保存编辑后的数据
    toast.success('简历保存成功');
    setResumeDetail(editedData);
    setIsEditMode(false);
    setEditedData(null);
  };

  // 更新编辑数据
  const handleDataChange = (
    field: string,
    value: string | number | string[] | unknown[]
  ) => {
    if (!editedData) return;
    setEditedData({
      ...editedData,
      [field]: value,
    });
  };

  // 打开岗位编辑弹窗
  const handleOpenJobEdit = () => {
    setIsEditingJob(true);
  };

  // 取消岗位编辑
  const handleCancelJobEdit = () => {
    setIsEditingJob(false);
    // 恢复原始选择
    if (resumeDetail?.job_positions && resumeDetail.job_positions.length > 0) {
      setSelectedJobIds(
        resumeDetail.job_positions.map((jp) => jp.job_position_id)
      );
    } else {
      setSelectedJobIds([]);
    }
  };

  // 保存岗位选择
  const handleSaveJobs = async () => {
    if (!resumeDetail || selectedJobIds.length === 0) {
      toast.error('请至少选择一个岗位');
      return;
    }

    setSavingJobs(true);
    try {
      // 调用更新API
      await updateResume(resumeDetail.id, {
        job_position_ids: selectedJobIds,
      });

      // 重新获取简历详情以更新界面
      const updatedDetail = await getResumeDetail(resumeDetail.id);
      setResumeDetail(updatedDetail);

      setIsEditingJob(false);
      toast.success('岗位更新成功');
    } catch (error) {
      console.error('更新岗位失败:', error);
      toast.error('更新岗位失败');
    } finally {
      setSavingJobs(false);
    }
  };

  if (!isOpen) return null;

  const displayData = isEditMode ? editedData : resumeDetail;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-[700px] h-[90vh] overflow-hidden p-0 gap-0">
        <DialogTitle className="sr-only">简历预览</DialogTitle>
        {/* 单一框体 */}
        <div className="bg-white rounded-xl overflow-hidden flex flex-col h-full">
          {/* 顶部标题栏 */}
          <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
            <h2 className="text-xl font-semibold text-[#1E293B]">简历预览</h2>
            <Button
              variant="ghost"
              size="sm"
              onClick={onClose}
              className="h-8 w-8 p-0 text-gray-500 hover:text-gray-900 hover:bg-gray-100"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>

          {/* 内容区域 */}
          {isLoading ? (
            <div className="flex items-center justify-center h-[600px]">
              <div className="text-gray-500">加载中...</div>
            </div>
          ) : displayData ? (
            <>
              {/* 渐变头部区域 */}
              <div
                className="px-8 py-5 border-b border-gray-200"
                style={{
                  background:
                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }}
              >
                {/* 姓名和操作按钮 */}
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3">
                    {/* 头像图标 */}
                    <div
                      className="w-12 h-12 rounded-full flex items-center justify-center flex-shrink-0"
                      style={{
                        background:
                          'linear-gradient(135deg, rgba(255,255,255,0.3) 0%, rgba(255,255,255,0.1) 100%)',
                        border: '2px solid rgba(255,255,255,0.6)',
                        boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
                      }}
                    >
                      <svg
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <circle
                          cx="12"
                          cy="8"
                          r="4"
                          fill="white"
                          opacity="0.9"
                        />
                        <path
                          d="M6 18.5C6 15.5 8.5 13 12 13C15.5 13 18 15.5 18 18.5V20H6V18.5Z"
                          fill="white"
                          opacity="0.9"
                        />
                      </svg>
                    </div>

                    <div className="flex items-center gap-2">
                      {isEditMode ? (
                        <Input
                          value={displayData.name}
                          onChange={(e) =>
                            handleDataChange('name', e.target.value)
                          }
                          className="text-lg font-semibold text-white bg-white/20 border-white/30 h-8 w-48"
                          placeholder="姓名"
                        />
                      ) : (
                        <h1 className="text-lg font-semibold text-white">
                          {displayData.name}
                        </h1>
                      )}
                      {/* 岗位名称展示 - 从job_positions获取,加粗 */}
                      {displayData.job_positions &&
                        displayData.job_positions.length > 0 && (
                          <div className="flex items-center gap-2">
                            <div className="relative group">
                              <span className="text-sm font-semibold text-white/90">
                                {displayData.job_positions.length === 1
                                  ? displayData.job_positions[0].job_title
                                  : `${displayData.job_positions[0].job_title} 等${displayData.job_positions.length}个岗位`}
                              </span>
                              {/* Hover展示所有岗位 */}
                              {displayData.job_positions.length > 1 && (
                                <div className="absolute left-0 top-full mt-2 hidden group-hover:block z-10">
                                  <div className="bg-white rounded-lg shadow-lg p-3 min-w-[200px]">
                                    <div className="space-y-1">
                                      {displayData.job_positions.map(
                                        (jobPos, idx) => (
                                          <div
                                            key={jobPos.id || idx}
                                            className="text-sm text-[#1E293B] py-1"
                                          >
                                            {jobPos.job_title}
                                          </div>
                                        )
                                      )}
                                    </div>
                                  </div>
                                </div>
                              )}
                            </div>
                            {/* 编辑岗位图标 */}
                            <button
                              onClick={handleOpenJobEdit}
                              className="w-5 h-5 rounded flex items-center justify-center hover:bg-white/20 transition-colors"
                              title="编辑岗位"
                            >
                              <Edit2 className="w-3.5 h-3.5 text-white/80 hover:text-white" />
                            </button>
                          </div>
                        )}
                    </div>
                  </div>

                  {/* 操作按钮 - 缩小尺寸 */}
                  <div className="flex gap-1">
                    <Button
                      variant="outline"
                      className="h-6 px-2 bg-white/20 border-white/30 text-white hover:bg-white/30 opacity-50 cursor-not-allowed"
                      onClick={handleAIAnalysis}
                      disabled
                    >
                      <Sparkles className="w-3 h-3 mr-1" />
                      <span className="text-xs font-medium">AI分析</span>
                    </Button>
                    <Button
                      variant="outline"
                      className="h-6 px-2 bg-white/20 border-white/30 text-white hover:bg-white/30"
                      onClick={handleViewOriginal}
                      disabled={!displayData.resume_file_url}
                    >
                      <Eye className="w-3 h-3 mr-1" />
                      <span className="text-xs font-medium">查看</span>
                    </Button>
                    {/* 编辑按钮 - 已注释 */}
                    {/* <Button
                      variant="outline"
                      className="h-6 px-2 bg-white/20 border-white/30 text-white hover:bg-white/30"
                      onClick={handleEdit}
                      disabled={isEditMode}
                    >
                      <Edit className="w-3 h-3 mr-1" />
                      <span className="text-xs font-medium">编辑</span>
                    </Button> */}
                    {/* 下载按钮 - 已注释 */}
                    {/* <Button
                      variant="outline"
                      className="h-6 px-2 bg-white/20 border-white/30 text-white hover:bg-white/30"
                      onClick={handleDownload}
                      disabled={isDownloading || !displayData.resume_file_url}
                    >
                      <Download className="w-3 h-3 mr-1" />
                      <span className="text-xs font-medium">下载</span>
                    </Button> */}
                  </div>
                </div>

                {/* 个人信息 - icon+内容形式 */}
                <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
                  {(displayData.phone || isEditMode) && (
                    <div className="flex items-center gap-1.5">
                      <Phone className="w-3.5 h-3.5 text-white" />
                      {isEditMode ? (
                        <Input
                          value={displayData.phone || ''}
                          onChange={(e) =>
                            handleDataChange('phone', e.target.value)
                          }
                          className="text-sm text-white bg-white/20 border-white/30 h-6 w-32"
                          placeholder="电话"
                        />
                      ) : (
                        <span className="text-sm text-white">
                          {displayData.phone}
                        </span>
                      )}
                    </div>
                  )}
                  {(displayData.email || isEditMode) && (
                    <div className="flex items-center gap-1.5">
                      <Mail className="w-3.5 h-3.5 text-white" />
                      {isEditMode ? (
                        <Input
                          value={displayData.email || ''}
                          onChange={(e) =>
                            handleDataChange('email', e.target.value)
                          }
                          className="text-sm text-white bg-white/20 border-white/30 h-6 w-40"
                          placeholder="邮箱"
                        />
                      ) : (
                        <span className="text-sm text-white">
                          {displayData.email}
                        </span>
                      )}
                    </div>
                  )}
                  {((displayData as unknown as { age?: number }).age ||
                    isEditMode) && (
                    <div className="flex items-center gap-1.5">
                      <Calendar className="w-3.5 h-3.5 text-white" />
                      {isEditMode ? (
                        <Input
                          type="number"
                          value={
                            (displayData as unknown as { age?: number }).age ||
                            ''
                          }
                          onChange={(e) =>
                            handleDataChange('age', parseInt(e.target.value))
                          }
                          className="text-sm text-white bg-white/20 border-white/30 h-6 w-20"
                          placeholder="年龄"
                        />
                      ) : (
                        <span className="text-sm text-white">
                          {(displayData as unknown as { age: number }).age}岁
                        </span>
                      )}
                    </div>
                  )}
                  {/* 性别 - 只展示"男"或"女" */}
                  {(displayData.gender === '男' ||
                    displayData.gender === '女' ||
                    isEditMode) && (
                    <div className="flex items-center gap-1.5">
                      <User className="w-3.5 h-3.5 text-white" />
                      {isEditMode ? (
                        <Input
                          value={displayData.gender || ''}
                          onChange={(e) =>
                            handleDataChange('gender', e.target.value)
                          }
                          className="text-sm text-white bg-white/20 border-white/30 h-6 w-20"
                          placeholder="性别"
                        />
                      ) : (
                        <span className="text-sm text-white">
                          {displayData.gender}
                        </span>
                      )}
                    </div>
                  )}
                  {(displayData.current_city || isEditMode) && (
                    <div className="flex items-center gap-1.5">
                      <MapPin className="w-3 h-3.5 text-white" />
                      {isEditMode ? (
                        <Input
                          value={displayData.current_city || ''}
                          onChange={(e) =>
                            handleDataChange('current_city', e.target.value)
                          }
                          className="text-sm text-white bg-white/20 border-white/30 h-6 w-32"
                          placeholder="工作城市"
                        />
                      ) : (
                        <span className="text-sm text-white">
                          {displayData.current_city}
                        </span>
                      )}
                    </div>
                  )}
                </div>
              </div>

              {/* 内容区域 - 所有内容垂直滚动展示 */}
              <div className="flex-1 overflow-y-auto px-6 py-4 bg-white space-y-6 relative">
                {/* 科技感装饰线条 - 弱化背景线条 */}
                <div className="absolute inset-0 opacity-5 pointer-events-none">
                  <div className="absolute top-0 left-0 w-full h-px bg-gradient-to-r from-transparent via-blue-500 to-transparent"></div>
                  <div className="absolute top-1/3 left-0 w-full h-px bg-gradient-to-r from-transparent via-purple-500 to-transparent"></div>
                  <div className="absolute top-2/3 left-0 w-full h-px bg-gradient-to-r from-transparent via-blue-500 to-transparent"></div>
                  <div className="absolute top-0 left-1/4 w-px h-full bg-gradient-to-b from-transparent via-blue-500 to-transparent"></div>
                  <div className="absolute top-0 left-3/4 w-px h-full bg-gradient-to-b from-transparent via-purple-500 to-transparent"></div>
                </div>
                {/* 求职意向 */}
                <JobIntentionSection
                  resumeDetail={displayData}
                  isEditMode={isEditMode}
                  onDataChange={handleDataChange}
                />

                {/* 个人简介 */}
                <IntroSection
                  resumeDetail={displayData}
                  isEditMode={isEditMode}
                  onDataChange={handleDataChange}
                />

                {/* 教育背景 */}
                <EducationSection
                  resumeDetail={displayData}
                  isEditMode={isEditMode}
                  onDataChange={handleDataChange}
                />

                {/* 工作经验 */}
                <ExperienceSection
                  resumeDetail={displayData}
                  isEditMode={isEditMode}
                  onDataChange={handleDataChange}
                />

                {/* 项目经验 */}
                <ProjectsSection
                  resumeDetail={displayData}
                  isEditMode={isEditMode}
                  onDataChange={handleDataChange}
                />

                {/* 技能专长 */}
                <SkillsSection
                  resumeDetail={displayData}
                  isEditMode={isEditMode}
                  onDataChange={handleDataChange}
                />

                {/* 荣誉与资格证书 */}
                <CertificatesSection
                  resumeDetail={displayData}
                  isEditMode={isEditMode}
                  onDataChange={handleDataChange}
                />

                {/* 其他 */}
                <OtherSection
                  resumeDetail={displayData}
                  isEditMode={isEditMode}
                  onDataChange={handleDataChange}
                />
              </div>

              {/* 编辑模式底部按钮 */}
              {isEditMode && (
                <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-white">
                  <Button
                    variant="outline"
                    onClick={handleCancelEdit}
                    className="h-9 px-4"
                  >
                    取消
                  </Button>
                  <Button
                    onClick={handleSaveEdit}
                    className="h-9 px-4 bg-[#3B82F6] text-white hover:bg-[#2563EB]"
                  >
                    <Save className="w-4 h-4 mr-1.5" />
                    保存
                  </Button>
                </div>
              )}
            </>
          ) : (
            <div className="flex items-center justify-center h-[600px]">
              <div className="text-gray-500">暂无数据</div>
            </div>
          )}
        </div>
      </DialogContent>

      {/* 岗位编辑弹窗 */}
      {isEditingJob && (
        <Dialog open={isEditingJob} onOpenChange={handleCancelJobEdit}>
          <DialogContent className="max-w-md">
            <DialogTitle className="text-lg font-semibold text-gray-900">
              编辑岗位
            </DialogTitle>

            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-700">
                  选择岗位 <span className="text-red-500">*</span>
                </label>
                <MultiSelect
                  options={jobOptions}
                  selected={selectedJobIds}
                  onChange={setSelectedJobIds}
                  placeholder={loadingJobs ? '加载岗位中...' : '请选择岗位'}
                  multiple={true}
                  searchPlaceholder="搜索岗位名称..."
                  disabled={loadingJobs || savingJobs}
                  selectCountLabel="岗位"
                />
              </div>

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

            {/* 底部按钮 */}
            <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
              <Button
                variant="outline"
                onClick={handleCancelJobEdit}
                disabled={savingJobs}
                className="px-4"
              >
                取消
              </Button>
              <Button
                onClick={handleSaveJobs}
                disabled={savingJobs || selectedJobIds.length === 0}
                className="px-4 text-white"
                style={{
                  background: savingJobs
                    ? '#D1D5DB'
                    : 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }}
              >
                {savingJobs ? '保存中...' : '保存'}
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      )}
    </Dialog>
  );
}

// ==================== 内容区域组件 ====================

// 个人简介
function IntroSection({
  resumeDetail,
  isEditMode,
  onDataChange,
}: {
  resumeDetail: ResumeDetail;
  isEditMode?: boolean;
  onDataChange?: (
    field: string,
    value: string | number | string[] | unknown[]
  ) => void;
}) {
  // 个人简介字段检查 - 假设有 personal_summary 或 summary 字段
  const hasDescription =
    (resumeDetail as { personal_summary?: string }).personal_summary ||
    (resumeDetail as { summary?: string }).summary;

  // 如果没有个人简介内容且不是编辑模式,则不显示整个section
  if (!hasDescription && !isEditMode) {
    return null;
  }

  const handleChange = (value: string) => {
    if (onDataChange) {
      // 尝试更新 personal_summary 或 summary 字段
      if (
        (resumeDetail as { personal_summary?: string }).personal_summary !==
        undefined
      ) {
        onDataChange('personal_summary', value);
      } else {
        onDataChange('summary', value);
      }
    }
  };

  return (
    <div className="space-y-3">
      {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
      <div className="flex items-center gap-3">
        <div
          className="w-1 h-7 rounded"
          style={{
            background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
          }}
        />
        <h3 className="text-lg font-bold text-[#1E293B]">个人简介</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-blue-300 via-purple-200 to-transparent"></div>
      </div>

      {/* 内容区 */}
      <div className="relative pl-8">
        {/* 左侧装饰 */}
        <div className="absolute left-0 top-0">
          {/* 圆形图标 - 渐变色背景 */}
          <div
            className="w-6 h-6 rounded-full flex items-center justify-center"
            style={{
              background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
            }}
          >
            <User className="w-3.5 h-3.5 text-white" />
          </div>
        </div>

        {isEditMode ? (
          <Textarea
            value={(hasDescription as string) || ''}
            onChange={(e) => handleChange(e.target.value)}
            className="text-sm font-medium text-[#64748B] min-h-[100px]"
            placeholder="请输入个人简介..."
          />
        ) : (
          <p className="text-sm font-medium text-[#64748B]">{hasDescription}</p>
        )}
      </div>
    </div>
  );
}

// 求职意向
function JobIntentionSection({
  resumeDetail,
  isEditMode,
  onDataChange,
}: {
  resumeDetail: ResumeDetail;
  isEditMode?: boolean;
  onDataChange?: (
    field: string,
    value: string | number | string[] | unknown[]
  ) => void;
}) {
  const expectedSalary = (
    resumeDetail as unknown as { expected_salary?: string }
  ).expected_salary;
  const expectedCity = resumeDetail.current_city;
  const employmentStatus = resumeDetail.employment_status;

  // 格式化职业状态显示
  const getEmploymentStatusText = (status?: string) => {
    const statusMap: Record<string, string> = {
      employed: '在职',
      unemployed: '离职',
      job_seeking: '求职中',
    };
    return status ? statusMap[status] || status : '--';
  };

  return (
    <div className="space-y-3">
      {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
      <div className="flex items-center gap-3">
        <div
          className="w-1 h-7 rounded"
          style={{
            background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
          }}
        />
        <h3 className="text-lg font-bold text-[#1E293B]">求职意向</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-green-300 via-teal-200 to-transparent"></div>
      </div>

      {/* 三项内容一排展示 - 左右对齐 */}
      <div className="flex items-center justify-between">
        {/* 期望薪资 */}
        <div className="flex items-center gap-2">
          <DollarSign className="w-4 h-4 text-[#9CA3AF]" />
          <span className="text-sm font-medium text-[#64748B]">期望薪资</span>
          {isEditMode ? (
            <Input
              value={expectedSalary || ''}
              onChange={(e) =>
                onDataChange?.('expected_salary', e.target.value)
              }
              className="text-sm font-medium text-[#64748B] h-7 w-32"
              placeholder="--"
            />
          ) : (
            <span className="text-sm font-medium text-[#64748B]">
              {expectedSalary || '--'}
            </span>
          )}
        </div>

        {/* 期望城市 */}
        <div className="flex items-center gap-2">
          <MapPin className="w-4 h-4 text-[#9CA3AF]" />
          <span className="text-sm font-medium text-[#64748B]">期望城市</span>
          {isEditMode ? (
            <Input
              value={expectedCity || ''}
              onChange={(e) => onDataChange?.('current_city', e.target.value)}
              className="text-sm font-medium text-[#64748B] h-7 w-32"
              placeholder="--"
            />
          ) : (
            <span className="text-sm font-medium text-[#64748B]">
              {expectedCity || '--'}
            </span>
          )}
        </div>

        {/* 职业状态 */}
        <div className="flex items-center gap-2">
          <Clock className="w-4 h-4 text-[#9CA3AF]" />
          <span className="text-sm font-medium text-[#64748B]">职业状态</span>
          <span className="text-sm font-medium text-[#64748B]">
            {getEmploymentStatusText(employmentStatus)}
          </span>
        </div>
      </div>
    </div>
  );
}

// 工作经验
function ExperienceSection({
  resumeDetail,
  isEditMode,
  onDataChange,
}: {
  resumeDetail: ResumeDetail;
  isEditMode?: boolean;
  onDataChange?: (
    field: string,
    value: string | number | string[] | unknown[]
  ) => void;
}) {
  const experiences = resumeDetail.experiences || [];

  if (experiences.length === 0 && !isEditMode) {
    return null;
  }

  const handleExperienceChange = (
    index: number,
    field: string,
    value: string
  ) => {
    if (!onDataChange) return;
    const newExperiences = [...experiences];
    newExperiences[index] = {
      ...newExperiences[index],
      [field]: value,
    };
    onDataChange('experiences', newExperiences);
  };

  return (
    <div className="space-y-3">
      {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
      <div className="flex items-center gap-3">
        <div
          className="w-1 h-7 rounded"
          style={{
            background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
          }}
        />
        <h3 className="text-lg font-bold text-[#1E293B]">工作经验</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-blue-300 via-purple-200 to-transparent"></div>
      </div>

      {/* 工作经历列表 */}
      <div className="space-y-6">
        {experiences.map((exp, index) => (
          <div key={exp.id} className="relative pl-8">
            {/* 左侧装饰 - 更优雅的设计 */}
            <div className="absolute left-0 top-0">
              {/* 圆形图标 - 渐变色背景 */}
              <div
                className="w-6 h-6 rounded-full flex items-center justify-center"
                style={{
                  background:
                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }}
              >
                <Briefcase className="w-3.5 h-3.5 text-white" />
              </div>
              {/* 连接线 - 如果不是最后一个 */}
              {index < experiences.length - 1 && (
                <div className="absolute left-[11px] top-6 w-0.5 h-[calc(100%+1.5rem)] bg-[#E5E7EB]" />
              )}
            </div>

            <div className="space-y-2">
              {/* 公司名称 + 工作职责 + 时间 */}
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-2">
                  {isEditMode ? (
                    <>
                      <Input
                        value={exp.company}
                        onChange={(e) =>
                          handleExperienceChange(
                            index,
                            'company',
                            e.target.value
                          )
                        }
                        className="text-sm font-bold text-[#1E293B] h-7 w-40"
                        placeholder="公司名称"
                      />
                      <Input
                        value={exp.position}
                        onChange={(e) =>
                          handleExperienceChange(
                            index,
                            'position',
                            e.target.value
                          )
                        }
                        className="text-sm font-bold text-[#1E293B] h-7 w-32"
                        placeholder="职位"
                      />
                    </>
                  ) : (
                    <>
                      <span className="text-sm font-bold text-[#1E293B]">
                        {exp.company}
                      </span>
                      <span className="text-sm font-bold text-[#1E293B]">
                        {exp.position}
                      </span>
                    </>
                  )}
                </div>
                {/* 时间右对齐 - 过滤undefined */}
                {exp.start_date &&
                  !exp.start_date.includes('undefined') &&
                  (!exp.end_date || !exp.end_date.includes('undefined')) && (
                    <div className="text-sm text-[#64748B] whitespace-nowrap">
                      {exp.start_date.split('-').slice(0, 2).join('年') + '月'}{' '}
                      -{' '}
                      {exp.end_date
                        ? exp.end_date.split('-').slice(0, 2).join('年') + '月'
                        : '至今'}
                    </div>
                  )}
              </div>

              {/* 工作经验内容 */}
              {(exp.description || isEditMode) &&
                (isEditMode ? (
                  <Textarea
                    value={exp.description || ''}
                    onChange={(e) =>
                      handleExperienceChange(
                        index,
                        'description',
                        e.target.value
                      )
                    }
                    className="text-sm font-medium leading-relaxed text-[#64748B] min-h-[80px]"
                    placeholder="工作描述..."
                  />
                ) : (
                  <p className="text-sm font-medium leading-relaxed text-[#64748B]">
                    {exp.description}
                  </p>
                ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// 教育背景
function EducationSection({
  resumeDetail,
  // isEditMode, // TODO: 未来可扩展编辑功能
  // onDataChange // TODO: 未来可扩展编辑功能
}: {
  resumeDetail: ResumeDetail;
  isEditMode?: boolean;
  onDataChange?: (
    field: string,
    value: string | number | string[] | unknown[]
  ) => void;
}) {
  const educations = resumeDetail.educations || [];

  // 学历类型映射
  const degreeMap: Record<string, string> = {
    bachelor: '本科',
    master: '硕士',
    doctor: '博士',
    junior_college: '大专',
  };

  // 大学类型映射
  const universityTypeMap: Record<string, string> = {
    '211': '211',
    '985': '985',
    double_first_class: '双一流',
    qs_top100: 'QS排名前100',
  };

  // 大学类型排序优先级（QS排名前100 > 双一流 > 985 > 211）
  const universityTypePriority: Record<string, number> = {
    qs_top100: 4,
    double_first_class: 3,
    '985': 2,
    '211': 1,
    ordinary: 0,
  };

  // 获取排序后的大学类型标签
  const getSortedUniversityTypes = (
    types?: Array<
      'ordinary' | '211' | '985' | 'double_first_class' | 'qs_top100'
    >
  ): string[] => {
    if (!types || types.length === 0) return [];
    // 过滤掉ordinary类型，按优先级排序，然后映射为显示文本
    return types
      .filter((type) => type !== 'ordinary')
      .sort((a, b) => universityTypePriority[b] - universityTypePriority[a])
      .map((type) => universityTypeMap[type] || type);
  };

  if (educations.length === 0) {
    return null;
  }

  return (
    <div className="space-y-3">
      {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
      <div className="flex items-center gap-3">
        <div
          className="w-1 h-7 rounded"
          style={{
            background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
          }}
        />
        <h3 className="text-lg font-bold text-[#1E293B]">教育背景</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-green-300 via-teal-200 to-transparent"></div>
      </div>

      {/* 教育经历列表 */}
      <div className="space-y-6">
        {educations.map((edu, index) => {
          const eduExtended = edu as unknown as {
            gpa?: string;
            papers?: string | string[];
            clubs?: string | string[];
          };

          return (
            <div key={edu.id} className="relative pl-8">
              {/* 左侧装饰 - 更优雅的设计 */}
              <div className="absolute left-0 top-0">
                {/* 圆形图标 - 渐变色背景 */}
                <div
                  className="w-6 h-6 rounded-full flex items-center justify-center"
                  style={{
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }}
                >
                  <GraduationCap className="w-3.5 h-3.5 text-white" />
                </div>
                {/* 连接线 - 如果不是最后一个 */}
                {index < educations.length - 1 && (
                  <div className="absolute left-[11px] top-6 w-0.5 h-[calc(100%+1.5rem)] bg-[#E5E7EB]" />
                )}
              </div>

              <div className="space-y-3">
                {/* 学校名称 + 大学类型标签 + 就读时间 */}
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2 flex-wrap flex-1">
                    <p className="text-sm font-bold text-[#1E293B]">
                      {edu.school}
                    </p>
                    {/* 大学类型标签 */}
                    {getSortedUniversityTypes(edu.university_types).map(
                      (typeLabel, idx) => (
                        <span
                          key={idx}
                          className="inline-block px-2.5 py-0.5 text-xs rounded text-white font-medium"
                          style={{
                            background:
                              'linear-gradient(to right, #1181C7, #647094)',
                          }}
                        >
                          {typeLabel}
                        </span>
                      )
                    )}
                  </div>
                  {/* 就读时间 - 过滤undefined */}
                  {edu.start_date &&
                    edu.end_date &&
                    !edu.start_date.includes('undefined') &&
                    !edu.end_date.includes('undefined') && (
                      <div className="text-sm text-[#64748B] whitespace-nowrap">
                        {edu.start_date.split('-').slice(0, 2).join('年') +
                          '月'}{' '}
                        -{' '}
                        {edu.end_date.split('-').slice(0, 2).join('年') + '月'}
                      </div>
                    )}
                </div>

                {/* 专业 + 学历标签 + GPA */}
                <div className="flex items-center gap-2">
                  <h4 className="text-sm font-medium text-[#64748B]">
                    {edu.major}
                  </h4>
                  {/* 学历标签 */}
                  <span
                    className="inline-block px-2.5 py-0.5 text-xs rounded text-white font-medium"
                    style={{
                      background: 'linear-gradient(to right, #51B3F0, #59799D)',
                    }}
                  >
                    {degreeMap[edu.degree] || edu.degree}
                  </span>
                  {/* GPA展示 - 在学历标签后面一行展示 */}
                  {eduExtended.gpa && (
                    <>
                      <GraduationCap className="w-4 h-4 text-[#9CA3AF]" />
                      <span className="text-sm font-bold text-[#1E293B]">
                        GPA {eduExtended.gpa}
                      </span>
                    </>
                  )}
                </div>
              </div>

              {/* 论文发表 */}
              {eduExtended.papers && (
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <FileText className="w-4 h-4 text-[#3B82F6]" />
                    <span className="text-sm font-medium text-[#64748B]">
                      论文发表
                    </span>
                  </div>
                  <div className="pl-6 space-y-1">
                    {Array.isArray(eduExtended.papers) ? (
                      eduExtended.papers.map((paper, idx) => (
                        <p key={idx} className="text-base text-[#1E293B]">
                          • {paper}
                        </p>
                      ))
                    ) : (
                      <p className="text-base text-[#1E293B]">
                        {eduExtended.papers}
                      </p>
                    )}
                  </div>
                </div>
              )}

              {/* 社团和组织经历 */}
              {eduExtended.clubs && (
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <Users className="w-4 h-4 text-[#3B82F6]" />
                    <span className="text-sm font-medium text-[#64748B]">
                      社团和组织经历
                    </span>
                  </div>
                  <div className="pl-6 space-y-1">
                    {Array.isArray(eduExtended.clubs) ? (
                      eduExtended.clubs.map((club, idx) => (
                        <p key={idx} className="text-base text-[#1E293B]">
                          • {club}
                        </p>
                      ))
                    ) : (
                      <p className="text-base text-[#1E293B]">
                        {eduExtended.clubs}
                      </p>
                    )}
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

// 技能专长
function SkillsSection({
  resumeDetail,
  // isEditMode, // TODO: 未来可扩展编辑功能
  // onDataChange // TODO: 未来可扩展编辑功能
}: {
  resumeDetail: ResumeDetail;
  isEditMode?: boolean;
  onDataChange?: (
    field: string,
    value: string | number | string[] | unknown[]
  ) => void;
}) {
  const skills = resumeDetail.skills || [];
  const [isExpanded, setIsExpanded] = useState(false);

  if (skills.length === 0) {
    return null;
  }

  // 获取进度条宽度 (缩短到30%)
  const getProgressWidth = (level?: string) => {
    switch (level) {
      case '精通':
        return '27%'; // 90% * 30% = 27% (最高)
      case '高级':
        return '24%'; // 80% * 30% = 24%
      case '熟练':
        return '22.5%'; // 75% * 30% = 22.5%
      case '中级':
        return '18%'; // 60% * 30% = 18%
      case '了解':
      case '初级':
        return '15%'; // 50% * 30% = 15% (最低)
      default:
        return '25.5%'; // 85% * 30% = 25.5%
    }
  };

  // 按类型分类技能(如果有category字段)
  const technicalSkills = skills.filter(
    (s) =>
      (s as { category?: string }).category === 'technical' ||
      !(s as { category?: string }).category
  );
  const otherSkills = skills.filter(
    (s) => (s as { category?: string }).category === 'other'
  );

  // 计算技能显示行数（每行3个）
  const skillsPerRow = 3;
  const maxRowsBeforeCollapse = 3;
  const maxSkillsBeforeCollapse = skillsPerRow * maxRowsBeforeCollapse;
  const shouldShowExpandButton =
    technicalSkills.length > maxSkillsBeforeCollapse;
  const displayedSkills =
    shouldShowExpandButton && !isExpanded
      ? technicalSkills.slice(0, maxSkillsBeforeCollapse)
      : technicalSkills;

  return (
    <div className="space-y-6">
      {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
      <div className="flex items-center gap-3">
        <div
          className="w-1 h-7 rounded"
          style={{
            background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
          }}
        />
        <h3 className="text-lg font-bold text-[#1E293B]">技能专长</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-purple-300 via-pink-200 to-transparent"></div>
      </div>

      {/* 技术技能 - 带进度条 */}
      {technicalSkills.length > 0 && (
        <div className="space-y-3">
          {/* 每行3个技能 */}
          <div className="grid grid-cols-3 gap-x-6 gap-y-4">
            {displayedSkills.map((skill) => (
              <div key={skill.id} className="space-y-2">
                {/* 技能名称 + 等级 */}
                <div className="flex items-center justify-between gap-2">
                  <span
                    className="text-sm font-medium text-[#64748B] truncate flex-1"
                    title={skill.skill_name}
                  >
                    {skill.skill_name && skill.skill_name.length > 10
                      ? skill.skill_name.substring(0, 10) + '...'
                      : skill.skill_name}
                  </span>
                  <span className="text-sm font-medium text-[#64748B] flex-shrink-0">
                    {skill.level || '精通'}
                  </span>
                </div>

                {/* 进度条 */}
                <div className="h-2.5 bg-[#E5E7EB] rounded-full overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all"
                    style={{
                      width: getProgressWidth(skill.level),
                      background:
                        'linear-gradient(90deg, #7bb8ff 0%, #3F3663 100%)',
                    }}
                  />
                </div>

                {/* 技术标签 */}
                {(skill as unknown as { tags?: string[] }).tags && (
                  <div className="flex flex-wrap gap-1">
                    {(skill as unknown as { tags: string[] }).tags.map(
                      (tag, idx) => (
                        <span
                          key={idx}
                          className="inline-block px-2 py-0.5 text-xs bg-[#F3F4F6] text-[#64748B] rounded"
                        >
                          {tag}
                        </span>
                      )
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* 展开/收起按钮 */}
          {shouldShowExpandButton && (
            <div className="flex justify-center pt-2">
              <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-[#7bb8ff] hover:text-[#6aa8ee] hover:bg-gray-50 rounded-lg transition-colors"
              >
                <span>
                  {isExpanded ? '收起' : `展开全部 (${technicalSkills.length})`}
                </span>
                <svg
                  className={`w-4 h-4 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 9l-7 7-7-7"
                  />
                </svg>
              </button>
            </div>
          )}
        </div>
      )}

      {/* 其他技能 - 带圆点列表 */}
      {otherSkills.length > 0 && (
        <div className="space-y-3">
          <h4 className="text-lg font-semibold text-[#1E293B]">其他技能</h4>
          <div className="grid grid-cols-2 gap-x-6 gap-y-2">
            {otherSkills.map((skill) => (
              <div key={skill.id} className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-[#3B82F6] flex-shrink-0 mt-2" />
                <span className="text-base text-[#1E293B]">
                  {skill.skill_name}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 如果没有分类,显示所有技能为其他技能 */}
      {technicalSkills.length === 0 &&
        otherSkills.length === 0 &&
        skills.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-lg font-semibold text-[#1E293B]">技能列表</h4>
            <div className="grid grid-cols-2 gap-x-6 gap-y-2">
              {skills.map((skill) => (
                <div key={skill.id} className="flex items-start gap-2">
                  <div className="w-1.5 h-1.5 rounded-full bg-[#3B82F6] flex-shrink-0 mt-2" />
                  <span className="text-base text-[#1E293B]">
                    {skill.skill_name} {skill.level ? `(${skill.level})` : ''}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
    </div>
  );
}

// 项目经历
function ProjectsSection({
  resumeDetail,
  // isEditMode, // TODO: 未来可扩展编辑功能
  // onDataChange // TODO: 未来可扩展编辑功能
}: {
  resumeDetail: ResumeDetail;
  isEditMode?: boolean;
  onDataChange?: (
    field: string,
    value: string | number | string[] | unknown[]
  ) => void;
}) {
  const projects = resumeDetail.projects || [];

  // 如果没有项目经历,不显示此区域
  if (projects.length === 0) {
    return null;
  }

  return (
    <div className="space-y-3">
      {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
      <div className="flex items-center gap-3">
        <div
          className="w-1 h-7 rounded"
          style={{
            background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
          }}
        />
        <h3 className="text-lg font-bold text-[#1E293B]">项目经历</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-blue-300 via-cyan-200 to-transparent"></div>
      </div>

      {/* 项目列表 */}
      <div className="space-y-6">
        {projects.map((project, index) => {
          const projectExtended = project as unknown as {
            role?: string;
            achievements?: string | string[];
          };

          return (
            <div key={project.id} className="relative pl-8">
              {/* 左侧装饰 - 更优雅的设计 */}
              <div className="absolute left-0 top-0">
                {/* 圆形图标 - 渐变色背景 */}
                <div
                  className="w-6 h-6 rounded-full flex items-center justify-center"
                  style={{
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }}
                >
                  <FolderGit2 className="w-3.5 h-3.5 text-white" />
                </div>
                {/* 连接线 - 如果不是最后一个 */}
                {index < projects.length - 1 && (
                  <div className="absolute left-[11px] top-6 w-0.5 h-[calc(100%+1.5rem)] bg-[#E5E7EB]" />
                )}
              </div>

              <div className="space-y-3">
                {/* 项目名称 + 角色 + 时间 */}
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <h4 className="text-sm font-bold text-[#1E293B]">
                      {project.name}
                    </h4>
                    {projectExtended.role && (
                      <span className="text-sm font-bold text-[#1E293B]">
                        {projectExtended.role}
                      </span>
                    )}
                  </div>
                  {/* 时间右对齐 - 过滤undefined */}
                  {project.start_date &&
                    !project.start_date.includes('undefined') &&
                    (!project.end_date ||
                      !project.end_date.includes('undefined')) && (
                      <div className="text-sm text-[#64748B] whitespace-nowrap">
                        {project.start_date.split('-').slice(0, 2).join('年') +
                          '月'}{' '}
                        -{' '}
                        {project.end_date
                          ? project.end_date.split('-').slice(0, 2).join('年') +
                            '月'
                          : '至今'}
                      </div>
                    )}
                </div>

                {/* 公司名称 */}
                {project.company && (
                  <p className="text-sm font-medium text-[#64748B]">
                    {project.company}
                  </p>
                )}

                {/* 项目描述 */}
                {project.description && (
                  <p className="text-sm font-medium leading-relaxed text-[#64748B]">
                    {project.description}
                  </p>
                )}

                {/* 主要职责 */}
                {(project.responsibilities || projectExtended.achievements) && (
                  <div className="space-y-2">
                    <h5 className="text-sm font-medium text-[#64748B]">
                      主要职责
                    </h5>
                    <ul className="space-y-2">
                      {/* responsibilities 字段 */}
                      {project.responsibilities &&
                        (typeof project.responsibilities === 'string'
                          ? [project.responsibilities]
                          : (project.responsibilities as unknown as string[]) ||
                            []
                        ).map((resp, idx) => (
                          <li
                            key={`resp-${idx}`}
                            className="flex items-start gap-2"
                          >
                            <div className="w-1.5 h-1.5 rounded-full bg-[#3B82F6] flex-shrink-0 mt-2" />
                            <span className="text-sm font-medium text-[#64748B]">
                              {resp}
                            </span>
                          </li>
                        ))}
                      {/* achievements 字段 */}
                      {projectExtended.achievements &&
                        (typeof projectExtended.achievements === 'string'
                          ? [projectExtended.achievements]
                          : projectExtended.achievements || []
                        ).map((ach, idx) => (
                          <li
                            key={`ach-${idx}`}
                            className="flex items-start gap-2"
                          >
                            <div className="w-1.5 h-1.5 rounded-full bg-[#3B82F6] flex-shrink-0 mt-2" />
                            <span className="text-sm font-medium text-[#64748B]">
                              {ach}
                            </span>
                          </li>
                        ))}
                    </ul>
                  </div>
                )}

                {/* 关键词标签 (technologies) */}
                {(project.technologies ||
                  (project as { tech_stack?: string }).tech_stack) &&
                  (() => {
                    const techData =
                      project.technologies ||
                      (project as { tech_stack?: string }).tech_stack ||
                      '';

                    // 处理不同的数据格式
                    let techArray: string[] = [];
                    if (Array.isArray(techData)) {
                      // 如果是数组，直接使用
                      techArray = techData.filter((t) => t && t.trim());
                    } else if (typeof techData === 'string') {
                      // 如果是字符串，使用中文逗号"，"、英文逗号","、顿号"、"作为分隔符
                      techArray = techData
                        .split(/[，,、]/)
                        .map((t) => t.trim())
                        .filter((t) => t);
                    }

                    if (techArray.length === 0) return null;

                    return (
                      <div className="flex flex-wrap gap-1.5">
                        {techArray.map((tech, idx) => (
                          <span
                            key={idx}
                            className="inline-block px-3 py-1 text-xs text-white rounded-full"
                            style={{
                              background:
                                'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                            }}
                          >
                            {tech}
                          </span>
                        ))}
                      </div>
                    );
                  })()}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// 荣誉与资格证书
function CertificatesSection({
  resumeDetail,
  // isEditMode, // TODO: 未来可扩展编辑功能
  // onDataChange // TODO: 未来可扩展编辑功能
}: {
  resumeDetail: ResumeDetail;
  isEditMode?: boolean;
  onDataChange?: (
    field: string,
    value: string | number | string[] | unknown[]
  ) => void;
}) {
  const certificatesData = resumeDetail as unknown as {
    honors_certificates?: string; // 使用honors_certificates字段
  };

  // 获取荣誉与资格证书内容
  const honorsContent = certificatesData.honors_certificates;

  // 如果没有内容，不显示此区域
  if (!honorsContent || !honorsContent.trim()) {
    return null;
  }

  // 将内容按中文逗号"，"、英文逗号","、中文分号"；"、顿号"、"分割成数组，并过滤掉软件著作相关内容
  const honorsList = honorsContent
    .split(/[，,；、]/)
    .filter((item) => item.trim())
    .map((item) => item.trim())
    .filter((item) => !item.includes('软件著作')); // 过滤掉包含"软件著作"的内容

  // 如果过滤后没有内容，不显示此区域
  if (honorsList.length === 0) {
    return null;
  }

  return (
    <div className="space-y-3">
      {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
      <div className="flex items-center gap-3">
        <div
          className="w-1 h-7 rounded"
          style={{
            background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
          }}
        />
        <h3 className="text-lg font-bold text-[#1E293B]">荣誉与资格证书</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-yellow-300 via-amber-200 to-transparent"></div>
      </div>

      {/* 内容列表 - 每行展示两个 */}
      <div className="grid grid-cols-2 gap-x-6 gap-y-3">
        {honorsList.map((honor, index) => (
          <div key={index} className="relative pl-8">
            {/* 左侧装饰 */}
            <div className="absolute left-0 top-0">
              {/* 圆形图标 - 渐变色背景 */}
              <div
                className="w-6 h-6 rounded-full flex items-center justify-center"
                style={{
                  background:
                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }}
              >
                <Award className="w-3.5 h-3.5 text-white" />
              </div>
            </div>
            <p className="text-sm font-medium text-[#64748B]">{honor}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

// 其他
function OtherSection({
  resumeDetail,
}: {
  resumeDetail: ResumeDetail;
  isEditMode?: boolean;
  onDataChange?: (
    field: string,
    value: string | number | string[] | unknown[]
  ) => void;
}) {
  const otherData = resumeDetail as unknown as {
    other_info?: string; // 使用other_info字段
    honors_certificates?: string; // 用于提取软件著作内容
  };

  // 获取其他信息内容
  const otherContent = otherData.other_info;
  const honorsContent = otherData.honors_certificates;

  // 从other_info按中文分号"；"分割
  let otherList: string[] = [];
  if (otherContent && otherContent.trim()) {
    otherList = otherContent
      .split('；')
      .filter((item) => item.trim())
      .map((item) => item.trim())
      .filter((item) => !item.startsWith('求职意向')); // 过滤掉以"求职意向"开头的内容
  }

  // 从honors_certificates中提取软件著作相关内容
  if (honorsContent && honorsContent.trim()) {
    const softwareWorks = honorsContent
      .split(/[，；、]/)
      .filter((item) => item.trim())
      .map((item) => item.trim())
      .filter((item) => item.includes('软件著作')); // 只保留包含"软件著作"的内容

    // 将软件著作内容添加到otherList
    otherList = [...otherList, ...softwareWorks];
  }

  // 如果过滤后没有内容，不显示此区域
  if (otherList.length === 0) {
    return null;
  }

  return (
    <div className="space-y-3">
      {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
      <div className="flex items-center gap-3">
        <div
          className="w-1 h-7 rounded"
          style={{
            background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
          }}
        />
        <h3 className="text-lg font-bold text-[#1E293B]">其他</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-gray-300 via-slate-200 to-transparent"></div>
      </div>

      {/* 内容列表 */}
      <div className="space-y-3">
        {otherList.map((item, index) => (
          <div key={index} className="relative pl-8">
            {/* 左侧装饰 */}
            <div className="absolute left-0 top-0">
              {/* 圆形图标 - 渐变色背景 */}
              <div
                className="w-6 h-6 rounded-full flex items-center justify-center"
                style={{
                  background:
                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }}
              >
                <Info className="w-3.5 h-3.5 text-white" />
              </div>
            </div>
            <p className="text-sm font-medium text-[#64748B]">{item}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
