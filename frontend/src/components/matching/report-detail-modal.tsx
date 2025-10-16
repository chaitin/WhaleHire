import { useEffect, useState, useCallback } from 'react';
import { X, CheckCircle2, AlertCircle } from 'lucide-react';
import { Dialog, DialogContent } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { getScreeningResult } from '@/services/screening';
import { ScreeningResult, MatchLevel } from '@/types/screening';

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

  // 获取匹配等级的样式和文本
  const getMatchLevelInfo = (level: MatchLevel) => {
    switch (level) {
      case 'excellent':
        return {
          text: '非常匹配',
          icon: CheckCircle2,
          color: 'text-green-600',
          bgColor: 'bg-green-50',
        };
      case 'good':
        return {
          text: '高匹配',
          icon: CheckCircle2,
          color: 'text-blue-600',
          bgColor: 'bg-blue-50',
        };
      case 'fair':
        return {
          text: '一般匹配',
          icon: AlertCircle,
          color: 'text-orange-600',
          bgColor: 'bg-orange-50',
        };
      case 'poor':
        return {
          text: '低匹配',
          icon: AlertCircle,
          color: 'text-red-600',
          bgColor: 'bg-red-50',
        };
      default:
        return {
          text: '未知',
          icon: AlertCircle,
          color: 'text-gray-600',
          bgColor: 'bg-gray-50',
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
      <div className="grid grid-cols-3 gap-5">
        {Object.entries(result.dimension_scores).map(([key, score]) => (
          <div
            key={key}
            className="rounded-xl bg-[#F9F9F9] p-6 flex flex-col gap-2"
          >
            <div className="text-center">
              <div className="text-[#667EEA] text-[32px] font-bold leading-tight">
                {Math.round(score)}
              </div>
            </div>
            <div className="text-center text-sm text-[#666666]">
              {getDimensionName(key)}
            </div>
          </div>
        ))}
      </div>
    );
  };

  // 渲染详情区块
  const renderDetailSection = (
    title: string,
    icon: React.ReactNode,
    content: React.ReactNode
  ) => {
    return (
      <div className="flex flex-col gap-5">
        <div className="flex items-center gap-2.5">
          <div className="flex items-center justify-center text-2xl">
            {icon}
          </div>
          <h3 className="text-lg font-semibold text-[#333333]">{title}</h3>
        </div>
        {content}
      </div>
    );
  };

  // 渲染优势和待提升点
  const renderStrengthsAndWeaknesses = () => {
    if (!result) return null;

    const strengths: string[] = [];
    const weaknesses: string[] = [];

    // 从各个维度提取优势和不足
    if (result.skill_detail?.llm_analysis) {
      strengths.push(
        ...result.skill_detail.llm_analysis.strength_areas.map((s) => `✓ ${s}`)
      );
      weaknesses.push(
        ...result.skill_detail.llm_analysis.gap_areas.map((g) => `⚠ ${g}`)
      );
    }

    if (result.responsibility_detail) {
      const matched = result.responsibility_detail.matched_responsibilities;
      if (matched.length > 0) {
        strengths.push(`✓ 匹配 ${matched.length} 项职责要求`);
      }
      const unmatched = result.responsibility_detail.unmatched_responsibilities;
      if (unmatched.length > 0) {
        weaknesses.push(`⚠ ${unmatched.length} 项职责待加强`);
      }
    }

    if (result.experience_detail?.years_match) {
      const years = result.experience_detail.years_match;
      if (years.gap <= 0) {
        strengths.push(`✓ 工作年限${years.actual_years}年，满足要求`);
      } else {
        weaknesses.push(`⚠ 工作年限差距${years.gap.toFixed(1)}年`);
      }
    }

    return renderDetailSection(
      '关键匹配点分析',
      '✨',
      <div className="flex gap-5">
        {/* 优势匹配点 */}
        {strengths.length > 0 && (
          <div className="flex-1 rounded-xl bg-[#E8F5E9] border-l-4 border-[#4CAF50] p-5 flex flex-col gap-2.5">
            <div className="font-semibold text-[#333333]">✅ 优势匹配点</div>
            <div className="flex flex-col gap-0">
              {strengths.map((item, index) => (
                <div
                  key={index}
                  className={cn(
                    'py-2.5 text-sm text-[#666666]',
                    index < strengths.length - 1 && 'border-b border-black/5'
                  )}
                >
                  {item}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* 待提升点 */}
        {weaknesses.length > 0 && (
          <div className="flex-1 rounded-xl bg-[#FFF3E0] border-l-4 border-[#FF9800] p-5 flex flex-col gap-2.5">
            <div className="font-semibold text-[#333333]">📈 待提升点</div>
            <div className="flex flex-col gap-0">
              {weaknesses.map((item, index) => (
                <div
                  key={index}
                  className={cn(
                    'py-2.5 text-sm text-[#666666]',
                    index < weaknesses.length - 1 && 'border-b border-black/5'
                  )}
                >
                  {item}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    );
  };

  // 渲染推荐建议
  const renderRecommendations = () => {
    if (!result?.recommendations || result.recommendations.length === 0) {
      return null;
    }

    return renderDetailSection(
      '简历优化建议',
      '💡',
      <div className="flex flex-col gap-[15px]">
        {result.recommendations.map((rec, index) => (
          <div
            key={index}
            className="rounded-xl bg-[#F5F7FF] border-l-4 border-[#667EEA] p-5 flex flex-col gap-2"
          >
            <div className="font-semibold text-[#333333]">
              {index + 1}. {rec.split(':')[0] || `建议 ${index + 1}`}
            </div>
            <div className="text-sm text-[#666666]">
              {rec.split(':').slice(1).join(':').trim() || rec}
            </div>
          </div>
        ))}
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
    return null;
  }

  const matchLevelInfo = getMatchLevelInfo(result.match_level);
  const MatchLevelIcon = matchLevelInfo.icon;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[1200px] max-h-[90vh] p-0 overflow-hidden flex flex-col">
        {/* 顶部标题栏 - 渐变背景 */}
        <div className="bg-gradient-to-r from-[#667EEA] to-[#764BA2] px-10 py-7 flex items-center justify-between">
          <h2 className="text-[28px] font-semibold text-white">报告详情</h2>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onOpenChange(false)}
            className="text-white hover:bg-white/20 hover:text-white"
          >
            <X className="h-5 w-5" />
          </Button>
        </div>

        {/* 滚动内容区域 */}
        <div className="flex-1 overflow-y-auto px-10 py-10">
          <div className="flex flex-col gap-7">
            {/* 总体匹配分数卡片 */}
            <div className="rounded-xl bg-gradient-to-br from-[#667EEA] to-[#764BA2] px-3 py-1 flex flex-col gap-2.5">
              <div className="flex items-center justify-center h-[84.5px]">
                <span className="text-[72px] font-bold text-white leading-tight">
                  {Math.round(result.overall_score)}
                </span>
                <span className="text-[36px] font-semibold text-white leading-tight ml-1">
                  分
                </span>
              </div>
              <div className="text-center opacity-90">
                <div className="text-[11px] font-semibold text-white flex items-center justify-center gap-1">
                  <span>综合匹配度 - {matchLevelInfo.text}</span>
                  <MatchLevelIcon className="h-3 w-3" />
                </div>
              </div>
              <div className="text-center opacity-90 pb-1.5">
                <div className="text-sm font-semibold text-white">
                  {resumeName && `投递人：${resumeName}`}
                </div>
              </div>
              {result.matched_at && (
                <div className="text-center opacity-90 pb-1">
                  <div className="text-xs text-white">
                    匹配完成时间：
                    {new Date(result.matched_at).toLocaleString('zh-CN')}
                  </div>
                </div>
              )}
            </div>

            {/* 各维度匹配分数 */}
            {result.dimension_scores && (
              <div className="flex flex-col gap-5">
                {renderDimensionScores()}
              </div>
            )}

            {/* 关键匹配点分析 */}
            {renderStrengthsAndWeaknesses()}

            {/* 基本信息匹配详情 */}
            {result.basic_detail && result.basic_detail.evidence.length > 0 && (
              <div className="flex flex-col gap-3">
                <h4 className="text-base font-semibold text-[#333333]">
                  基本信息匹配详情
                </h4>
                <div className="rounded-lg bg-gray-50 p-4">
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
                <div className="rounded-lg bg-gray-50 p-4 space-y-3">
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
                <div className="rounded-lg bg-gray-50 p-4 space-y-3">
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
                  <div className="rounded-lg bg-gray-50 p-4">
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
                <div className="rounded-lg bg-gray-50 p-4 space-y-3">
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
            {result.skill_detail && (
              <div className="flex flex-col gap-3">
                <h4 className="text-base font-semibold text-[#333333]">
                  技能匹配详情
                </h4>
                <div className="rounded-lg bg-gray-50 p-4 space-y-3">
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
                      <ul className="mt-1 space-y-1">
                        {result.skill_detail.extra_skills.map((skill, idx) => (
                          <li key={idx} className="text-gray-600">
                            • {skill}
                          </li>
                        ))}
                      </ul>
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
            className="rounded-[25px] px-7 py-4 bg-[#667EEA] text-white hover:bg-[#667EEA]/90"
          >
            完成
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
