import { useState, useEffect, useCallback } from 'react';
import { X, User, Phone, Mail, MapPin, GraduationCap, Briefcase, Award, Calendar } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

import { ResumeDetail, ResumeStatus } from '@/types/resume';
import { getResumeDetail } from '@/services/resume';
import { formatDate } from '@/lib/utils';

interface ResumePreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  resumeId: string | null;
}

const statusLabels: Record<ResumeStatus, string> = {
  [ResumeStatus.PENDING]: '待解析',
  [ResumeStatus.PROCESSING]: '解析中',
  [ResumeStatus.COMPLETED]: '待筛选',
  [ResumeStatus.FAILED]: '解析失败',
  [ResumeStatus.ARCHIVED]: '已归档',
};

const statusColors: Record<ResumeStatus, string> = {
  [ResumeStatus.PENDING]: 'bg-yellow-100 text-yellow-800',
  [ResumeStatus.PROCESSING]: 'bg-blue-100 text-blue-800',
  [ResumeStatus.COMPLETED]: 'bg-green-100 text-green-800',
  [ResumeStatus.FAILED]: 'bg-red-100 text-red-800',
  [ResumeStatus.ARCHIVED]: 'bg-gray-100 text-gray-800',
};

export function ResumePreviewModal({ isOpen, onClose, resumeId }: ResumePreviewModalProps) {
  const [resumeDetail, setResumeDetail] = useState<ResumeDetail | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
    onClose();
  };

  if (!isOpen) return null;

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] p-0">
        <DialogHeader className="px-6 py-4 border-b">
          <div className="flex items-center justify-between">
            <DialogTitle className="text-xl font-semibold">简历预览</DialogTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={handleClose}
              className="h-8 w-8"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        </DialogHeader>

        <div className="flex-1 px-6 py-4 overflow-y-auto max-h-[calc(90vh-120px)]">
          {isLoading && (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              <span className="ml-2 text-gray-600">加载中...</span>
            </div>
          )}

          {error && (
            <div className="flex items-center justify-center py-12">
              <div className="text-center">
                <div className="text-red-600 mb-2">{error}</div>
                <Button variant="outline" onClick={fetchResumeDetail}>
                  重试
                </Button>
              </div>
            </div>
          )}

          {resumeDetail && (
            <div className="space-y-6">
              {/* 基本信息 */}
              <div className="bg-gray-50 rounded-lg p-6">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
                      <User className="h-6 w-6 text-blue-600" />
                    </div>
                    <div>
                      <h2 className="text-2xl font-bold text-gray-900">{resumeDetail.name}</h2>
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColors[resumeDetail.status]}`}>
                        {statusLabels[resumeDetail.status]}
                      </span>
                    </div>
                  </div>
                  <div className="text-sm text-gray-500">
                    上传时间: {formatDate(resumeDetail.created_at, 'datetime')}
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {resumeDetail.email && (
                    <div className="flex items-center gap-2">
                      <Mail className="h-4 w-4 text-gray-500" />
                      <span className="text-sm">{resumeDetail.email}</span>
                    </div>
                  )}
                  {resumeDetail.phone && (
                    <div className="flex items-center gap-2">
                      <Phone className="h-4 w-4 text-gray-500" />
                      <span className="text-sm">{resumeDetail.phone}</span>
                    </div>
                  )}
                  {resumeDetail.current_city && (
                    <div className="flex items-center gap-2">
                      <MapPin className="h-4 w-4 text-gray-500" />
                      <span className="text-sm">{resumeDetail.current_city}</span>
                    </div>
                  )}
                  {resumeDetail.gender && (
                    <div className="flex items-center gap-2">
                      <User className="h-4 w-4 text-gray-500" />
                      <span className="text-sm">性别: {resumeDetail.gender}</span>
                    </div>
                  )}
                  {resumeDetail.birthday && (
                    <div className="flex items-center gap-2">
                      <Calendar className="h-4 w-4 text-gray-500" />
                      <span className="text-sm">生日: {formatDate(resumeDetail.birthday)}</span>
                    </div>
                  )}
                  {resumeDetail.highest_education && (
                    <div className="flex items-center gap-2">
                      <GraduationCap className="h-4 w-4 text-gray-500" />
                      <span className="text-sm">学历: {resumeDetail.highest_education}</span>
                    </div>
                  )}
                  {resumeDetail.years_experience !== undefined && (
                    <div className="flex items-center gap-2">
                      <Briefcase className="h-4 w-4 text-gray-500" />
                      <span className="text-sm">工作经验: {resumeDetail.years_experience}年</span>
                    </div>
                  )}
                </div>
              </div>

              {/* 工作经历 */}
              {resumeDetail.experiences && resumeDetail.experiences.length > 0 && (
                <div>
                  <div className="flex items-center gap-2 mb-4">
                    <Briefcase className="h-5 w-5 text-gray-700" />
                    <h3 className="text-lg font-semibold text-gray-900">工作经历</h3>
                  </div>
                  <div className="space-y-4">
                    {resumeDetail.experiences.map((exp, index) => (
                      <div key={exp.id || index} className="border rounded-lg p-4">
                        <div className="flex justify-between items-start mb-2">
                          <div>
                            <h4 className="font-semibold text-gray-900">{exp.position}</h4>
                            <p className="text-gray-600">{exp.company}</p>
                          </div>
                          <div className="text-sm text-gray-500">
                            {formatDate(exp.start_date)} - {exp.end_date ? formatDate(exp.end_date) : '至今'}
                          </div>
                        </div>
                        {exp.description && (
                          <p className="text-sm text-gray-700 mt-2">{exp.description}</p>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* 教育背景 */}
              {resumeDetail.educations && resumeDetail.educations.length > 0 && (
                <div>
                  <div className="flex items-center gap-2 mb-4">
                    <GraduationCap className="h-5 w-5 text-gray-700" />
                    <h3 className="text-lg font-semibold text-gray-900">教育背景</h3>
                  </div>
                  <div className="space-y-4">
                    {resumeDetail.educations.map((edu, index) => (
                      <div key={edu.id || index} className="border rounded-lg p-4">
                        <div className="flex justify-between items-start mb-2">
                          <div>
                            <h4 className="font-semibold text-gray-900">{edu.school}</h4>
                            <p className="text-gray-600">{edu.major} - {edu.degree}</p>
                          </div>
                          <div className="text-sm text-gray-500">
                            {formatDate(edu.start_date)} - {edu.end_date ? formatDate(edu.end_date) : '至今'}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* 技能信息 */}
              {resumeDetail.skills && resumeDetail.skills.length > 0 && (
                <div>
                  <div className="flex items-center gap-2 mb-4">
                    <Award className="h-5 w-5 text-gray-700" />
                    <h3 className="text-lg font-semibold text-gray-900">技能信息</h3>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {resumeDetail.skills.map((skill, index) => (
                      <div key={skill.id || index} className="border rounded-lg p-4">
                        <div className="flex justify-between items-start mb-2">
                          <h4 className="font-semibold text-gray-900">{skill.skill_name}</h4>
                          {skill.level && (
                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">{skill.level}</span>
                          )}
                        </div>
                        {skill.description && (
                          <p className="text-sm text-gray-700">{skill.description}</p>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* 解析失败信息 */}
              {resumeDetail.status === ResumeStatus.FAILED && resumeDetail.error_message && (
                <div>
                  <div className="border-t border-gray-200 my-4"></div>
                  <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                    <h3 className="text-lg font-semibold text-red-800 mb-2">解析失败原因</h3>
                    <p className="text-sm text-red-700">{resumeDetail.error_message}</p>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}