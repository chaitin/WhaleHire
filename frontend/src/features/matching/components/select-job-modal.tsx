import { useState, useEffect, useCallback } from 'react';
import { X, Search, ChevronLeft, ChevronRight } from 'lucide-react';
import { Dialog, DialogContent, DialogTitle } from '@/ui/dialog';
import { Button } from '@/ui/button';
import { Input } from '@/ui/input';
import { Checkbox } from '@/ui/checkbox';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/ui/select';
import { cn } from '@/lib/utils';
import { JobProfile } from '@/types/job-profile';
import { listJobProfiles } from '@/services/job-profile';
import { listDepartments, Department } from '@/services/department';

interface SelectJobModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onNext: (selectedJobIds: string[]) => void;
}

// 流程步骤配置
const STEPS = [
  {
    id: 1,
    name: '选择岗位',
    icon: () => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M20 7h-9" />
        <path d="M14 17H5" />
        <circle cx="17" cy="17" r="3" />
        <circle cx="7" cy="7" r="3" />
      </svg>
    ),
    active: true,
  },
  {
    id: 2,
    name: '选择简历',
    icon: () => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
        <circle cx="12" cy="7" r="4" />
      </svg>
    ),
    active: false,
  },
  {
    id: 3,
    name: '权重配置',
    icon: () => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M12 3v18" />
        <path d="m8 15-4-4 4-4" />
        <path d="M16 9l4 4-4 4" />
      </svg>
    ),
    active: false,
  },
  {
    id: 4,
    name: '匹配处理',
    icon: () => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
        <circle cx="12" cy="12" r="3" />
      </svg>
    ),
    active: false,
  },
  {
    id: 5,
    name: '匹配结果',
    icon: () => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <line x1="12" x2="12" y1="20" y2="10" />
        <line x1="18" x2="18" y1="20" y2="4" />
        <line x1="6" x2="6" y1="20" y2="16" />
      </svg>
    ),
    active: false,
  },
];

export function SelectJobModal({
  open,
  onOpenChange,
  onNext,
}: SelectJobModalProps) {
  const [searchKeyword, setSearchKeyword] = useState('');
  const [departmentFilter, setDepartmentFilter] = useState<string>('');
  const [locationFilter, setLocationFilter] = useState<string>('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(20);
  const [selectedJobId, setSelectedJobId] = useState<string>('');
  const [jobs, setJobs] = useState<JobProfile[]>([]);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [locations, setLocations] = useState<string[]>([]);

  // 加载筛选条件选项
  const loadFilterOptions = async () => {
    try {
      // 加载部门列表
      const deptResponse = await listDepartments({ size: 100 });
      setDepartments(deptResponse.items || []);

      // 加载地点列表（从岗位数据中提取唯一地点）
      const jobsResponse = await listJobProfiles({ page: 1, page_size: 1000 });
      const uniqueLocations = Array.from(
        new Set(
          (jobsResponse.items || [])
            .map((job) => job.location)
            .filter((loc): loc is string => !!loc)
        )
      );
      setLocations(uniqueLocations);
    } catch (error) {
      console.error('加载筛选选项失败:', error);
    }
  };

  // 加载岗位数据
  const loadJobs = useCallback(async () => {
    setLoading(true);
    try {
      const response = await listJobProfiles({
        page: currentPage,
        page_size: pageSize,
        department_id: departmentFilter || undefined,
        locations: locationFilter ? [locationFilter] : undefined,
        keyword: searchKeyword || undefined,
      });

      console.log('📋 岗位列表响应:', response);
      console.log('📋 response.items:', response.items);
      console.log('📋 response.total_count:', response.total_count);

      const items = response.items || [];
      setJobs(items);

      // API直接返回total_count
      const totalCount = response.total_count || 0;
      const calculatedTotalPages =
        totalCount > 0 ? Math.ceil(totalCount / pageSize) : 1;

      console.log(
        `📊 总数: ${totalCount}, 每页: ${pageSize}, 总页数: ${calculatedTotalPages}`
      );
      setTotal(totalCount);
      setTotalPages(calculatedTotalPages);
    } catch (error) {
      console.error('加载岗位列表失败:', error);
      setJobs([]);
      setTotal(0);
      setTotalPages(0);
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, departmentFilter, locationFilter, searchKeyword]);

  // 加载部门和地点选项
  useEffect(() => {
    if (open) {
      loadFilterOptions();
    }
  }, [open]);

  // 加载岗位列表
  useEffect(() => {
    if (open) {
      loadJobs();
    }
  }, [open, loadJobs]);

  // 处理搜索
  const handleSearch = () => {
    setCurrentPage(1);
    loadJobs();
  };

  // 处理下一步
  const handleNext = () => {
    if (!selectedJobId) {
      // TODO: 可以添加提示用户选择岗位的 toast 消息
      return;
    }
    // 将选中的岗位ID作为数组传递
    onNext([selectedJobId]);
  };

  // 渲染分页按钮
  const renderPaginationButtons = () => {
    const pages: JSX.Element[] = [];

    // 如果总页数为0,不渲染任何分页按钮
    if (totalPages === 0) {
      return pages;
    }

    const maxVisiblePages = 5;

    let startPage = Math.max(1, currentPage - Math.floor(maxVisiblePages / 2));
    const endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);

    if (endPage - startPage < maxVisiblePages - 1) {
      startPage = Math.max(1, endPage - maxVisiblePages + 1);
    }

    console.log(
      `🔢 渲染分页按钮: startPage=${startPage}, endPage=${endPage}, totalPages=${totalPages}`
    );

    for (let i = startPage; i <= endPage; i++) {
      pages.push(
        <Button
          key={i}
          variant="outline"
          size="sm"
          onClick={() => setCurrentPage(i)}
          className={cn(
            'h-8 w-8 p-0',
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

    return pages;
  };

  // 格式化时间
  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleDateString('zh-CN');
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-w-[920px] p-0 gap-0 bg-white rounded-xl"
        onInteractOutside={(e) => e.preventDefault()}
      >
        <DialogTitle className="sr-only">创建新匹配任务 - 选择岗位</DialogTitle>
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
                          'w-8 h-8 rounded-2xl flex items-center justify-center text-[#999999]',
                          step.active ? 'bg-white' : 'bg-white opacity-60'
                        )}
                      >
                        {step.icon()}
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
              {/* 筛选器 */}
              <div className="flex items-center gap-3">
                <Select
                  value={departmentFilter}
                  onValueChange={setDepartmentFilter}
                >
                  <SelectTrigger className="w-[144px] h-[33.5px] text-sm">
                    <SelectValue placeholder="请选择所属部门" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部部门</SelectItem>
                    {departments.map((dept) => (
                      <SelectItem key={dept.id} value={dept.id}>
                        {dept.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select
                  value={locationFilter}
                  onValueChange={setLocationFilter}
                >
                  <SelectTrigger className="w-[144px] h-[33.5px] text-sm">
                    <SelectValue placeholder="请选择工作地点" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部地点</SelectItem>
                    {locations.map((location) => (
                      <SelectItem key={location} value={location}>
                        {location}
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
                    placeholder="请输入岗位名称搜索"
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

            {/* 岗位表格 */}
            <div className="bg-white rounded-md border border-[#E8E8E8] overflow-hidden mb-4">
              <table className="w-full">
                <thead className="bg-[#F5F5F5]">
                  <tr className="border-b border-[#E8E8E8]">
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666] w-12">
                      选择
                    </th>
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                      岗位名称
                    </th>
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                      所属部门
                    </th>
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                      工作地点
                    </th>
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                      创建时间
                    </th>
                    <th className="px-3 py-3.5 text-left text-[13px] font-medium text-[#666666]">
                      创建人
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
                  ) : jobs.length === 0 ? (
                    <tr>
                      <td
                        colSpan={6}
                        className="px-3 py-8 text-center text-sm text-[#999999]"
                      >
                        暂无岗位数据
                      </td>
                    </tr>
                  ) : (
                    jobs.map((job) => (
                      <tr
                        key={job.id}
                        className="border-b border-[#F0F0F0] last:border-b-0 hover:bg-[#FAFAFA] cursor-pointer transition-colors"
                        onClick={() => setSelectedJobId(job.id)}
                      >
                        <td className="px-3 py-4">
                          <Checkbox
                            checked={selectedJobId === job.id}
                            onCheckedChange={() => setSelectedJobId(job.id)}
                            className="data-[state=checked]:bg-primary data-[state=checked]:border-primary"
                          />
                        </td>
                        <td className="px-3 py-4 text-sm font-medium text-[#333333]">
                          {job.name}
                        </td>
                        <td className="px-3 py-4 text-sm text-[#333333]">
                          {job.department}
                        </td>
                        <td className="px-3 py-4 text-sm text-[#333333]">
                          {job.location || '-'}
                        </td>
                        <td className="px-3 py-4 text-sm text-[#333333]">
                          {formatDate(job.created_at)}
                        </td>
                        <td className="px-3 py-4 text-sm text-[#333333]">
                          {job.creator_name || '-'}
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
                  显示第 {total > 0 ? (currentPage - 1) * pageSize + 1 : 0} 到{' '}
                  {Math.min(currentPage * pageSize, total)} 条，共 {total}{' '}
                  条结果
                </div>

                {/* 每页条数 */}
                <Select
                  value={pageSize.toString()}
                  onValueChange={(val) => console.log(val)}
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
              </div>

              {/* 分页按钮 */}
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                  disabled={currentPage === 1}
                  className="h-8 w-8 p-0"
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
                  className="h-8 w-8 p-0"
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
