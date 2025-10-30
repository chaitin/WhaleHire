import { useState } from 'react';
import { Dialog, DialogContent } from '@/ui/dialog';
import { Button } from '@/ui/button';
import {
  X,
  MapPin,
  Briefcase,
  GraduationCap,
  Share2,
  Clock,
  Sparkles,
  Target,
  CheckCircle2,
  Star,
  ChevronDown,
  ChevronUp,
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

// 工作性质转换
const getWorkTypeLabel = (workType: string): string => {
  const workTypeMap: Record<string, string> = {
    full_time: '全职',
    part_time: '兼职',
    internship: '实习',
    outsourcing: '外包',
  };
  return workTypeMap[workType] || workType;
};

export function JobProfilePreview({
  open,
  onOpenChange,
  jobProfile,
}: JobProfilePreviewProps) {
  // 加分项展开/收起状态
  const [isBonusExpanded, setIsBonusExpanded] = useState(false);

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

  // 从description中提取任职要求(如果存在)
  const getRequirementsFromDescription = () => {
    if (!jobProfile.description) return [];

    // 查找"任职要求"后的内容
    const requirementMatch = jobProfile.description.match(
      /任职要求[：:]\s*([\s\S]*?)(?=\n\n|$)/
    );
    if (requirementMatch && requirementMatch[1]) {
      // 按行分割,过滤空行
      return requirementMatch[1]
        .split('\n')
        .map((line) => line.trim())
        .filter((line) => line && line.length > 0);
    }
    return [];
  };

  // 获取技能列表 - 用于加分项展示
  // 直接使用所有skills作为加分项展示
  const bonusSkills = jobProfile.skills || [];

  // 任职要求 - 从description中提取
  const requiredSkills = getRequirementsFromDescription();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[700px] h-[90vh] overflow-hidden p-0 gap-0">
        {/* 单一框体 */}
        <div className="bg-white rounded-xl overflow-hidden flex flex-col h-full">
          {/* 顶部标题栏 */}
          <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
            <h2 className="text-xl font-semibold text-[#1E293B]">岗位预览</h2>
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
              background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
            }}
          >
            {/* 整体布局容器 */}
            <div className="relative">
              {/* 岗位名称和状态 */}
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3 flex-1">
                  {/* 岗位图标 - 使用目标靶心图标代表岗位 */}
                  <div
                    className="w-12 h-12 rounded-full flex items-center justify-center flex-shrink-0"
                    style={{
                      background:
                        'linear-gradient(135deg, rgba(255,255,255,0.3) 0%, rgba(255,255,255,0.1) 100%)',
                      border: '2px solid rgba(255,255,255,0.6)',
                      boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
                    }}
                  >
                    <Target className="w-6 h-6 text-white" />
                  </div>

                  <div className="flex items-center justify-between flex-1 gap-3">
                    {/* 左侧:岗位名称、部门和状态标签 */}
                    <div className="flex items-center gap-2">
                      <h1 className="text-xl font-semibold text-white">
                        {jobProfile.name}
                      </h1>
                      {/* 部门显示 */}
                      {jobProfile.department && (
                        <span className="text-sm text-white/90">
                          {jobProfile.department}
                        </span>
                      )}
                      {/* 状态标签 - 字体加大加粗 */}
                      <div className="relative inline-flex items-center">
                        <div className="absolute inset-0 bg-gradient-to-r from-green-400 to-emerald-500 rounded blur-[1px] opacity-50"></div>
                        <div className="relative px-1.5 py-[2px] bg-gradient-to-r from-green-400 to-emerald-500 rounded">
                          <span className="text-[9px] font-bold text-white tracking-wide">
                            已发布
                          </span>
                        </div>
                      </div>
                    </div>
                    {/* 右侧:薪资范围右对齐 */}
                    <span
                      className="text-lg font-bold"
                      style={{ color: '#FF8C42' }}
                    >
                      {formatSalary()}
                    </span>
                  </div>
                </div>
              </div>

              {/* 关键信息标签 - 简化显示:只显示icon+内容 */}
              <div className="flex flex-wrap items-center gap-x-4 gap-y-2 mb-2">
                {/* 工作地点 */}
                {jobProfile.location && (
                  <div className="flex items-center gap-1.5">
                    <MapPin className="w-3 h-3.5 text-white" />
                    <span className="text-sm text-white">
                      {jobProfile.location}
                    </span>
                  </div>
                )}
                {/* 工作性质 */}
                {jobProfile.work_type && (
                  <div className="flex items-center gap-1.5">
                    <Briefcase className="w-3.5 h-3.5 text-white" />
                    <span className="text-sm text-white">
                      {getWorkTypeLabel(jobProfile.work_type)}
                    </span>
                  </div>
                )}
                {/* 学历要求 */}
                {educationLabel && educationLabel !== '不限' && (
                  <div className="flex items-center gap-1.5">
                    <GraduationCap className="w-3.5 h-3.5 text-white" />
                    <span className="text-sm text-white">{educationLabel}</span>
                  </div>
                )}
                {/* 工作经验 */}
                {experienceLabel && experienceLabel !== '不限' && (
                  <div className="flex items-center gap-1.5">
                    <Clock className="w-3.5 h-3.5 text-white" />
                    <span className="text-sm text-white">
                      {experienceLabel}
                    </span>
                  </div>
                )}
              </div>

              {/* 操作按钮 - 放在框体的右下角 */}
              <div className="absolute bottom-0 right-0 flex items-center gap-1">
                {/* 结构体 - 可点击（已注释，不展示） */}
                {/* <Button
                  variant="outline"
                  className="h-6 px-2 bg-white/20 border-white/30 text-white hover:bg-white/30"
                  onClick={() => {
                    alert('结构体功能开发中');
                  }}
                >
                  <Grid className="w-3 h-3 mr-1" />
                  <span className="text-xs font-medium">结构体</span>
                </Button> */}
                {/* AI关联 - 置灰 */}
                <Button
                  variant="outline"
                  className="h-6 px-2 bg-white/20 border-white/30 text-white hover:bg-white/30 opacity-50 cursor-not-allowed"
                  disabled
                >
                  <Sparkles className="w-3 h-3 mr-1" />
                  <span className="text-xs font-medium">AI关联</span>
                </Button>
                {/* 分享 - 置灰 */}
                <Button
                  variant="outline"
                  className="h-6 px-2 bg-white/20 border-white/30 text-white hover:bg-white/30 opacity-50 cursor-not-allowed"
                  disabled
                >
                  <Share2 className="w-3 h-3 mr-1" />
                  <span className="text-xs font-medium">分享</span>
                </Button>
              </div>
            </div>
          </div>

          {/* 内容区域 - 添加科技感线条 */}
          <div className="flex-1 overflow-y-auto px-6 py-4 bg-white space-y-6 relative">
            {/* 科技感装饰线条 */}
            <div className="absolute inset-0 opacity-10 pointer-events-none">
              <div className="absolute top-0 left-0 w-full h-px bg-gradient-to-r from-transparent via-blue-500 to-transparent"></div>
              <div className="absolute top-1/3 left-0 w-full h-px bg-gradient-to-r from-transparent via-purple-500 to-transparent"></div>
              <div className="absolute top-2/3 left-0 w-full h-px bg-gradient-to-r from-transparent via-blue-500 to-transparent"></div>
              <div className="absolute top-0 left-1/4 w-px h-full bg-gradient-to-b from-transparent via-blue-500 to-transparent"></div>
              <div className="absolute top-0 left-3/4 w-px h-full bg-gradient-to-b from-transparent via-purple-500 to-transparent"></div>
            </div>

            {/* 岗位职责 */}
            {responsibilities.length > 0 && (
              <div className="relative space-y-3">
                {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
                <div className="flex items-center gap-3">
                  <div
                    className="w-1 h-7 rounded"
                    style={{
                      background:
                        'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                    }}
                  />
                  <h3 className="text-lg font-bold text-[#1E293B]">岗位职责</h3>
                  <div className="flex-1 h-px bg-gradient-to-r from-blue-300 via-purple-200 to-transparent"></div>
                </div>
                <ul className="space-y-2.5 relative">
                  {/* 左侧垂直装饰线 */}
                  <div className="absolute left-0 top-0 bottom-0 w-px bg-gradient-to-b from-blue-200 via-purple-100 to-transparent opacity-30"></div>
                  {responsibilities.map((resp, index) => (
                    <li key={index} className="flex items-start gap-2.5 pl-3">
                      {/* 使用Target图标代表岗位职责 */}
                      <div className="w-5 h-5 flex items-center justify-center flex-shrink-0 mt-0.5">
                        <Target className="w-4 h-4 text-[#3B82F6]" />
                      </div>
                      <span className="text-sm font-medium text-[#64748B] leading-relaxed flex-1">
                        {resp}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* 任职要求 - 从description中提取 */}
            {requiredSkills.length > 0 && (
              <div className="relative space-y-3">
                {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
                <div className="flex items-center gap-3">
                  <div
                    className="w-1 h-7 rounded"
                    style={{
                      background:
                        'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                    }}
                  />
                  <h3 className="text-lg font-bold text-[#1E293B]">任职要求</h3>
                  <div className="flex-1 h-px bg-gradient-to-r from-green-300 via-teal-200 to-transparent"></div>
                </div>
                <ul className="space-y-2 relative">
                  {/* 左侧垂直装饰线 */}
                  <div className="absolute left-0 top-0 bottom-0 w-px bg-gradient-to-b from-green-200 via-teal-100 to-transparent opacity-30"></div>
                  {requiredSkills.map((requirement, index) => (
                    <li key={index} className="flex items-start gap-2.5 pl-3">
                      {/* 使用CheckCircle2图标代表任职要求 */}
                      <div className="w-4 h-4 flex items-center justify-center flex-shrink-0 mt-0.5">
                        <CheckCircle2 className="w-4 h-4 text-[#10B981]" />
                      </div>
                      <span className="text-sm font-medium text-[#64748B] leading-relaxed">
                        {requirement}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* 加分项 - 标签形式展示 */}
            {bonusSkills.length > 0 && (
              <div className="relative space-y-3">
                {/* 标题区 - 带左侧蓝色装饰条和右侧延伸线 */}
                <div className="flex items-center gap-3">
                  <div
                    className="w-1 h-7 rounded"
                    style={{
                      background:
                        'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                    }}
                  />
                  <h3 className="text-lg font-bold text-[#1E293B]">加分项</h3>
                  <div className="flex-1 h-px bg-gradient-to-r from-yellow-300 via-amber-200 to-transparent"></div>
                </div>
                {/* 标签展示区域 */}
                <div className="pl-3 space-y-2">
                  <div className="flex flex-wrap gap-2">
                    {(() => {
                      // 每行大约可以显示5-6个标签，三行约15个标签
                      const maxItemsBeforeCollapse = 15;
                      const shouldShowExpandButton =
                        bonusSkills.length > maxItemsBeforeCollapse;
                      const displayedSkills =
                        shouldShowExpandButton && !isBonusExpanded
                          ? bonusSkills.slice(0, maxItemsBeforeCollapse)
                          : bonusSkills;

                      return displayedSkills.map((skill, index) => (
                        <span
                          key={skill.id || index}
                          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-white text-sm font-medium"
                          style={{
                            background:
                              'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                          }}
                        >
                          <Star className="w-3.5 h-3.5" />
                          {skill.skill || skill.skill_name || ''}
                        </span>
                      ));
                    })()}
                  </div>
                  {/* 展开/收起按钮 */}
                  {bonusSkills.length > 15 && (
                    <button
                      onClick={() => setIsBonusExpanded(!isBonusExpanded)}
                      className="flex items-center gap-1 text-sm text-[#7bb8ff] hover:text-[#6aa8ee] transition-colors"
                    >
                      {isBonusExpanded ? (
                        <>
                          <ChevronUp className="w-4 h-4" />
                          <span>收起</span>
                        </>
                      ) : (
                        <>
                          <ChevronDown className="w-4 h-4" />
                          <span>展开</span>
                        </>
                      )}
                    </button>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
