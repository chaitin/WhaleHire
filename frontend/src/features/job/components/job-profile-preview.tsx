import { useState } from 'react';
import { Dialog, DialogContent } from '@/ui/dialog';
import { Button } from '@/ui/button';
import {
  X,
  MapPin,
  Briefcase,
  Building2,
  GraduationCap,
  Award,
  Target,
  FileText,
  LayoutList,
} from 'lucide-react';
import type { JobProfileDetail } from '@/types/job-profile';
import { cn } from '@/lib/utils';

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
  const [displayMode, setDisplayMode] = useState<'markdown' | 'structured'>(
    'markdown'
  );

  if (!jobProfile) return null;

  // 格式化薪资范围
  const formatSalary = () => {
    if (jobProfile.salary_min && jobProfile.salary_max) {
      return `${jobProfile.salary_min}K-${jobProfile.salary_max}K`;
    } else if (jobProfile.salary_min) {
      return `${jobProfile.salary_min}K以上`;
    } else if (jobProfile.salary_max) {
      return `${jobProfile.salary_max}K以下`;
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
      <DialogContent className="max-w-[800px] max-h-[90vh] overflow-hidden p-0 gap-0">
        {/* 顶部标题栏 */}
        <div className="bg-white border-b border-gray-200 px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <h2 className="text-[18px] font-semibold text-gray-900">
                岗位详情
              </h2>
              {/* 切换视图按钮 - 只显示图标 */}
              <div className="flex items-center gap-2 ml-4">
                <button
                  onClick={() => setDisplayMode('markdown')}
                  className={cn(
                    'p-2 rounded-lg transition-all',
                    displayMode === 'markdown'
                      ? 'text-white'
                      : 'text-gray-600 bg-gray-100 hover:bg-gray-200'
                  )}
                  style={
                    displayMode === 'markdown'
                      ? { backgroundColor: '#7bb8ff' }
                      : undefined
                  }
                  title="原文展示"
                >
                  <FileText className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setDisplayMode('structured')}
                  className={cn(
                    'p-2 rounded-lg transition-all',
                    displayMode === 'structured'
                      ? 'text-white'
                      : 'text-gray-600 bg-gray-100 hover:bg-gray-200'
                  )}
                  style={
                    displayMode === 'structured'
                      ? { backgroundColor: '#7bb8ff' }
                      : undefined
                  }
                  title="结构化展示"
                >
                  <LayoutList className="w-4 h-4" />
                </button>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onOpenChange(false)}
              className="h-8 w-8 p-0 text-gray-500 hover:text-gray-900 hover:bg-gray-100"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* 岗位头部信息 */}
        <div
          className="px-6 py-6 border-b border-gray-200"
          style={{ backgroundColor: '#f8fbff' }}
        >
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-[26px] font-bold text-gray-900">
              {jobProfile.name}
            </h1>
            <div className="flex items-center gap-2">
              <span className="text-[28px] font-bold text-[#FF6B35]">
                {formatSalary()}
              </span>
              <span className="text-[14px] text-gray-500">·月薪</span>
            </div>
          </div>

          {/* 关键信息标签 */}
          <div className="flex flex-wrap items-center gap-3">
            {jobProfile.location && (
              <div
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[14px]"
                style={{ backgroundColor: '#e6f4ff', color: '#7bb8ff' }}
              >
                <MapPin className="h-4 w-4" />
                <span>{jobProfile.location}</span>
              </div>
            )}
            {experienceLabel && (
              <div
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[14px]"
                style={{ backgroundColor: '#e6f4ff', color: '#7bb8ff' }}
              >
                <Briefcase className="h-4 w-4" />
                <span>{experienceLabel}</span>
              </div>
            )}
            {educationLabel && (
              <div
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[14px]"
                style={{ backgroundColor: '#e6f4ff', color: '#7bb8ff' }}
              >
                <GraduationCap className="h-4 w-4" />
                <span>{educationLabel}</span>
              </div>
            )}
            {workTypeLabel && (
              <div
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[14px]"
                style={{ backgroundColor: '#e6f4ff', color: '#7bb8ff' }}
              >
                <Briefcase className="h-4 w-4" />
                <span>{workTypeLabel}</span>
              </div>
            )}
            {jobProfile.department && (
              <div
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[14px]"
                style={{ backgroundColor: '#e6f4ff', color: '#7bb8ff' }}
              >
                <Building2 className="h-4 w-4" />
                <span>{jobProfile.department}</span>
              </div>
            )}
          </div>
        </div>

        {/* 内容区域 */}
        <div className="overflow-y-auto max-h-[calc(90vh-280px)] bg-gray-50">
          <div className="p-6 space-y-5">
            {displayMode === 'markdown' ? (
              /* 原文展示模式 */
              <div className="bg-white rounded-xl p-6 border border-gray-200">
                {(() => {
                  // 优先使用description字段(AI生成的原文)
                  if (jobProfile.description) {
                    const description = jobProfile.description;
                    const sections: Array<{
                      title: string;
                      content: string;
                      icon: React.ReactNode;
                    }> = [];

                    // 匹配岗位职责
                    const responsibilityMatch = description.match(
                      /岗位职责[：:]([\s\S]*?)(?=任职要求[：:]|加分项[：:]|$)/
                    );
                    if (responsibilityMatch) {
                      sections.push({
                        title: '岗位职责',
                        content: responsibilityMatch[1].trim(),
                        icon: (
                          <Target
                            className="w-5 h-5"
                            style={{ color: '#7bb8ff' }}
                          />
                        ),
                      });
                    }

                    // 匹配任职要求
                    const requirementMatch = description.match(
                      /任职要求[：:]([\s\S]*?)(?=加分项[：:]|$)/
                    );
                    if (requirementMatch) {
                      sections.push({
                        title: '任职要求',
                        content: requirementMatch[1].trim(),
                        icon: (
                          <Award
                            className="w-5 h-5"
                            style={{ color: '#7bb8ff' }}
                          />
                        ),
                      });
                    }

                    // 匹配加分项
                    const bonusMatch =
                      description.match(/加分项[：:]([\s\S]*?)$/);
                    if (bonusMatch) {
                      sections.push({
                        title: '加分项',
                        content: bonusMatch[1].trim(),
                        icon: <Award className="w-5 h-5 text-green-600" />,
                      });
                    }

                    // 如果匹配到了结构
                    if (sections.length > 0) {
                      return (
                        <div className="space-y-5">
                          {sections.map((section, idx) => (
                            <div key={idx}>
                              <div className="flex items-center gap-2 mb-3">
                                {section.icon}
                                <h3 className="text-[18px] font-bold text-gray-900">
                                  {section.title}
                                </h3>
                              </div>
                              <ul className="space-y-2 ml-2">
                                {section.content
                                  .split('\n')
                                  .filter((line) => line.trim())
                                  .map((line, lineIdx) => {
                                    // 移除行首的 - 或 • 或数字序号
                                    const cleanLine = line
                                      .replace(/^[-•\d+.]\s*/, '')
                                      .trim();
                                    if (!cleanLine) return null;
                                    return (
                                      <li
                                        key={lineIdx}
                                        className="flex items-start gap-2 text-[15px] text-gray-700"
                                      >
                                        <span
                                          className="inline-block w-1.5 h-1.5 rounded-full mt-2 flex-shrink-0"
                                          style={{ backgroundColor: '#7bb8ff' }}
                                        ></span>
                                        <span className="leading-relaxed">
                                          {cleanLine}
                                        </span>
                                      </li>
                                    );
                                  })}
                              </ul>
                            </div>
                          ))}
                        </div>
                      );
                    }

                    // 如果没有匹配到结构,显示原文
                    return (
                      <pre className="whitespace-pre-wrap font-sans text-[15px] text-gray-700 leading-relaxed">
                        {description}
                      </pre>
                    );
                  }

                  // 如果没有description,使用结构化数据生成展示(手动编辑模式)
                  if (
                    responsibilities.length > 0 ||
                    requiredSkills.length > 0 ||
                    bonusSkills.length > 0
                  ) {
                    return (
                      <div className="space-y-5">
                        {/* 岗位职责 */}
                        {responsibilities.length > 0 && (
                          <div>
                            <div className="flex items-center gap-2 mb-3">
                              <Target
                                className="w-5 h-5"
                                style={{ color: '#7bb8ff' }}
                              />
                              <h3 className="text-[18px] font-bold text-gray-900">
                                岗位职责
                              </h3>
                            </div>
                            <ul className="space-y-2 ml-2">
                              {responsibilities.map((resp, index) => (
                                <li
                                  key={index}
                                  className="flex items-start gap-2 text-[15px] text-gray-700"
                                >
                                  <span
                                    className="inline-block w-1.5 h-1.5 rounded-full mt-2 flex-shrink-0"
                                    style={{ backgroundColor: '#7bb8ff' }}
                                  ></span>
                                  <span className="leading-relaxed">
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
                            <div className="flex items-center gap-2 mb-3">
                              <Award
                                className="w-5 h-5"
                                style={{ color: '#7bb8ff' }}
                              />
                              <h3 className="text-[18px] font-bold text-gray-900">
                                任职要求
                              </h3>
                            </div>
                            <ul className="space-y-2 ml-2">
                              {requiredSkills.map((skill, index) => (
                                <li
                                  key={index}
                                  className="flex items-start gap-2 text-[15px] text-gray-700"
                                >
                                  <span
                                    className="inline-block w-1.5 h-1.5 rounded-full mt-2 flex-shrink-0"
                                    style={{ backgroundColor: '#7bb8ff' }}
                                  ></span>
                                  <span className="leading-relaxed">
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
                            <div className="flex items-center gap-2 mb-3">
                              <Award className="w-5 h-5 text-green-600" />
                              <h3 className="text-[18px] font-bold text-gray-900">
                                加分项
                              </h3>
                            </div>
                            <ul className="space-y-2 ml-2">
                              {bonusSkills.map((skill, index) => (
                                <li
                                  key={index}
                                  className="flex items-start gap-2 text-[15px] text-gray-700"
                                >
                                  <span
                                    className="inline-block w-1.5 h-1.5 rounded-full mt-2 flex-shrink-0"
                                    style={{ backgroundColor: '#7bb8ff' }}
                                  ></span>
                                  <span className="leading-relaxed">
                                    {skill.skill}
                                  </span>
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}
                      </div>
                    );
                  }

                  // 完全没有数据
                  return (
                    <div className="text-center py-12 text-gray-400">
                      <FileText className="w-12 h-12 mx-auto mb-3 opacity-50" />
                      <p className="text-[14px]">暂无岗位描述</p>
                    </div>
                  );
                })()}
              </div>
            ) : (
              /* 结构化展示模式 */
              <div className="space-y-5">
                {/* 岗位职责 */}
                {responsibilities.length > 0 && (
                  <div className="bg-white rounded-xl p-6 border border-gray-200">
                    <div className="flex items-center gap-2 mb-4">
                      <Target
                        className="w-5 h-5"
                        style={{ color: '#7bb8ff' }}
                      />
                      <h3 className="text-[18px] font-bold text-gray-900">
                        岗位职责
                      </h3>
                    </div>
                    <ul className="space-y-2.5">
                      {responsibilities.map((responsibility, index) => (
                        <li
                          key={index}
                          className="flex items-start gap-2 text-[15px] text-gray-700"
                        >
                          <span
                            className="inline-block w-1.5 h-1.5 rounded-full mt-2 flex-shrink-0"
                            style={{ backgroundColor: '#7bb8ff' }}
                          ></span>
                          <span className="leading-relaxed">
                            {responsibility}
                          </span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {/* 任职要求 */}
                {(requiredSkills.length > 0 || bonusSkills.length > 0) && (
                  <div className="bg-white rounded-xl p-6 border border-gray-200">
                    <div className="flex items-center gap-2 mb-4">
                      <Award className="w-5 h-5" style={{ color: '#7bb8ff' }} />
                      <h3 className="text-[18px] font-bold text-gray-900">
                        任职要求
                      </h3>
                    </div>
                    <div className="space-y-4">
                      {requiredSkills.length > 0 && (
                        <div>
                          <ul className="space-y-2">
                            {requiredSkills.map((skill, index) => (
                              <li
                                key={index}
                                className="flex items-start gap-2 text-[15px] text-gray-700"
                              >
                                <span
                                  className="inline-block w-1.5 h-1.5 rounded-full mt-2 flex-shrink-0"
                                  style={{ backgroundColor: '#7bb8ff' }}
                                ></span>
                                <span className="leading-relaxed">
                                  {skill.skill}
                                </span>
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {/* 加分项 */}
                {bonusSkills.length > 0 && (
                  <div className="bg-white rounded-xl p-6 border border-gray-200">
                    <div className="flex items-center gap-2 mb-4">
                      <Award className="w-5 h-5 text-green-600" />
                      <h3 className="text-[18px] font-bold text-gray-900">
                        加分项
                      </h3>
                    </div>
                    <ul className="space-y-2">
                      {bonusSkills.map((skill, index) => (
                        <li
                          key={index}
                          className="flex items-start gap-2 text-[15px] text-gray-700"
                        >
                          <span
                            className="inline-block w-1.5 h-1.5 rounded-full mt-2 flex-shrink-0"
                            style={{ backgroundColor: '#7bb8ff' }}
                          ></span>
                          <span className="leading-relaxed">{skill.skill}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
