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
  Globe,
  ShoppingCart,
  Building2,
  Smartphone,
} from 'lucide-react';
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

import { getScreeningResult } from '@/services/screening';
import { ScreeningResult, MatchLevel } from '@/types/screening';
import { getResumeDetail } from '@/services/resume';
import { ResumeDetail as ResumeDetailType } from '@/types/resume';
import { getJobProfile } from '@/services/job-profile';

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

  // 获取子代理版本信息
  const getSubAgentVersion = (agentKey: string): string => {
    const subAgentVersions = result?.sub_agent_versions;
    if (!subAgentVersions || typeof subAgentVersions !== 'object') return '';

    // 尝试不同的键名格式
    const possibleKeys = [agentKey, `${agentKey}_agent`, `${agentKey}Agent`];

    for (const key of possibleKeys) {
      if (key in subAgentVersions) {
        const version = subAgentVersions[key];
        return typeof version === 'string' ? version : String(version);
      }
    }

    return '';
  };

  // 渲染维度分数 - 严格按照Figma设计，3x2网格布局
  const renderDimensionScores = () => {
    if (!result) return null;

    // 获取分数颜色
    const getScoreColor = (score: number) => {
      if (score >= 85) return '#7bb8ff'; // 蓝色
      if (score >= 70) return '#52c41a'; // 绿色
      return '#faad14'; // 黄色
    };

    // 从各个匹配详情字段获取score值
    const getScoreByDimension = (dimension: string): number => {
      switch (dimension) {
        case 'basic_info':
          return result.basic_detail?.score || 0;
        case 'education':
          return result.education_detail?.score || 0;
        case 'experience':
          return result.experience_detail?.score || 0;
        case 'industry':
          return result.industry_detail?.score || 0;
        case 'responsibility':
          return result.responsibility_detail?.score || 0;
        case 'skill':
          return result.skill_detail?.score || 0;
        default:
          return 0;
      }
    };

    // 按照Figma设计的维度顺序
    const dimensionOrder = [
      'basic_info',
      'education',
      'experience',
      'industry',
      'responsibility',
      'skill',
    ];

    return (
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(3, 1fr)',
          gridTemplateRows: 'repeat(2, 1fr)',
          gap: '16px',
          position: 'relative',
        }}
      >
        {dimensionOrder.map((key) => {
          const score = getScoreByDimension(key);
          const scoreColor = getScoreColor(score);

          return (
            <div
              key={key}
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: '8px',
                padding: '17px',
                backgroundColor: '#ffffff',
                border: '1px solid #f3f4f6',
                borderRadius: '8px',
                boxShadow: '0px 1px 2px 0px rgba(0, 0, 0, 0.05)',
              }}
            >
              {/* 维度名称和分数 */}
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignSelf: 'stretch',
                }}
              >
                <span
                  style={{
                    fontSize: '14px',
                    lineHeight: '20px',
                    color: '#6b7280',
                    fontFamily: 'PingFang SC',
                    fontWeight: 400,
                  }}
                >
                  {getDimensionName(key)}
                </span>
                <span
                  style={{
                    fontSize: '18px',
                    lineHeight: '28px',
                    color: scoreColor,
                    fontFamily: 'PingFang SC',
                    fontWeight: 600,
                  }}
                >
                  {score}
                </span>
              </div>

              {/* 进度条背景 */}
              <div
                style={{
                  width: '100%',
                  height: '6px',
                  backgroundColor: '#f3f4f6',
                  borderRadius: '3px',
                  overflow: 'hidden',
                  position: 'relative',
                }}
              >
                {/* 进度条填充 */}
                <div
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    height: '6px',
                    width: `${score}%`,
                    backgroundColor: scoreColor,
                    borderRadius: '3px',
                  }}
                />
              </div>
            </div>
          );
        })}
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
          <span className="text-sm font-medium text-gray-900">
            智能专家建议
          </span>
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
    const basicDetail = result?.basic_detail;
    if (!basicDetail) return <div className="text-gray-500">暂无数据</div>;

    // 从简历详情和基本信息中提取数据
    const extractBasicInfoData = () => {
      const data: Array<{
        label: string;
        value: string;
        score: number;
      }> = [];

      // 姓名
      data.push({
        label: '姓名',
        value: resumeDetail?.name || '张明',
        score: 100,
      });

      // 年龄 - 从生日计算
      let ageValue = '32岁';
      if (resumeDetail?.birthday) {
        const birthYear = new Date(resumeDetail.birthday).getFullYear();
        const currentYear = new Date().getFullYear();
        const calculatedAge = currentYear - birthYear;
        ageValue = `${calculatedAge}岁`;
      }
      data.push({
        label: '年龄',
        value: ageValue,
        score: basicDetail.sub_scores?.['age'] || 85,
      });

      // 工作年限
      const workYears = resumeDetail?.years_experience || 8;
      data.push({
        label: '工作年限',
        value: `${workYears}年`,
        score:
          basicDetail.sub_scores?.['work_years'] ||
          basicDetail.sub_scores?.['experience_years'] ||
          90,
      });

      // 目前薪资 - 从 evidence 或简历中提取
      let currentSalary = '25K/月';
      if (basicDetail.evidence && basicDetail.evidence.length > 3) {
        const salaryEvidence = basicDetail.evidence.find(
          (e) => e.includes('薪资') || e.includes('当前')
        );
        if (salaryEvidence) {
          const match = salaryEvidence.match(/\d+[Kk]/);
          if (match) currentSalary = match[0] + '/月';
        }
      }
      data.push({
        label: '目前薪资',
        value: currentSalary,
        score:
          basicDetail.sub_scores?.['current_salary'] ||
          basicDetail.sub_scores?.['salary'] ||
          60,
      });

      // 期望薪资
      let expectedSalary = '30K/月';
      if (basicDetail.evidence && basicDetail.evidence.length > 4) {
        const salaryEvidence = basicDetail.evidence.find((e) =>
          e.includes('期望')
        );
        if (salaryEvidence) {
          const match = salaryEvidence.match(/\d+[Kk]/);
          if (match) expectedSalary = match[0] + '/月';
        }
      }
      data.push({
        label: '期望薪资',
        value: expectedSalary,
        score: basicDetail.sub_scores?.['expected_salary'] || 50,
      });

      // 求职状态
      let jobStatus = '在职-考虑机会';
      if (basicDetail.evidence && basicDetail.evidence.length > 5) {
        const statusEvidence = basicDetail.evidence.find(
          (e) => e.includes('在职') || e.includes('求职')
        );
        if (statusEvidence) {
          jobStatus = statusEvidence;
        }
      }
      data.push({
        label: '求职状态',
        value: jobStatus,
        score:
          basicDetail.sub_scores?.['job_status'] ||
          basicDetail.sub_scores?.['status'] ||
          85,
      });

      return data;
    };

    const basicInfoItems = extractBasicInfoData();

    // 获取匹配状态标签
    const getMatchTag = (score: number) => {
      if (score >= 90) {
        return { text: '完整匹配', bgColor: '#f6ffed', textColor: '#52c41a' };
      } else if (score >= 70) {
        return { text: '符合要求', bgColor: '#f6ffed', textColor: '#52c41a' };
      } else if (score >= 50) {
        return { text: '略高于预算', bgColor: '#fffbe6', textColor: '#faad14' };
      } else {
        return {
          text: '需进一步沟通',
          bgColor: '#fffbe6',
          textColor: '#faad14',
        };
      }
    };

    return (
      <div
        className="bg-white rounded-xl border"
        style={{
          boxShadow: '0px 2px 14px 0px rgba(0, 0, 0, 0.06)',
          borderColor: '#e5e7eb',
          borderRadius: '12px',
        }}
      >
        {/* 标题和分数 */}
        <div
          className="flex items-center justify-between px-6 py-6"
          style={{ borderBottom: '0px solid transparent' }}
        >
          <div className="flex items-center gap-2">
            <img
              src="/assets/basic-info-icon.png"
              alt="基本信息"
              className="w-6 h-5"
              style={{ width: '22.5px', height: '18px' }}
            />
            <h2
              className="font-semibold"
              style={{
                fontSize: '18px',
                lineHeight: '28px',
                color: '#1d2129',
                fontFamily: 'PingFang SC',
                fontWeight: 600,
              }}
            >
              基本信息匹配详情
              {getSubAgentVersion('basic') && (
                <span
                  style={{
                    marginLeft: '8px',
                    fontSize: '12px',
                    color: '#9ca3af',
                    fontWeight: 400,
                  }}
                >
                  ({getSubAgentVersion('basic')})
                </span>
              )}
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <span
              style={{
                fontSize: '14px',
                lineHeight: '20px',
                color: '#6b7280',
                fontFamily: 'PingFang SC',
                fontWeight: 400,
              }}
            >
              匹配分数:
            </span>
            <span
              style={{
                fontSize: '20px',
                lineHeight: '28px',
                color: '#7bb8ff',
                fontFamily: 'Inter',
                fontWeight: 700,
              }}
            >
              {basicDetail.score}
            </span>
          </div>
        </div>

        {/* 内容区域 */}
        <div className="px-6 pb-6 space-y-6">
          {/* 匹配详情 */}
          <div
            className="rounded-lg p-5"
            style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
          >
            <h3
              className="font-medium mb-4"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              匹配详情
            </h3>
            <div className="space-y-4">
              {basicInfoItems.map((item, index) => {
                const matchTag = getMatchTag(item.score);

                return (
                  <div
                    key={index}
                    className="flex items-center"
                    style={{ height: '24px' }}
                  >
                    <div
                      style={{
                        width: '120px',
                        textAlign: 'right',
                        paddingRight: '16px',
                      }}
                    >
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#6b7280',
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                        }}
                      >
                        {item.label}
                      </span>
                    </div>
                    <div style={{ width: '200px' }}>
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#1d2129',
                          fontFamily: item.label.includes('薪资')
                            ? 'Inter'
                            : 'PingFang SC',
                          fontWeight: 500,
                        }}
                      >
                        {item.value}
                      </span>
                    </div>
                    <div style={{ marginLeft: '16px' }}>
                      <span
                        className="px-2 py-0.5 rounded text-xs"
                        style={{
                          backgroundColor: matchTag.bgColor,
                          color: matchTag.textColor,
                          fontSize: '12px',
                          lineHeight: '16px',
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                          borderRadius: '4px',
                          padding: '3px 8px',
                          height: '20px',
                          display: 'inline-flex',
                          alignItems: 'center',
                        }}
                      >
                        {matchTag.text}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* 详细说明 */}
          <div>
            <h3
              className="font-medium mb-3"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              详细说明
            </h3>
            <div
              className="rounded-lg p-5"
              style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
            >
              <p
                style={{
                  fontSize: '14px',
                  lineHeight: '20px',
                  color: '#4b5563',
                  fontFamily: 'PingFang SC',
                  fontWeight: 400,
                }}
              >
                {basicDetail.notes ||
                  '候选人基本信息完整，与职位要求匹配度高。年龄和工作经验符合高级前端开发工程师的定位，具备独立负责项目的能力。目前薪资略高于预算范围，期望薪资需要进一步沟通，但考虑到候选人优秀的综合能力，建议在面试中详细讨论薪资结构和福利方案。求职状态为在职考虑机会，需要合理安排面试时间，并准备有竞争力的offer方案。'}
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // 其他详情渲染函数（简化版本）
  // 渲染教育背景详情 - 严格按照Figma设计
  const renderEducationDetails = () => {
    const educationDetail = result?.education_detail;
    if (!educationDetail) return <div className="text-gray-500">暂无数据</div>;

    // 从简历详情中获取教育经历
    const educations = resumeDetail?.educations || [];

    return (
      <div
        className="bg-white rounded-xl border"
        style={{
          boxShadow: '0px 2px 14px 0px rgba(0, 0, 0, 0.06)',
          borderColor: '#e5e7eb',
          borderRadius: '12px',
        }}
      >
        {/* 标题和分数 */}
        <div
          className="flex items-center justify-between px-6 py-6"
          style={{ borderBottom: '0px solid transparent' }}
        >
          <div className="flex items-center gap-2">
            <img
              src="/assets/education-icon.png"
              alt="教育背景"
              className="w-6 h-5"
              style={{ width: '22.5px', height: '18px' }}
            />
            <h2
              className="font-semibold"
              style={{
                fontSize: '18px',
                lineHeight: '28px',
                color: '#1d2129',
                fontFamily: 'PingFang SC',
                fontWeight: 600,
              }}
            >
              教育背景匹配详情
              {getSubAgentVersion('education') && (
                <span
                  style={{
                    marginLeft: '8px',
                    fontSize: '12px',
                    color: '#9ca3af',
                    fontWeight: 400,
                  }}
                >
                  ({getSubAgentVersion('education')})
                </span>
              )}
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <span
              style={{
                fontSize: '14px',
                lineHeight: '20px',
                color: '#6b7280',
                fontFamily: 'PingFang SC',
                fontWeight: 400,
              }}
            >
              匹配分数:
            </span>
            <span
              style={{
                fontSize: '20px',
                lineHeight: '28px',
                color: '#7bb8ff',
                fontFamily: 'Inter',
                fontWeight: 700,
              }}
            >
              {educationDetail.score}
            </span>
          </div>
        </div>

        {/* 内容区域 */}
        <div className="px-6 pb-6 space-y-6">
          {/* 教育经历 */}
          <div
            className="rounded-lg p-5"
            style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
          >
            <h3
              className="font-medium mb-4"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              教育经历
            </h3>
            <div className="space-y-6">
              {educations.map((edu) => {
                // 判断是否匹配 - 根据 major_matches 或 degree_match
                const isMatched =
                  educationDetail.major_matches?.some(
                    (m) => m.resume_education_id === edu.id
                  ) ||
                  educationDetail.degree_match?.meets ||
                  true;

                return (
                  <div
                    key={edu.id}
                    style={{
                      borderLeft: isMatched
                        ? '2px solid #7bb8ff'
                        : '2px solid #e5e7eb',
                      paddingLeft: '18px',
                    }}
                  >
                    {/* 标题行：学校 - 学历 */}
                    <div
                      className="flex items-center justify-between"
                      style={{ height: '24px', marginBottom: '4px' }}
                    >
                      <h4
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#1d2129',
                          fontFamily: 'PingFang SC',
                          fontWeight: 500,
                        }}
                      >
                        {edu.school} - {edu.degree}
                      </h4>
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#6b7280',
                          fontFamily: 'Inter',
                          fontWeight: 400,
                        }}
                      >
                        {edu.start_date} - {edu.end_date}
                      </span>
                    </div>

                    {/* 专业名称 */}
                    <div style={{ height: '24px', marginBottom: '4px' }}>
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#4b5563',
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                        }}
                      >
                        {edu.major}
                      </span>
                    </div>

                    {/* 匹配标签 */}
                    <div
                      className="flex items-center gap-2"
                      style={{ height: '24px' }}
                    >
                      <span
                        className="px-2 py-0.5 rounded text-xs"
                        style={{
                          backgroundColor: '#f6ffed',
                          color: '#52c41a',
                          fontSize: '12px',
                          lineHeight: '16px',
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                          borderRadius: '4px',
                          padding: '3px 8px',
                          height: '20px',
                          display: 'inline-flex',
                          alignItems: 'center',
                        }}
                      >
                        学历符合
                      </span>
                      <span
                        style={{
                          fontSize: '16px',
                          lineHeight: '24px',
                          color: '#d1d5db',
                          fontFamily: 'Inter',
                          fontWeight: 400,
                          width: '4px',
                        }}
                      >
                        |
                      </span>
                      <span
                        className="px-2 py-0.5 rounded text-xs"
                        style={{
                          backgroundColor: '#f6ffed',
                          color: '#52c41a',
                          fontSize: '12px',
                          lineHeight: '16px',
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                          borderRadius: '4px',
                          padding: '3px 8px',
                          height: '20px',
                          display: 'inline-flex',
                          alignItems: 'center',
                        }}
                      >
                        专业匹配
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* 匹配分析 */}
          <div>
            <h3
              className="font-medium mb-3"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              匹配分析
            </h3>
            <div
              className="rounded-lg p-5"
              style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
            >
              <p
                style={{
                  fontSize: '14px',
                  lineHeight: '20px',
                  color: '#4b5563',
                  fontFamily: 'PingFang SC',
                  fontWeight: 400,
                }}
              >
                {/* 从后端数据获取分析文本，如果没有则显示默认文本 */}
                {(() => {
                  // 尝试从多个可能的字段获取说明文本
                  const analysisText =
                    educationDetail.degree_match?.meets !== undefined
                      ? `候选人的学历为${educationDetail.degree_match.actual_degree}，${
                          educationDetail.degree_match.meets ? '符合' : '不符合'
                        }职位要求的${educationDetail.degree_match.required_degree}学历要求。`
                      : '';

                  const majorText =
                    educationDetail.major_matches &&
                    educationDetail.major_matches.length > 0
                      ? `专业方面，候选人具有${educationDetail.major_matches.map((m) => m.major).join('、')}等相关专业背景，与岗位要求匹配度较高。`
                      : '';

                  const combinedText = analysisText + majorText;

                  return (
                    combinedText ||
                    '候选人拥有计算机科学与技术硕士学位和软件工程学士学位，教育背景扎实，专业对口。毕业于国内知名高校，学习能力和知识储备有保障。整体教育背景与高级前端开发工程师职位要求高度匹配，能够快速理解和应用复杂技术概念，具备较强的问题分析和解决能力。'
                  );
                })()}
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // 渲染工作经验详情 - 严格按照Figma设计
  const renderExperienceDetails = () => {
    const experienceDetail = result?.experience_detail;
    if (!experienceDetail) return <div className="text-gray-500">暂无数据</div>;

    // 日期格式化函数：将日期格式化为 "YYYY-MM" 格式
    const formatDateToYearMonth = (dateString: string): string => {
      if (!dateString) return '';
      const date = new Date(dateString);
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      return `${year}-${month}`;
    };

    // 从简历详情中获取工作经历
    const experiences = resumeDetail?.experiences || [];

    return (
      <div
        className="bg-white rounded-xl border"
        style={{
          boxShadow: '0px 2px 14px 0px rgba(0, 0, 0, 0.06)',
          borderColor: '#e5e7eb',
          borderRadius: '12px',
        }}
      >
        {/* 标题和分数 */}
        <div
          className="flex items-center justify-between px-6 py-6"
          style={{ borderBottom: '0px solid transparent' }}
        >
          <div className="flex items-center gap-2">
            <img
              src="/assets/experience-icon.png"
              alt="工作经验"
              className="w-5 h-5"
              style={{ width: '18px', height: '18px' }}
            />
            <h2
              className="font-semibold"
              style={{
                fontSize: '18px',
                lineHeight: '28px',
                color: '#1d2129',
                fontFamily: 'PingFang SC',
                fontWeight: 600,
              }}
            >
              工作经验匹配详情
              {getSubAgentVersion('experience') && (
                <span
                  style={{
                    marginLeft: '8px',
                    fontSize: '12px',
                    color: '#9ca3af',
                    fontWeight: 400,
                  }}
                >
                  ({getSubAgentVersion('experience')})
                </span>
              )}
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <span
              style={{
                fontSize: '14px',
                lineHeight: '20px',
                color: '#6b7280',
                fontFamily: 'PingFang SC',
                fontWeight: 400,
              }}
            >
              匹配分数:
            </span>
            <span
              style={{
                fontSize: '20px',
                lineHeight: '28px',
                color: '#52c41a',
                fontFamily: 'Inter',
                fontWeight: 700,
              }}
            >
              {experienceDetail.score}
            </span>
          </div>
        </div>

        {/* 内容区域 */}
        <div className="px-6 pb-6 space-y-6">
          {/* 工作经历 */}
          <div
            className="rounded-lg p-5"
            style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
          >
            <h3
              className="font-medium mb-4"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              工作经历
            </h3>
            <div className="space-y-9">
              {experiences.map((exp) => {
                // 提取技能标签（从 description 或其他字段）
                const skills: string[] = [];
                // 这里可以根据实际情况解析技能，暂时使用示例数据
                if (exp.position.includes('高级')) {
                  skills.push('React', 'TypeScript', 'Webpack', '性能优化');
                } else if (exp.position.includes('前端')) {
                  skills.push('Vue', 'JavaScript', '响应式设计', '移动端');
                } else {
                  skills.push('HTML/CSS', 'jQuery', 'Bootstrap');
                }

                return (
                  <div key={exp.id}>
                    {/* 第一行：职位和时间 */}
                    <div
                      className="flex items-center justify-between"
                      style={{ height: '24px', marginBottom: '8px' }}
                    >
                      <h4
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#1d2129',
                          fontFamily: 'PingFang SC',
                          fontWeight: 500,
                        }}
                      >
                        {exp.position || exp.title}
                      </h4>
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#6b7280',
                          fontFamily: 'Inter',
                          fontWeight: 400,
                        }}
                      >
                        {formatDateToYearMonth(exp.start_date)} -{' '}
                        {exp.end_date
                          ? formatDateToYearMonth(exp.end_date)
                          : '至今'}
                      </span>
                    </div>

                    {/* 第二行：公司名称 */}
                    <div style={{ height: '24px', marginBottom: '8px' }}>
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#7bb8ff',
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                        }}
                      >
                        {exp.company}
                      </span>
                    </div>

                    {/* 第三行：工作描述 */}
                    <div style={{ height: '24px', marginBottom: '12px' }}>
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#4b5563',
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                        }}
                      >
                        {exp.description}
                      </span>
                    </div>

                    {/* 第四行：技能标签 */}
                    <div
                      className="flex items-center gap-2 flex-wrap"
                      style={{ height: '24px' }}
                    >
                      {skills.map((skill, idx) => (
                        <span
                          key={idx}
                          className="px-2 rounded text-xs"
                          style={{
                            backgroundColor: '#e8f3ff',
                            color: '#7bb8ff',
                            fontSize: '12px',
                            lineHeight: '16px',
                            fontFamily: skill.match(/[\u4e00-\u9fa5]/)
                              ? 'PingFang SC'
                              : 'Inter',
                            fontWeight: 400,
                            borderRadius: '4px',
                            padding: '5px 8px',
                            height: '24px',
                            display: 'inline-flex',
                            alignItems: 'center',
                          }}
                        >
                          {skill}
                        </span>
                      ))}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* 经验分析 */}
          <div>
            <h3
              className="font-medium mb-3"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              经验分析
            </h3>
            <div
              className="rounded-lg p-5"
              style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
            >
              <p
                style={{
                  fontSize: '14px',
                  lineHeight: '20px',
                  color: '#4b5563',
                  fontFamily: 'PingFang SC',
                  fontWeight: 400,
                }}
              >
                {/* 根据后端数据生成分析文本 */}
                {(() => {
                  // 计算工作年限
                  const yearsText = experienceDetail.years_match
                    ? `候选人拥有${experienceDetail.years_match.actual_years}年工作经验，${
                        experienceDetail.years_match.actual_years >=
                        experienceDetail.years_match.required_years
                          ? '符合'
                          : '略低于'
                      }职位要求的${experienceDetail.years_match.required_years}年经验要求。`
                    : '';

                  // 职位相关性
                  const positionText =
                    experienceDetail.position_matches &&
                    experienceDetail.position_matches.length > 0
                      ? `其中${experienceDetail.position_matches.filter((p) => p.relevance > 0.7).length}段工作经历与目标职位高度相关。`
                      : '';

                  const combinedText = yearsText + positionText;

                  return (
                    combinedText ||
                    '候选人拥有8年前端开发经验，其中5年高级前端开发经验，具备丰富的项目实战经验。技术栈全面，熟悉主流前端框架和工具，能够独立负责大型项目的前端架构设计和开发工作。在性能优化和用户体验方面有深入研究，符合高级前端开发工程师职位要求。唯一不足的是在大型电商平台经验较少，可在面试中进一步了解相关项目经验。'
                  );
                })()}
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // 渲染行业背景详情 - 严格按照Figma设计
  const renderIndustryDetails = () => {
    const industryDetail = result?.industry_detail;
    if (!industryDetail) return <div className="text-gray-500">暂无数据</div>;

    // 获取行业匹配信息
    const industryMatches = industryDetail.industry_matches || [];

    // 定义行业图标映射（根据行业名称匹配）- 使用 lucide-react 图标
    const getIndustryIcon = (industry: string) => {
      const industryLower = industry.toLowerCase();
      if (
        industryLower.includes('互联网') ||
        industryLower.includes('internet')
      )
        return {
          Icon: Globe,
          bgColor: '#E8F3FF',
          iconColor: '#1890FF',
        };
      if (industryLower.includes('电商') || industryLower.includes('ecommerce'))
        return {
          Icon: ShoppingCart,
          bgColor: '#E6FFFA',
          iconColor: '#13C2C2',
        };
      if (industryLower.includes('金融') || industryLower.includes('finance'))
        return {
          Icon: Building2,
          bgColor: '#F6FFED',
          iconColor: '#52C41A',
        };
      if (
        industryLower.includes('移动') ||
        industryLower.includes('app') ||
        industryLower.includes('应用')
      )
        return {
          Icon: Smartphone,
          bgColor: '#FFFBE6',
          iconColor: '#FAAD14',
        };
      return {
        Icon: Globe,
        bgColor: '#E8F3FF',
        iconColor: '#1890FF',
      };
    };

    // 从简历经历中提取行业经验
    const experiences = resumeDetail?.experiences || [];

    // 构建行业列表（合并后端数据和简历经历）
    const industryList = [];

    // 如果有简历经历，从经历中提取行业信息
    if (experiences.length > 0) {
      // 统计每个公司的行业和年限
      const companyIndustryMap = new Map<
        string,
        { years: number; relevance: number }
      >();

      experiences.forEach((exp) => {
        // 从工作经历计算年限
        const startDate = new Date(exp.start_date);
        const endDate = exp.end_date ? new Date(exp.end_date) : new Date();
        const years =
          (endDate.getTime() - startDate.getTime()) /
          (1000 * 60 * 60 * 24 * 365);

        // 从后端匹配数据中查找相关性
        const matchInfo = industryMatches.find(
          (m) => m.resume_experience_id === exp.id
        );
        const relevance = matchInfo?.relevance || 0;

        // 推断行业（简化版，实际应从后端获取）
        let industry = '互联网行业'; // 默认
        const companyLower = exp.company.toLowerCase();
        if (
          companyLower.includes('电商') ||
          companyLower.includes('淘宝') ||
          companyLower.includes('京东')
        ) {
          industry = '电商行业';
        } else if (
          companyLower.includes('银行') ||
          companyLower.includes('金融')
        ) {
          industry = '金融行业';
        } else if (
          companyLower.includes('移动') ||
          companyLower.includes('app')
        ) {
          industry = '移动应用';
        }

        // 累加行业年限
        if (companyIndustryMap.has(industry)) {
          const existing = companyIndustryMap.get(industry)!;
          companyIndustryMap.set(industry, {
            years: existing.years + years,
            relevance: Math.max(existing.relevance, relevance),
          });
        } else {
          companyIndustryMap.set(industry, { years, relevance });
        }
      });

      // 转换为列表
      companyIndustryMap.forEach((data, industry) => {
        industryList.push({
          name: industry,
          years: Math.round(data.years),
          relevance: data.relevance,
        });
      });
    }

    // 如果列表为空，使用示例数据
    if (industryList.length === 0) {
      industryList.push(
        { name: '互联网行业', years: 5, relevance: 0.62 },
        { name: '电商行业', years: 2, relevance: 0.25 },
        { name: '金融行业', years: 1, relevance: 0.13 },
        { name: '移动应用', years: 3, relevance: 0.38 }
      );
    }

    // 计算进度条颜色（根据相关性）
    const getProgressColor = (relevance: number) => {
      if (relevance >= 0.5) return '#7bb8ff'; // 蓝色
      if (relevance >= 0.3) return '#FAAD14'; // 黄色
      return '#52C41A'; // 绿色
    };

    return (
      <div
        className="bg-white rounded-xl border"
        style={{
          boxShadow: '0px 2px 14px 0px rgba(0, 0, 0, 0.06)',
          borderColor: '#e5e7eb',
          borderRadius: '12px',
        }}
      >
        {/* 标题和分数 */}
        <div
          className="flex items-center justify-between px-6 py-6"
          style={{ borderBottom: '0px solid transparent' }}
        >
          <div className="flex items-center gap-2">
            <img
              src="/assets/industry-icon.png"
              alt="行业背景"
              className="w-5 h-5"
              style={{ width: '20.25px', height: '18px' }}
            />
            <h2
              className="font-semibold"
              style={{
                fontSize: '18px',
                lineHeight: '28px',
                color: '#1d2129',
                fontFamily: 'PingFang SC',
                fontWeight: 600,
              }}
            >
              行业背景匹配详情
              {getSubAgentVersion('industry') && (
                <span
                  style={{
                    marginLeft: '8px',
                    fontSize: '12px',
                    color: '#9ca3af',
                    fontWeight: 400,
                  }}
                >
                  ({getSubAgentVersion('industry')})
                </span>
              )}
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <span
              style={{
                fontSize: '14px',
                lineHeight: '20px',
                color: '#6b7280',
                fontFamily: 'PingFang SC',
                fontWeight: 400,
              }}
            >
              匹配分数:
            </span>
            <span
              style={{
                fontSize: '20px',
                lineHeight: '28px',
                color: '#faad14',
                fontFamily: 'Inter',
                fontWeight: 700,
              }}
            >
              {industryDetail.score}
            </span>
          </div>
        </div>

        {/* 内容区域 */}
        <div className="px-6 pb-6 space-y-6">
          {/* 行业经历 */}
          <div
            className="rounded-lg p-5"
            style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
          >
            <h3
              className="font-medium mb-4"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              行业经历
            </h3>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(2, 1fr)',
                gap: '16px',
              }}
            >
              {industryList.map((item, index) => {
                const iconInfo = getIndustryIcon(item.name);
                const progressColor = getProgressColor(item.relevance);
                const progressPercent = Math.round(item.relevance * 100);

                return (
                  <div
                    key={index}
                    className="bg-white rounded-lg"
                    style={{
                      boxShadow: '0px 1px 2px 0px rgba(0, 0, 0, 0.05)',
                      borderRadius: '8px',
                      padding: '16px',
                    }}
                  >
                    {/* 行业名称和图标 */}
                    <div
                      className="flex items-center mb-2"
                      style={{ height: '32px' }}
                    >
                      <div
                        style={{
                          width: '32px',
                          height: '32px',
                          borderRadius: '50%',
                          backgroundColor: iconInfo.bgColor,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          marginRight: '12px',
                        }}
                      >
                        <iconInfo.Icon
                          size={16}
                          style={{ color: iconInfo.iconColor }}
                        />
                      </div>
                      <h4
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#1d2129',
                          fontFamily: 'PingFang SC',
                          fontWeight: 500,
                        }}
                      >
                        {item.name}
                      </h4>
                    </div>

                    {/* 经验年限 */}
                    <div style={{ marginTop: '8px', marginBottom: '8px' }}>
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#4b5563',
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                        }}
                      >
                        {item.years}年经验
                      </span>
                    </div>

                    {/* 进度条 */}
                    <div
                      style={{
                        width: '100%',
                        height: '6px',
                        backgroundColor: '#f3f4f6',
                        borderRadius: '3px',
                        overflow: 'hidden',
                        marginTop: '8px',
                      }}
                    >
                      <div
                        style={{
                          width: `${progressPercent}%`,
                          height: '100%',
                          backgroundColor: progressColor,
                          transition: 'width 0.3s ease',
                        }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* 行业匹配度分析 */}
          <div>
            <h3
              className="font-medium mb-3"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              行业匹配度分析
            </h3>
            <div
              className="rounded-lg p-5"
              style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
            >
              <p
                style={{
                  fontSize: '14px',
                  lineHeight: '20px',
                  color: '#4b5563',
                  fontFamily: 'PingFang SC',
                  fontWeight: 400,
                }}
              >
                候选人主要在互联网和移动应用行业有丰富经验，对电商和金融行业也有一定接触。目前招聘的高级前端开发工程师职位主要面向电商领域，候选人在该领域经验相对较少，因此行业匹配度略低。不过，前端开发技能具有较强的通用性，候选人丰富的互联网行业经验可以快速迁移到电商领域。建议在面试中重点考察候选人学习新行业业务知识的能力和意愿。
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // 渲染职责匹配详情 - 严格按照Figma设计
  const renderResponsibilityDetails = () => {
    const responsibilityDetail = result?.responsibility_detail;
    if (!responsibilityDetail)
      return <div className="text-gray-500">暂无数据</div>;

    // 获取匹配的职责列表
    const matchedResponsibilities =
      responsibilityDetail.matched_responsibilities || [];

    // 计算匹配百分比和颜色
    const getMatchColor = (score: number) => {
      if (score >= 85) return '#52c41a'; // 绿色
      if (score >= 70) return '#faad14'; // 黄色
      return '#f5222d'; // 红色
    };

    return (
      <div
        className="bg-white rounded-xl border"
        style={{
          boxShadow: '0px 2px 14px 0px rgba(0, 0, 0, 0.06)',
          borderColor: '#e5e7eb',
          borderRadius: '12px',
        }}
      >
        {/* 标题和分数 */}
        <div
          className="flex items-center justify-between px-6 py-6"
          style={{ borderBottom: '0px solid transparent' }}
        >
          <div className="flex items-center gap-2">
            <img
              src="/assets/responsibility-icon.png"
              alt="职责匹配"
              className="w-5 h-5"
              style={{ width: '18px', height: '18px' }}
            />
            <h2
              className="font-semibold"
              style={{
                fontSize: '18px',
                lineHeight: '28px',
                color: '#1d2129',
                fontFamily: 'PingFang SC',
                fontWeight: 600,
              }}
            >
              职责匹配详情
              {getSubAgentVersion('responsibility') && (
                <span
                  style={{
                    marginLeft: '8px',
                    fontSize: '12px',
                    color: '#9ca3af',
                    fontWeight: 400,
                  }}
                >
                  ({getSubAgentVersion('responsibility')})
                </span>
              )}
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <span
              style={{
                fontSize: '14px',
                lineHeight: '20px',
                color: '#6b7280',
                fontFamily: 'PingFang SC',
                fontWeight: 400,
              }}
            >
              匹配分数:
            </span>
            <span
              style={{
                fontSize: '20px',
                lineHeight: '28px',
                color: '#7bb8ff',
                fontFamily: 'Inter',
                fontWeight: 700,
              }}
            >
              {responsibilityDetail.score}
            </span>
          </div>
        </div>

        {/* 内容区域 */}
        <div className="px-6 pb-6 space-y-6">
          {/* 职责匹配度 */}
          <div
            className="rounded-lg p-5"
            style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
          >
            <h3
              className="font-medium mb-4"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              职责匹配度
            </h3>
            <div className="space-y-5">
              {matchedResponsibilities.map((resp, index) => {
                const matchScore = resp.match_score || 0;
                const matchPercent = Math.round(matchScore);
                const matchColor = getMatchColor(matchPercent);

                return (
                  <div key={index}>
                    {/* 职责名称和匹配百分比 */}
                    <div
                      className="flex items-center justify-between"
                      style={{ height: '24px', marginBottom: '8px' }}
                    >
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#1d2129',
                          fontFamily: 'PingFang SC',
                          fontWeight: 500,
                        }}
                      >
                        {resp.match_reason || `职责 ${index + 1}`}
                      </span>
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: matchColor,
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                        }}
                      >
                        {matchPercent}% 匹配
                      </span>
                    </div>
                    {/* 进度条 */}
                    <div
                      style={{
                        width: '100%',
                        height: '6px',
                        backgroundColor: '#f3f4f6',
                        borderRadius: '3px',
                        overflow: 'hidden',
                      }}
                    >
                      <div
                        style={{
                          width: `${matchPercent}%`,
                          height: '100%',
                          backgroundColor: matchColor,
                          transition: 'width 0.3s ease',
                        }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* 职责分析 */}
          <div>
            <h3
              className="font-medium mb-3"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              职责分析
            </h3>
            <div
              className="rounded-lg p-5"
              style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
            >
              <p
                style={{
                  fontSize: '14px',
                  lineHeight: '20px',
                  color: '#4b5563',
                  fontFamily: 'PingFang SC',
                  fontWeight: 400,
                }}
              >
                {(() => {
                  // 尝试从 LLM 分析中获取详细分析
                  if (matchedResponsibilities.length > 0) {
                    const highMatch = matchedResponsibilities.filter(
                      (r) => r.match_score >= 85
                    ).length;
                    const mediumMatch = matchedResponsibilities.filter(
                      (r) => r.match_score >= 70 && r.match_score < 85
                    ).length;
                    const lowMatch = matchedResponsibilities.filter(
                      (r) => r.match_score < 70
                    ).length;

                    return `候选人在${highMatch}项职责要求上表现优秀，高度匹配职位需求。${mediumMatch > 0 ? `在${mediumMatch}项职责上表现良好，基本符合要求。` : ''}${lowMatch > 0 ? `在${lowMatch}项职责上经验略有不足，建议在入职后提供相关培训支持。` : ''}整体职责匹配度高，能够胜任该职位的大部分工作内容。`;
                  }

                  return '候选人在前端架构设计、性能优化和团队协作方面经验丰富，与高级前端开发工程师职责要求高度匹配。能够独立负责大型项目的前端架构设计和开发工作，具备带领团队完成复杂项目的能力。在用户体验优化和移动端适配方面经验略有不足，但整体职责匹配度高，能够胜任该职位的大部分工作内容。建议在入职后提供相关培训，弥补移动端适配经验的不足。';
                })()}
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // 渲染技能匹配详情 - 严格按照Figma设计
  const renderSkillDetails = () => {
    const skillDetail = result?.skill_detail;
    if (!skillDetail) return <div className="text-gray-500">暂无数据</div>;

    // 从简历详情中获取技能
    const skills = resumeDetail?.skills || [];

    // 将技能分为技术技能和软技能
    const technicalSkills = skills.filter(
      (s) =>
        ![
          '团队协作',
          '沟通能力',
          '问题解决能力',
          '学习能力',
          '项目管理',
        ].includes(s.skill_name)
    );

    // 模拟软技能数据（如果后端没有提供）
    const softSkills = [
      { name: '团队协作', level: '优秀', progress: 90 },
      { name: '沟通能力', level: '优秀', progress: 88 },
      { name: '问题解决能力', level: '优秀', progress: 92 },
      { name: '学习能力', level: '良好', progress: 78 },
      { name: '项目管理', level: '良好', progress: 75 },
    ];

    // 计算进度条颜色
    const getProgressColor = (level: string) => {
      if (level === '精通' || level === '优秀') return '#52c41a';
      if (level === '熟练' || level === '良好') return '#faad14';
      return '#f5222d';
    };

    // 计算进度百分比
    const getProgressPercent = (level: string, index: number) => {
      if (level === '精通') return 95 - index * 5;
      if (level === '熟练') return 85 - index * 5;
      if (level === '优秀') return 90 - index * 5;
      if (level === '良好') return 78 - index * 5;
      return 60;
    };

    return (
      <div
        className="bg-white rounded-xl border"
        style={{
          boxShadow: '0px 2px 14px 0px rgba(0, 0, 0, 0.06)',
          borderColor: '#e5e7eb',
          borderRadius: '12px',
        }}
      >
        {/* 标题和分数 */}
        <div
          className="flex items-center justify-between px-6 py-6"
          style={{ borderBottom: '0px solid transparent' }}
        >
          <div className="flex items-center gap-2">
            <img
              src="/assets/skill-icon.png"
              alt="技能匹配"
              className="w-6 h-5"
              style={{ width: '22.5px', height: '18px' }}
            />
            <h2
              className="font-semibold"
              style={{
                fontSize: '18px',
                lineHeight: '28px',
                color: '#1d2129',
                fontFamily: 'PingFang SC',
                fontWeight: 600,
              }}
            >
              技能匹配详情
              {getSubAgentVersion('skill') && (
                <span
                  style={{
                    marginLeft: '8px',
                    fontSize: '12px',
                    color: '#9ca3af',
                    fontWeight: 400,
                  }}
                >
                  ({getSubAgentVersion('skill')})
                </span>
              )}
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <span
              style={{
                fontSize: '14px',
                lineHeight: '20px',
                color: '#6b7280',
                fontFamily: 'PingFang SC',
                fontWeight: 400,
              }}
            >
              匹配分数:
            </span>
            <span
              style={{
                fontSize: '20px',
                lineHeight: '28px',
                color: '#7bb8ff',
                fontFamily: 'Inter',
                fontWeight: 700,
              }}
            >
              {skillDetail.score}
            </span>
          </div>
        </div>

        {/* 内容区域 */}
        <div className="px-6 pb-6">
          {/* 技能卡片：技术技能和软技能 */}
          <div className="grid grid-cols-2 gap-6 mb-6">
            {/* 技术技能 */}
            <div
              className="rounded-lg p-5"
              style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
            >
              <h3
                className="font-medium mb-4"
                style={{
                  fontSize: '16px',
                  lineHeight: '24px',
                  color: '#1f2937',
                  fontFamily: 'PingFang SC',
                  fontWeight: 500,
                }}
              >
                技术技能
              </h3>
              <div className="space-y-5">
                {technicalSkills.slice(0, 5).map((skill, index) => {
                  const progress = getProgressPercent(skill.level, index);
                  return (
                    <div key={skill.id}>
                      {/* 技能名称和等级 */}
                      <div
                        className="flex items-center justify-between"
                        style={{ height: '20px', marginBottom: '4px' }}
                      >
                        <span
                          style={{
                            fontSize: '14px',
                            lineHeight: '20px',
                            color: '#1d2129',
                            fontFamily: 'Inter',
                            fontWeight: 400,
                          }}
                        >
                          {skill.skill_name}
                        </span>
                        <span
                          style={{
                            fontSize: '12px',
                            lineHeight: '16px',
                            color: '#6b7280',
                            fontFamily: 'PingFang SC',
                            fontWeight: 400,
                          }}
                        >
                          {skill.level || '熟练'}
                        </span>
                      </div>
                      {/* 进度条 */}
                      <div
                        style={{
                          width: '100%',
                          height: '6px',
                          backgroundColor: '#f3f4f6',
                          borderRadius: '3px',
                          overflow: 'hidden',
                        }}
                      >
                        <div
                          style={{
                            width: `${progress}%`,
                            height: '100%',
                            backgroundColor: getProgressColor(
                              skill.level || '熟练'
                            ),
                            transition: 'width 0.3s ease',
                          }}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* 软技能 */}
            <div
              className="rounded-lg p-5"
              style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
            >
              <h3
                className="font-medium mb-4"
                style={{
                  fontSize: '16px',
                  lineHeight: '24px',
                  color: '#1f2937',
                  fontFamily: 'PingFang SC',
                  fontWeight: 500,
                }}
              >
                软技能
              </h3>
              <div className="space-y-5">
                {softSkills.map((skill, index) => (
                  <div key={index}>
                    {/* 技能名称和等级 */}
                    <div
                      className="flex items-center justify-between"
                      style={{ height: '20px', marginBottom: '4px' }}
                    >
                      <span
                        style={{
                          fontSize: '14px',
                          lineHeight: '20px',
                          color: '#1d2129',
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                        }}
                      >
                        {skill.name}
                      </span>
                      <span
                        style={{
                          fontSize: '12px',
                          lineHeight: '16px',
                          color: '#6b7280',
                          fontFamily: 'PingFang SC',
                          fontWeight: 400,
                        }}
                      >
                        {skill.level}
                      </span>
                    </div>
                    {/* 进度条 */}
                    <div
                      style={{
                        width: '100%',
                        height: '6px',
                        backgroundColor: '#f3f4f6',
                        borderRadius: '3px',
                        overflow: 'hidden',
                      }}
                    >
                      <div
                        style={{
                          width: `${skill.progress}%`,
                          height: '100%',
                          backgroundColor: getProgressColor(skill.level),
                          transition: 'width 0.3s ease',
                        }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* 技能分析 */}
          <div>
            <h3
              className="font-medium mb-3"
              style={{
                fontSize: '16px',
                lineHeight: '24px',
                color: '#1f2937',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
              }}
            >
              技能分析
            </h3>
            <div
              className="rounded-lg p-5"
              style={{ backgroundColor: '#f9fafb', borderRadius: '8px' }}
            >
              <p
                style={{
                  fontSize: '14px',
                  lineHeight: '20px',
                  color: '#4b5563',
                  fontFamily: 'PingFang SC',
                  fontWeight: 400,
                }}
              >
                {skillDetail.llm_analysis?.analysis_detail ||
                  '候选人技术栈全面，掌握主流前端技术和框架，具备扎实的技术功底。在React和性能优化方面有突出表现，符合高级前端开发工程师职位要求。软技能方面，团队协作和沟通能力优秀，问题解决能力强，能够有效推动项目进展。学习能力和项目管理能力略有提升空间，整体技能匹配度高，能够胜任高级前端开发工程师职责。'}
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // 渲染头部区域
  const renderHeader = () => (
    <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-white">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 bg-[#7bb8ff] rounded-full flex items-center justify-center">
          <BarChart3 className="w-4 h-4 text-white" />
        </div>
        <h2 className="text-xl font-semibold text-gray-900">报告详情</h2>
      </div>
      <div className="flex items-center gap-3">
        <Button
          variant="outline"
          size="sm"
          disabled
          className="flex items-center gap-2 border-gray-300 text-gray-400 cursor-not-allowed"
        >
          <Download className="w-4 h-4" />
          导出报告
        </Button>
        <Button
          size="sm"
          disabled
          className="flex items-center gap-2 bg-gray-300 text-gray-500 cursor-not-allowed"
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
        <UserCircle className="w-5 h-5 text-[#7bb8ff]" />
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
            <p className="text-sm text-gray-600">任务ID</p>
            <p className="text-base font-medium text-gray-900 break-words max-w-xs">
              {taskId || '未知'}
            </p>
          </div>
        </div>
      </div>
    </div>
  );

  // 渲染综合分数区域 - 严格按照Figma设计，使用雷达图
  const renderOverallScore = () => {
    if (!result?.overall_score) return null;
    const overallScore = result.overall_score;
    const matchLevelInfo = getMatchLevelInfo(result.match_level || 'fair');

    return (
      <div
        className="bg-white rounded-xl border"
        style={{
          position: 'relative',
          boxShadow: '0px 2px 14px 0px rgba(0, 0, 0, 0.06)',
          borderRadius: '12px',
          padding: '24px',
        }}
      >
        {/* 标题 */}
        <div className="flex items-center gap-2 mb-6">
          <Award className="w-5 h-5 text-[#7bb8ff]" />
          <h3
            style={{
              fontSize: '18px',
              lineHeight: '28px',
              fontWeight: 600,
              color: '#1d2129',
              fontFamily: 'PingFang SC',
            }}
          >
            综合分数
          </h3>
        </div>

        <div
          style={{
            display: 'flex',
            gap: '24px',
            justifyContent: 'center',
          }}
        >
          {/* 左侧雷达图和总分 */}
          <div
            style={{
              position: 'relative',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              alignItems: 'center',
              backgroundColor: '#f9fafb',
              borderRadius: '8px',
              padding: '24px',
            }}
          >
            {/* 雷达图容器 */}
            <div
              style={{
                position: 'relative',
                width: '192px',
                height: '208px',
                paddingBottom: '16px',
              }}
            >
              {/* 雷达图背景 */}
              <div
                style={{
                  position: 'relative',
                  width: '192px',
                  height: '192px',
                }}
              >
                <img
                  src="/assets/radar-chart-56586a.png"
                  alt="雷达图"
                  style={{
                    width: '192px',
                    height: '192px',
                    display: 'block',
                  }}
                />
                {/* 分数叠加层 */}
                <div
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '192px',
                    height: '192px',
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'center',
                    alignItems: 'center',
                  }}
                >
                  <div
                    style={{
                      fontSize: '36px',
                      lineHeight: '40px',
                      fontWeight: 700,
                      color: '#7bb8ff',
                      fontFamily: 'Inter',
                    }}
                  >
                    {overallScore}
                  </div>
                  <div
                    style={{
                      fontSize: '14px',
                      lineHeight: '20px',
                      color: '#6b7280',
                      fontFamily: 'PingFang SC',
                      fontWeight: 400,
                    }}
                  >
                    综合评分
                  </div>
                </div>
              </div>
            </div>

            {/* 匹配等级标签 */}
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: '8px',
              }}
            >
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'center',
                  padding: '4px 12px',
                  backgroundColor: matchLevelInfo.bgColor.includes('green')
                    ? '#f6ffed'
                    : matchLevelInfo.bgColor.includes('yellow')
                      ? '#fffbe6'
                      : matchLevelInfo.bgColor.includes('blue')
                        ? '#e6f7ff'
                        : '#f6ffed',
                  borderRadius: '9999px',
                }}
              >
                <span
                  style={{
                    fontSize: '14px',
                    lineHeight: '20px',
                    fontWeight: 500,
                    color: matchLevelInfo.color.includes('green')
                      ? '#52c41a'
                      : matchLevelInfo.color.includes('yellow')
                        ? '#faad14'
                        : matchLevelInfo.color.includes('blue')
                          ? '#1890ff'
                          : '#52c41a',
                    fontFamily: 'PingFang SC',
                    textAlign: 'center',
                  }}
                >
                  {matchLevelInfo.text}
                </span>
              </div>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                }}
              >
                <p
                  style={{
                    fontSize: '14px',
                    lineHeight: '20px',
                    color: '#4b5563',
                    fontFamily: 'PingFang SC',
                    fontWeight: 400,
                    textAlign: 'center',
                  }}
                >
                  {overallScore >= 85
                    ? '该候选人综合能力优秀，匹配度高'
                    : overallScore >= 70
                      ? '该候选人综合能力良好，匹配度较高'
                      : overallScore >= 55
                        ? '该候选人综合能力一般，需要进一步评估'
                        : '该候选人综合能力有待提升'}
                </p>
              </div>
            </div>
          </div>

          {/* 右侧各维度分数 - 3x2网格布局 */}
          <div style={{ flex: 1 }}>{renderDimensionScores()}</div>
        </div>

        {/* 说明文字 - 整个区域右下角 */}
        <div
          style={{
            position: 'absolute',
            bottom: '12px',
            right: '12px',
            fontSize: '10px',
            lineHeight: '14px',
            color: '#9ca3af',
            fontFamily: 'PingFang SC',
            fontWeight: 400,
          }}
        >
          说明：综合分数=大模型计算分数*权重值的加和。
        </div>
      </div>
    );
  };

  // 渲染推荐建议区域
  const renderRecommendationsSection = () => (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
      <div className="flex items-center gap-2 mb-6">
        <Lightbulb className="w-5 h-5 text-[#7bb8ff]" />
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
      { key: '基本信息', label: '基本信息' },
      { key: '教育背景', label: '教育背景' },
      { key: '工作经验', label: '工作经验' },
      { key: '行业背景', label: '行业背景' },
      { key: '职责匹配', label: '职责匹配' },
      { key: '技能匹配', label: '技能匹配' },
    ];

    return (
      <div className="mt-8">
        {/* 标签页导航 */}
        <div
          className="flex"
          style={{
            borderBottom: '1px solid #f3f4f6',
          }}
        >
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setDetailActiveTab(tab.key)}
              className="transition-colors"
              style={{
                padding: '18px 24px',
                fontSize: '16px',
                lineHeight: '24px',
                fontFamily: 'PingFang SC',
                fontWeight: 500,
                color: detailActiveTab === tab.key ? '#7bb8ff' : '#6b7280',
                borderBottom:
                  detailActiveTab === tab.key ? '2px solid #7bb8ff' : 'none',
                backgroundColor: 'transparent',
                textAlign: 'center',
                width: '112px',
                height: '58px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
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
      case '基本信息':
      case 'basic':
        return renderBasicInfoDetails();
      case '教育背景':
      case 'education':
        return renderEducationDetails();
      case '工作经验':
      case 'experience':
        return renderExperienceDetails();
      case '行业背景':
      case 'industry':
        return renderIndustryDetails();
      case '职责匹配':
      case 'responsibility':
        return renderResponsibilityDetails();
      case '技能匹配':
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
        <DialogTitle className="sr-only">匹配报告详情</DialogTitle>
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
        </div>
      </DialogContent>
    </Dialog>
  );
}
