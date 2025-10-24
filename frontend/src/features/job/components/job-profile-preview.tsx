import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/ui/dialog';
import { Button } from '@/ui/button';
import {
  X,
  MapPin,
  Briefcase,
  Building2,
  CheckCircle2,
  Gift,
  GraduationCap,
  ClipboardList,
  Award,
} from 'lucide-react';
import type { JobProfileDetail } from '@/types/job-profile';

interface JobProfilePreviewProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  jobProfile: JobProfileDetail | null;
}

// 学历类型转换
const getEducationLabel = (educationType: string): string => {
  const educationMap: Record<string, string> = {
    unlimited: '不限',
    junior_college: '大专',
    bachelor: '本科',
    master: '硕士',
    doctor: '博士',
  };
  return educationMap[educationType] || educationType;
};

// 工作经验转换
const getExperienceLabel = (experienceType: string): string => {
  const experienceMap: Record<string, string> = {
    unlimited: '不限',
    fresh_graduate: '应届毕业生',
    under_one_year: '1年以下',
    one_to_three_years: '1-3年',
    three_to_five_years: '3-5年',
    five_to_ten_years: '5-10年',
    over_ten_years: '10年以上',
  };
  return experienceMap[experienceType] || experienceType;
};

export function JobProfilePreview({
  open,
  onOpenChange,
  jobProfile,
}: JobProfilePreviewProps) {
  if (!jobProfile) return null;

  // 格式化薪资范围
  const salaryRange =
    jobProfile.salary_min && jobProfile.salary_max
      ? `${jobProfile.salary_min / 1000}K-${jobProfile.salary_max / 1000}K·15薪`
      : '面议';

  // 获取学历要求
  const educationLabel =
    jobProfile.education_requirements &&
    jobProfile.education_requirements.length > 0
      ? getEducationLabel(jobProfile.education_requirements[0].education_type)
      : '不限';

  // 获取工作经验要求
  const experienceLabel =
    jobProfile.experience_requirements &&
    jobProfile.experience_requirements.length > 0
      ? getExperienceLabel(
          jobProfile.experience_requirements[0].experience_type
        )
      : '不限';

  // 获取岗位职责列表
  const responsibilities =
    jobProfile.responsibilities?.map((r) => r.responsibility) || [];

  // 获取技能列表
  const requiredSkills =
    jobProfile.skills
      ?.filter((s) => s.type === 'required')
      .map((s) => s.skill) || [];
  const bonusSkills =
    jobProfile.skills?.filter((s) => s.type === 'bonus').map((s) => s.skill) ||
    [];

  // 获取工作性质标签
  const getWorkTypeLabel = (workType: string): string => {
    const workTypeMap: Record<string, string> = {
      full_time: '全职',
      part_time: '兼职',
      internship: '实习',
      outsourcing: '外包',
    };
    return workTypeMap[workType] || workType;
  };

  const workTypeLabel = jobProfile.work_type
    ? getWorkTypeLabel(jobProfile.work_type)
    : '';

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[900px] max-h-[90vh] overflow-hidden p-0 gap-0">
        {/* 顶部标题区域 - 白色背景 */}
        <div className="bg-white border-b border-gray-100">
          <div className="flex items-center justify-between px-6 py-2.5">
            <DialogHeader className="space-y-0">
              <DialogTitle className="text-[16px] font-semibold text-[#1F2937]">
                岗位画像预览
              </DialogTitle>
            </DialogHeader>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onOpenChange(false)}
              className="h-6 w-6 p-0 text-[#546E7A] hover:text-[#1F2937] hover:bg-gray-100"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* 岗位名称区域 - 淡蓝色渐变背景 */}
        <div className="bg-gradient-to-r from-[#E3F2FD] via-[#BBDEFB] to-[#E3F2FD] px-6 py-5">
          <div className="flex items-start justify-between mb-3">
            <h1 className="text-[24px] font-semibold text-[#1F2937]">
              {jobProfile.name}
            </h1>
            <div className="text-[26px] font-semibold text-[#FF6B35] whitespace-nowrap ml-4">
              {salaryRange}
            </div>
          </div>

          {/* 标签信息 - 一行展示：工作地点、工作经验要求、所属部门、工作性质 */}
          <div className="flex items-center gap-4 flex-wrap mb-3">
            {jobProfile.location && (
              <div className="flex items-center gap-1.5 text-[14px] text-[#546E7A]">
                <MapPin className="h-4 w-4" />
                <span>{jobProfile.location}</span>
              </div>
            )}
            {experienceLabel && experienceLabel !== '不限' && (
              <div className="flex items-center gap-1.5 text-[14px] text-[#546E7A]">
                <Briefcase className="h-4 w-4" />
                <span>{experienceLabel}</span>
              </div>
            )}
            {jobProfile.department && (
              <div className="flex items-center gap-1.5 text-[14px] text-[#546E7A]">
                <Building2 className="h-4 w-4" />
                <span>{jobProfile.department}</span>
              </div>
            )}
            {workTypeLabel && (
              <div className="flex items-center gap-1.5 text-[14px] text-[#546E7A]">
                <CheckCircle2 className="h-4 w-4" />
                <span>{workTypeLabel}</span>
              </div>
            )}
            {educationLabel && educationLabel !== '不限' && (
              <div className="flex items-center gap-1.5 text-[14px] text-[#546E7A]">
                <GraduationCap className="h-4 w-4" />
                <span>{educationLabel}及以上</span>
              </div>
            )}
          </div>
        </div>

        {/* 白色内容区域 */}
        <div className="overflow-y-auto max-h-[calc(90vh-200px)] bg-[#F5F5F5]">
          <div className="p-6 space-y-5">
            {/* 岗位职责及要求 */}
            {responsibilities.length > 0 && (
              <div className="bg-white rounded-lg p-5 shadow-sm">
                <h3 className="text-[16px] font-semibold text-[#1F2937] mb-4 flex items-center gap-2">
                  <ClipboardList className="h-4 w-4 text-[#7bb8ff]" />
                  岗位职责及要求
                </h3>
                <div className="space-y-3">
                  {responsibilities.map((responsibility, index) => (
                    <div key={index} className="flex items-start gap-2">
                      <span className="text-[14px] text-[#6B7280] mt-1 flex-shrink-0">
                        {index + 1}.
                      </span>
                      <p className="text-[14px] text-[#374151] leading-relaxed">
                        {responsibility}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* 技能要求 - 小标签形式 */}
            {(requiredSkills.length > 0 || bonusSkills.length > 0) && (
              <div className="bg-white rounded-lg p-5 shadow-sm">
                <h3 className="text-[16px] font-semibold text-[#1F2937] mb-4 flex items-center gap-2">
                  <Award className="h-4 w-4 text-[#7bb8ff]" />
                  技能要求
                </h3>
                <div className="space-y-3">
                  {requiredSkills.length > 0 && (
                    <div>
                      <div className="text-[12px] text-[#6B7280] mb-2 font-medium">
                        必备技能
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {requiredSkills.map((skill, index) => (
                          <span
                            key={index}
                            className="inline-flex items-center px-3 py-1.5 bg-[#FEE2E2] text-[#DC2626] rounded-md text-[13px] font-medium"
                          >
                            {skill}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                  {bonusSkills.length > 0 && (
                    <div>
                      <div className="text-[12px] text-[#6B7280] mb-2 font-medium">
                        加分技能
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {bonusSkills.map((skill, index) => (
                          <span
                            key={index}
                            className="inline-flex items-center px-3 py-1.5 bg-[#DBEAFE] text-[#2563EB] rounded-md text-[13px] font-medium"
                          >
                            {skill}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* 其他信息 */}
            <div className="bg-white rounded-lg p-5 shadow-sm">
              <h3 className="text-[16px] font-semibold text-[#1F2937] mb-4 flex items-center gap-2">
                <Gift className="h-4 w-4 text-[#7bb8ff]" />
                其他信息
              </h3>
              {jobProfile.description ? (
                <div className="text-[14px] text-[#374151] leading-relaxed whitespace-pre-wrap">
                  {jobProfile.description}
                </div>
              ) : (
                <div className="text-[14px] text-[#9CA3AF] leading-relaxed">
                  暂无其他信息
                </div>
              )}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
