import { useState, useEffect, useCallback } from 'react';
import {
  X,
  Search,
  ChevronLeft,
  ChevronRight,
  Briefcase,
  User,
  Scale,
  Settings,
  BarChart3,
} from 'lucide-react';
import { Dialog, DialogContent, DialogTitle } from '@/ui/dialog';
import { Button } from '@/ui/button';
import { Input } from '@/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/ui/select';
import { Checkbox } from '@/ui/checkbox';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/ui/tooltip';
import { cn } from '@/lib/utils';
import { Resume, ResumeStatus } from '@/types/resume';
import { getResumeList } from '@/services/resume';
import { listJobProfiles } from '@/services/job-profile';
import type { JobProfileDetail } from '@/types/job-profile';

interface SelectResumeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onNext: (selectedResumeIds: string[]) => void;
  onPrevious: () => void;
  selectedJobIds: string[]; // 第一步选择的岗位ID列表
}

// 流程步骤配置
const STEPS = [
  { id: 1, name: '选择岗位', icon: Briefcase, active: false },
  { id: 2, name: '选择简历', icon: User, active: true },
  { id: 3, name: '权重配置', icon: Scale, active: false },
  { id: 4, name: '匹配处理', icon: Settings, active: false },
  { id: 5, name: '匹配结果', icon: BarChart3, active: false },
];

export function SelectResumeModal({
  open,
  onOpenChange,
  onNext,
  onPrevious,
  selectedJobIds,
}: SelectResumeModalProps) {
  const [searchKeyword, setSearchKeyword] = useState('');
  // 岗位筛选：默认使用第一步选择的岗位ID（如果只有一个），否则为'all'
  const [jobFilter, setJobFilter] = useState<string>('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [selectedResumeIds, setSelectedResumeIds] = useState<string[]>([]);
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [totalPages, setTotalPages] = useState(0);
  const [totalResults, setTotalResults] = useState(0);
  const [loading, setLoading] = useState(false);
  const [jobProfiles, setJobProfiles] = useState<JobProfileDetail[]>([]); // 岗位画像列表

  // 加载岗位画像列表 - 显示所有岗位供用户选择
  const loadJobProfiles = useCallback(async () => {
    try {
      const response = await listJobProfiles({ page: 1, page_size: 100 });
      // 显示所有岗位,不仅限于第一步选择的岗位
      const profiles = response.items || [];
      console.log('📋 加载岗位画像列表:', profiles.length, '个岗位');
      console.log(
        '📋 岗位详情:',
        profiles.map((p) => ({ id: p.id, name: p.name }))
      );
      setJobProfiles(profiles);
    } catch (error) {
      console.error('加载岗位列表失败:', error);
      setJobProfiles([]);
    }
  }, []);

  // 加载简历数据（服务端分页）- 使用第一步选择的岗位进行筛选
  const loadResumes = useCallback(async () => {
    setLoading(true);
    try {
      // 构建请求参数
      const params: {
        page: number;
        size: number;
        keywords?: string;
        status: string;
        job_position_id?: string; // 后端只支持单个岗位ID筛选
      } = {
        page: currentPage,
        size: pageSize,
        keywords: searchKeyword || undefined,
        status: ResumeStatus.COMPLETED, // 只显示已完成解析的简历
      };

      // 如果有选择的岗位ID,使用它来筛选简历
      // 后端只支持单个job_position_id，所以只取第一个岗位ID
      // 重要：只有当用户明确选择了某个岗位（不是"all"）时，才传递岗位ID进行筛选
      if (jobFilter && jobFilter !== 'all') {
        params.job_position_id = jobFilter;
        console.log('📋 使用jobFilter筛选简历:', jobFilter);
      }
      // 如果用户选择了"全部岗位"，则不传递job_position_id参数，这样会显示所有简历

      console.log('📋 简历列表请求参数:', params);
      const response = await getResumeList(params);

      console.log('📋 简历列表响应:', response);
      console.log('📋 返回的简历数量:', response.resumes?.length);

      // 调试：检查返回的简历关联的岗位ID
      if (response.resumes && response.resumes.length > 0) {
        response.resumes.forEach((resume, index) => {
          console.log(
            `📋 简历${index + 1} [${resume.name}] 关联的岗位:`,
            resume.job_ids,
            resume.job_names
          );
        });
      }

      setResumes(response.resumes || []);
      const totalCount = response.total_count || 0;
      setTotalResults(totalCount);
      setTotalPages(Math.ceil(totalCount / pageSize));
    } catch (error) {
      console.error('加载简历列表失败:', error);
      setResumes([]);
      setTotalResults(0);
      setTotalPages(0);
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, searchKeyword, jobFilter]);

  // 当弹窗打开时，初始化岗位筛选条件；关闭时重置状态
  useEffect(() => {
    if (open) {
      console.log('📋 选择简历弹窗打开，第一步选择的岗位IDs:', selectedJobIds);
      // 如果第一步选择了岗位，将其设置为默认筛选条件
      if (selectedJobIds && selectedJobIds.length === 1) {
        console.log('📋 初始化岗位筛选条件为单个岗位:', selectedJobIds[0]);
        setJobFilter(selectedJobIds[0]);
      } else if (selectedJobIds && selectedJobIds.length > 1) {
        console.log(
          '📋 第一步选择了多个岗位，筛选条件默认为全部:',
          selectedJobIds
        );
        setJobFilter('all');
      } else {
        console.log('📋 第一步未选择岗位，筛选条件默认为全部');
        setJobFilter('all');
      }
    } else {
      // 弹窗关闭时重置状态
      setJobFilter('all');
      setSearchKeyword('');
      setCurrentPage(1);
      setSelectedResumeIds([]);
    }
  }, [open, selectedJobIds]);

  // 加载岗位列表
  useEffect(() => {
    if (open) {
      loadJobProfiles();
    }
  }, [open, loadJobProfiles]);

  // 当弹窗打开且jobFilter初始化完成后，加载简历列表
  useEffect(() => {
    if (open) {
      console.log('📋 触发简历列表加载，当前jobFilter:', jobFilter);
      loadResumes();
    }
  }, [open, jobFilter, loadResumes]);

  // 处理搜索
  const handleSearch = () => {
    setCurrentPage(1);
    loadResumes();
  };

  // 处理复选框切换
  const handleCheckboxChange = (resumeId: string, checked: boolean) => {
    if (checked) {
      setSelectedResumeIds([...selectedResumeIds, resumeId]);
    } else {
      setSelectedResumeIds(selectedResumeIds.filter((id) => id !== resumeId));
    }
  };

  // 处理下一步
  const handleNext = () => {
    onNext(selectedResumeIds);
  };

  // 处理上一步
  const handlePrevious = () => {
    onPrevious();
  };

  // 渲染分页按钮
  const renderPaginationButtons = () => {
    const pages = [];
    const maxVisiblePages = 3;

    if (totalPages <= maxVisiblePages + 2) {
      // 如果总页数较少，显示所有页码
      for (let i = 1; i <= totalPages; i++) {
        pages.push(
          <Button
            key={i}
            variant="outline"
            size="sm"
            onClick={() => setCurrentPage(i)}
            className={cn(
              'h-8 w-8 p-0 text-sm',
              i === currentPage && 'text-white border-primary hover:text-white'
            )}
            style={
              i === currentPage
                ? {
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }
                : undefined
            }
          >
            {i}
          </Button>
        );
      }
    } else {
      // 始终显示第1页
      pages.push(
        <Button
          key={1}
          variant="outline"
          size="sm"
          onClick={() => setCurrentPage(1)}
          className={cn(
            'h-8 w-8 p-0 text-sm',
            1 === currentPage && 'text-white border-primary hover:text-white'
          )}
          style={
            1 === currentPage
              ? {
                  background:
                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }
              : undefined
          }
        >
          1
        </Button>
      );

      // 中间的页码
      if (currentPage > 2) {
        pages.push(
          <Button
            key={2}
            variant="outline"
            size="sm"
            onClick={() => setCurrentPage(2)}
            className={cn(
              'h-8 w-8 p-0 text-sm',
              2 === currentPage && 'text-white border-primary hover:text-white'
            )}
            style={
              2 === currentPage
                ? {
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }
                : undefined
            }
          >
            2
          </Button>
        );
      }

      if (currentPage > 3) {
        pages.push(
          <Button
            key={3}
            variant="outline"
            size="sm"
            onClick={() => setCurrentPage(3)}
            className={cn(
              'h-8 w-8 p-0 text-sm',
              3 === currentPage && 'text-white border-primary hover:text-white'
            )}
            style={
              3 === currentPage
                ? {
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }
                : undefined
            }
          >
            3
          </Button>
        );
      }

      // 省略号
      if (currentPage > 3 && currentPage < totalPages - 1) {
        pages.push(
          <div
            key="ellipsis"
            className="flex items-center justify-center h-8 w-8"
          >
            <span className="text-sm text-[#999999]">...</span>
          </div>
        );
      }

      // 最后一页
      pages.push(
        <Button
          key={totalPages}
          variant="outline"
          size="sm"
          onClick={() => setCurrentPage(totalPages)}
          className={cn(
            'h-8 w-8 p-0 text-sm',
            totalPages === currentPage &&
              'text-white border-primary hover:text-white'
          )}
          style={
            totalPages === currentPage
              ? {
                  background:
                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }
              : undefined
          }
        >
          {totalPages}
        </Button>
      );
    }

    return pages;
  };

  // 格式化时间
  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000)
      .toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
      })
      .replace(/\//g, '-');
  };

  // 格式化工作经验
  const formatExperience = (years?: number) => {
    if (!years) return '-';
    return `${years}年`;
  };

  const renderJobPositionsCell = (resume: Resume) => {
    const jobPositions = resume.job_positions || [];

    if (jobPositions.length === 0) {
      return <span className="text-[#999999]">-</span>;
    }

    if (jobPositions.length === 1) {
      return (
        <span className="text-sm text-[#333333]">
          {jobPositions[0].job_title}
        </span>
      );
    }

    const firstJobTitle = jobPositions[0].job_title;
    const remainingCount = jobPositions.length - 1;

    return (
      <div className="text-sm text-[#333333]">
        {firstJobTitle}等
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="text-primary cursor-pointer underline decoration-dotted">
                {remainingCount}
              </span>
            </TooltipTrigger>
            <TooltipContent className="max-w-xs">
              <div className="text-sm">
                <div className="font-medium mb-1">所有岗位：</div>
                <div className="flex flex-col gap-1">
                  {jobPositions.map((jp, index) => (
                    <div key={index}>{jp.job_title}</div>
                  ))}
                </div>
              </div>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
        个岗位
      </div>
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-w-[920px] p-0 gap-0 bg-white rounded-xl"
        onInteractOutside={(e) => e.preventDefault()}
      >
        <DialogTitle className="sr-only">创建新匹配任务 - 选择简历</DialogTitle>
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
                          ? 'text-[#7bb8ff] font-semibold'
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

          {/* 筛选和搜索 */}
          <div className="bg-[#FAFAFA] rounded-lg p-5 mb-4">
            <div className="flex items-center justify-between gap-3 mb-4">
              {/* 筛选器 - 仅显示岗位筛选 */}
              <div className="flex items-center gap-3">
                <Select value={jobFilter} onValueChange={setJobFilter}>
                  <SelectTrigger className="w-[200px] h-[33.5px] text-sm">
                    <SelectValue placeholder="请选择岗位" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部岗位</SelectItem>
                    {jobProfiles.map((job) => (
                      <SelectItem key={job.id} value={job.id}>
                        {job.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* 搜索框 */}
              <div className="flex items-center">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[#999999]" />
                  <Input
                    placeholder="请输入姓名或技能搜索"
                    value={searchKeyword}
                    onChange={(e) => setSearchKeyword(e.target.value)}
                    className="w-[300px] h-9 pl-10 pr-3 text-sm border-r-0 rounded-r-none"
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') handleSearch();
                    }}
                  />
                </div>
                <Button
                  onClick={handleSearch}
                  className="h-9 px-6 rounded-l-none text-white"
                  style={{
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }}
                >
                  <Search className="h-3.5 w-3.5 mr-1 text-white" />
                  搜索
                </Button>
              </div>
            </div>

            {/* 简历表格 */}
            <div className="bg-white rounded-md border border-[#E8E8E8] overflow-hidden mb-4">
              <table className="w-full">
                <thead className="bg-[#F5F5F5]">
                  <tr className="border-b border-[#E8E8E8]">
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666] w-[42px]">
                      <Checkbox
                        checked={
                          selectedResumeIds.length === resumes.length &&
                          resumes.length > 0
                        }
                        onCheckedChange={(checked) => {
                          if (checked) {
                            setSelectedResumeIds(resumes.map((r) => r.id));
                          } else {
                            setSelectedResumeIds([]);
                          }
                        }}
                      />
                    </th>
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                      姓名
                    </th>
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                      期望岗位
                    </th>
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                      工作经验
                    </th>
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                      学历
                    </th>
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                      上传时间
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {loading ? (
                    <tr>
                      <td
                        colSpan={6}
                        className="px-3 py-8 text-center text-sm text-[#999999]"
                      >
                        加载中...
                      </td>
                    </tr>
                  ) : resumes.length === 0 ? (
                    <tr>
                      <td
                        colSpan={6}
                        className="px-3 py-8 text-center text-sm text-[#999999]"
                      >
                        暂无简历数据
                      </td>
                    </tr>
                  ) : (
                    resumes.map((resume) => (
                      <tr
                        key={resume.id}
                        className="border-b border-[#F0F0F0] last:border-b-0"
                      >
                        <td className="px-3 py-4 text-sm">
                          <Checkbox
                            checked={selectedResumeIds.includes(resume.id)}
                            onCheckedChange={(checked) =>
                              handleCheckboxChange(
                                resume.id,
                                checked as boolean
                              )
                            }
                          />
                        </td>
                        <td className="px-3 py-4 text-sm font-medium text-[#333333]">
                          {resume.name}
                        </td>
                        <td className="px-3 py-4 text-sm text-[#333333]">
                          {renderJobPositionsCell(resume)}
                        </td>
                        <td className="px-3 py-4 text-sm text-[#333333]">
                          {formatExperience(resume.years_experience)}
                        </td>
                        <td className="px-3 py-4 text-sm text-[#333333]">
                          {resume.highest_education || '-'}
                        </td>
                        <td className="px-3 py-4 text-sm text-[#333333]">
                          {formatDate(resume.created_at)}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>

            {/* 分页控件 */}
            <div className="flex items-center justify-between">
              {/* 分页信息和每页条数 */}
              <div className="flex items-center gap-3">
                <div className="text-[13px] text-[#666666]">
                  显示第 {(currentPage - 1) * pageSize + 1} 到{' '}
                  {Math.min(currentPage * pageSize, totalResults)} 条，共{' '}
                  {totalResults} 条结果
                </div>

                {/* 每页条数 */}
                <Select
                  value={pageSize.toString()}
                  onValueChange={(val) => setPageSize(Number(val))}
                >
                  <SelectTrigger className="w-[85.5px] h-[25px] text-[13px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="10">10条/页</SelectItem>
                    <SelectItem value="20">20条/页</SelectItem>
                    <SelectItem value="50">50条/页</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 分页按钮 */}
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                  disabled={currentPage === 1}
                  className={cn(
                    'h-8 w-8 p-0 text-sm',
                    currentPage === 1 && 'opacity-50 cursor-not-allowed'
                  )}
                >
                  <ChevronLeft className="h-4 w-4 text-[#999999]" />
                </Button>

                {renderPaginationButtons()}

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setCurrentPage(Math.min(totalPages, currentPage + 1))
                  }
                  disabled={currentPage >= totalPages}
                  className={cn(
                    'h-8 w-8 p-0 text-sm',
                    currentPage >= totalPages && 'opacity-50 cursor-not-allowed'
                  )}
                >
                  <ChevronRight className="h-4 w-4 text-[#999999]" />
                </Button>
              </div>
            </div>
          </div>
        </div>

        {/* 底部：操作按钮 */}
        <div className="border-t border-[#E8E8E8] px-6 py-4">
          {/* 操作按钮 */}
          <div className="flex justify-end gap-3">
            <Button
              variant="outline"
              onClick={() => onOpenChange(false)}
              className="px-6 bg-[#F5F5F5] border-0 hover:bg-[#E8E8E8] text-[#666666]"
            >
              取消
            </Button>
            <Button
              variant="outline"
              onClick={handlePrevious}
              className="px-6 bg-[#F5F5F5] border-0 hover:bg-[#E8E8E8] text-[#666666]"
            >
              上一步
            </Button>
            <Button
              onClick={handleNext}
              className="px-6 text-white"
              style={{
                background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
              }}
            >
              下一步
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
