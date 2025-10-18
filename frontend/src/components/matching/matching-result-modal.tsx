import { useState, useEffect } from 'react';
import {
  ChevronLeft,
  ChevronRight,
  Briefcase,
  User,
  Scale,
  Settings,
  BarChart3,
  X,
  FileText,
} from 'lucide-react';
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { getScreeningTask } from '@/services/screening';
import { GetScreeningTaskResp, MatchLevel } from '@/types/screening';
import { ReportDetailModal } from '@/components/matching/report-detail-modal';
// 新增：按ID获取简历与岗位画像
import { getResumeDetail } from '@/services/resume';
import { getJobProfile } from '@/services/job-profile';

interface MatchingResultModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onBackToHome?: () => void;
  taskId: string | null;
}

// 流程步骤配置
const STEPS = [
  { id: 1, name: '选择岗位', icon: Briefcase, active: false },
  { id: 2, name: '选择简历', icon: User, active: false },
  { id: 3, name: '权重配置', icon: Scale, active: false },
  { id: 4, name: '匹配处理', icon: Settings, active: false },
  { id: 5, name: '匹配结果', icon: BarChart3, active: true },
];

export function MatchingResultModal({
  open,
  onOpenChange,
  onBackToHome,
  taskId,
}: MatchingResultModalProps) {
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [taskData, setTaskData] = useState<GetScreeningTaskResp | null>(null);
  const [loading, setLoading] = useState(false);
  // const [selectedLevel, setSelectedLevel] = useState<'all' | MatchLevel>('all');
  const [isReportOpen, setIsReportOpen] = useState(false);
  const [selectedResumeId, setSelectedResumeId] = useState<string | null>(null);
  const [selectedResumeName, setSelectedResumeName] = useState<string>('');
  // 新增：姓名与岗位名映射缓存
  const [resumeNameMap, setResumeNameMap] = useState<Record<string, string>>(
    {}
  );
  const [jobPositionName, setJobPositionName] = useState<string>('');

  // 获取任务详情
  useEffect(() => {
    if (!open || !taskId) return;

    const fetchTaskData = async () => {
      setLoading(true);
      try {
        const data = await getScreeningTask(taskId);
        setTaskData(data);
        console.log('任务详情:', data);
      } catch (error) {
        console.error('获取任务详情失败:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchTaskData();
  }, [open, taskId]);

  // 新增：根据 job_position_id 获取岗位名称（回退用任务里的名称）
  useEffect(() => {
    const id = taskData?.task?.job_position_id;
    const fallbackName = taskData?.task?.job_position_name || '';
    if (!id) {
      setJobPositionName(fallbackName);
      return;
    }
    (async () => {
      try {
        const profile = await getJobProfile(id);
        setJobPositionName(profile?.name || fallbackName);
      } catch (e) {
        console.error('获取岗位画像失败:', e);
        setJobPositionName(fallbackName);
      }
    })();
  }, [taskData?.task?.job_position_id, taskData?.task?.job_position_name]);

  // 计算匹配等级（基于分数阈值）
  const calcMatchLevel = (score: number): MatchLevel => {
    if (score >= 85) return 'excellent';
    if (score >= 70) return 'good';
    if (score >= 50) return 'fair';
    return 'poor';
  };

  // 计算匹配结果数据（按总分倒序排序并赋排名）
  const matchingResults =
    taskData?.resumes
      .map((resume) => ({
        id: resume.resume_id,
        candidateName: resume.resume_name,
        position: jobPositionName || taskData!.task.job_position_name,
        totalScore: resume.score || 0,
        matchLevel: calcMatchLevel(resume.score || 0),
        status: resume.status,
      }))
      .sort((a, b) => b.totalScore - a.totalScore)
      .map((r, idx) => ({ ...r, rank: idx + 1 })) || [];

  const totalMatches = matchingResults.length;
  const levelCounts = {
    excellent: matchingResults.filter((r) => r.matchLevel === 'excellent')
      .length,
    good: matchingResults.filter((r) => r.matchLevel === 'good').length,
    fair: matchingResults.filter((r) => r.matchLevel === 'fair').length,
    poor: matchingResults.filter((r) => r.matchLevel === 'poor').length,
  };

  // 分页
  const totalPages = Math.ceil(matchingResults.length / pageSize);
  const startIndex = (currentPage - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const currentResults = matchingResults.slice(startIndex, endIndex);

  // 新增：当前页简历姓名懒加载缓存
  useEffect(() => {
    if (!open) return;
    const missingIds = currentResults
      .map((r) => r.id)
      .filter((id) => !resumeNameMap[id]);
    if (missingIds.length === 0) return;

    (async () => {
      const updates: Record<string, string> = {};
      for (const id of missingIds) {
        try {
          const detail = await getResumeDetail(id);
          updates[id] = detail.name || '';
        } catch (e) {
          console.error('获取简历详情失败:', e);
          const fallback =
            currentResults.find((r) => r.id === id)?.candidateName || '';
          updates[id] = fallback;
        }
      }
      setResumeNameMap((prev) => ({ ...prev, ...updates }));
    })();
  }, [
    open,
    startIndex,
    endIndex,
    pageSize,
    taskData?.resumes,
    currentResults,
    resumeNameMap,
  ]);

  const handlePageChange = (page: number) => {
    if (page >= 1 && page <= totalPages) {
      setCurrentPage(page);
    }
  };

  // 渲染分页按钮
  const renderPaginationButtons = () => {
    const pages = [] as JSX.Element[];
    const maxVisiblePages = 5;

    let startPage = Math.max(1, currentPage - Math.floor(maxVisiblePages / 2));
    const endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);

    if (endPage - startPage < maxVisiblePages - 1) {
      startPage = Math.max(1, endPage - maxVisiblePages + 1);
    }

    for (let i = startPage; i <= endPage; i++) {
      pages.push(
        <Button
          key={i}
          variant="outline"
          size="sm"
          onClick={() => setCurrentPage(i)}
          className={cn(
            'h-8 w-8 p-0',
            i === currentPage &&
              'bg-primary text-white border-primary hover:bg-primary hover:text-white'
          )}
        >
          {i}
        </Button>
      );
    }

    return pages;
  };

  const getMatchLevelColor = (level: MatchLevel) => {
    switch (level) {
      case 'excellent':
        return 'bg-primary/10 text-primary';
      case 'good':
        return 'bg-[#E6F0FF] text-[#2563EB]';
      case 'fair':
        return 'bg-[#FFF3E6] text-[#F59E0B]';
      case 'poor':
        return 'bg-[#F3F4F6] text-[#6B7280]';
    }
  };

  const getLevelLabel = (level: MatchLevel) => {
    switch (level) {
      case 'excellent':
        return '非常匹配';
      case 'good':
        return '高匹配';
      case 'fair':
        return '一般匹配';
      case 'poor':
        return '低匹配';
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-w-[920px] p-0 gap-0 bg-white rounded-xl"
        onInteractOutside={(e) => e.preventDefault()}
      >
        <DialogTitle className="sr-only">创建新匹配任务 - 匹配结果</DialogTitle>
        {/* 头部 */}
        <div className="flex items-center justify-between border-b border-[#E8E8E8] px-6 py-5">
          <h2 className="text-lg font-semibold text-[#333333]">
            创建新匹配任务
          </h2>
          <Button
            variant="ghost"
            size="sm"
            className="h-8 w-8 p-0 rounded-md bg-[#F5F5F5] hover:bg-[#E8E8E8]"
            onClick={() => onOpenChange(false)}
          >
            <X className="h-4 w-4 text-[#999999]" />
          </Button>
        </div>

        {/* 内容区域 */}
        <div className="overflow-y-auto max-h-[calc(100vh-160px)] px-6 py-6">
          {loading ? (
            <div className="flex items-center justify-center py-20">
              <div className="text-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
                <p className="text-gray-600">加载匹配结果中...</p>
              </div>
            </div>
          ) : (
            <>
              {/* 流程指示器 */}
              <div className="mb-8">
                <h3 className="text-base font-semibold text-[#333333] mb-6">
                  匹配任务创建流程
                </h3>

                <div className="relative flex items-center justify-between">
                  {/* 连接线 */}
                  <div
                    className="absolute top-7 left-0 right-0 h-0.5 bg-[#E8E8E8]"
                    style={{ marginLeft: '28px', marginRight: '28px' }}
                  />

                  {/* 步骤项 */}
                  {STEPS.map((step, index) => {
                    const IconComponent = step.icon;
                    return (
                      <div
                        key={step.id}
                        className="relative flex flex-col items-center z-10"
                        style={{ width: '108.5px' }}
                      >
                        {/* 图标 */}
                        <div
                          className={cn(
                            'w-14 h-14 rounded-full flex items-center justify-center mb-3',
                            step.active
                              ? 'bg-[#D1FAE5] shadow-[0_0_0_4px_rgba(16,185,129,0.1)]'
                              : 'bg-[#E8E8E8]'
                          )}
                        >
                          <div
                            className={cn(
                              'w-8 h-8 rounded-2xl flex items-center justify-center',
                              step.active ? 'bg-white' : 'bg-white opacity-60'
                            )}
                          >
                            <IconComponent className="h-5 w-5 text-[#999999]" />
                          </div>
                        </div>

                        {/* 步骤名称 */}
                        <span
                          className={cn(
                            'text-sm text-center',
                            step.active
                              ? 'text-[#36CFC9] font-semibold'
                              : 'text-[#666666]'
                          )}
                        >
                          {step.name}
                        </span>
                        {/* 描述 */}
                        <span className="text-xs text-center text-[#999999] mt-0.5">
                          {index === 0 && '选择需要匹配的岗位'}
                          {index === 1 && '选择需要匹配的简历'}
                          {index === 2 && '配置匹配权重'}
                          {index === 3 && 'AI智能匹配分析'}
                          {index === 4 && '查看匹配报告'}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* 整体数据概览 */}
              <div className="bg-[#FAFAFA] rounded-lg p-5 mb-4">
                <h3 className="mb-4 text-base font-semibold text-[#333333]">
                  整体数据概览
                </h3>
                <div className="grid grid-cols-5 gap-4">
                  {/* 匹配总数 */}
                  <div className="flex items-center gap-3 rounded-lg border border-[#E8E8E8] bg-white p-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#E0E7FF]">
                      <BarChart3 className="h-5 w-5 text-[#6366F1]" />
                    </div>
                    <div className="flex flex-col gap-0.5">
                      <div className="text-2xl font-bold leading-tight text-[#333333]">
                        {totalMatches}
                      </div>
                      <div className="text-xs text-[#666666]">匹配总数</div>
                      <div className="text-[11px] font-medium text-[#999999]">
                        100%
                      </div>
                    </div>
                  </div>

                  {/* 非常匹配 */}
                  <div className="flex items-center gap-3 rounded-lg border border-[#E8E8E8] bg-white p-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#D1FAE5]">
                      <BarChart3 className="h-5 w-5 text-[#10B981]" />
                    </div>
                    <div className="flex flex-col gap-0.5">
                      <div className="text-2xl font-bold leading-tight text-[#333333]">
                        {levelCounts.excellent}
                      </div>
                      <div className="text-xs text-[#666666]">非常匹配</div>
                      <div className="text-[11px] font-medium text-[#999999]">
                        {totalMatches > 0
                          ? (
                              (levelCounts.excellent / totalMatches) *
                              100
                            ).toFixed(1)
                          : 0}
                        %
                      </div>
                    </div>
                  </div>

                  {/* 高匹配 */}
                  <div className="flex items-center gap-3 rounded-lg border border-[#E8E8E8] bg-white p-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#E6F0FF]">
                      <BarChart3 className="h-5 w-5 text-[#3B82F6]" />
                    </div>
                    <div className="flex flex-col gap-0.5">
                      <div className="text-2xl font-bold leading-tight text-[#333333]">
                        {levelCounts.good}
                      </div>
                      <div className="text-xs text-[#666666]">高匹配</div>
                      <div className="text-[11px] font-medium text-[#999999]">
                        {totalMatches > 0
                          ? ((levelCounts.good / totalMatches) * 100).toFixed(1)
                          : 0}
                        %
                      </div>
                    </div>
                  </div>

                  {/* 一般匹配 */}
                  <div className="flex items-center gap-3 rounded-lg border border-[#E8E8E8] bg-white p-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#FFF3E6]">
                      <BarChart3 className="h-5 w-5 text-[#F59E0B]" />
                    </div>
                    <div className="flex flex-col gap-0.5">
                      <div className="text-2xl font-bold leading-tight text-[#333333]">
                        {levelCounts.fair}
                      </div>
                      <div className="text-xs text-[#666666]">一般匹配</div>
                      <div className="text-[11px] font-medium text-[#999999]">
                        {totalMatches > 0
                          ? ((levelCounts.fair / totalMatches) * 100).toFixed(1)
                          : 0}
                        %
                      </div>
                    </div>
                  </div>

                  {/* 低匹配 */}
                  <div className="flex items-center gap-3 rounded-lg border border-[#E8E8E8] bg-white p-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#F3F4F6]">
                      <BarChart3 className="h-5 w-5 text-[#6B7280]" />
                    </div>
                    <div className="flex flex-col gap-0.5">
                      <div className="text-2xl font-bold leading-tight text-[#333333]">
                        {levelCounts.poor}
                      </div>
                      <div className="text-xs text-[#666666]">低匹配</div>
                      <div className="text-[11px] font-medium text-[#999999]">
                        {totalMatches > 0
                          ? ((levelCounts.poor / totalMatches) * 100).toFixed(1)
                          : 0}
                        %
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* 等级筛选页签 - 已按需求删除 */}
              {/* 匹配结果详情 */}
              <div className="bg-[#FAFAFA] rounded-lg p-5 mb-4">
                <h3 className="text-base font-semibold text-[#333333] mb-4">
                  匹配结果详情
                </h3>

                {/* 表格 */}
                <div className="bg-white rounded-md border border-[#E8E8E8] overflow-hidden mb-4">
                  <table className="w-full">
                    <thead className="bg-[#F5F5F5]">
                      <tr className="border-b border-[#E8E8E8]">
                        <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                          排名
                        </th>
                        <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                          候选人信息
                        </th>
                        <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                          应聘岗位
                        </th>
                        <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                          匹配总分
                        </th>
                        <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                          匹配度等级
                        </th>
                        <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                          操作
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {currentResults.map((result) => (
                        <tr
                          key={result.id}
                          className="border-b border-[#F0F0F0] last:border-b-0 hover:bg-[#FAFAFA] transition-colors"
                        >
                          <td className="px-3 py-4 text-sm text-[#333333]">
                            {result.rank}
                          </td>
                          <td className="px-3 py-4 text-sm font-medium text-[#333333]">
                            {resumeNameMap[result.id] || result.candidateName}
                          </td>
                          <td className="px-3 py-4 text-sm text-[#333333]">
                            <div className="flex flex-col gap-1">
                              <span>
                                {jobPositionName ||
                                  result.position ||
                                  '未知岗位'}
                              </span>
                              {taskData?.task?.job_position_id && (
                                <span className="text-xs text-gray-500">
                                  ID: {taskData.task.job_position_id}
                                </span>
                              )}
                            </div>
                          </td>
                          <td className="px-3 py-4">
                            <span className="text-lg font-bold text-primary">
                              {result.totalScore}
                            </span>
                          </td>
                          <td className="px-3 py-4">
                            <span
                              className={cn(
                                'inline-flex rounded-full px-2.5 py-1 text-xs font-medium',
                                getMatchLevelColor(result.matchLevel)
                              )}
                            >
                              {getLevelLabel(result.matchLevel)}
                            </span>
                          </td>
                          <td className="px-3 py-4">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 w-8 p-0 text-[#36CFC9] hover:text-[#36CFC9]/80 hover:bg-[#D1FAE5]/50"
                              onClick={() => {
                                setSelectedResumeId(result.id);
                                setSelectedResumeName(
                                  resumeNameMap[result.id] ||
                                    result.candidateName
                                );
                                setIsReportOpen(true);
                              }}
                              title="查看报告详情"
                            >
                              <FileText className="h-4 w-4" />
                            </Button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {/* 分页控件 */}
                <div className="flex items-center justify-between">
                  {/* 分页信息 */}
                  <div className="text-[13px] text-[#666666]">
                    显示第 {startIndex + 1} 到{' '}
                    {Math.min(endIndex, matchingResults.length)} 条，共{' '}
                    {matchingResults.length} 条结果
                  </div>

                  {/* 分页按钮 */}
                  <div className="flex items-center gap-2">
                    {/* 每页条数 */}
                    <Select
                      value={String(pageSize)}
                      onValueChange={(value) => setPageSize(Number(value))}
                    >
                      <SelectTrigger className="w-[85.5px] h-[25px] text-[13px]">
                        <SelectValue placeholder="20条/页" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="10">10条/页</SelectItem>
                        <SelectItem value="20">20条/页</SelectItem>
                        <SelectItem value="50">50条/页</SelectItem>
                      </SelectContent>
                    </Select>

                    {/* 页码按钮 */}
                    <div className="flex items-center gap-1">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handlePageChange(currentPage - 1)}
                        disabled={currentPage === 1}
                        className="h-8 w-8 p-0"
                      >
                        <ChevronLeft className="h-4 w-4 text-[#999999]" />
                      </Button>

                      {renderPaginationButtons()}

                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handlePageChange(currentPage + 1)}
                        disabled={currentPage >= totalPages}
                        className="h-8 w-8 p-0"
                      >
                        <ChevronRight className="h-4 w-4 text-[#999999]" />
                      </Button>
                    </div>
                  </div>
                </div>
              </div>

              {/* 报告详情弹窗 */}
              <ReportDetailModal
                open={isReportOpen}
                onOpenChange={setIsReportOpen}
                taskId={taskId}
                resumeId={selectedResumeId}
                resumeName={selectedResumeName}
              />
            </>
          )}
        </div>

        {/* 底部：操作按钮 */}
        <div className="border-t border-[#E8E8E8] px-6 py-4">
          <div className="flex justify-end gap-3">
            <Button
              variant="outline"
              onClick={() => onOpenChange(false)}
              className="px-6 bg-[#F5F5F5] border-0 hover:bg-[#E8E8E8] text-[#666666]"
            >
              关闭
            </Button>
            <Button
              onClick={() => {
                onOpenChange(false);
                onBackToHome?.();
              }}
              className="px-6 bg-primary hover:bg-primary/90 text-white"
            >
              返回首页
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
