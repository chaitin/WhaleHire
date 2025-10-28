import { useState, useEffect } from 'react';
import {
  X,
  Download,
  Edit,
  Mail,
  Phone,
  MapPin,
  Eye,
  Briefcase,
  DollarSign,
  Building2,
  Calendar,
} from 'lucide-react';
import { Dialog, DialogContent } from '@/ui/dialog';
import { Button } from '@/ui/button';
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
  onEdit,
}: ResumePreviewModalProps) {
  const [resumeDetail, setResumeDetail] = useState<ResumeDetail | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);

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

  // 编辑简历
  const handleEdit = () => {
    onEdit(resumeId);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-[700px] h-[90vh] overflow-hidden p-0 gap-0">
        {/* 单一框体 */}
        <div className="bg-white rounded-xl overflow-hidden flex flex-col h-full">
          {/* 顶部标题栏 */}
          <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
            <h2 className="text-2xl font-semibold text-[#1E293B]">简历预览</h2>
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
          ) : resumeDetail ? (
            <>
              {/* 渐变头部区域 */}
              <div
                className="px-8 py-5 border-b border-gray-200"
                style={{
                  background:
                    'linear-gradient(135deg, #8b5cf6 0%, #7bb8ff 100%)',
                }}
              >
                {/* 姓名、职位、联系方式同一排 */}
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3 max-w-[65%]">
                    <h1 className="text-lg font-semibold text-white">
                      {resumeDetail.name}
                    </h1>
                    {resumeDetail.experiences?.[0]?.position && (
                      <div className="px-2.5 py-0.5 bg-white/20 rounded-full">
                        <span className="text-xs font-medium text-white">
                          {resumeDetail.experiences[0].position}
                        </span>
                      </div>
                    )}
                  </div>

                  {/* 操作按钮 */}
                  <div className="flex gap-1.5">
                    <Button
                      variant="outline"
                      className="h-7 px-2.5 bg-white/20 border-white/30 text-white hover:bg-white/30"
                      onClick={handleViewOriginal}
                      disabled={!resumeDetail.resume_file_url}
                    >
                      <Eye className="w-3 h-3 mr-1.5" />
                      <span className="text-xs font-medium">查看</span>
                    </Button>
                    <Button
                      variant="outline"
                      className="h-7 px-2.5 bg-white/20 border-white/30 text-white hover:bg-white/30"
                      onClick={handleEdit}
                    >
                      <Edit className="w-3 h-3 mr-1.5" />
                      <span className="text-xs font-medium">编辑</span>
                    </Button>
                    <Button
                      variant="outline"
                      className="h-7 px-2.5 bg-white/20 border-white/30 text-white hover:bg-white/30"
                      onClick={handleDownload}
                      disabled={isDownloading || !resumeDetail.resume_file_url}
                    >
                      <Download className="w-3 h-3 mr-1.5" />
                      <span className="text-xs font-medium">下载</span>
                    </Button>
                  </div>
                </div>

                {/* 关键信息标签 */}
                <div className="flex items-center gap-3">
                  {resumeDetail.email && (
                    <div className="flex items-center gap-1.5">
                      <Mail className="w-3.5 h-3.5 text-white" />
                      <span className="text-sm text-white">
                        {resumeDetail.email}
                      </span>
                    </div>
                  )}
                  {resumeDetail.phone && (
                    <div className="flex items-center gap-1.5">
                      <Phone className="w-3.5 h-3.5 text-white" />
                      <span className="text-sm text-white">
                        {resumeDetail.phone}
                      </span>
                    </div>
                  )}
                  {resumeDetail.current_city && (
                    <div className="flex items-center gap-1.5">
                      <MapPin className="w-3 h-3.5 text-white" />
                      <span className="text-sm text-white">
                        {resumeDetail.current_city}
                      </span>
                    </div>
                  )}
                </div>
              </div>

              {/* 内容区域 - 所有内容垂直滚动展示 */}
              <div className="flex-1 overflow-y-auto px-6 py-4 bg-white space-y-6">
                {/* 个人简介 */}
                <IntroSection resumeDetail={resumeDetail} />

                {/* 工作经验 */}
                <ExperienceSection resumeDetail={resumeDetail} />

                {/* 教育背景 */}
                <EducationSection resumeDetail={resumeDetail} />

                {/* 项目经验 */}
                <ProjectsSection resumeDetail={resumeDetail} />

                {/* 技能专长 */}
                <SkillsSection resumeDetail={resumeDetail} />

                {/* 求职意向 */}
                <JobIntentionSection resumeDetail={resumeDetail} />

                {/* 证书与培训 */}
                <CertificatesSection />
              </div>
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
function IntroSection({ resumeDetail }: { resumeDetail: ResumeDetail }) {
  // 个人简介字段检查 - 假设有 personal_summary 或 summary 字段
  const hasDescription =
    (resumeDetail as { personal_summary?: string }).personal_summary ||
    (resumeDetail as { summary?: string }).summary;

  // 如果没有个人简介内容,则不显示整个section
  if (!hasDescription) {
    return null;
  }

  return (
    <div>
      <h4 className="text-base font-semibold text-[#1E293B] mb-3">个人简介</h4>
      <p className="text-sm text-[#1E293B] leading-[1.429]">{hasDescription}</p>
    </div>
  );
}

// 求职意向
function JobIntentionSection({ resumeDetail }: { resumeDetail: ResumeDetail }) {
  return (
    <div>
      <h4 className="text-base font-semibold text-[#1E293B] mb-3">求职意向</h4>
      <div className="flex items-center justify-between">
        {/* 期望职位 */}
        <div className="flex items-center gap-2">
          <Briefcase className="w-4 h-4 text-gray-400 flex-shrink-0" />
          <span className="text-sm font-medium text-gray-600">期望职位</span>
          <span className="text-sm text-gray-900">
            {resumeDetail.experiences?.[0]?.position || '--'}
          </span>
        </div>

        {/* 期望年薪 */}
        <div className="flex items-center gap-2">
          <DollarSign className="w-4 h-4 text-gray-400 flex-shrink-0" />
          <span className="text-sm font-medium text-gray-600">期望年薪</span>
          <span className="text-sm text-gray-900">
            {(resumeDetail as { expected_salary?: string }).expected_salary ||
              '--'}
          </span>
        </div>

        {/* 行业偏好 */}
        <div className="flex items-center gap-2">
          <Building2 className="w-4 h-4 text-gray-400 flex-shrink-0" />
          <span className="text-sm font-medium text-gray-600">行业偏好</span>
          <span className="text-sm text-gray-900">
            {(resumeDetail as { industry_preference?: string })
              .industry_preference || '--'}
          </span>
        </div>

        {/* 到岗时间 */}
        <div className="flex items-center gap-2">
          <Calendar className="w-4 h-4 text-gray-400 flex-shrink-0" />
          <span className="text-sm font-medium text-gray-600">到岗时间</span>
          <span className="text-sm text-gray-900">
            {(resumeDetail as { available_date?: string }).available_date ||
              '--'}
          </span>
        </div>
      </div>
    </div>
  );
}

// 工作经验
function ExperienceSection({ resumeDetail }: { resumeDetail: ResumeDetail }) {
  const experiences = resumeDetail.experiences || [];

  return (
    <div>
      <h4 className="text-base font-semibold text-[#1E293B] mb-3">工作经验</h4>

      {experiences.length > 0 ? (
        <ul className="space-y-2.5">
          {experiences.map((exp, index) => (
            <li key={exp.id} className="flex items-start gap-2.5">
              <div className="w-5 h-5 rounded-full bg-[#7bb8ff] flex items-center justify-center flex-shrink-0 mt-0.5">
                <span className="text-xs font-medium text-white">
                  {index + 1}
                </span>
              </div>
              <div className="flex-1">
                <div className="flex items-center justify-between gap-2 mb-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold text-[#1E293B]">
                      {exp.company}
                    </span>
                    <span className="text-gray-400">|</span>
                    <span className="text-sm text-[#1E293B]">
                      {exp.position}
                    </span>
                  </div>
                  <span className="text-xs text-gray-600">
                    {exp.start_date?.split('-').slice(0, 2).join('-')}
                    {exp.end_date
                      ? ` - ${exp.end_date.split('-').slice(0, 2).join('-')}`
                      : ' - 至今'}
                  </span>
                </div>
                {exp.description && (
                  <p className="text-sm text-[#1E293B] leading-[1.429]">
                    {exp.description}
                  </p>
                )}
              </div>
            </li>
          ))}
        </ul>
      ) : (
        <div className="text-center text-gray-500 py-6">暂无工作经验</div>
      )}
    </div>
  );
}

// 教育背景
function EducationSection({ resumeDetail }: { resumeDetail: ResumeDetail }) {
  const educations = resumeDetail.educations || [];

  // 学历类型映射
  const degreeMap: Record<string, string> = {
    bachelor: '本科',
    master: '硕士',
    doctor: '博士',
    junior_college: '大专',
  };

  return (
    <div>
      <h4 className="text-base font-semibold text-[#1E293B] mb-3">教育背景</h4>

      {educations.length > 0 ? (
        <ul className="space-y-2">
          {educations.map((edu) => (
            <li key={edu.id} className="flex items-start gap-2">
              <div className="w-1.5 h-1.5 rounded-full bg-[#7bb8ff] flex-shrink-0 mt-2"></div>
              <div className="flex-1">
                <div className="flex items-start justify-between gap-2">
                  <div className="flex-1">
                    <span className="text-sm text-[#1E293B] leading-[1.429]">
                      {edu.school} - {degreeMap[edu.degree] || edu.degree} ·{' '}
                      {edu.major}
                    </span>
                  </div>
                  <span className="text-xs text-gray-600 whitespace-nowrap">
                    {edu.start_date?.split('-').slice(0, 2).join('-')} -{' '}
                    {edu.end_date?.split('-').slice(0, 2).join('-')}
                  </span>
                </div>
              </div>
            </li>
          ))}
        </ul>
      ) : (
        <div className="text-center text-gray-500 py-6">暂无教育背景</div>
      )}
    </div>
  );
}

// 技能专长
function SkillsSection({ resumeDetail }: { resumeDetail: ResumeDetail }) {
  const skills = resumeDetail.skills || [];

  if (skills.length === 0) {
    return null;
  }

  return (
    <div>
      <h4 className="text-base font-semibold text-[#1E293B] mb-3">技能专长</h4>
      <ul className="space-y-2">
        {skills.map((skill) => (
          <li key={skill.id} className="flex items-start gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-[#7bb8ff] flex-shrink-0 mt-2"></div>
            <span className="text-sm text-[#1E293B] leading-[1.429]">
              {skill.skill_name} {skill.level ? `(${skill.level})` : ''}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}

// 项目经验
function ProjectsSection({ resumeDetail }: { resumeDetail: ResumeDetail }) {
  const projects = resumeDetail.projects || [];

  // 如果没有项目经验,不显示此区域
  if (projects.length === 0) {
    return null;
  }

  return (
    <div>
      <h4 className="text-base font-semibold text-[#1E293B] mb-3">项目经验</h4>

      <ul className="space-y-2.5">
        {projects.map((project, index) => (
          <li key={project.id} className="flex items-start gap-2.5">
            <div className="w-5 h-5 rounded-full bg-[#7bb8ff] flex items-center justify-center flex-shrink-0 mt-0.5">
              <span className="text-xs font-medium text-white">
                {index + 1}
              </span>
            </div>
            <div className="flex-1">
              <div className="flex items-center justify-between gap-2 mb-1">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-semibold text-[#1E293B]">
                    {project.name}
                  </span>
                  {project.company && (
                    <>
                      <span className="text-gray-400">|</span>
                      <span className="text-sm text-[#1E293B]">
                        {project.company}
                      </span>
                    </>
                  )}
                </div>
                <span className="text-xs text-gray-600">
                  {project.start_date?.split('-').slice(0, 2).join('-')} -{' '}
                  {project.end_date
                    ? project.end_date.split('-').slice(0, 2).join('-')
                    : '至今'}
                </span>
              </div>
              {project.description && (
                <p className="text-sm text-[#1E293B] leading-[1.429]">
                  {project.description}
                </p>
              )}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}

// 证书与培训
function CertificatesSection() {
  // 由于后端可能没有证书相关字段，这里先用空数组
  const certificates: Array<{
    id: string;
    name: string;
    issuer: string;
    date: string;
  }> = [];

  // 如果没有证书，不显示此区域
  if (certificates.length === 0) {
    return null;
  }

  return (
    <div>
      <h4 className="text-base font-semibold text-[#1E293B] mb-3">
        证书与培训
      </h4>
      <ul className="space-y-2">
        {certificates.map((cert) => (
          <li key={cert.id} className="flex items-start gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-[#7bb8ff] flex-shrink-0 mt-2"></div>
            <span className="text-sm text-[#1E293B] leading-[1.429]">
              {cert.name} - {cert.issuer} ({cert.date})
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}
