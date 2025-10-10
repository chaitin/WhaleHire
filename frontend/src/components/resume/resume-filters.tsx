import React from 'react';
import { Search } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { ResumeStatusFilter, ResumeStatus } from '@/types/resume';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ResumeFilters } from '@/types/resume';

// 状态选项
const statusOptions = [
  { value: 'all', label: '所有状态' },
  { value: ResumeStatus.PROCESSING, label: '解析中' },
  { value: ResumeStatus.COMPLETED, label: '待筛选' }, // 解析成功翻译为待筛选
];

interface ResumeFiltersProps {
  filters: ResumeFilters;
  onFiltersChange: (filters: ResumeFilters, options?: { shouldSearch?: boolean }) => void;
  onSearch?: () => void;
}

export function ResumeFiltersComponent({
  filters,
  onFiltersChange,
  onSearch,
}: ResumeFiltersProps) {
  const handleStatusChange = (value: string) => {
    onFiltersChange({ ...filters, status: value as ResumeStatusFilter }, { shouldSearch: true });
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onFiltersChange({ ...filters, keywords: e.target.value }, { shouldSearch: false });
  };

  return (
    <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
      {/* 左侧筛选器 */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        <Select value={filters.status} onValueChange={handleStatusChange}>
          <SelectTrigger className="h-11 w-full rounded-lg border-[#D1D5DB] sm:w-40">
            <SelectValue placeholder="所有状态" />
          </SelectTrigger>
          <SelectContent className="rounded-lg border border-[#E5E7EB]">
            {statusOptions.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* 右侧搜索区域 */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="relative sm:w-64">
          <Search className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="搜索姓名、电话..."
            value={filters.keywords ?? ''}
            onChange={handleSearchChange}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault();
                onSearch?.();
              }
            }}
            className="h-11 w-full rounded-lg border-[#D1D5DB] pl-11"
          />
        </div>
        <Button
          type="button"
          className="h-11 rounded-lg px-6"
          onClick={onSearch}
        >
          搜索
        </Button>
      </div>
    </div>
  );
}
