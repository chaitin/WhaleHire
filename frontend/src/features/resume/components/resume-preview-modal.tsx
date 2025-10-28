import { useState, useEffect } from 'react';
import {
  X,
  Download,
  Edit,
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
} from 'lucide-react';
import { Dialog, DialogContent } from '@/ui/dialog';
import { Button } from '@/ui/button';
import { Input } from '@/ui/input';
import { Textarea } from '@/ui/textarea';
import { ResumeDetail } from '@/types/resume';
import { getResumeDetail, downloadResumeFile } from '@/services/resume';
import { toast } from '@/ui/toast';

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
  const [isDownloading, setIsDownloading] = useState(false);
  const [isEditMode, setIsEditMode] = useState(false);
  const [editedData, setEditedData] = useState<ResumeDetail | null>(null);

  // 获取简历详情
  useEffect(() => {
    const fetchResumeDetail = async () => {
      setIsLoading(true);
      try {
        const detail = await getResumeDetail(resumeId);
        setResumeDetail(detail);
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

  // 下载简历
  const handleDownload = async () => {
    if (!resumeDetail?.resume_file_url) {
      toast.error('简历文件不存在');
      return;
    }

    setIsDownloading(true);
    try {
      await downloadResumeFile(resumeDetail, resumeDetail.name);
      toast.success('简历下载成功');
    } catch (error) {
      console.error('下载简历失败:', error);
      toast.error('下载简历失败');
    } finally {
      setIsDownloading(false);
    }
  };

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

  // 进入编辑模式
  const handleEdit = () => {
    setIsEditMode(true);
    setEditedData(resumeDetail);
  };

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

  if (!isOpen) return null;

  const displayData = isEditMode ? editedData : resumeDetail;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-[700px] h-[90vh] overflow-hidden p-0 gap-0">
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
                    <Button
                      variant="outline"
                      className="h-6 px-2 bg-white/20 border-white/30 text-white hover:bg-white/30"
                      onClick={handleEdit}
                      disabled={isEditMode}
                    >
                      <Edit className="w-3 h-3 mr-1" />
                      <span className="text-xs font-medium">编辑</span>
                    </Button>
                    <Button
                      variant="outline"
                      className="h-6 px-2 bg-white/20 border-white/30 text-white hover:bg-white/30"
                      onClick={handleDownload}
                      disabled={isDownloading || !displayData.resume_file_url}
                    >
                      <Download className="w-3 h-3 mr-1" />
                      <span className="text-xs font-medium">下载</span>
                    </Button>
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
                  {(displayData.gender || isEditMode) && (
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

              {/* 内容区域 - 所有内容垂直滚动展示,添加科技感线条 */}
              <div className="flex-1 overflow-y-auto px-6 py-4 bg-white space-y-6 relative">
                {/* 科技感装饰线条 - 提高透明度使其更清晰 */}
                <div className="absolute inset-0 opacity-20 pointer-events-none">
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

                {/* 个人简介 */}
                <IntroSection
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
        <h3 className="text-xl font-bold text-[#1E293B]">个人简介</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-blue-300 via-purple-200 to-transparent"></div>
      </div>

      {/* 内容区 */}
      <div className="space-y-3">
        {isEditMode ? (
          <Textarea
            value={(hasDescription as string) || ''}
            onChange={(e) => handleChange(e.target.value)}
            className="text-base leading-[1.625] text-[#1E293B] min-h-[100px]"
            placeholder="请输入个人简介..."
          />
        ) : (
          <p className="text-base leading-[1.625] text-[#1E293B]">
            {hasDescription}
          </p>
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
  const availableDate = (resumeDetail as unknown as { available_date?: string })
    .available_date;

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

        {/* 到岗时间 */}
        <div className="flex items-center gap-2">
          <Clock className="w-4 h-4 text-[#9CA3AF]" />
          <span className="text-sm font-medium text-[#64748B]">到岗时间</span>
          {isEditMode ? (
            <Input
              value={availableDate || ''}
              onChange={(e) => onDataChange?.('available_date', e.target.value)}
              className="text-sm font-medium text-[#64748B] h-7 w-32"
              placeholder="--"
            />
          ) : (
            <span className="text-sm font-medium text-[#64748B]">
              {availableDate || '--'}
            </span>
          )}
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
                {/* 时间右对齐 */}
                <div className="text-sm text-[#64748B] whitespace-nowrap">
                  {exp.start_date?.split('-').slice(0, 2).join('年') + '月'} -{' '}
                  {exp.end_date
                    ? exp.end_date.split('-').slice(0, 2).join('年') + '月'
                    : '至今'}
                </div>
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
                {/* 学校名称 + 就读时间 */}
                <div className="flex items-center justify-between">
                  <p className="text-sm font-bold text-[#1E293B]">
                    {edu.school}
                  </p>
                  <div className="text-sm text-[#64748B] whitespace-nowrap">
                    {edu.start_date?.split('-').slice(0, 2).join('年') + '月'} -{' '}
                    {edu.end_date?.split('-').slice(0, 2).join('年') + '月'}
                  </div>
                </div>

                {/* 专业 + 学历标签 */}
                <div className="flex items-center gap-2">
                  <h4 className="text-sm font-medium text-[#64748B]">
                    {edu.major}
                  </h4>
                  {/* 学历标签 */}
                  <span className="inline-block px-2.5 py-0.5 text-xs bg-[#EFF6FF] text-[#3B82F6] rounded-full">
                    {degreeMap[edu.degree] || edu.degree}
                  </span>
                </div>
              </div>

              {/* GPA展示 */}
              {eduExtended.gpa && (
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <GraduationCap className="w-4 h-4 text-[#3B82F6]" />
                    <span className="text-sm font-medium text-[#64748B]">
                      GPA
                    </span>
                  </div>
                  <p className="text-base text-[#1E293B] pl-6">
                    {eduExtended.gpa}
                  </p>
                </div>
              )}

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
            {technicalSkills.map((skill) => (
              <div key={skill.id} className="space-y-2">
                {/* 技能名称 + 等级 */}
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-[#64748B]">
                    {skill.skill_name}
                  </span>
                  <span className="text-sm font-medium text-[#64748B]">
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
                  {/* 时间右对齐 */}
                  <div className="text-sm text-[#64748B] whitespace-nowrap">
                    {project.start_date?.split('-').slice(0, 2).join('年') +
                      '月'}{' '}
                    -{' '}
                    {project.end_date
                      ? project.end_date.split('-').slice(0, 2).join('年') +
                        '月'
                      : '至今'}
                  </div>
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
                      // 如果是字符串，支持多种分隔符：逗号、顿号、分号
                      // 先统一替换为逗号，再分隔
                      const normalizedData = techData.replace(/[、；;]/g, ',');
                      techArray = normalizedData
                        .split(',')
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
    certificates?:
      | Array<{
          id: string;
          name: string;
          issuer?: string;
          date?: string;
        }>
      | string;
    honors?: string;
  };

  // 处理证书数据
  let certificates: Array<{
    id: string;
    name: string;
    issuer?: string;
    date?: string;
  }> = [];

  if (Array.isArray(certificatesData.certificates)) {
    certificates = certificatesData.certificates;
  } else if (
    typeof certificatesData.certificates === 'string' &&
    certificatesData.certificates.trim()
  ) {
    // 如果是字符串，将其转换为数组
    certificates = certificatesData.certificates
      .split('\n')
      .filter((c) => c.trim())
      .map((cert, idx) => ({
        id: `cert-${idx}`,
        name: cert.trim(),
      }));
  }

  // 如果有honors字段也添加进来
  if (certificatesData.honors && typeof certificatesData.honors === 'string') {
    const honorsList = certificatesData.honors
      .split('\n')
      .filter((h) => h.trim());
    honorsList.forEach((honor, idx) => {
      certificates.push({
        id: `honor-${idx}`,
        name: honor.trim(),
      });
    });
  }

  // 如果没有证书或荣誉，不显示此区域
  if (certificates.length === 0) {
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
        <h3 className="text-xl font-bold text-[#1E293B]">荣誉与资格证书</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-yellow-300 via-amber-200 to-transparent"></div>
      </div>

      {/* 证书卡片列表 */}
      <div className="space-y-3">
        {certificates.map((cert) => (
          <div
            key={cert.id}
            className="flex items-start gap-3 p-3 border border-gray-200 rounded-lg bg-white hover:shadow-sm transition-shadow"
          >
            {/* 图标 */}
            <div className="w-9 h-10 bg-blue-50/60 rounded-lg flex items-center justify-center flex-shrink-0">
              <svg
                className="w-5 h-4 text-[#3B82F6]"
                fill="none"
                viewBox="0 0 20 16"
              >
                <path
                  fill="currentColor"
                  d="M4 0a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V2a2 2 0 0 0-2-2H4zm0 2h12v12H4V2zm2 2v2h8V4H6zm0 4v2h8V8H6zm0 4v2h5v-2H6z"
                />
              </svg>
            </div>

            {/* 证书信息 */}
            <div className="flex-1">
              <h4 className="text-base font-semibold text-[#1E293B] mb-1">
                {cert.name}
              </h4>
              {cert.issuer && (
                <p className="text-sm text-[#64748B] mb-0.5">{cert.issuer}</p>
              )}
              {cert.date && (
                <p className="text-sm text-[#64748B]">{cert.date}</p>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// 其他
function OtherSection({
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
  const otherData = resumeDetail as unknown as {
    other?: string;
    additional_info?: string;
    remarks?: string;
  };

  // 合并所有"其他"相关字段
  const otherContent = [
    otherData.other,
    otherData.additional_info,
    otherData.remarks,
  ]
    .filter((content) => content && content.trim())
    .join('\n\n');

  // 如果没有内容且不是编辑模式，不显示此区域
  if (!otherContent && !isEditMode) {
    return null;
  }

  const handleOtherChange = (value: string) => {
    if (onDataChange) {
      // 更新 other 字段作为主要字段
      onDataChange('other', value);
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
        <h3 className="text-xl font-bold text-[#1E293B]">其他</h3>
        <div className="flex-1 h-px bg-gradient-to-r from-gray-300 via-slate-200 to-transparent"></div>
      </div>

      {/* 内容区 */}
      <div className="space-y-2">
        {isEditMode ? (
          <Textarea
            value={otherContent || ''}
            onChange={(e) => handleOtherChange(e.target.value)}
            className="text-base leading-relaxed text-[#1E293B] min-h-[80px]"
            placeholder="请输入其他信息..."
          />
        ) : (
          <p className="text-base leading-relaxed text-[#1E293B] whitespace-pre-wrap">
            {otherContent}
          </p>
        )}
      </div>
    </div>
  );
}
