import { useEffect, useState, useCallback } from 'react';
import {
  X,
  CheckCircle2,
  AlertCircle,
  User,
  Lightbulb,
  Award,
  BarChart3,
  Download,
  Share2,
  UserCircle,
  Briefcase,
  Clock,
  FileText,
  Check,
  GraduationCap,
} from 'lucide-react';
import { Dialog, DialogContent } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

import { getScreeningResult } from '@/services/screening';
import { ScreeningResult, MatchLevel } from '@/types/screening';
import { getResumeDetail } from '@/services/resume';
import { ResumeDetail as ResumeDetailType } from '@/types/resume';
import { getJobProfile } from '@/services/job-profile';

// 本地类型定义，避免使用 any
interface BasicInfoItem {
  name: string;
  value: string;
  matched: boolean;
}
interface BasicInfoSection {
  score: number;
  items?: BasicInfoItem[];
  explanation?: string;
}

interface EducationItem {
  matched: boolean;
  school: string;
  major: string;
  degree: string;
  start_time?: string;
  end_time?: string;
  description?: string;
}
interface EducationSection {
  score: number;
  items?: EducationItem[];
  explanation?: string;
}

interface ExperienceItem {
  matched: boolean;
  company: string;
  position: string;
  duration?: string;
  description?: string;
}
interface ExperienceSection {
  score: number;
  items?: ExperienceItem[];
  explanation?: string;
}

interface IndustryItem {
  matched: boolean;
  name: string;
  description?: string;
}
interface IndustrySection {
  score: number;
  items?: IndustryItem[];
  explanation?: string;
}

interface ResponsibilityItem {
  matched: boolean;
  name: string;
  description?: string;
}
interface ResponsibilitySection {
  score: number;
  items?: ResponsibilityItem[];
  explanation?: string;
}

interface SkillItem {
  matched: boolean;
  name: string;
  level?: string;
}
interface SkillSection {
  score: number;
  items?: SkillItem[];
  explanation?: string;
}

interface DetailedAnalysis {
  basic_info?: BasicInfoSection;
  education?: EducationSection;
  experience?: ExperienceSection;
  industry?: IndustrySection;
  responsibility?: ResponsibilitySection;
  skills?: SkillSection;
}

type ScreeningResultWithDetail = ScreeningResult & {
  detailed_analysis?: DetailedAnalysis;
};

interface ReportDetailModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  taskId: string | null;
  resumeId: string | null;
  resumeName?: string;
}

export function ReportDetailModal({
  open,
  onOpenChange,
  taskId,
  resumeId,
  resumeName,
}: ReportDetailModalProps) {
  const [result, setResult] = useState<ScreeningResult | null>(null);
  // removed unused state: loading
  // removed unused state: error
  const [resumeDetail, setResumeDetail] = useState<ResumeDetailType | null>(
    null
  );
  const [jobPositionName, setJobPositionName] = useState<string | null>(null);
  // removed unused state: activeTab
  const [detailActiveTab, setDetailActiveTab] = useState('基本信息');

  const loadReportDetail = useCallback(async () => {
    if (!taskId || !resumeId) return;

    try {
      const response = await getScreeningResult(taskId, resumeId);
      setResult(response.result);
    } catch (err) {
      console.error('加载报告详情失败:', err);
    }
  }, [taskId, resumeId]);

  // 加载报告详情
  useEffect(() => {
    if (open && taskId && resumeId) {
      loadReportDetail();
    }
  }, [open, taskId, resumeId, loadReportDetail]);

  // 加载简历详情
  useEffect(() => {
    if (open && resumeId) {
      (async () => {
        try {
          const detail = await getResumeDetail(resumeId);
          setResumeDetail(detail);
        } catch (e) {
          console.error('加载简历详情失败:', e);
        }
      })();
    } else {
      setResumeDetail(null);
    }
  }, [open, resumeId]);

  // 解析应聘岗位名称
  useEffect(() => {
    if (!result?.job_position_id) {
      setJobPositionName(null);
      return;
    }
    const jobId = result.job_position_id;

    const matchedFromResume = resumeDetail?.job_positions?.find(
      (jp) => jp.job_position_id === jobId
    );
    if (matchedFromResume?.job_title) {
      setJobPositionName(matchedFromResume.job_title);
      return;
    }

    (async () => {
      try {
        const job = await getJobProfile(jobId);
        setJobPositionName(job.name || null);
      } catch (e) {
        console.error('加载岗位名称失败:', e);
        setJobPositionName(null);
      }
    })();
  }, [result, resumeDetail]);

  // 进度轮询
  useEffect(() => {
    // 已移除进度轮询逻辑（不再维护本地进度状态）
    return () => {};
  }, [open, taskId, resumeId]);

  // 获取匹配等级的样式和文本
  const getMatchLevelInfo = (level: MatchLevel) => {
    switch (level) {
      case 'excellent':
        return {
          text: '非常匹配',
          icon: CheckCircle2,
          color: 'text-green-600',
          bgColor: 'bg-green-100',
        };
      case 'good':
        return {
          text: '高匹配',
          icon: CheckCircle2,
          color: 'text-blue-600',
          bgColor: 'bg-blue-100',
        };
      case 'fair':
        return {
          text: '一般匹配',
          icon: AlertCircle,
          color: 'text-orange-600',
          bgColor: 'bg-orange-100',
        };
      case 'poor':
        return {
          text: '低匹配',
          icon: AlertCircle,
          color: 'text-red-600',
          bgColor: 'bg-red-100',
        };
      default:
        return {
          text: '未知',
          icon: AlertCircle,
          color: 'text-gray-600',
          bgColor: 'bg-gray-100',
        };
    }
  };

  // removed unused function: getOverallScoreBg

  // 获取维度名称
  const getDimensionName = (key: string): string => {
    const dimensionMap: Record<string, string> = {
      basic_info: '基本信息',
      education: '教育背景',
      experience: '工作经验',
      industry: '行业背景',
      responsibility: '职责匹配',
      skill: '技能匹配',
    };
    return dimensionMap[key] || key;
  };

  // 渲染维度分数
  const renderDimensionScores = () => {
    if (!result?.dimension_scores) return null;

    return (
      <div className="grid grid-cols-2 gap-4">
        {Object.entries(result.dimension_scores).map(([key, score]) => (
          <div
            key={key}
            className="border border-gray-200 rounded-lg p-4 shadow-sm"
          >
            <div className="flex justify-between items-center mb-2">
              <span className="text-sm text-gray-600">
                {getDimensionName(key)}
              </span>
              <span className="text-lg font-bold text-[#36cfc9]">{score}</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-1.5">
              <div
                className={`h-1.5 rounded-full ${
                  score >= 85
                    ? 'bg-[#36cfc9]'
                    : score >= 70
                      ? 'bg-[#faad14]'
                      : score >= 55
                        ? 'bg-[#f5222d]'
                        : 'bg-gray-400'
                }`}
                style={{ width: `${score}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    );
  };

  // 渲染推荐建议
  const renderRecommendations = () => {
    if (!result?.recommendations || result.recommendations.length === 0)
      return null;

    return (
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
            <User className="w-4 h-4 text-blue-600" />
          </div>
          <span className="text-sm font-medium text-gray-900">专家建议</span>
        </div>
        <div className="space-y-3">
          {result.recommendations.map((rec, index) => (
            <div key={index} className="flex gap-3">
              <div className="w-5 h-5 bg-green-100 rounded-full flex items-center justify-center mt-0.5 flex-shrink-0">
                <CheckCircle2 className="w-3 h-3 text-green-600" />
              </div>
              <p className="text-sm text-gray-700 leading-relaxed">{rec}</p>
            </div>
          ))}
        </div>
      </div>
    );
  };

  // removed unused function: renderDetailedMatchInfo

  // 渲染基本信息详情 - 严格按照Figma设计
  const renderBasicInfoDetails = () => {
    const basicInfo = (result as ScreeningResultWithDetail | null)
      ?.detailed_analysis?.basic_info;
    if (!basicInfo) return <div className="text-gray-500">暂无数据</div>;

    return (
      <div
        className="bg-white rounded-xl shadow-sm border border-gray-100 p-6"
        style={{ boxShadow: '0px 2px 14px 0px #0000000f' }}
      >
        {/* 标题和分数 */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 bg-[#36cfc9] rounded-full flex items-center justify-center">
              <User className="w-4 h-4 text-white" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900">
              基本信息匹配详情
            </h3>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600">匹配分数:</span>
            <span
              className="text-xl font-bold"
              style={{
                color:
                  basicInfo.score >= 80
                    ? '#52c41a'
                    : basicInfo.score >= 60
                      ? '#faad14'
                      : '#f5222d',
              }}
            >
              {basicInfo.score}
            </span>
          </div>
        </div>

        {/* 匹配详情 */}
        <div className="bg-gray-50 rounded-lg p-4 mb-6">
          <h4 className="font-medium text-gray-900 mb-4">匹配详情</h4>
          <div className="space-y-4">
            {basicInfo.items?.map((item: BasicInfoItem, index: number) => (
              <div
                key={index}
                className="flex items-center justify-between py-3 border-b border-gray-200 last:border-b-0"
              >
                <span className="text-gray-700 font-medium">{item.name}</span>
                <div className="flex items-center gap-3">
                  <span className="text-gray-900">{item.value}</span>
                  <span
                    className="px-3 py-1 rounded-full text-sm font-medium"
                    style={{
                      backgroundColor: item.matched ? '#f6ffed' : '#fff1f0',
                      color: item.matched ? '#52c41a' : '#f5222d',
                    }}
                  >
                    {item.matched ? '匹配' : '不匹配'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* 详细说明 */}
        {basicInfo.explanation && (
          <div className="bg-blue-50 rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">详细说明</h4>
            <p className="text-gray-700 text-sm leading-relaxed">
              {basicInfo.explanation}
            </p>
          </div>
        )}
      </div>
    );
  };

  // 其他详情渲染函数（简化版本）
  // 渲染教育背景详情 - 严格按照Figma设计
  const renderEducationDetails = () => {
    const education = (result as ScreeningResultWithDetail | null)
      ?.detailed_analysis?.education;
    if (!education) return <div className="text-gray-500">暂无数据</div>;

    return (
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
        {/* 标题和分数 */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 bg-[#36cfc9] rounded-full flex items-center justify-center">
              <GraduationCap className="w-4 h-4 text-white" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900">
              教育背景匹配详情
            </h3>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600">匹配分数:</span>
            <span
              className="text-xl font-bold"
              style={{
                color:
                  education.score >= 80
                    ? '#52c41a'
                    : education.score >= 60
                      ? '#faad14'
                      : '#f5222d',
              }}
            >
              {education.score}
            </span>
          </div>
        </div>

        {/* 教育经历列表 */}
        <div className="space-y-4 mb-6">
          {education.items?.map((item: EducationItem, index: number) => (
            <div key={index} className="bg-gray-50 rounded-lg p-4">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <div
                    className="w-6 h-6 rounded-full flex items-center justify-center"
                    style={{
                      backgroundColor: item.matched ? '#f6ffed' : '#fff1f0',
                    }}
                  >
                    {item.matched ? (
                      <Check className="w-4 h-4" style={{ color: '#52c41a' }} />
                    ) : (
                      <X className="w-4 h-4" style={{ color: '#f5222d' }} />
                    )}
                  </div>
                  <div>
                    <div className="font-medium text-gray-900">
                      {item.school}
                    </div>
                    <div className="text-sm text-gray-600">
                      {item.major} · {item.degree}
                    </div>
                  </div>
                </div>
                <div
                  className="px-3 py-1 rounded-full text-sm font-medium"
                  style={{
                    backgroundColor: item.matched ? '#f6ffed' : '#fff1f0',
                    color: item.matched ? '#52c41a' : '#f5222d',
                  }}
                >
                  {item.matched ? '匹配' : '不匹配'}
                </div>
              </div>
              <div className="text-sm text-gray-600 mb-2">
                {item.start_time} - {item.end_time}
              </div>
              {item.description && (
                <div className="text-sm text-gray-700 bg-white rounded-lg p-3">
                  {item.description}
                </div>
              )}
            </div>
          ))}
        </div>

        {/* 匹配分析 */}
        {education.explanation && (
          <div className="bg-blue-50 rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">匹配分析</h4>
            <p className="text-gray-700 text-sm leading-relaxed">
              {education.explanation}
            </p>
          </div>
        )}
      </div>
    );
  };

  // 渲染工作经验详情 - 严格按照Figma设计
  const renderExperienceDetails = () => {
    const experience = (result as ScreeningResultWithDetail | null)
      ?.detailed_analysis?.experience;
    if (!experience) return <div className="text-gray-500">暂无数据</div>;

    return (
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
        {/* 标题和分数 */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 bg-[#36cfc9] rounded-full flex items-center justify-center">
              <Briefcase className="w-4 h-4 text-white" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900">
              工作经验匹配详情
            </h3>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600">匹配分数:</span>
            <span
              className={`text-xl font-bold ${
                experience.score >= 80
                  ? 'text-green-600'
                  : experience.score >= 60
                    ? 'text-yellow-600'
                    : 'text-red-600'
              }`}
            >
              {experience.score}
            </span>
          </div>
        </div>

        {/* 工作经历列表 */}
        <div className="space-y-4 mb-6">
          {experience.items?.map((item: ExperienceItem, index: number) => (
            <div key={index} className="bg-gray-50 rounded-lg p-4">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <div
                    className="w-6 h-6 rounded-full flex items-center justify-center"
                    style={{
                      backgroundColor: item.matched ? '#f6ffed' : '#fff1f0',
                    }}
                  >
                    {item.matched ? (
                      <Check className="w-4 h-4" style={{ color: '#52c41a' }} />
                    ) : (
                      <X className="w-4 h-4" style={{ color: '#f5222d' }} />
                    )}
                  </div>
                  <div>
                    <div className="font-medium text-gray-900">
                      {item.company}
                    </div>
                    <div className="text-sm text-gray-600">{item.position}</div>
                  </div>
                </div>
                <div
                  className={`px-3 py-1 rounded-full text-sm font-medium ${
                    item.matched
                      ? 'bg-green-100 text-green-800'
                      : 'bg-red-100 text-red-800'
                  }`}
                >
                  {item.matched ? '匹配' : '不匹配'}
                </div>
              </div>
              <div className="text-sm text-gray-600 mb-2">{item.duration}</div>
              <div className="text-sm text-gray-700 bg-white rounded-lg p-3 mb-3">
                {item.description}
              </div>
            </div>
          ))}
        </div>

        {/* 匹配分析 */}
        {experience.explanation && (
          <div className="bg-blue-50 rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">匹配分析</h4>
            <p className="text-gray-700 text-sm leading-relaxed">
              {experience.explanation}
            </p>
          </div>
        )}
      </div>
    );
  };

  // 渲染行业背景详情 - 严格按照Figma设计
  const renderIndustryDetails = () => {
    const industry = (result as ScreeningResultWithDetail | null)
      ?.detailed_analysis?.industry;
    if (!industry) return <div className="text-gray-500">暂无数据</div>;

    return (
      <div className="space-y-6">
        {/* 标题和分数 */}
        <div className="flex items-center justify-between">
          <h4 className="text-lg font-semibold text-gray-900">行业背景详情</h4>
          <div className="flex items-center space-x-2">
            <span className="text-sm text-gray-600">匹配分数：</span>
            <span
              className={`px-4 py-2 rounded-lg text-lg font-bold ${
                industry.score >= 80
                  ? 'bg-green-100 text-green-800'
                  : industry.score >= 60
                    ? 'bg-yellow-100 text-yellow-800'
                    : 'bg-red-100 text-red-800'
              }`}
            >
              {industry.score}分
            </span>
          </div>
        </div>

        {/* 行业项目列表 */}
        <div className="space-y-4">
          {industry.items?.map((item: IndustryItem, index: number) => (
            <div
              key={index}
              className="p-5 bg-white rounded-xl border border-gray-200 shadow-sm"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex-1">
                  <div className="text-lg font-semibold text-gray-900 mb-1">
                    {item.name}
                  </div>
                  <div className="text-sm text-gray-600 mb-2">
                    {item.description}
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <div
                    className={`px-3 py-1 rounded-lg text-sm font-medium ${
                      item.matched
                        ? 'bg-green-100 text-green-700'
                        : 'bg-red-100 text-red-700'
                    }`}
                  >
                    {item.matched ? '匹配' : '不匹配'}
                  </div>
                  {item.matched ? (
                    <div className="w-6 h-6 bg-green-100 rounded-full flex items-center justify-center">
                      <Check className="w-4 h-4 text-green-600" />
                    </div>
                  ) : (
                    <div className="w-6 h-6 bg-red-100 rounded-full flex items-center justify-center">
                      <X className="w-4 h-4 text-red-600" />
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* 匹配分析 */}
        {industry.explanation && (
          <div className="p-5 bg-blue-50 rounded-xl border border-blue-100">
            <div className="flex items-start space-x-3">
              <div className="w-5 h-5 bg-blue-100 rounded-full flex items-center justify-center mt-0.5">
                <svg
                  className="w-3 h-3 text-blue-600"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div>
                <div className="text-sm font-medium text-blue-900 mb-1">
                  匹配分析
                </div>
                <div className="text-sm text-blue-800">
                  {industry.explanation}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    );
  };

  // 渲染职责匹配详情 - 严格按照Figma设计
  const renderResponsibilityDetails = () => {
    const responsibility = (result as ScreeningResultWithDetail | null)
      ?.detailed_analysis?.responsibility;
    if (!responsibility) return <div className="text-gray-500">暂无数据</div>;

    return (
      <div className="space-y-6">
        {/* 标题和分数 */}
        <div className="flex items-center justify-between">
          <h4 className="text-lg font-semibold text-gray-900">职责匹配详情</h4>
          <div className="flex items-center space-x-2">
            <span className="text-sm text-gray-600">匹配分数：</span>
            <span
              className={`px-4 py-2 rounded-lg text-lg font-bold ${
                responsibility.score >= 80
                  ? 'bg-green-100 text-green-800'
                  : responsibility.score >= 60
                    ? 'bg-yellow-100 text-yellow-800'
                    : 'bg-red-100 text-red-800'
              }`}
            >
              {responsibility.score}分
            </span>
          </div>
        </div>

        {/* 职责项目列表 */}
        <div className="space-y-4">
          {responsibility.items?.map(
            (item: ResponsibilityItem, index: number) => (
              <div
                key={index}
                className="p-5 bg-white rounded-xl border border-gray-200 shadow-sm"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <div className="text-lg font-semibold text-gray-900 mb-1">
                      {item.name}
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <div
                      className={`px-3 py-1 rounded-lg text-sm font-medium ${
                        item.matched
                          ? 'bg-green-100 text-green-700'
                          : 'bg-red-100 text-red-700'
                      }`}
                    >
                      {item.matched ? '匹配' : '不匹配'}
                    </div>
                    {item.matched ? (
                      <div className="w-6 h-6 bg-green-100 rounded-full flex items-center justify-center">
                        <Check className="w-4 h-4 text-green-600" />
                      </div>
                    ) : (
                      <div className="w-6 h-6 bg-red-100 rounded-full flex items-center justify-center">
                        <X className="w-4 h-4 text-red-600" />
                      </div>
                    )}
                  </div>
                </div>
                <div className="text-sm text-gray-700 leading-relaxed">
                  {item.description}
                </div>
              </div>
            )
          )}
        </div>

        {/* 匹配分析 */}
        {responsibility.explanation && (
          <div className="p-5 bg-blue-50 rounded-xl border border-blue-100">
            <div className="flex items-start space-x-3">
              <div className="w-5 h-5 bg-blue-100 rounded-full flex items-center justify-center mt-0.5">
                <svg
                  className="w-3 h-3 text-blue-600"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div>
                <div className="text-sm font-medium text-blue-900 mb-1">
                  匹配分析
                </div>
                <div className="text-sm text-blue-800">
                  {responsibility.explanation}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    );
  };

  // 渲染技能匹配详情 - 严格按照Figma设计
  const renderSkillDetails = () => {
    const skills = (result as ScreeningResultWithDetail | null)
      ?.detailed_analysis?.skills;
    if (!skills) return <div className="text-gray-500">暂无数据</div>;

    return (
      <div className="space-y-6">
        {/* 标题和分数 */}
        <div className="flex items-center justify-between">
          <h4 className="text-lg font-semibold text-gray-900">技能匹配详情</h4>
          <div className="flex items-center space-x-2">
            <span className="text-sm text-gray-600">匹配分数：</span>
            <span
              className={`px-4 py-2 rounded-lg text-lg font-bold ${
                skills.score >= 80
                  ? 'bg-green-100 text-green-800'
                  : skills.score >= 60
                    ? 'bg-yellow-100 text-yellow-800'
                    : 'bg-red-100 text-red-800'
              }`}
            >
              {skills.score}分
            </span>
          </div>
        </div>

        {/* 技能项目列表 */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {skills.items?.map((item: SkillItem, index: number) => (
            <div
              key={index}
              className="p-5 bg-white rounded-xl border border-gray-200 shadow-sm"
            >
              <div className="flex items-center justify-between mb-3">
                <div className="text-lg font-semibold text-gray-900">
                  {item.name}
                </div>
                <div className="flex items-center space-x-2">
                  <div
                    className={`px-3 py-1 rounded-lg text-sm font-medium ${
                      item.matched
                        ? 'bg-green-100 text-green-700'
                        : 'bg-red-100 text-red-700'
                    }`}
                  >
                    {item.matched ? '匹配' : '不匹配'}
                  </div>
                  {item.matched ? (
                    <div className="w-6 h-6 bg-green-100 rounded-full flex items-center justify-center">
                      <Check className="w-4 h-4 text-green-600" />
                    </div>
                  ) : (
                    <div className="w-6 h-6 bg-red-100 rounded-full flex items-center justify-center">
                      <X className="w-4 h-4 text-red-600" />
                    </div>
                  )}
                </div>
              </div>
              <div className="text-sm text-gray-600">熟练度: {item.level}</div>
            </div>
          ))}
        </div>

        {/* 匹配分析 */}
        {skills.explanation && (
          <div className="p-5 bg-blue-50 rounded-xl border border-blue-100">
            <div className="flex items-start space-x-3">
              <div className="w-5 h-5 bg-blue-100 rounded-full flex items-center justify-center mt-0.5">
                <svg
                  className="w-3 h-3 text-blue-600"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div>
                <div className="text-sm font-medium text-blue-900 mb-1">
                  匹配分析
                </div>
                <div className="text-sm text-blue-800">
                  {skills.explanation}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    );
  };

  // 渲染头部区域
  const renderHeader = () => (
    <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-white">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 bg-[#36cfc9] rounded-full flex items-center justify-center">
          <BarChart3 className="w-4 h-4 text-white" />
        </div>
        <h2 className="text-xl font-semibold text-gray-900">报告详情</h2>
      </div>
      <div className="flex items-center gap-3">
        <Button
          variant="outline"
          size="sm"
          className="flex items-center gap-2 border-[#36cfc9] text-[#36cfc9] hover:bg-[#36cfc9] hover:text-white"
        >
          <Download className="w-4 h-4" />
          导出报告
        </Button>
        <Button
          size="sm"
          className="flex items-center gap-2 bg-[#36cfc9] text-white hover:bg-[#2db7b5]"
        >
          <Share2 className="w-4 h-4" />
          分享
        </Button>
        <button
          onClick={() => onOpenChange(false)}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <X className="w-5 h-5 text-gray-500" />
        </button>
      </div>
    </div>
  );

  // 渲染基本信息区域
  const renderBasicInfo = () => (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
      <div className="flex items-center gap-2 mb-6">
        <UserCircle className="w-5 h-5 text-[#36cfc9]" />
        <h3 className="text-lg font-semibold text-gray-900">基本信息</h3>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center">
            <User className="w-6 h-6 text-green-600" />
          </div>
          <div>
            <p className="text-sm text-gray-600">姓名</p>
            <p className="text-base font-medium text-gray-900">
              {resumeDetail?.name || '未知'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
            <Briefcase className="w-6 h-6 text-blue-600" />
          </div>
          <div>
            <p className="text-sm text-gray-600">应聘职位</p>
            <p className="text-base font-medium text-gray-900">
              {jobPositionName || '未知'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center">
            <Clock className="w-6 h-6 text-purple-600" />
          </div>
          <div>
            <p className="text-sm text-gray-600">报告生成时间</p>
            <p className="text-base font-medium text-gray-900">
              {result?.created_at
                ? new Date(result.created_at).toLocaleString('zh-CN')
                : '未知'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 bg-orange-100 rounded-full flex items-center justify-center">
            <FileText className="w-6 h-6 text-orange-600" />
          </div>
          <div>
            <p className="text-sm text-gray-600">简历版本</p>
            <p className="text-base font-medium text-gray-900">
              {resumeName || 'V1.0'}
            </p>
          </div>
        </div>
      </div>
    </div>
  );

  // 渲染综合分数区域
  const renderOverallScore = () => {
    if (!result?.overall_score) return null;
    const overallScore = result.overall_score;
    const matchLevelInfo = getMatchLevelInfo(result.match_level || 'fair');

    return (
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
        <div className="flex items-center gap-2 mb-6">
          <Award className="w-5 h-5 text-[#36cfc9]" />
          <h3 className="text-lg font-semibold text-gray-900">综合分数</h3>
        </div>
        <div className="flex flex-col lg:flex-row gap-8">
          {/* 左侧总分圆环 */}
          <div className="flex flex-col items-center">
            <div className="relative w-32 h-32">
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="text-center">
                  <div className="text-4xl font-bold text-[#36cfc9]">
                    {overallScore}
                  </div>
                  <div className="text-sm text-gray-600">综合评分</div>
                </div>
              </div>
              <svg
                className="w-32 h-32 transform -rotate-90"
                viewBox="0 0 120 120"
              >
                <circle
                  cx="60"
                  cy="60"
                  r="54"
                  fill="none"
                  stroke="#f3f4f6"
                  strokeWidth="8"
                />
                <circle
                  cx="60"
                  cy="60"
                  r="54"
                  fill="none"
                  stroke="#36cfc9"
                  strokeWidth="8"
                  strokeDasharray={`${(overallScore / 100) * 339.29} 339.29`}
                  strokeLinecap="round"
                />
              </svg>
            </div>
            <div
              className={`mt-4 px-3 py-1 rounded-full text-sm font-medium ${matchLevelInfo.bgColor} ${matchLevelInfo.color}`}
            >
              {matchLevelInfo.text}
            </div>
            <p className="mt-2 text-sm text-gray-600 text-center max-w-xs">
              {overallScore >= 85
                ? '该候选人综合能力优秀，匹配度高'
                : overallScore >= 70
                  ? '该候选人综合能力良好，匹配度较高'
                  : overallScore >= 55
                    ? '该候选人综合能力一般，需要进一步评估'
                    : '该候选人综合能力有待提升'}
            </p>
          </div>

          {/* 右侧各维度分数 */}
          <div className="flex-1">{renderDimensionScores()}</div>
        </div>
      </div>
    );
  };

  // 渲染推荐建议区域
  const renderRecommendationsSection = () => (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
      <div className="flex items-center gap-2 mb-6">
        <Lightbulb className="w-5 h-5 text-[#36cfc9]" />
        <h3 className="text-lg font-semibold text-gray-900">简历推荐建议</h3>
      </div>
      {renderRecommendations()}

      {/* 各版块匹配详情展示 */}
      {renderDetailedMatchTabs()}
    </div>
  );

  // removed unused function: renderDetailedMatchSection

  // 渲染详细匹配标签页导航
  const renderDetailedMatchTabs = () => {
    const tabs = [
      { key: 'basic', label: '基本信息' },
      { key: 'education', label: '教育背景' },
      { key: 'experience', label: '工作经验' },
      { key: 'industry', label: '行业背景' },
      { key: 'responsibility', label: '职责匹配' },
      { key: 'skill', label: '技能匹配' },
    ];

    return (
      <div className="mt-8">
        {/* 标签页导航 */}
        <div className="flex border-b border-[#f3f4f6]">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setDetailActiveTab(tab.key)}
              className={`px-6 py-4 text-base font-medium transition-colors ${
                detailActiveTab === tab.key
                  ? 'border-b-2 border-[#36cfc9] text-[#36cfc9]'
                  : 'text-[#6b7280]'
              }`}
              style={{
                fontFamily:
                  '"PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", SimHei, Arial, Helvetica, sans-serif',
              }}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* 标签页内容 */}
        <div className="mt-6">{renderDetailedMatchTabContent()}</div>
      </div>
    );
  };

  // 渲染详细匹配标签页内容
  const renderDetailedMatchTabContent = () => {
    switch (detailActiveTab) {
      case 'basic':
        return renderBasicInfoDetails();
      case 'education':
        return renderEducationDetails();
      case 'experience':
        return renderExperienceDetails();
      case 'industry':
        return renderIndustryDetails();
      case 'responsibility':
        return renderResponsibilityDetails();
      case 'skill':
        return renderSkillDetails();
      default:
        return renderBasicInfoDetails();
    }
  };

  // 主渲染
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-6xl max-h-[90vh] p-0 overflow-hidden">
        {/* 头部 */}
        {renderHeader()}

        {/* 内容区域 */}
        <div className="overflow-y-auto max-h-[calc(90vh-80px)]">
          <div className="p-6 space-y-6">
            {/* 基本信息 */}
            {renderBasicInfo()}

            {/* 综合分数 */}
            {result?.overall_score && renderOverallScore()}

            {/* 推荐建议 */}
            {result?.recommendations &&
              result.recommendations.length > 0 &&
              renderRecommendationsSection()}

            {/* 详细匹配信息通过“简历推荐建议”标签展示，已移除冗余调用 */}
          </div>

          {/* 底部版权信息 */}
          <div className="px-6 py-4 border-t border-gray-200 bg-gray-50">
            <p className="text-center text-sm text-gray-500">
              © 2023 简历分析报告系统. 保留所有权利.
            </p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
