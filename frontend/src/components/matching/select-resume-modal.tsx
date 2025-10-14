import { useState } from 'react';
import { X, Search, ChevronLeft, ChevronRight, Briefcase, User, Scale, Settings, BarChart3 } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogOverlay,
  DialogPortal,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { cn } from '@/lib/utils';
import { Resume, ResumeStatus } from '@/types/resume';

interface SelectResumeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onNext: (selectedResumeIds: string[]) => void;
  onPrevious: () => void;
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
}: SelectResumeModalProps) {
  const [searchKeyword, setSearchKeyword] = useState('');
  const [experienceFilter, setExperienceFilter] = useState<string>('');
  const [educationFilter, setEducationFilter] = useState<string>('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [selectedResumeIds, setSelectedResumeIds] = useState<string[]>([]);

  // Mock 数据 - 后续需要替换为真实 API 调用
  const mockResumes: Resume[] = [
    {
      id: '1',
      name: '张伟',
      email: 'zhangwei@example.com',
      phone: '13800138001',
      highest_education: '本科',
      years_experience: 5,
      status: ResumeStatus.COMPLETED,
      uploader_id: 'user1',
      uploader_name: '李华',
      created_at: Date.now() / 1000 - 86400 * 30,
      updated_at: Date.now() / 1000 - 86400 * 30,
    },
    {
      id: '2',
      name: '李娜',
      email: 'lina@example.com',
      phone: '13800138002',
      highest_education: '硕士',
      years_experience: 3,
      status: ResumeStatus.COMPLETED,
      uploader_id: 'user1',
      uploader_name: '王芳',
      created_at: Date.now() / 1000 - 86400 * 33,
      updated_at: Date.now() / 1000 - 86400 * 33,
    },
    {
      id: '3',
      name: '王强',
      email: 'wangqiang@example.com',
      phone: '13800138003',
      highest_education: '本科',
      years_experience: 4,
      status: ResumeStatus.COMPLETED,
      uploader_id: 'user1',
      uploader_name: '赵强',
      created_at: Date.now() / 1000 - 86400 * 36,
      updated_at: Date.now() / 1000 - 86400 * 36,
    },
    {
      id: '4',
      name: '刘芳',
      email: 'liufang@example.com',
      phone: '13800138004',
      highest_education: '硕士',
      years_experience: 6,
      status: ResumeStatus.COMPLETED,
      uploader_id: 'user1',
      uploader_name: '陈静',
      created_at: Date.now() / 1000 - 86400 * 41,
      updated_at: Date.now() / 1000 - 86400 * 41,
    },
    {
      id: '5',
      name: '陈明',
      email: 'chenming@example.com',
      phone: '13800138005',
      highest_education: '本科',
      years_experience: 2,
      status: ResumeStatus.COMPLETED,
      uploader_id: 'user1',
      uploader_name: '张明',
      created_at: Date.now() / 1000 - 86400 * 46,
      updated_at: Date.now() / 1000 - 86400 * 46,
    },
  ];

  const totalPages = 5; // Mock 总页数
  const totalResults = 24; // Mock 总数据量

  // 处理搜索
  const handleSearch = () => {
    console.log('搜索:', searchKeyword);
    // TODO: 实现搜索功能
  };

  // 处理复选框切换
  const handleCheckboxChange = (resumeId: string, checked: boolean) => {
    if (checked) {
      setSelectedResumeIds([...selectedResumeIds, resumeId]);
    } else {
      setSelectedResumeIds(selectedResumeIds.filter(id => id !== resumeId));
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
              i === currentPage && 'bg-primary text-white border-primary hover:bg-primary hover:text-white'
            )}
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
            1 === currentPage && 'bg-primary text-white border-primary hover:bg-primary hover:text-white'
          )}
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
              2 === currentPage && 'bg-primary text-white border-primary hover:bg-primary hover:text-white'
            )}
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
              3 === currentPage && 'bg-primary text-white border-primary hover:bg-primary hover:text-white'
            )}
          >
            3
          </Button>
        );
      }

      // 省略号
      if (currentPage > 3 && currentPage < totalPages - 1) {
        pages.push(
          <div key="ellipsis" className="flex items-center justify-center h-8 w-8">
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
            totalPages === currentPage && 'bg-primary text-white border-primary hover:bg-primary hover:text-white'
          )}
        >
          {totalPages}
        </Button>
      );
    }

    return pages;
  };

  // 格式化时间
  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit'
    }).replace(/\//g, '-');
  };

  // 格式化工作经验
  const formatExperience = (years?: number) => {
    if (!years) return '-';
    return `${years}年`;
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPortal>
        <DialogOverlay className="bg-black/50" />
        <DialogContent className="max-w-[920px] p-0 gap-0 bg-white rounded-xl overflow-hidden">
          {/* 头部 */}
          <div className="flex items-center justify-between border-b border-[#E8E8E8] px-6 py-5">
            <h2 className="text-lg font-semibold text-[#333333]">创建新匹配任务</h2>
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
                <div className="absolute top-7 left-0 right-0 h-0.5 bg-[#E8E8E8]" style={{ marginLeft: '28px', marginRight: '28px' }} />
                
                {/* 步骤项 */}
                {STEPS.map((step, index) => {
                  const IconComponent = step.icon;
                  return (
                    <div key={step.id} className="relative flex flex-col items-center z-10" style={{ width: '108.5px' }}>
                      {/* 图标 */}
                      <div className={cn(
                        'w-14 h-14 rounded-full flex items-center justify-center mb-3',
                        step.active
                          ? 'bg-[#D1FAE5] shadow-[0_0_0_4px_rgba(16,185,129,0.1)]'
                          : 'bg-[#E8E8E8]'
                      )}>
                        <div className={cn(
                          'w-8 h-8 rounded-2xl flex items-center justify-center',
                          step.active ? 'bg-white' : 'bg-white opacity-60'
                        )}>
                          <IconComponent className="h-5 w-5 text-[#999999]" />
                        </div>
                      </div>
                      
                      {/* 步骤名称 */}
                      <span className={cn(
                        'text-sm text-center',
                        step.active ? 'text-[#10B981] font-semibold' : 'text-[#666666]'
                      )}>
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
                  <Select value={experienceFilter} onValueChange={setExperienceFilter}>
                    <SelectTrigger className="w-[144px] h-[33.5px] text-sm">
                      <SelectValue placeholder="请选择工作经验" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">全部经验</SelectItem>
                      <SelectItem value="0-1">1年以下</SelectItem>
                      <SelectItem value="1-3">1-3年</SelectItem>
                      <SelectItem value="3-5">3-5年</SelectItem>
                      <SelectItem value="5+">5年以上</SelectItem>
                    </SelectContent>
                  </Select>

                  <Select value={educationFilter} onValueChange={setEducationFilter}>
                    <SelectTrigger className="w-[116px] h-[33.5px] text-sm">
                      <SelectValue placeholder="请选择学历" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">全部学历</SelectItem>
                      <SelectItem value="高中">高中</SelectItem>
                      <SelectItem value="专科">专科</SelectItem>
                      <SelectItem value="本科">本科</SelectItem>
                      <SelectItem value="硕士">硕士</SelectItem>
                      <SelectItem value="博士">博士</SelectItem>
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
                    className="h-9 px-6 rounded-l-none bg-primary hover:bg-primary/90"
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
                          checked={selectedResumeIds.length === mockResumes.length && mockResumes.length > 0}
                          onCheckedChange={(checked) => {
                            if (checked) {
                              setSelectedResumeIds(mockResumes.map(r => r.id));
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
                    {mockResumes.map((resume) => (
                      <tr key={resume.id} className="border-b border-[#F0F0F0] last:border-b-0">
                        <td className="px-3 py-4 text-sm">
                          <Checkbox 
                            checked={selectedResumeIds.includes(resume.id)}
                            onCheckedChange={(checked) => handleCheckboxChange(resume.id, checked as boolean)}
                          />
                        </td>
                        <td className="px-3 py-4 text-sm font-medium text-[#333333]">
                          {resume.name}
                        </td>
                        <td className="px-3 py-4 text-sm text-[#333333]">
                          {resume.id === '1' ? '前端开发工程师' : 
                           resume.id === '2' ? '产品经理' :
                           resume.id === '3' ? 'UI设计师' :
                           resume.id === '4' ? '后端开发工程师' :
                           '数据分析师'}
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
                    ))}
                  </tbody>
                </table>
              </div>

              {/* 分页控件 */}
              <div className="flex items-center justify-between">
                {/* 分页信息和每页条数 */}
                <div className="flex items-center gap-3">
                  <div className="text-[13px] text-[#666666]">
                    显示第 {(currentPage - 1) * pageSize + 1} 到{' '}
                    {Math.min(currentPage * pageSize, totalResults)} 条，共 {totalResults} 条结果
                  </div>
                  
                  {/* 每页条数 */}
                  <Select value={pageSize.toString()} onValueChange={(val) => setPageSize(Number(val))}>
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
                    onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
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
                className="px-6 bg-primary hover:bg-primary/90 text-white"
              >
                下一步
              </Button>
            </div>
          </div>
        </DialogContent>
      </DialogPortal>
    </Dialog>
  );
}

