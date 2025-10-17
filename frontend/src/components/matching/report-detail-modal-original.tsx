import { useEffect, useState, useCallback } from 'react';
import {
  X,
  CheckCircle2,
  AlertCircle,
  Sparkles,
  User,
  Briefcase,
  FolderOpen,
  Award,
  Lightbulb,
} from 'lucide-react';
import { Dialog, DialogContent } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

import { getScreeningResult } from '@/services/screening';
import { ScreeningResult, MatchLevel } from '@/types/screening';
import { getResumeDetail } from '@/services/resume';
import { ResumeDetail as ResumeDetailType } from '@/types/resume';
import { getJobProfile } from '@/services/job-profile';
import { getResumeProgress } from '@/services/screening';
import { GetResumeProgressResp } from '@/types/screening';

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
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [resumeDetail, setResumeDetail] = useState<ResumeDetailType | null>(
    null
  );
  const [jobPositionName, setJobPositionName] = useState<string | null>(null);
  const [progress, setProgress] = useState<GetResumeProgressResp | null>(null);

  const loadReportDetail = useCallback(async () => {
    if (!taskId || !resumeId) return;

    setLoading(true);
    setError(null);

    try {
      const response = await getScreeningResult(taskId, resumeId);
      setResult(response.result);
    } catch (err) {
      console.error('加载报告详情失败:', err);
      setError(err instanceof Error ? err.message : '加载失败');
    } finally {
      setLoading(false);
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

  // 解析应聘岗位名称（优先从简历的岗位关联中匹配当前任务岗位ID，失败则回退到岗位画像接口）
  useEffect(() => {
    if (!result?.job_position_id) {
      setJobPositionName(null);
      return;
    }
    const jobId = result.job_position_id;

    // 优先从简历详情的 job_positions 中匹配
    const matchedFromResume = resumeDetail?.job_positions?.find(
      (jp) => jp.job_position_id === jobId
    );
    if (matchedFromResume?.job_title) {
      setJobPositionName(matchedFromResume.job_title);
      return;
    }

    // 回退到根据岗位ID请求岗位画像详情
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

  // 进度轮询：打开且有 taskId/resumeId 时，每 8s 轮询一次
  useEffect(() => {
    if (!open || !taskId || !resumeId) {
      setProgress(null);
      return;
    }

    let timer: number | undefined;
    let cancelled = false;

    const poll = async () => {
      try {
        const resp = await getResumeProgress(taskId, resumeId);
        if (cancelled) return;
        setProgress(resp);
        const cont = resp.status === 'in_progress' || resp.status === 'pending';
        if (cont) {
          timer = window.setTimeout(poll, 8000);
        }
      } catch (e) {
        if (cancelled) return;
        console.error('获取简历进度失败:', e);
        // 回退延迟后重试
        timer = window.setTimeout(poll, 10000);
      }
    };

    // 立即拉取一次
    poll();

    return () => {
      cancelled = true;
      if (timer) window.clearTimeout(timer);
    };
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

  // 获取维度名称
  const getDimensionName = (key: string): string => {
    const nameMap: Record<string, string> = {
      skill: '技能匹配',
      responsibility: '职责匹配',
      experience: '经验匹配',
      education: '教育背景',
      industry: '行业匹配',
      basic: '基本信息',
    };
    return nameMap[key] || key;
  };

  // 渲染维度分数卡片
  const renderDimensionScores = () => {
    if (!result?.dimension_scores) return null;

    return (
      <div className="grid grid-cols-2 md:grid-cols-3 gap-6">
        {Object.entries(result.dimension_scores).map(([key, score]) => (
          <div
            key={key}
            className="rounded-xl border border-[#E5E7EB] bg-white p-5 shadow-sm flex flex-col items-center gap-2.5"
          >
            <div className="text-primary text-[32px] font-bold leading-tight">
              {Math.round(score)}
            </div>
            <div className="text-sm font-medium text-gray-700">
              {getDimensionName(key)}
            </div>
          </div>
        ))}
      </div>
    );
  };

  // 已移除未使用的 renderDetailSection 以通过 ESLint 检查

  // 渲染推荐建议
  const renderRecommendations = () => {
    if (!result?.recommendations || result.recommendations.length === 0)
      return null;

    return (
      <div className="mt-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
          <Lightbulb className="h-5 w-5 text-gray-500" />
          推荐建议
        </h3>
        <div className="space-y-3">
          {result.recommendations.map((rec, index) => (
            <div key={index} className="border-l-2 border-blue-200 pl-4 py-2">
              <div className="font-medium text-gray-900 mb-1">
                {rec.split(':')[0] || `建议 ${index + 1}`}
              </div>
              <div className="text-sm text-gray-700 leading-relaxed">
                {rec.split(':').slice(1).join(':').trim() || rec}
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  };

  if (loading) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-[1200px] max-h-[90vh] p-0 overflow-hidden">
          <div className="flex items-center justify-center h-96">
            <div className="text-gray-500">加载中...</div>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  if (error) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-[1200px] max-h-[90vh] p-0 overflow-hidden">
          <div className="flex flex-col items-center justify-center h-96 gap-4">
            <div className="text-red-500">加载失败: {error}</div>
            <Button onClick={loadReportDetail}>重试</Button>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  if (!result) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-[1200px] max-h-[90vh] p-0 overflow-hidden">
          <div className="flex flex-col items-center justify-center h-96 gap-4">
            <div className="text-gray-500">暂无数据或未生成报告</div>
            <Button onClick={loadReportDetail}>重试</Button>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  const matchLevelInfo = getMatchLevelInfo(result.match_level);
  const MatchLevelIcon = matchLevelInfo.icon;

  // 已移除未使用的 renderResumeInfo 以通过 ESLint 检查

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[1200px] max-h-[90vh] p-0 overflow-hidden flex flex-col">
        {/* 顶部标题栏 */}
        <div className="border-b border-[#E5E7EB] bg-white px-10 py-6 flex items-center justify-between">
          <h2 className="text-[20px] font-semibold text-gray-900">报告详情</h2>
          <div className="flex items-center gap-3">
            {progress && (
              <div className="flex items-center gap-2">
                <span
                  className={`inline-flex items-center rounded-full px-3 py-1 text-xs ${
                    progress.status === 'completed'
                      ? 'bg-[#D1FAE5] text-[#36CFC9]'
                      : 'bg-muted text-muted-foreground'
                  }`}
                >
                  {progress.status === 'completed'
                    ? '已完成'
                    : progress.status === 'in_progress'
                      ? `处理中 · ${Math.round(progress.progress_percent ?? 0)}%`
                      : progress.status === 'pending'
                        ? '排队中'
                        : progress.status === 'failed'
                          ? '处理失败'
                          : '未知状态'}
                </span>
                {typeof progress.progress_percent === 'number' && (
                  <div className="w-40 h-2 rounded bg-muted">
                    <div
                      className="h-2 rounded bg-primary"
                      style={{
                        width: `${Math.max(0, Math.min(100, progress.progress_percent))}%`,
                      }}
                    />
                  </div>
                )}
              </div>
            )}
            <Button
              variant="ghost"
              size="icon"
              onClick={() => onOpenChange(false)}
              className="text-gray-600 hover:bg-gray-100"
            >
              <X className="h-5 w-5" />
            </Button>
          </div>
        </div>

        {/* 滚动内容区域 */}
        <div className="flex-1 overflow-y-auto px-10 py-10">
          <div className="flex flex-col gap-7">
            {/* 顶部信息区：左侧基本信息，右侧总体分数 */}
            <div className="flex flex-col md:flex-row gap-7">
              <div className="flex-1">
                <div className="rounded-xl border border-[#E5E7EB] bg-white p-6">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                    <User className="h-5 w-5 text-gray-500" />
                    基本信息
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-600">姓名:</span>
                      <span className="font-medium">
                        {resumeDetail?.name || resumeName || '未知'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">性别:</span>
                      <span className="font-medium">
                        {resumeDetail?.gender || '未填写'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">学历:</span>
                      <span className="font-medium">
                        {resumeDetail?.highest_education || '未填写'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">工作年限:</span>
                      <span className="font-medium">
                        {resumeDetail?.years_experience
                          ? `${resumeDetail.years_experience}年`
                          : '未填写'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">应聘岗位:</span>
                      <span className="font-medium">
                        {jobPositionName || '未填写'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">电话:</span>
                      <span className="font-medium">
                        {resumeDetail?.phone || '未填写'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">邮箱:</span>
                      <span className="font-medium">
                        {resumeDetail?.email || '未填写'}
                      </span>
                    </div>
                  </div>
                </div>

                {/* 工作经历 */}
                <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 mt-6">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                    <Briefcase className="h-5 w-5 text-gray-500" />
                    工作经历
                  </h3>
                  {resumeDetail?.experiences &&
                  resumeDetail.experiences.length > 0 ? (
                    <div className="space-y-4">
                      {resumeDetail.experiences.map((work, index) => (
                        <div
                          key={index}
                          className="border-l-2 border-green-200 pl-4 py-2"
                        >
                          <div className="flex justify-between items-start mb-2">
                            <div>
                              <div className="font-medium text-gray-900">
                                {work.company}
                              </div>
                              <div className="text-sm text-gray-600">
                                {work.position || work.title}
                              </div>
                            </div>
                            <div className="text-sm text-gray-500">
                              {work.start_date} - {work.end_date || '至今'}
                            </div>
                          </div>
                          {work.description && (
                            <div className="text-sm text-gray-700 leading-relaxed">
                              {work.description}
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-sm text-gray-500">
                      暂无工作经历信息
                    </div>
                  )}
                </div>

                {/* 项目经验 */}
                <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 mt-6">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                    <FolderOpen className="h-5 w-5 text-gray-500" />
                    项目经验
                  </h3>
                  {resumeDetail?.projects &&
                  resumeDetail.projects.length > 0 ? (
                    <div className="space-y-4">
                      {resumeDetail.projects.map((project, index) => (
                        <div
                          key={index}
                          className="border-l-2 border-purple-200 pl-4 py-2"
                        >
                          <div className="flex justify-between items-start mb-2">
                            <div>
                              <div className="font-medium text-gray-900">
                                {project.name || project.project_name}
                              </div>
                              {project.role && (
                                <div className="text-sm text-gray-600">
                                  {project.role}
                                </div>
                              )}
                            </div>
                            <div className="text-sm text-gray-500">
                              {project.start_date} -{' '}
                              {project.end_date || '至今'}
                            </div>
                          </div>
                          {project.description && (
                            <div className="text-sm text-gray-700 leading-relaxed">
                              {project.description}
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-sm text-gray-500">
                      暂无项目经验信息
                    </div>
                  )}
                </div>

                {/* 技能 */}
                <div className="rounded-xl border border-[#E5E7EB] bg-white p-6">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                    <Award className="h-5 w-5 text-gray-500" />
                    技能
                  </h3>
                  <div className="space-y-4">
                    {resumeDetail?.skills && resumeDetail.skills.length > 0 ? (
                      <div>
                        <div className="flex flex-wrap gap-2">
                          {resumeDetail.skills.map((skill, index) => (
                            <span
                              key={index}
                              className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full"
                            >
                              {skill.skill_name}
                            </span>
                          ))}
                        </div>
                        {/* 修复：补充关闭外层 div */}
                      </div>
                    ) : (
                      <div className="text-sm text-gray-500">暂无技能信息</div>
                    )}
                  </div>
                </div>
              </div>

              <div className="flex-1">
                <div className="rounded-xl border border-[#E5E7EB] bg-[#D1FAE5] p-6 flex flex-col items-center gap-2">
                  <div className="flex items-end justify-center">
                    <span className="text-[64px] font-bold text-gray-600 leading-none">
                      {Math.round(result.overall_score)}
                    </span>
                    <span className="ml-2 text-[28px] font-semibold text-gray-600 leading-none">
                      分
                    </span>
                  </div>
                  <div className="text-xs font-semibold text-gray-600 flex items-center gap-1">
                    <span>综合匹配度 - {matchLevelInfo.text}</span>
                    <MatchLevelIcon className="h-3 w-3 text-gray-600" />
                  </div>
                  {resumeName && (
                    <div className="text-sm text-gray-700">
                      投递人：{resumeName}
                    </div>
                  )}
                  {jobPositionName && (
                    <div className="text-sm text-gray-700">
                      应聘岗位：{jobPositionName}
                    </div>
                  )}
                  {result.matched_at && (
                    <div className="text-xs text-gray-500">
                      匹配完成时间：
                      {new Date(result.matched_at).toLocaleString('zh-CN')}
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* 各维度匹配分数 */}
            {result.dimension_scores && (
              <div className="flex flex-col gap-5">
                {renderDimensionScores()}
              </div>
            )}

            {/* 关键匹配点分析 - 按要求移除亮点，不再展示 */}
            {/* （已移除 renderStrengthsAndWeaknesses() 调用） */}

            {/* 基本信息匹配详情 */}
            {result.basic_detail && result.basic_detail.evidence.length > 0 && (
              <div className="flex flex-col gap-3">
                <h4 className="text-base font-semibold text-[#333333]">
                  基本信息匹配详情
                </h4>
                <div className="rounded-xl border border-[#E5E7EB] bg-white p-6">
                  <div className="text-sm text-gray-700 space-y-1">
                    {result.basic_detail.evidence.map((item, idx) => (
                      <div key={idx}>• {item}</div>
                    ))}
                  </div>
                  {result.basic_detail.notes && (
                    <div className="mt-2 text-sm text-gray-600">
                      备注：{result.basic_detail.notes}
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* 教育背景匹配详情 */}
            {result.education_detail && (
              <div className="flex flex-col gap-3">
                <h4 className="text-base font-semibold text-[#333333]">
                  教育背景匹配详情
                </h4>
                <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 space-y-3">
                  {result.education_detail.degree_match && (
                    <div className="text-sm">
                      <span className="font-medium text-gray-700">
                        学历要求：
                      </span>
                      <span className="text-gray-600">
                        {result.education_detail.degree_match.required_degree} →{' '}
                        实际学历：
                        {result.education_detail.degree_match.actual_degree}
                        {result.education_detail.degree_match.meets
                          ? ' ✓ 符合'
                          : ' ✗ 不符合'}
                      </span>
                    </div>
                  )}
                  {result.education_detail.major_matches &&
                    result.education_detail.major_matches.length > 0 && (
                      <div className="text-sm">
                        <span className="font-medium text-gray-700">
                          专业匹配：
                        </span>
                        <ul className="mt-1 space-y-1">
                          {result.education_detail.major_matches.map(
                            (major, idx) => (
                              <li key={idx} className="text-gray-600">
                                • {major.major} (相关度:{' '}
                                {(major.relevance * 100).toFixed(0)}%)
                              </li>
                            )
                          )}
                        </ul>
                      </div>
                    )}
                </div>
              </div>
            )}

            {/* 工作经验匹配详情 */}
            {result.experience_detail && (
              <div className="flex flex-col gap-3">
                <h4 className="text-base font-semibold text-[#333333]">
                  工作经验匹配详情
                </h4>
                <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 space-y-3">
                  {result.experience_detail.years_match && (
                    <div className="text-sm">
                      <span className="font-medium text-gray-700">
                        工作年限：
                      </span>
                      <span className="text-gray-600">
                        要求{' '}
                        {result.experience_detail.years_match.required_years}{' '}
                        年， 实际{' '}
                        {result.experience_detail.years_match.actual_years} 年
                        {result.experience_detail.years_match.gap > 0 &&
                          ` (差距 ${result.experience_detail.years_match.gap.toFixed(1)} 年)`}
                      </span>
                    </div>
                  )}
                  {result.experience_detail.position_matches &&
                    result.experience_detail.position_matches.length > 0 && (
                      <div className="text-sm">
                        <span className="font-medium text-gray-700">
                          相关职位：
                        </span>
                        <ul className="mt-1 space-y-1">
                          {result.experience_detail.position_matches.map(
                            (pos, idx) => (
                              <li key={idx} className="text-gray-600">
                                • {pos.position} (相关度:{' '}
                                {(pos.relevance * 100).toFixed(0)}%)
                              </li>
                            )
                          )}
                        </ul>
                      </div>
                    )}
                </div>
              </div>
            )}

            {/* 行业背景匹配详情 */}
            {result.industry_detail &&
              result.industry_detail.industry_matches &&
              result.industry_detail.industry_matches.length > 0 && (
                <div className="flex flex-col gap-3">
                  <h4 className="text-base font-semibold text-[#333333]">
                    行业背景匹配详情
                  </h4>
                  <div className="rounded-xl border border-[#E5E7EB] bg-white p-6">
                    <ul className="text-sm space-y-1">
                      {result.industry_detail.industry_matches.map(
                        (ind, idx) => (
                          <li key={idx} className="text-gray-600">
                            • {ind.company} - {ind.industry} (相关度:{' '}
                            {(ind.relevance * 100).toFixed(0)}%)
                          </li>
                        )
                      )}
                    </ul>
                  </div>
                </div>
              )}

            {/* 职责匹配详情 */}
            {result.responsibility_detail && (
              <div className="flex flex-col gap-3">
                <h4 className="text-base font-semibold text-[#333333]">
                  职责匹配详情
                </h4>
                <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 space-y-3">
                  {result.responsibility_detail.matched_responsibilities
                    .length > 0 && (
                    <div className="text-sm">
                      <span className="font-medium text-gray-700">
                        匹配的职责：
                      </span>
                      <ul className="mt-1 space-y-1">
                        {result.responsibility_detail.matched_responsibilities.map(
                          (resp, idx) => (
                            <li key={idx} className="text-gray-600">
                              • {resp.match_reason} (分数:{' '}
                              {resp.match_score.toFixed(0)})
                            </li>
                          )
                        )}
                      </ul>
                    </div>
                  )}
                  {result.responsibility_detail.unmatched_responsibilities
                    .length > 0 && (
                    <div className="text-sm">
                      <span className="font-medium text-gray-700">
                        未匹配的职责：
                      </span>
                      <ul className="mt-1 space-y-1">
                        {result.responsibility_detail.unmatched_responsibilities.map(
                          (resp, idx) => (
                            <li key={idx} className="text-gray-600">
                              • {resp.description}
                            </li>
                          )
                        )}
                      </ul>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* 技能匹配详情 */}
            {result.skill_detail &&
              result.skill_detail.missing_skills &&
              result.skill_detail.missing_skills.length > 0 && (
                <div className="flex flex-col gap-3">
                  <h4 className="text-base font-semibold text-[#333333]">
                    技能匹配详情
                  </h4>
                  <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 space-y-3">
                    {result.skill_detail.matched_skills.length > 0 && (
                      <div className="text-sm">
                        <span className="font-medium text-gray-700">
                          匹配的技能：
                        </span>
                        <ul className="mt-1 space-y-1">
                          {result.skill_detail.matched_skills.map(
                            (skill, idx) => (
                              <li key={idx} className="text-gray-600">
                                • 分数: {skill.score.toFixed(0)} | 匹配类型:{' '}
                                {skill.match_type}
                                {skill.llm_analysis &&
                                  ` | ${skill.llm_analysis.match_reason}`}
                              </li>
                            )
                          )}
                        </ul>
                      </div>
                    )}
                    {result.skill_detail.missing_skills.length > 0 && (
                      <div className="text-sm">
                        <span className="font-medium text-gray-700">
                          缺失的技能：
                        </span>
                        <ul className="mt-1 space-y-1">
                          {result.skill_detail.missing_skills.map(
                            (skill, idx) => (
                              <li key={idx} className="text-gray-600">
                                • {skill.name}{' '}
                                {skill.required_level &&
                                  `(要求: ${skill.required_level})`}
                              </li>
                            )
                          )}
                        </ul>
                      </div>
                    )}
                    {result.skill_detail.extra_skills.length > 0 && (
                      <div className="text-sm">
                        <span className="font-medium text-gray-700">
                          额外的技能：
                        </span>
                        <div className="mt-2 flex flex-wrap gap-2">
                          {result.skill_detail.extra_skills.map(
                            (skill, idx) => (
                              <div
                                key={idx}
                                className="inline-flex items-center gap-1 rounded-md border border-[#E5E7EB] bg-white px-2.5 py-1 shadow-sm"
                              >
                                <Sparkles className="w-3.5 h-3.5 text-primary" />
                                <span className="text-[12px] text-gray-700">
                                  {skill}
                                </span>
                              </div>
                            )
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}

            {/* 推荐建议 */}
            {renderRecommendations()}
          </div>
        </div>

        {/* 底部按钮栏 */}
        <div className="border-t border-[#E0E0E0] px-10 py-7 flex items-center justify-end">
          <Button
            onClick={() => onOpenChange(false)}
            className="rounded-[25px] px-7 py-4 bg-primary text-primary-foreground hover:bg-primary/90"
          >
            完成
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
