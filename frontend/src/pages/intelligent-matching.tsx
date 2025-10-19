import { useState, useEffect, useCallback } from 'react';
import {
  Users,
  Clock,
  CheckCircle2,
  Plus,
  Search,
  ChevronLeft,
  ChevronRight,
  Eye,
  Trash2,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import {
  MatchingTask,
  MatchingStats,
  MatchingFilters,
  Pagination,
  MatchingTaskStatus,
} from '@/types/matching';
// Mock数据已移除，统一使用真实接口数据
// import {
//   getMockMatchingTasks,
//   getMockMatchingStats,
// } from '@/data/mockMatchingData';
import { MatchingDetailDrawer } from '@/components/matching/matching-detail-drawer';
import { SelectJobModal } from '@/components/matching/select-job-modal';
import { SelectResumeModal } from '@/components/matching/select-resume-modal';
import {
  ConfigWeightModal,
  WeightConfig,
} from '@/components/matching/config-weight-modal';
import { MatchingProcessModal } from '@/components/matching/matching-process-modal';
import { MatchingResultModal } from '@/components/matching/matching-result-modal';
import {
  createScreeningTask,
  startScreeningTask,
  listScreeningTasks,
  deleteScreeningTask,
} from '@/services/screening';
import {
  DimensionWeights,
  ScreeningTaskStatus,
  ScreeningTaskItem,
} from '@/types/screening';
import { listJobProfiles } from '@/services/job-profile';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';

// 状态映射函数 - 将后端状态转换为中文显示
const getStatusLabel = (status: string): string => {
  switch (status) {
    case 'pending':
      return '创建完成';
    case 'in_progress':
      return '匹配中';
    case 'completed':
      return '匹配完成';
    case 'failed':
      return '匹配失败';
    default:
      return '未知状态';
  }
};

// 获取状态筛选选项
const getStatusFilterOptions = (): { value: string; label: string }[] => {
  return [
    { value: 'all', label: '全部状态' },
    { value: 'completed', label: '匹配完成' },
    { value: 'failed', label: '匹配失败' },
  ];
};

export function IntelligentMatchingPage() {
  const [stats, setStats] = useState<MatchingStats>({
    total: 0,
    inProgress: 0,
    completed: 0,
  });
  const [tasks, setTasks] = useState<MatchingTask[]>([]);
  const [pagination, setPagination] = useState<Pagination>({
    current: 1,
    pageSize: 10,
    total: 0,
    totalPages: 0,
  });
  const [filters, setFilters] = useState<MatchingFilters>({
    position: 'all',
    status: 'all',
    keywords: '',
  });
  const [searchInput, setSearchInput] = useState('');
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [isDetailDrawerOpen, setIsDetailDrawerOpen] = useState(false);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isSelectResumeModalOpen, setIsSelectResumeModalOpen] = useState(false);
  const [isConfigWeightModalOpen, setIsConfigWeightModalOpen] = useState(false);
  const [isMatchingProcessModalOpen, setIsMatchingProcessModalOpen] =
    useState(false);
  const [isMatchingResultModalOpen, setIsMatchingResultModalOpen] =
    useState(false);
  const [selectedJobIds, setSelectedJobIds] = useState<string[]>([]);
  const [selectedResumeIds, setSelectedResumeIds] = useState<string[]>([]);
  const [currentTaskId, setCurrentTaskId] = useState<string | null>(null);
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [taskToDelete, setTaskToDelete] = useState<{
    id: string;
    taskId: string;
  } | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [jobPositions, setJobPositions] = useState<
    { id: string; name: string }[]
  >([]); // 岗位列表（仅任务中出现的岗位）
  const [jobProfileMap, setJobProfileMap] = useState<Record<string, string>>(
    {}
  ); // 岗位ID到名称映射

  // 加载岗位列表（用于名称映射）
  const loadJobPositions = useCallback(async () => {
    try {
      const response = await listJobProfiles({ page: 1, page_size: 100 });
      if (response?.items) {
        const map: Record<string, string> = {};
        response.items.forEach((item: { id: string; name: string }) => {
          map[item.id] = item.name;
        });
        setJobProfileMap(map);
      }
    } catch (error) {
      console.error('加载岗位列表失败:', error);
    }
  }, []);

  // 加载任务列表
  const loadTasks = useCallback(async () => {
    try {
      const response = await listScreeningTasks({
        page: pagination.current,
        size: pagination.pageSize,
        status:
          filters.status !== 'all'
            ? (filters.status as ScreeningTaskStatus)
            : undefined,
      });

      const backendTasks: ScreeningTaskItem[] = Array.isArray(response?.tasks)
        ? response.tasks
        : [];
      if (!Array.isArray(response?.tasks)) {
        console.warn(
          'listScreeningTasks 返回 tasks 非数组，使用空数组回退',
          response
        );
      }

      // 从任务数据收集岗位ID集合，用于构建岗位筛选下拉
      const jobIdSet = new Set<string>(
        backendTasks
          .map((t: ScreeningTaskItem) => t.job_position_id)
          .filter(Boolean)
      );
      const positionsFromTasks: { id: string; name: string }[] = Array.from(
        jobIdSet
      ).map((id) => ({
        id,
        name: jobProfileMap[id] || '未知岗位',
      }));
      setJobPositions(positionsFromTasks);

      // 转换API响应到前端格式，并使用岗位映射获取岗位名称
      const tasksWithJobNames: MatchingTask[] = backendTasks.map(
        (task: ScreeningTaskItem) => {
          const jobPositionName =
            jobProfileMap[task.job_position_id] || '未知岗位';

          let frontendStatus: MatchingTaskStatus;
          if (task.status === 'pending' || task.status === 'in_progress') {
            frontendStatus = 'in_progress';
          } else if (task.status === 'completed') {
            frontendStatus = 'completed';
          } else {
            frontendStatus = 'failed';
          }

          return {
            id: task.id,
            taskId: task.task_id,
            jobPositions: [jobPositionName],
            resumeCount: task.resume_count,
            status: frontendStatus,
            creator: task.creator_name || '',
            createdAt: task.created_at,
          } as MatchingTask;
        }
      );

      // 应用前端筛选：岗位与创建人
      const posId = filters.position || 'all';
      const keyword = (filters.keywords || '').trim();
      const filteredTasks = tasksWithJobNames.filter((t) => {
        const matchesPosition =
          posId === 'all'
            ? true
            : t.jobPositions.includes(jobProfileMap[posId] || '');
        const matchesCreator =
          keyword === '' || (t.creator || '').includes(keyword);
        return matchesPosition && matchesCreator;
      });

      // 按创建时间倒序
      filteredTasks.sort((a, b) => (b.createdAt || 0) - (a.createdAt || 0));

      setTasks(filteredTasks);

      const pageSize = pagination.pageSize;
      const totalVal = filteredTasks.length; // 使用筛选后的数量
      const totalPages = Math.ceil((totalVal || 0) / pageSize);

      // 使用函数式更新避免依赖整个 pagination 对象
      setPagination((prev) => ({
        current: prev.current,
        pageSize: prev.pageSize,
        total: totalVal || 0,
        totalPages: totalPages || 0,
      }));

      const statsData = {
        total:
          typeof response?.total === 'number'
            ? response.total
            : backendTasks.length,
        inProgress: backendTasks.reduce(
          (sum: number, task: ScreeningTaskItem) =>
            sum + (task.resume_count || 0),
          0
        ),
        completed: backendTasks.filter(
          (t: ScreeningTaskItem) => t.status === 'completed'
        ).length,
      };
      setStats(statsData);
    } catch (error) {
      console.error('加载任务列表失败:', error);
      setTasks([]);
      setStats({ total: 0, inProgress: 0, completed: 0 });
      // 使用函数式更新
      setPagination((prev) => ({ ...prev, total: 0, totalPages: 0 }));
    }
    // 注意：这里只依赖 pagination.current 和 pagination.pageSize，避免无限循环
    // 因为 setPagination 会更新 pagination 对象，如果依赖整个 pagination 会导致无限循环
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pagination.current, pagination.pageSize, filters, jobProfileMap]);

  // 首次加载岗位名称映射（仅执行一次）
  useEffect(() => {
    loadJobPositions();
  }, [loadJobPositions]);

  // 加载任务数据与应用筛选（随分页与筛选条件、岗位映射变化刷新）
  useEffect(() => {
    loadTasks();
  }, [loadTasks]);

  // 处理搜索
  const handleSearch = () => {
    setFilters({
      ...filters,
      keywords: searchInput,
    });
    setPagination({ ...pagination, current: 1 });
  };

  // 处理回车搜索
  const handleSearchKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  // 处理分页
  const handlePageChange = (page: number) => {
    setPagination({ ...pagination, current: page });
  };

  // 当筛选器变更时，重置到第一页
  useEffect(() => {
    setPagination((prev) => ({ ...prev, current: 1 }));
  }, [filters.status, filters.keywords]);

  // 处理选择岗位完成
  const handleJobsSelected = (jobIds: string[]) => {
    console.log('选择的岗位IDs:', jobIds);
    setSelectedJobIds(jobIds);
    setIsCreateModalOpen(false);
    setIsSelectResumeModalOpen(true);
  };

  // 处理选择简历完成
  const handleResumesSelected = (resumeIds: string[]) => {
    console.log('选择的岗位IDs:', selectedJobIds);
    console.log('选择的简历IDs:', resumeIds);
    setSelectedResumeIds(resumeIds);
    setIsSelectResumeModalOpen(false);
    setIsConfigWeightModalOpen(true);
  };

  // 处理从选择简历返回到选择岗位
  const handleBackToSelectJob = () => {
    setIsSelectResumeModalOpen(false);
    setIsCreateModalOpen(true);
  };

  // 处理权重配置完成
  const handleWeightConfigured = async (weights: WeightConfig) => {
    console.log('选择的岗位IDs:', selectedJobIds);
    console.log('选择的简历IDs:', selectedResumeIds);
    console.log('配置的权重:', weights);

    try {
      // 构建维度权重对象
      const dimensionWeights: DimensionWeights = {
        basic: weights.basicInfo / 100,
        responsibility: weights.responsibilities / 100,
        skill: weights.skills / 100,
        education: weights.education / 100,
        experience: weights.experience / 100,
        industry: weights.industry / 100,
      };

      // 创建筛选任务
      const response = await createScreeningTask({
        job_position_id: selectedJobIds[0], // 选择第一个岗位ID
        resume_ids: selectedResumeIds,
        dimension_weights: dimensionWeights,
      });

      console.log('任务创建成功，任务ID:', response.task_id);
      setCurrentTaskId(response.task_id);

      // 启动筛选任务
      await startScreeningTask(response.task_id);
      console.log('任务启动成功');

      // 进入匹配处理界面
      setIsConfigWeightModalOpen(false);
      setIsMatchingProcessModalOpen(true);
    } catch (error) {
      console.error('创建或启动任务失败:', error);
      alert('创建任务失败，请重试');
    }
  };

  // 处理从权重配置返回到选择简历
  const handleBackToSelectResume = () => {
    setIsConfigWeightModalOpen(false);
    setIsSelectResumeModalOpen(true);
  };

  // 处理从匹配处理返回到权重配置
  const handleBackToConfigWeight = () => {
    setIsMatchingProcessModalOpen(false);
    setIsConfigWeightModalOpen(true);
  };

  // 处理删除任务
  const handleDeleteTask = async () => {
    if (!taskToDelete) return;

    setDeleting(true);
    try {
      await deleteScreeningTask(taskToDelete.id);
      // 删除成功后重新加载任务列表
      await loadTasks();
      setIsDeleteConfirmOpen(false);
      setTaskToDelete(null);
    } catch (error) {
      console.error('删除任务失败:', error);
      alert('删除任务失败，请重试');
    } finally {
      setDeleting(false);
    }
  };

  // 获取状态样式
  const getStatusStyle = (status: string) => {
    switch (status) {
      case 'completed':
        return {
          bg: 'bg-green-50',
          text: 'text-green-700',
          border: 'border-green-200',
          label: '匹配完成',
          icon: CheckCircle2,
        };
      case 'running':
      case 'in_progress':
        return {
          bg: 'bg-orange-50',
          text: 'text-orange-700',
          border: 'border-orange-200',
          label: '匹配中',
          icon: Clock,
        };
      case 'pending':
        return {
          bg: 'bg-blue-50',
          text: 'text-blue-700',
          border: 'border-blue-200',
          label: '创建完成',
          icon: CheckCircle2,
        };
      case 'failed':
        return {
          bg: 'bg-red-50',
          text: 'text-red-700',
          border: 'border-red-200',
          label: '匹配失败',
          icon: CheckCircle2,
        };
      default:
        return {
          bg: 'bg-gray-50',
          text: 'text-gray-700',
          border: 'border-gray-200',
          label: '未知',
          icon: CheckCircle2,
        };
    }
  };

  // 渲染分页按钮
  const renderPaginationButtons = () => {
    const pages = [];
    const maxVisiblePages = 5;
    const halfVisible = Math.floor(maxVisiblePages / 2);

    let startPage = Math.max(1, pagination.current - halfVisible);
    let endPage = Math.min(
      pagination.totalPages,
      pagination.current + halfVisible
    );

    // 调整起始页和结束页，确保显示足够的页码
    if (endPage - startPage + 1 < maxVisiblePages) {
      if (startPage === 1) {
        endPage = Math.min(
          pagination.totalPages,
          startPage + maxVisiblePages - 1
        );
      } else if (endPage === pagination.totalPages) {
        startPage = Math.max(1, endPage - maxVisiblePages + 1);
      }
    }

    for (let i = startPage; i <= endPage; i++) {
      pages.push(
        <Button
          key={i}
          variant="outline"
          size="sm"
          onClick={() => handlePageChange(i)}
          className={cn(
            'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0 text-sm font-normal text-[#374151] pagination-btn',
            i === pagination.current &&
              'border-primary bg-primary text-primary-foreground'
          )}
        >
          {i}
        </Button>
      );
    }

    return pages;
  };

  return (
    <div className="flex h-full flex-col gap-6 px-6 pb-6 pt-6 page-content">
      {/* 页面标题和操作区域 */}
      <div className="flex items-center justify-between rounded-xl border border-[#E5E7EB] bg-white px-6 py-5 shadow-sm transition-all duration-200 hover:shadow-md">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">智能匹配</h1>
          <p className="text-sm text-muted-foreground">
            创建匹配任务，智能推荐合适的候选人
          </p>
        </div>
        <Button
          className="gap-2 rounded-lg px-4 py-2 shadow-sm btn-primary"
          onClick={() => setIsCreateModalOpen(true)}
        >
          <Plus className="h-4 w-4" />
          创建任务
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-3 gap-6">
        {/* 匹配总数 */}
        <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm card-hover">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">匹配任务总数</p>
              <p className="text-2xl font-bold text-blue-600 mt-1">
                {stats.total}
              </p>
            </div>
            <div className="rounded-lg bg-blue-50 p-3 transition-transform duration-200 hover:scale-110">
              <Users className="h-5 w-5 text-blue-600" />
            </div>
          </div>
        </div>

        {/* 匹配简历总数 */}
        <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm card-hover">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">匹配简历总数</p>
              <p className="text-2xl font-bold text-orange-600 mt-1">
                {stats.inProgress}
              </p>
            </div>
            <div className="rounded-lg bg-orange-50 p-3 transition-transform duration-200 hover:scale-110">
              <Clock className="h-5 w-5 text-orange-600" />
            </div>
          </div>
        </div>

        {/* 匹配已完成 */}
        <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm card-hover">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">匹配已完成</p>
              <p className="text-2xl font-bold text-green-600 mt-1">
                {stats.completed}
              </p>
            </div>
            <div className="rounded-lg bg-green-50 p-3 transition-transform duration-200 hover:scale-110">
              <CheckCircle2 className="h-5 w-5 text-green-600" />
            </div>
          </div>
        </div>
      </div>

      {/* 筛选和搜索区域 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            {/* 状态筛选 */}
            <Select
              value={filters.status}
              onValueChange={(value) =>
                setFilters({ ...filters, status: value })
              }
            >
              <SelectTrigger className="w-32">
                <SelectValue placeholder="全部状态" />
              </SelectTrigger>
              <SelectContent>
                {getStatusFilterOptions().map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* 岗位筛选 */}
            <Select
              value={filters.position}
              onValueChange={(value) =>
                setFilters({ ...filters, position: value })
              }
            >
              <SelectTrigger className="w-32">
                <SelectValue placeholder="全部岗位" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部岗位</SelectItem>
                {jobPositions.map((position) => (
                  <SelectItem key={position.id} value={position.id}>
                    {position.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* 搜索区域 */}
          <div className="flex items-center gap-3">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400 transition-colors" />
              <Input
                placeholder="可根据创建人搜索"
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                className="pl-10 w-64 input-focus"
                onKeyPress={handleSearchKeyPress}
              />
            </div>
            <Button onClick={handleSearch} className="gap-2 btn-secondary">
              <Search className="h-4 w-4" />
              搜索
            </Button>
          </div>
        </div>
      </div>

      {/* 任务列表表格 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white shadow-sm overflow-hidden">
        {/* 表格标题 */}
        <div className="border-b border-[#E5E7EB] px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">
            匹配任务列表 ({tasks.length})
          </h3>
        </div>

        {/* 表格内容 */}
        <div className="overflow-x-auto">
          <table className="w-full table-auto min-w-max">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                  任务序号
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                  匹配岗位
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                  简历数
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                  状态
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                  创建人
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                  创建时间
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {tasks.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-6 py-16">
                    <div className="flex flex-col items-center justify-center space-y-3">
                      <div className="text-gray-400">
                        <svg
                          className="h-12 w-12 mx-auto"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={1}
                            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                          />
                        </svg>
                      </div>
                      <div className="text-sm text-gray-500">
                        暂无匹配任务数据
                      </div>
                      <div className="text-xs text-gray-400">
                        请尝试调整搜索条件或创建新的匹配任务
                      </div>
                    </div>
                  </td>
                </tr>
              ) : (
                tasks.map((task, index) => {
                  const statusStyle = getStatusStyle(task.status);

                  return (
                    <tr key={task.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 text-sm font-medium text-gray-900 whitespace-nowrap">
                        {index +
                          1 +
                          (pagination.current - 1) * pagination.pageSize}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500">
                        <div className="max-w-xs break-words">
                          {task.jobPositions.join(', ')}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500 whitespace-nowrap">
                        {task.resumeCount}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={cn(
                            'inline-flex items-center rounded-full px-3 py-1 text-xs font-medium border transition-all duration-200 hover:scale-105',
                            statusStyle.bg,
                            statusStyle.text,
                            statusStyle.border
                          )}
                        >
                          {getStatusLabel(task.status)}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500 whitespace-nowrap">
                        {task.creator}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500 whitespace-nowrap">
                        {new Date(task.createdAt * 1000).toLocaleDateString(
                          'zh-CN'
                        )}
                      </td>
                      <td className="px-6 py-4 text-sm font-medium whitespace-nowrap">
                        <div className="flex items-center gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0 text-gray-400 hover:text-gray-600 action-btn"
                            onClick={() => {
                              setSelectedTaskId(task.taskId);
                              setIsDetailDrawerOpen(true);
                            }}
                            title="查看详情"
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0 text-gray-400 hover:text-red-600 action-btn"
                            onClick={() => {
                              setTaskToDelete({
                                id: task.id,
                                taskId: task.taskId,
                              });
                              setIsDeleteConfirmOpen(true);
                            }}
                            title="删除任务"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* 分页 */}
        <div className="border-t border-[#E5E7EB] px-6 py-6">
          <div className="flex items-center justify-between">
            <div className="text-sm text-[#6B7280]">
              显示
              <span className="text-[#6B7280]">
                {tasks.length > 0
                  ? (pagination.current - 1) * pagination.pageSize + 1
                  : 0}
              </span>{' '}
              到{' '}
              <span className="text-[#6B7280]">
                {tasks.length > 0
                  ? Math.min(
                      pagination.current * pagination.pageSize,
                      pagination.total
                    )
                  : 0}
              </span>{' '}
              条，共 <span className="text-[#6B7280]">{pagination.total}</span>{' '}
              条结果
            </div>

            {/* 分页按钮 - 只在有数据时显示 */}
            {tasks.length > 0 && (
              <div className="flex flex-wrap items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(pagination.current - 1)}
                  disabled={pagination.current === 1}
                  className={cn(
                    'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0 pagination-btn',
                    pagination.current === 1 && 'opacity-50'
                  )}
                >
                  <ChevronLeft className="h-4 w-4 text-[#6B7280]" />
                </Button>

                {renderPaginationButtons()}

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(pagination.current + 1)}
                  disabled={pagination.current >= pagination.totalPages}
                  className={cn(
                    'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0 pagination-btn',
                    pagination.current >= pagination.totalPages && 'opacity-50'
                  )}
                >
                  <ChevronRight className="h-4 w-4 text-[#6B7280]" />
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 匹配详情抽屉 */}
      <MatchingDetailDrawer
        open={isDetailDrawerOpen}
        onOpenChange={setIsDetailDrawerOpen}
        taskId={selectedTaskId}
      />

      {/* 创建任务模态框 - 选择岗位 */}
      <SelectJobModal
        open={isCreateModalOpen}
        onOpenChange={setIsCreateModalOpen}
        onNext={handleJobsSelected}
      />

      {/* 创建任务模态框 - 选择简历 */}
      <SelectResumeModal
        open={isSelectResumeModalOpen}
        onOpenChange={setIsSelectResumeModalOpen}
        onNext={handleResumesSelected}
        onPrevious={handleBackToSelectJob}
      />

      {/* 创建任务模态框 - 权重配置 */}
      <ConfigWeightModal
        open={isConfigWeightModalOpen}
        onOpenChange={setIsConfigWeightModalOpen}
        onNext={handleWeightConfigured}
        onPrevious={handleBackToSelectResume}
      />

      {/* 创建任务模态框 - 匹配处理中 */}
      <MatchingProcessModal
        open={isMatchingProcessModalOpen}
        onOpenChange={setIsMatchingProcessModalOpen}
        onPrevious={handleBackToConfigWeight}
        onComplete={() => {
          setIsMatchingProcessModalOpen(false);
          setIsMatchingResultModalOpen(true);
          // 刷新任务列表，确保创建完成后数据展示，并重置到第一页
          setPagination((prev) => ({ ...prev, current: 1 }));
          loadTasks();
        }}
        selectedJobCount={selectedJobIds.length}
        selectedResumeCount={selectedResumeIds.length}
        taskId={currentTaskId}
      />

      {/* 匹配结果模态框 */}
      <MatchingResultModal
        open={isMatchingResultModalOpen}
        onOpenChange={setIsMatchingResultModalOpen}
        onBackToHome={() => {
          setIsMatchingResultModalOpen(false);
          // 返回主页面后，重置到第一页以显示最新任务
          setPagination((prev) => ({ ...prev, current: 1 }));
          loadTasks();
        }}
        taskId={currentTaskId}
      />

      {/* 删除确认对话框 */}
      <ConfirmDialog
        open={isDeleteConfirmOpen}
        onOpenChange={setIsDeleteConfirmOpen}
        title="确认删除"
        description={
          taskToDelete ? (
            <div>
              确定要删除任务{' '}
              <span className="font-semibold text-gray-900">
                {taskToDelete.taskId}
              </span>{' '}
              吗？此操作不可恢复。
            </div>
          ) : (
            ''
          )
        }
        confirmText="确定"
        cancelText="取消"
        onConfirm={handleDeleteTask}
        variant="destructive"
        loading={deleting}
      />
    </div>
  );
}
