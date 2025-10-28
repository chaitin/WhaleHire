import { Dialog, DialogContent } from '@/ui/dialog';
import { Button } from '@/ui/button';
import {
  X,
  MapPin,
  Briefcase,
  GraduationCap,
  Share2,
  Clock,
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
  const formatSalary = () => {
    if (jobProfile.salary_min && jobProfile.salary_max) {
      return `${jobProfile.salary_min}K-${jobProfile.salary_max}K/月`;
    } else if (jobProfile.salary_min) {
      return `${jobProfile.salary_min}K以上/月`;
    } else if (jobProfile.salary_max) {
      return `${jobProfile.salary_max}K以下/月`;
    }
    return '薪资面议';
  };

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
    jobProfile.skills?.filter((s) => s.type === 'required') || [];
  const bonusSkills =
    jobProfile.skills?.filter((s) => s.type === 'bonus') || [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[700px] h-[90vh] overflow-hidden p-0 gap-0">
        {/* 单一框体 */}
        <div className="bg-white rounded-xl overflow-hidden flex flex-col h-full">
          {/* 顶部标题栏 */}
          <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
            <h2 className="text-2xl font-semibold text-[#1E293B]">岗位预览</h2>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onOpenChange(false)}
              className="h-8 w-8 p-0 text-gray-500 hover:text-gray-900 hover:bg-gray-100"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>

          {/* 岗位头部信息 - 渐变背景 */}
          <div
            className="px-8 py-5 border-b border-gray-200"
            style={{
              background: 'linear-gradient(135deg, #8b5cf6 0%, #7bb8ff 100%)',
            }}
          >
            {/* 岗位名称、薪资、标签和分享按钮同一排 */}
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-3 max-w-[65%]">
                <h1 className="text-lg font-semibold text-white">
                  {jobProfile.name}
                </h1>
                <div className="px-2.5 py-0.5 bg-white/20 rounded-full">
                  <span className="text-xs font-medium text-white">已发布</span>
                </div>
                {/* 薪资 - 与岗位名称同一排 */}
                <div className="flex items-center gap-1.5 ml-auto">
                  <Briefcase className="w-4 h-3.5 text-white" />
                  <span className="text-lg font-bold text-white">
                    {formatSalary()}
                  </span>
                </div>
              </div>

              {/* 分享按钮 */}
              <Button
                variant="outline"
                className="h-7 px-2.5 bg-white/20 border-white/30 text-white hover:bg-white/30"
                onClick={() => {
                  alert('分享功能开发中');
                }}
              >
                <Share2 className="w-3 h-3 mr-1.5" />
                <span className="text-xs font-medium">分享</span>
              </Button>
            </div>

            {/* 关键信息标签 */}
            <div className="flex items-center gap-3">
              {experienceLabel && (
                <div className="flex items-center gap-1.5">
                  <Clock className="w-3.5 h-3.5 text-white" />
                  <span className="text-sm text-white">{experienceLabel}</span>
                </div>
              )}
              {educationLabel && (
                <div className="flex items-center gap-1.5">
                  <GraduationCap className="w-4 h-3.5 text-white" />
                  <span className="text-sm text-white">{educationLabel}</span>
                </div>
              )}
              {jobProfile.location && (
                <div className="flex items-center gap-1.5">
                  <MapPin className="w-3 h-3.5 text-white" />
                  <span className="text-sm text-white">
                    {jobProfile.location}
                  </span>
                </div>
              )}
            </div>
          </div>

          {/* 内容区域 */}
          <div className="flex-1 overflow-y-auto px-6 py-4 bg-white space-y-6">
            {/* 岗位职责 - 移到最上方 */}
            {responsibilities.length > 0 && (
              <div>
                <h4 className="text-base font-semibold text-[#1E293B] mb-3">
                  岗位职责
                </h4>
                <ul className="space-y-2.5">
                  {responsibilities.map((resp, index) => (
                    <li key={index} className="flex items-start gap-2.5">
                      <div className="w-5 h-5 rounded-full bg-[#7bb8ff] flex items-center justify-center flex-shrink-0 mt-0.5">
                        <span className="text-xs font-medium text-white">
                          {index + 1}
                        </span>
                      </div>
                      <span className="text-sm text-[#1E293B] leading-[1.429] flex-1">
                        {resp}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* 任职要求 */}
            {requiredSkills.length > 0 && (
              <div>
                <h4 className="text-base font-semibold text-[#1E293B] mb-3">
                  任职要求
                </h4>
                <ul className="space-y-2">
                  {requiredSkills.map((skill, index) => (
                    <li key={index} className="flex items-start gap-2">
                      <div className="w-1.5 h-1.5 rounded-full bg-[#7bb8ff] flex-shrink-0 mt-2"></div>
                      <span className="text-sm text-[#1E293B] leading-[1.429]">
                        {skill.skill}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* 加分项 */}
            {bonusSkills.length > 0 && (
              <div>
                <h4 className="text-base font-semibold text-[#1E293B] mb-3">
                  加分项
                </h4>
                <ul className="space-y-2">
                  {bonusSkills.map((skill, index) => (
                    <li key={index} className="flex items-start gap-2">
                      <div className="w-1.5 h-1.5 rounded-full bg-[#7bb8ff] flex-shrink-0 mt-2"></div>
                      <span className="text-sm text-[#1E293B] leading-[1.429]">
                        {skill.skill}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
