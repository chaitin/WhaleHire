import { useState, useEffect } from 'react';
import {
  User,
  ChevronLeft,
  ChevronRight,
  FileText,
  Download,
  Award,
  TrendingUp,
  BarChart3,
  Clock,
  CheckCircle2,
} from 'lucide-react';
import { Sheet, SheetContent } from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { MatchingTaskDetail } from '@/types/matching';
import { ReportDetailModal } from './report-detail-modal';
import { getScreeningTask } from '@/services/screening';
import { getResumeDetail } from '@/services/resume';
import { getJobProfile } from '@/services/job-profile';

interface MatchingDetailDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  taskId: string | null;
}

// 计算匹配等级（与匹配结果页卡一致）
const calcMatchLevel = (score: number) => {
  if (score >= 85) return 'excellent';
  if (score >= 70) return 'good';
  if (score >= 50) return 'fair';
  return 'poor';
};

// 获取匹配等级颜色（用户要求的颜色）
const getMatchLevelColor = (level: string) => {
  switch (level) {
    case 'excellent':
      return 'bg-[#FDE7E9] text-[#EF4444]'; // 浅红色
    case 'good':
      return 'bg-[#FFF3E6] text-[#F59E0B]'; // 浅橙色
    case 'fair':
      return 'bg-[#FEF3C7] text-[#F59E0B]'; // 浅黄色
    case 'poor':
      return 'bg-[#F3F4F6] text-[#6B7280]'; // 灰色
    default:
      return 'bg-[#F3F4F6] text-[#6B7280]';
  }
};

// 获取匹配等级标签
const getLevelLabel = (level: string) => {
  switch (level) {
    case 'excellent':
      return '非常匹配';
    case 'good':
      return '高匹配';
    case 'fair':
      return '一般匹配';
    case 'poor':
      return '低匹配';
    default:
      return '低匹配';
  }
};

export function MatchingDetailDrawer({
  open,
  onOpenChange,
  taskId,
}: MatchingDetailDrawerProps) {
  const [taskDetail, setTaskDetail] = useState<MatchingTaskDetail | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [isReportModalOpen, setIsReportModalOpen] = useState(false);
  const [selectedResumeId, setSelectedResumeId] = useState<string | null>(null);
  const [selectedResumeName, setSelectedResumeName] = useState<string>('');
  const [overview, setOverview] = useState({
    total: 0,
    excellent: 0,
    good: 0,
    fair: 0,
    poor: 0,
  });

  // 加载任务详情（真实接口）
  useEffect(() => {
    if (!taskId || !open) return;

    const fetchDetail = async () => {
      try {
        const resp = await getScreeningTask(taskId);

        // 计算分页信息
        const total = resp.resumes.length;
        const pageSize = 3;
        const totalPages = Math.ceil(total / pageSize);
        const start = (currentPage - 1) * pageSize;
        const end = start + pageSize;

        // 构造任务头信息（根据 job_position_id 获取岗位名称）
        const jobProfileName = await (async () => {
          try {
            const profile = await getJobProfile(resp.task.job_position_id);
            return profile?.name || resp.task.job_position_name;
          } catch (e) {
            console.error('获取岗位画像失败:', e);
            return resp.task.job_position_name;
          }
        })();

        const header: MatchingTaskDetail = {
          id: resp.task.id,
          taskId: resp.task.task_id,
          jobPositions: [jobProfileName],
          resumeCount: total,
          status:
            resp.task.status === 'completed'
              ? 'completed'
              : resp.task.status === 'failed'
                ? 'failed'
                : 'in_progress',
          creator: resp.task.creator_name || resp.task.created_by,
          createdAt: resp.task.created_at,
          results: [],
          resultsPagination: {
            current: currentPage,
            pageSize,
            total,
            totalPages,
          },
        };

        // 当前页结果映射
        const sortedResumes = [...resp.resumes].sort(
          (a, b) => (b.score || 0) - (a.score || 0)
        );
        const pageResumes = sortedResumes.slice(start, end);
        const results = await Promise.all(
          pageResumes.map(async (r) => {
            // 获取简历详情以补充信息（兼容后端字段）
            let resumeName = r.resume_name;
            let highestEducation = '';
            let yearsExperience = 0;
            try {
              const detail = await getResumeDetail(r.resume_id);
              resumeName = detail.name || resumeName;
              highestEducation = detail.highest_education || '';
              yearsExperience = detail.years_experience || 0;
            } catch (e) {
              console.warn('获取简历详情失败，使用任务返回的名称作为回退:', e);
            }

            return {
              id: r.resume_id,
              resume: {
                id: r.resume_id,
                name: resumeName,
                age: 0,
                education: highestEducation || '—',
                experience: yearsExperience ? `${yearsExperience}年` : '—',
              },
              job: {
                id: resp.task.job_position_id,
                title: jobProfileName,
                department: '—',
                jobId: resp.task.job_position_id,
              },
              matchScore: r.score || 0,
            };
          })
        );

        // 概览统计数据（与匹配结果页卡一致）
        const allResumes = resp.resumes || [];
        const totalMatches = allResumes.length;
        const levelCounts = {
          excellent: allResumes.filter(
            (r) => calcMatchLevel(r.score || 0) === 'excellent'
          ).length,
          good: allResumes.filter(
            (r) => calcMatchLevel(r.score || 0) === 'good'
          ).length,
          fair: allResumes.filter(
            (r) => calcMatchLevel(r.score || 0) === 'fair'
          ).length,
          poor: allResumes.filter(
            (r) => calcMatchLevel(r.score || 0) === 'poor'
          ).length,
        };
        setOverview({
          total: totalMatches,
          excellent: levelCounts.excellent,
          good: levelCounts.good,
          fair: levelCounts.fair,
          poor: levelCounts.poor,
        });

        setTaskDetail({
          ...header,
          results,
        });
      } catch (err) {
        console.error('加载任务详情失败:', err);
        setTaskDetail(null);
      }
    };

    fetchDetail();
  }, [taskId, open, currentPage]);

  // 重置页码当打开新任务时
  useEffect(() => {
    if (open && taskId) {
      setCurrentPage(1);
    }
  }, [open, taskId]);

  if (!taskDetail) return null;

  // 获取状态样式与图标
  const getStatusStyle = (status: string) => {
    switch (status) {
      case 'completed':
        return {
          text: 'text-green-700',
          bg: 'bg-green-100',
          icon: CheckCircle2,
          label: '匹配完成',
        };
      case 'in_progress':
        return {
          text: 'text-orange-700',
          bg: 'bg-orange-100',
          icon: Clock,
          label: '匹配中',
        };
      case 'failed':
        return {
          text: 'text-red-700',
          bg: 'bg-red-100',
          icon: CheckCircle2,
          label: '匹配失败',
        };
      default:
        return {
          text: 'text-gray-700',
          bg: 'bg-gray-100',
          icon: CheckCircle2,
          label: '未知',
        };
    }
  };

  const statusStyle = getStatusStyle(taskDetail.status);
  const StatusIcon = statusStyle.icon;

  // 处理分页
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  // 渲染分页按钮
  const renderPaginationButtons = () => {
    const pages = [];
    const totalPages = taskDetail.resultsPagination.totalPages;
    const maxVisiblePages = 5;

    let startPage = Math.max(1, currentPage - 2);
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
          onClick={() => handlePageChange(i)}
          className={cn(
            'h-[30px] w-[33px] rounded-md border border-[#C9CDD4] bg-white p-0 text-sm text-[#374151] hover:border-[#6B7280] hover:text-[#111827]',
            i === currentPage &&
              'border-[#6B7280] bg-[#6B7280] text-white hover:bg-[#6B7280] hover:text-white'
          )}
        >
          {i}
        </Button>
      );
    }

    return pages;
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-[960px] p-0 overflow-y-auto">
        {/* 头部 */}
        <div className="border-b border-[#E5E6EB] px-6 py-4">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <div className="flex items-center gap-3 mb-2">
                <h3 className="text-lg font-medium text-[#1D2129] flex items-center gap-3">
                  <span>
                    匹配任务 {taskDetail.taskId} · {taskDetail.jobPositions[0]}
                  </span>
                  <span
                    className={cn(
                      'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs',
                      statusStyle.bg,
                      statusStyle.text
                    )}
                  >
                    <StatusIcon className="h-4 w-4" />
                    {statusStyle.label}
                  </span>
                </h3>
              </div>
              <p className="text-sm text-[#4E5969]">
                创建人：{taskDetail.creator || '—'} · 创建时间：
                {taskDetail.createdAt
                  ? new Date(taskDetail.createdAt)
                      .toLocaleString('zh-CN', {
                        year: 'numeric',
                        month: '2-digit',
                        day: '2-digit',
                        hour: '2-digit',
                        minute: '2-digit',
                        second: '2-digit',
                        hour12: false,
                      })
                      .replace(/\//g, '-')
                  : '—'}
              </p>
            </div>
          </div>
        </div>

        {/* 整体数据概览 */}
        <div className="px-6 py-6 border-b border-[#E5E6EB]">
          <div className="bg-[#FAFAFA] rounded-lg p-5">
            <h3 className="mb-4 text-base font-semibold text-[#333333]">
              整体数据概览
            </h3>
            <div className="grid grid-cols-5 gap-4">
              {/* 匹配总数：蓝色 */}
              <div className="flex items-center gap-3 rounded-lg border border-[#E8E8E8] bg-white p-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#E6F0FF]">
                  <FileText className="h-5 w-5 text-[#3B82F6]" />
                </div>
                <div className="flex flex-col gap-0.5">
                  <div className="text-2xl font-bold leading-tight text-[#333333]">
                    {overview.total}
                  </div>
                  <div className="text-xs text-[#666666]">匹配总数</div>
                </div>
              </div>

              {/* 非常匹配：浅红色 */}
              <div className="flex items-center gap-3 rounded-lg border border-[#E8E8E8] bg-white p-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#FDE7E9]">
                  <TrendingUp className="h-5 w-5 text-[#EF4444]" />
                </div>
                <div className="flex flex-col gap-0.5">
                  <div className="text-2xl font-bold leading-tight text-[#333333]">
                    {overview.excellent}
                  </div>
                  <div className="text-xs text-[#666666]">非常匹配</div>
                </div>
              </div>

              {/* 高匹配：浅橙色 */}
              <div className="flex items-center gap-3 rounded-lg border border-[#E8E8E8] bg-white p-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#FFF3E6]">
                  <BarChart3 className="h-5 w-5 text-[#F59E0B]" />
                </div>
                <div className="flex flex-col gap-0.5">
                  <div className="text-2xl font-bold leading-tight text-[#333333]">
                    {overview.good}
                  </div>
                  <div className="text-xs text-[#666666]">高匹配</div>
                </div>
              </div>

              {/* 一般匹配：浅绿色 */}
              <div className="flex items-center gap-3 rounded-lg border border-[#E8E8E8] bg-white p-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#D1FAE5]">
                  <Award className="h-5 w-5 text-[#7bb8ff]" />
                </div>
                <div className="flex flex-col gap-0.5">
                  <div className="text-2xl font-bold leading-tight text-[#333333]">
                    {overview.fair}
                  </div>
                  <div className="text-xs text-[#666666]">一般匹配</div>
                </div>
              </div>

              {/* 低匹配：浅灰色 */}
              <div className="flex items-center gap-3 rounded-lg border border-[#E8E8E8] bg-white p-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-[#F3F4F6]">
                  <BarChart3 className="h-5 w-5 text-[#6B7280]" />
                </div>
                <div className="flex flex-col gap-0.5">
                  <div className="text-2xl font-bold leading-tight text-[#333333]">
                    {overview.poor}
                  </div>
                  <div className="text-xs text-[#666666]">低匹配</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* 匹配结果 */}
        <div className="border-b border-[#E5E6EB] px-6 py-6">
          <h4 className="text-lg font-medium text-[#1D2129] mb-6">匹配结果</h4>

          {/* 表格 */}
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-[#F9FAFB]">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase tracking-wider text-[#4E5969]">
                    简历信息
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase tracking-wider text-[#4E5969]">
                    岗位信息
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase tracking-wider text-[#4E5969]">
                    <div className="flex items-center gap-1">
                      <span>匹配分数</span>
                      <Award className="h-4 w-4 text-[#4E5969]" />
                    </div>
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-medium uppercase tracking-wider text-[#4E5969]">
                    匹配等级
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-medium uppercase tracking-wider text-[#4E5969]">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#E5E6EB] bg-white">
                {taskDetail.results.map((result) => {
                  const matchLevel = calcMatchLevel(result.matchScore);
                  return (
                    <tr key={result.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[#E8E8E8]">
                            <User className="h-5 w-5 text-[#666666]" />
                          </div>
                          <div>
                            <div className="text-sm font-medium text-[#1D2129]">
                              {result.resume.name}
                            </div>
                            <div className="text-sm text-[#4E5969]">
                              {result.resume.education} ·{' '}
                              {result.resume.experience}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div>
                          <div className="text-sm font-medium text-[#1D2129] mb-1">
                            {result.job.title}
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2">
                          <Award className="h-4 w-4 text-[#7bb8ff]" />
                          <span className="text-sm font-medium text-[#1D2129]">
                            {result.matchScore}分
                          </span>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className={cn(
                            'inline-flex rounded-full px-2.5 py-1 text-xs font-medium',
                            getMatchLevelColor(matchLevel)
                          )}
                        >
                          {getLevelLabel(matchLevel)}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <div className="flex items-center justify-end gap-4">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0 text-gray-400 hover:text-gray-600"
                            onClick={() => {
                              setSelectedResumeId(result.id);
                              setSelectedResumeName(result.resume.name);
                              setIsReportModalOpen(true);
                            }}
                            title="查看报告"
                          >
                            <FileText className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0 text-gray-300 cursor-not-allowed"
                            disabled
                            title="下载功能暂不可用"
                          >
                            <Download className="h-4 w-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* 分页 */}
          <div className="mt-6 flex items-center justify-between">
            <div className="text-sm text-[#1D2129]">
              显示第{' '}
              <span className="font-normal text-[#1D2129]">
                {(currentPage - 1) * 3 + 1}
              </span>{' '}
              到{' '}
              <span className="font-normal text-[#1D2129]">
                {Math.min(currentPage * 3, taskDetail.resultsPagination.total)}
              </span>{' '}
              条，共{' '}
              <span className="font-normal text-[#1D2129]">
                {taskDetail.resultsPagination.total}
              </span>{' '}
              条结果
            </div>

            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="sm"
                onClick={() => handlePageChange(Math.max(1, currentPage - 1))}
                className={cn(
                  'h-[30px] w-[33px] rounded-md border border-[#C9CDD4] bg-white p-0',
                  currentPage === 1 && 'opacity-50'
                )}
                disabled={currentPage === 1}
              >
                <ChevronLeft className="h-4 w-4 text-[#6B7280]" />
              </Button>

              {renderPaginationButtons()}

              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  handlePageChange(
                    Math.min(
                      taskDetail.resultsPagination.totalPages,
                      currentPage + 1
                    )
                  )
                }
                className={cn(
                  'h-[30px] w-[33px] rounded-md border border-[#C9CDD4] bg-white p-0',
                  currentPage >= taskDetail.resultsPagination.totalPages &&
                    'opacity-50'
                )}
                disabled={
                  currentPage >= taskDetail.resultsPagination.totalPages
                }
              >
                <ChevronRight className="h-4 w-4 text-[#6B7280]" />
              </Button>
            </div>
          </div>
        </div>

        {/* 报告详情模态框 */}
        <ReportDetailModal
          open={isReportModalOpen}
          onOpenChange={setIsReportModalOpen}
          resumeId={selectedResumeId}
          resumeName={selectedResumeName}
          taskId={taskId}
        />
      </SheetContent>
    </Sheet>
  );
}
