import React, { useState, useEffect, useCallback } from 'react';
import { X, ChevronLeft, ChevronRight, Download, Edit, Save, XCircle, Mail, Phone, MapPin, Github, Plus, Trash2, RefreshCw } from 'lucide-react';
import { ResumeDetail, ResumeUpdateParams, Resume, ResumeStatus } from '@/types/resume';
import { getResumeDetail, updateResume, reparseResume, downloadResumeFile } from '@/services/resume';
import { Input } from '@/components/ui/input';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';
import { formatDate } from '@/lib/utils';

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
  onEdit,
  onNavigate,
  canNavigatePrev = false,
  canNavigateNext = false,
}: ResumePreviewModalProps) {
  const [resumeDetail, setResumeDetail] = useState<ResumeDetail | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [editFormData, setEditFormData] = useState<ResumeUpdateParams | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [isReparsing, setIsReparsing] = useState(false);
  const [reparseMessage, setReparseMessage] = useState<string | null>(null);
  const [showReparseConfirm, setShowReparseConfirm] = useState(false);
  const [showReparseResult, setShowReparseResult] = useState(false);
  const [reparseResult, setReparseResult] = useState<{ success: boolean; message: string } | null>(null);
  const [isDownloading, setIsDownloading] = useState(false);

  const fetchResumeDetail = useCallback(async () => {
    if (!resumeId) return;

    setIsLoading(true);
    setError(null);
    try {
      const detail = await getResumeDetail(resumeId);
      setResumeDetail(detail);
    } catch (err) {
      console.error('获取简历详情失败:', err);
      setError('获取简历详情失败，请稍后重试');
    } finally {
      setIsLoading(false);
    }
  }, [resumeId]);

  useEffect(() => {
    if (isOpen && resumeId) {
      fetchResumeDetail();
    }
  }, [isOpen, resumeId, fetchResumeDetail]);

  const handleClose = () => {
    setResumeDetail(null);
    setError(null);
    setIsEditing(false);
    setEditFormData(null);
    onClose();
  };

  // 处理简历切换
  const handleNavigate = (direction: 'prev' | 'next') => {
    if (onNavigate && !isEditing) {
      onNavigate(direction);
      // 当 resumeId prop 更新时，useEffect 会自动触发重新获取数据
    }
  };

  // 处理下载简历 - 统一使用 downloadResumeFile 函数，确保一致的错误处理和日志记录
  const handleDownload = async () => {
    if (!resumeDetail) {
      console.error('❌ 简历详情不存在，无法下载');
      return;
    }

    setIsDownloading(true);
    try {
      console.log('🔽 预览模态框开始下载简历:', { 
        resumeId: resumeDetail.id, 
        name: resumeDetail.name,
        hasFileUrl: !!resumeDetail.resume_file_url 
      });
      
      // 构造 Resume 对象调用统一的 downloadResumeFile 函数
      // 这样可以确保使用相同的错误处理、URL处理和日志记录逻辑
      const resume: Resume = {
        id: resumeDetail.id,
        name: resumeDetail.name || '',
        phone: resumeDetail.phone || '',
        email: resumeDetail.email || '',
        current_city: resumeDetail.current_city || '',
        status: ResumeStatus.COMPLETED,
        created_at: typeof resumeDetail.created_at === 'number' ? resumeDetail.created_at : Date.now(),
        updated_at: typeof resumeDetail.updated_at === 'number' ? resumeDetail.updated_at : Date.now(),
        uploader_name: resumeDetail.uploader_name || '',
        resume_file_url: resumeDetail.resume_file_url || '',
        uploader_id: resumeDetail.uploader_id || '',
      };
      
      await downloadResumeFile(resume);
      console.log('✅ 预览模态框下载完成');
    } catch (error) {
      console.error('❌ 预览模态框下载简历失败:', error);
      
      // 显示用户友好的错误信息
      let errorMessage = '下载失败，请稍后重试';
      if (error instanceof Error) {
        errorMessage = error.message;
      }
      
      // 这里可以添加 toast 通知或其他用户提示
      // 暂时使用 console.error，后续可以集成通知组件
      console.error('用户错误提示:', errorMessage);
    } finally {
      setIsDownloading(false);
    }
  };

  // 开始编辑
  const handleStartEdit = () => {
    if (resumeDetail) {
      setIsEditing(true);
      setEditFormData({
        name: resumeDetail.name || '',
        email: resumeDetail.email || '',
        phone: resumeDetail.phone || '',
        current_city: resumeDetail.current_city || '',
        experiences: resumeDetail.experiences?.map(exp => ({
          id: exp.id,
          action: 'update' as const,
          company: exp.company,
          position: exp.position,
          title: exp.title,
          start_date: exp.start_date,
          end_date: exp.end_date,
          description: exp.description,
        })) || [],
        educations: resumeDetail.educations?.map(edu => ({
          id: edu.id,
          action: 'update' as const,
          school: edu.school,
          major: edu.major,
          degree: edu.degree,
          start_date: edu.start_date,
          end_date: edu.end_date,
        })) || [],
        skills: resumeDetail.skills?.map(skill => ({
          id: skill.id,
          action: 'update' as const,
          skill_name: skill.skill_name,
          level: skill.level,
          description: skill.description,
        })) || [],
      });
    }
  };

  // 取消编辑
  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditFormData(null);
  };

  // 处理重新解析简历 - 显示确认对话框
  const handleReparse = () => {
    if (isReparsing) return;
    setShowReparseConfirm(true);
  };

  // 确认重新解析
  const handleConfirmReparse = async () => {
    if (!resumeId || isReparsing) return;

    setShowReparseConfirm(false);
    setIsReparsing(true);
    setReparseMessage(null);
    
    try {
      await reparseResume(resumeId);
      setReparseResult({
        success: true,
        message: '简历重新解析成功！系统已更新简历信息。'
      });
      
      // 延迟重新获取简历详情，给后端处理时间
      setTimeout(() => {
        fetchResumeDetail();
      }, 1000);
      
    } catch (err) {
      console.error('重新解析简历失败:', err);
      setReparseResult({
        success: false,
        message: '重新解析失败，请检查网络连接后重试。'
      });
    } finally {
      setIsReparsing(false);
      setShowReparseResult(true);
    }
  };

  // 保存编辑
  const handleSaveEdit = async () => {
    if (!editFormData || !resumeId) return;

    setIsSaving(true);
    try {
      await updateResume(resumeId, editFormData);
      await fetchResumeDetail(); // 重新获取数据
      setIsEditing(false);
      setEditFormData(null);
    } catch (err) {
      console.error('保存简历失败:', err);
      setError('保存简历失败，请稍后重试');
    } finally {
      setIsSaving(false);
    }
  };

  // 更新表单数据
  const updateFormData = (field: keyof ResumeUpdateParams, value: any) => {
    if (editFormData) {
      setEditFormData({
        ...editFormData,
        [field]: value,
      });
    }
  };

  // 更新工作经验
  const updateExperience = (index: number, field: string, value: string) => {
    if (editFormData && editFormData.experiences) {
      const newExperiences = [...editFormData.experiences];
      newExperiences[index] = {
        ...newExperiences[index],
        [field]: value,
      };
      setEditFormData({
        ...editFormData,
        experiences: newExperiences,
      });
    }
  };

  // 更新教育经历
  const updateEducation = (index: number, field: string, value: string) => {
    if (editFormData && editFormData.educations) {
      const newEducations = [...editFormData.educations];
      newEducations[index] = {
        ...newEducations[index],
        [field]: value,
      };
      setEditFormData({
        ...editFormData,
        educations: newEducations,
      });
    }
  };

  // 更新技能
  const updateSkill = (index: number, value: string) => {
    if (editFormData && editFormData.skills) {
      const newSkills = [...editFormData.skills];
      newSkills[index] = {
        ...newSkills[index],
        skill_name: value,
      };
      setEditFormData({
        ...editFormData,
        skills: newSkills,
      });
    }
  };

  // 添加技能
  const addSkill = () => {
    if (editFormData) {
      const newSkill = {
        id: `temp_${Date.now()}`, // 临时ID
        action: 'create' as const,
        skill_name: '',
        level: '',
        description: '',
      };
      setEditFormData({
        ...editFormData,
        skills: [...(editFormData.skills || []), newSkill],
      });
    }
  };

  // 删除技能
  const removeSkill = (index: number) => {
    if (editFormData && editFormData.skills) {
      const newSkills = [...editFormData.skills];
      const skillToRemove = newSkills[index];
      
      // 如果是已存在的技能，标记为删除；如果是新添加的，直接移除
      if (skillToRemove.id && !skillToRemove.id.startsWith('temp_')) {
        newSkills[index] = {
          ...skillToRemove,
          action: 'delete' as const,
        };
      } else {
        newSkills.splice(index, 1);
      }
      
      setEditFormData({
        ...editFormData,
        skills: newSkills,
      });
    }
  };

  if (!isOpen) return null;

  return (
    <TooltipProvider>
      <div className="fixed inset-0 z-50 flex items-center justify-center">
        {/* 半透明黑色背景 */}
        <div 
          className="absolute inset-0 bg-black bg-opacity-50"
          onClick={handleClose}
        />
        
        {/* 弹窗容器 - 增大尺寸 */}
        <div className="relative bg-white rounded-lg shadow-xl w-[1200px] max-h-[90vh] overflow-hidden">
        {/* 顶部标题栏 */}
        <div className="bg-gray-100 px-6 py-4 flex items-center justify-between">
          <h2 className="text-lg font-medium text-gray-900">简历详情预览</h2>
          <div className="flex items-center gap-2">
            <Tooltip>
              <TooltipTrigger asChild>
                <button 
                  onClick={() => handleNavigate('prev')}
                  disabled={!canNavigatePrev || isEditing}
                  className={`p-2 rounded transition-colors ${
                    canNavigatePrev && !isEditing
                      ? 'hover:bg-gray-200 text-gray-600' 
                      : 'text-gray-300 cursor-not-allowed'
                  }`}
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                <p>上一份简历</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <button 
                  onClick={() => handleNavigate('next')}
                  disabled={!canNavigateNext || isEditing}
                  className={`p-2 rounded transition-colors ${
                    canNavigateNext && !isEditing
                      ? 'hover:bg-gray-200 text-gray-600' 
                      : 'text-gray-300 cursor-not-allowed'
                  }`}
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                <p>下一份简历</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <button 
                  onClick={handleDownload}
                  disabled={!resumeDetail || isDownloading}
                  className={`p-2 rounded transition-colors ${
                    !resumeDetail || isDownloading
                      ? 'text-gray-400 cursor-not-allowed'
                      : 'hover:bg-gray-200 text-gray-600'
                  }`}
                >
                  <Download className="w-4 h-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{isDownloading ? '下载中...' : '下载简历'}</p>
              </TooltipContent>
            </Tooltip>

            {/* 重新解析按钮 */}
            <Tooltip>
              <TooltipTrigger asChild>
                <button 
                  onClick={handleReparse}
                  disabled={isReparsing}
                  className={`p-2 rounded ${
                    isReparsing 
                      ? 'text-gray-400 cursor-not-allowed' 
                      : 'hover:bg-gray-200 text-gray-600'
                  }`}
                >
                  <RefreshCw className={`w-4 h-4 ${isReparsing ? 'animate-spin' : ''}`} />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                <p>重新解析简历</p>
              </TooltipContent>
            </Tooltip>

            {/* 编辑按钮 */}
            {!isEditing && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <button 
                    onClick={handleStartEdit}
                    className="p-2 hover:bg-gray-200 rounded"
                  >
                    <Edit className="w-4 h-4 text-gray-600" />
                  </button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>编辑简历</p>
                </TooltipContent>
              </Tooltip>
            )}
            
            <Tooltip>
              <TooltipTrigger asChild>
                <button 
                  className="p-2 hover:bg-gray-200 rounded"
                  onClick={handleClose}
                >
                  <X className="w-4 h-4 text-gray-600" />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                <p>关闭</p>
              </TooltipContent>
            </Tooltip>
          </div>
        </div>

        {/* 内容区域 */}
        <div className="bg-gradient-to-br from-blue-50 to-purple-50 p-8 overflow-y-auto max-h-[calc(90vh-140px)]">
          {/* 重新解析消息提示 */}
          {reparseMessage && (
            <div className={`mb-4 p-3 rounded-lg text-center ${
              reparseMessage.includes('成功') 
                ? 'bg-green-100 text-green-700 border border-green-200' 
                : 'bg-red-100 text-red-700 border border-red-200'
            }`}>
              {reparseMessage}
            </div>
          )}

          {isLoading && (
            <div className="flex items-center justify-center py-16">
              <div className="text-gray-500 text-lg">加载中...</div>
            </div>
          )}

          {error && (
            <div className="flex items-center justify-center py-16">
              <div className="text-red-500 text-lg">{error}</div>
            </div>
          )}

          {resumeDetail && (
            <div className="bg-white rounded-lg shadow-sm p-8 max-w-4xl mx-auto">
              {/* 个人信息头部 */}
              <div className="text-center mb-8">
                {isEditing ? (
                  <div className="space-y-4">
                    <Input
                      value={editFormData?.name || ''}
                      onChange={(e) => updateFormData('name', e.target.value)}
                      className="text-3xl font-bold text-center"
                      placeholder="姓名"
                    />
                    <p className="text-lg text-green-600 font-medium">
                      高级前端开发工程师
                    </p>
                  </div>
                ) : (
                  <>
                    <h1 className="text-3xl font-bold text-gray-900 mb-2">
                      {resumeDetail.name || '张明'}
                    </h1>
                    <p className="text-lg text-green-600 font-medium mb-4">
                      高级前端开发工程师
                    </p>
                  </>
                )}
                
                <div className="flex items-center justify-center gap-6 text-sm text-gray-600">
                  {isEditing ? (
                    <>
                      <div className="flex items-center gap-1">
                        <Mail className="w-4 h-4" />
                        <Input
                          value={editFormData?.email || ''}
                          onChange={(e) => updateFormData('email', e.target.value)}
                          placeholder="邮箱"
                          className="w-40"
                        />
                      </div>
                      <div className="flex items-center gap-1">
                        <Phone className="w-4 h-4" />
                        <Input
                          value={editFormData?.phone || ''}
                          onChange={(e) => updateFormData('phone', e.target.value)}
                          placeholder="电话"
                          className="w-40"
                        />
                      </div>
                      <div className="flex items-center gap-1">
                        <MapPin className="w-4 h-4" />
                        <Input
                          value={editFormData?.current_city || ''}
                          onChange={(e) => updateFormData('current_city', e.target.value)}
                          placeholder="城市"
                          className="w-40"
                        />
                      </div>
                    </>
                  ) : (
                    <>
                      {resumeDetail.email && (
                        <div className="flex items-center gap-1">
                          <Mail className="w-4 h-4" />
                          <span>{resumeDetail.email}</span>
                        </div>
                      )}
                      {resumeDetail.phone && (
                        <div className="flex items-center gap-1">
                          <Phone className="w-4 h-4" />
                          <span>{resumeDetail.phone}</span>
                        </div>
                      )}
                      {resumeDetail.current_city && (
                        <div className="flex items-center gap-1">
                          <MapPin className="w-4 h-4" />
                          <span>{resumeDetail.current_city}</span>
                        </div>
                      )}
                      <div className="flex items-center gap-1">
                        <Github className="w-4 h-4" />
                        <span>github.com/zhangming</span>
                      </div>
                    </>
                  )}
                </div>
              </div>

              {/* 个人介绍模块 */}
              <div className="mb-8">
                <div className="border-l-4 border-green-500 pl-4 mb-4">
                  <h2 className="text-xl font-semibold text-gray-900">个人介绍</h2>
                </div>
                {isEditing ? (
                  <textarea
                    className="w-full p-3 border border-gray-300 rounded-md resize-none"
                    rows={4}
                    placeholder="请输入个人介绍..."
                    defaultValue="拥有5年以上前端开发经验，精通React、Vue等主流框架，熟悉TypeScript、Node.js等技术栈。具备良好的代码规范意识和团队协作能力，能够独立完成复杂项目的开发和维护。"
                  />
                ) : (
                  <p className="text-gray-700 leading-relaxed">
                    拥有5年以上前端开发经验，精通React、Vue等主流框架，熟悉TypeScript、Node.js等技术栈。具备良好的代码规范意识和团队协作能力，能够独立完成复杂项目的开发和维护。
                  </p>
                )}
              </div>

              {/* 工作经验模块 */}
              {((isEditing && editFormData?.experiences) || (!isEditing && resumeDetail.experiences)) && (
                <div className="mb-8">
                  <div className="border-l-4 border-green-500 pl-4 mb-6">
                    <h2 className="text-xl font-semibold text-gray-900">工作经验</h2>
                  </div>
                  <div className="space-y-6">
                    {(isEditing ? editFormData?.experiences : resumeDetail.experiences)?.map((exp, index) => (
                      <div key={exp.id || index} className="bg-gray-50 rounded-lg p-6">
                        <div className="flex justify-between items-start mb-3">
                          <div className="flex-1">
                            {isEditing ? (
                              <div className="space-y-2">
                                <Input
                                  value={exp.company || ''}
                                  onChange={(e) => updateExperience(index, 'company', e.target.value)}
                                  placeholder="公司名称"
                                  className="text-lg font-semibold"
                                />
                                <Input
                                  value={exp.position || ''}
                                  onChange={(e) => updateExperience(index, 'position', e.target.value)}
                                  placeholder="职位"
                                  className="text-green-600 font-medium"
                                />
                              </div>
                            ) : (
                              <>
                                <h3 className="text-lg font-semibold text-gray-900">{exp.company}</h3>
                                <p className="text-green-600 font-medium">{exp.position}</p>
                              </>
                            )}
                          </div>
                          <span className="text-sm text-gray-500 bg-white px-3 py-1 rounded">
                            {formatDate(exp.start_date)} - {exp.end_date ? formatDate(exp.end_date) : '至今'}
                          </span>
                        </div>
                        {isEditing ? (
                          <textarea
                            value={exp.description || ''}
                            onChange={(e) => updateExperience(index, 'description', e.target.value)}
                            placeholder="工作描述"
                            className="w-full p-3 border border-gray-300 rounded-md resize-none"
                            rows={3}
                          />
                        ) : (
                          exp.description && (
                            <div className="text-gray-700 leading-relaxed">
                              {exp.description.split('\n').map((line, lineIndex) => (
                                <p key={lineIndex} className="mb-2">• {line}</p>
                              ))}
                            </div>
                          )
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* 教育经历模块 */}
              {((isEditing && editFormData?.educations) || (!isEditing && resumeDetail.educations)) && (
                <div className="mb-8">
                  <div className="border-l-4 border-green-500 pl-4 mb-6">
                    <h2 className="text-xl font-semibold text-gray-900">教育经历</h2>
                  </div>
                  <div className="space-y-4">
                    {(isEditing ? editFormData?.educations : resumeDetail.educations)?.map((edu, index) => (
                      <div key={edu.id || index} className="bg-gray-50 rounded-lg p-6">
                        <div className="flex justify-between items-start">
                          <div className="flex-1">
                            {isEditing ? (
                              <div className="space-y-2">
                                <Input
                                  value={edu.school || ''}
                                  onChange={(e) => updateEducation(index, 'school', e.target.value)}
                                  placeholder="学校名称"
                                  className="text-lg font-semibold"
                                />
                                <Input
                                  value={edu.major || ''}
                                  onChange={(e) => updateEducation(index, 'major', e.target.value)}
                                  placeholder="专业"
                                  className="text-green-600 font-medium"
                                />
                                <Input
                                  value={edu.degree || ''}
                                  onChange={(e) => updateEducation(index, 'degree', e.target.value)}
                                  placeholder="学位"
                                  className="text-gray-600"
                                />
                              </div>
                            ) : (
                              <>
                                <h3 className="text-lg font-semibold text-gray-900">{edu.school}</h3>
                                <p className="text-green-600 font-medium">{edu.major}</p>
                                {edu.degree && (
                                  <p className="text-gray-600">{edu.degree}</p>
                                )}
                              </>
                            )}
                          </div>
                          <span className="text-sm text-gray-500 bg-white px-3 py-1 rounded">
                            {formatDate(edu.start_date)} - {edu.end_date ? formatDate(edu.end_date) : '至今'}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* 技能特长模块 */}
              {((isEditing && editFormData?.skills) || (!isEditing && resumeDetail.skills)) && (
                <div className="mb-8">
                  <div className="border-l-4 border-green-500 pl-4 mb-6">
                    <h2 className="text-xl font-semibold text-gray-900">技能特长</h2>
                  </div>
                  <div className="flex flex-wrap gap-3">
                    {(isEditing ? editFormData?.skills : resumeDetail.skills)
                      ?.filter(skill => !('action' in skill) || skill.action !== 'delete') // 过滤掉标记为删除的技能
                      ?.map((skill, index) => (
                      isEditing ? (
                        <div key={skill.id || index} className="flex items-center gap-2 bg-gray-50 rounded-lg p-2">
                          <Input
                            value={skill.skill_name || ''}
                            onChange={(e) => updateSkill(index, e.target.value)}
                            placeholder="技能名称"
                            className="w-32 h-8"
                          />
                          <button
                            onClick={() => removeSkill(index)}
                            className="p-1 hover:bg-red-100 rounded text-red-500"
                            title="删除技能"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      ) : (
                        <span
                          key={skill.id || index}
                          className="bg-green-100 text-green-800 px-4 py-2 rounded-full text-sm font-medium"
                        >
                          {skill.skill_name}
                        </span>
                      )
                    ))}
                    {/* 添加技能按钮 */}
                    {isEditing && (
                      <button
                        onClick={addSkill}
                        className="flex items-center gap-1 px-3 py-2 border-2 border-dashed border-green-300 text-green-600 rounded-lg hover:bg-green-50 transition-colors"
                        title="添加技能"
                      >
                        <Plus className="w-4 h-4" />
                        <span className="text-sm">添加技能</span>
                      </button>
                    )}
                  </div>
                </div>
              )}

              {/* 编辑模式下的操作按钮 */}
              {isEditing && (
                <div className="mb-6 flex justify-center gap-4 p-4 bg-gray-50 rounded-lg">
                  <button 
                    onClick={handleSaveEdit}
                    disabled={isSaving}
                    className="flex items-center gap-2 px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    <Save className="w-4 h-4" />
                    {isSaving ? '保存中...' : '保存修改'}
                  </button>
                  <button 
                    onClick={handleCancelEdit}
                    className="flex items-center gap-2 px-6 py-2 border border-gray-300 text-gray-600 rounded-lg hover:bg-gray-50 transition-colors"
                  >
                    <XCircle className="w-4 h-4" />
                    取消编辑
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* 底部操作栏 */}
        <div className="bg-white border-t px-6 py-4 flex justify-end">
          <button 
            className="px-6 py-2 bg-green-600 text-white rounded hover:bg-green-700"
            onClick={handleClose}
          >
            关闭
          </button>
        </div>
      </div>
    </div>

    {/* 重新解析确认对话框 */}
    <ConfirmDialog
      open={showReparseConfirm}
      onOpenChange={setShowReparseConfirm}
      title="确认重新解析"
      description="确定要重新解析简历吗？这将使用最新的解析算法重新处理简历内容。"
      confirmText="确认解析"
      cancelText="取消"
      onConfirm={handleConfirmReparse}
      loading={isReparsing}
    />

    {/* 重新解析结果对话框 */}
    {showReparseResult && reparseResult && (
      <div className="fixed inset-0 z-50 flex items-center justify-center">
        <div className="absolute inset-0 bg-black bg-opacity-50" onClick={() => setShowReparseResult(false)} />
        <div className="relative bg-white rounded-lg shadow-xl w-96 p-6">
          <div className="text-center">
            <div className={`mx-auto mb-4 w-12 h-12 rounded-full flex items-center justify-center ${
              reparseResult.success ? 'bg-green-100' : 'bg-red-100'
            }`}>
              {reparseResult.success ? (
                <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              ) : (
                <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              )}
            </div>
            <h3 className={`text-lg font-medium mb-2 ${
              reparseResult.success ? 'text-green-900' : 'text-red-900'
            }`}>
              {reparseResult.success ? "解析成功" : "解析失败"}
            </h3>
            <p className="text-gray-600 mb-6">{reparseResult.message}</p>
            <button
              onClick={() => setShowReparseResult(false)}
              className={`px-6 py-2 rounded-lg text-white font-medium ${
                reparseResult.success 
                  ? 'bg-green-600 hover:bg-green-700' 
                  : 'bg-red-600 hover:bg-red-700'
              }`}
            >
              确定
            </button>
          </div>
        </div>
      </div>
    )}
    </TooltipProvider>
  );
}