import * as React from 'react';
import { Check, X, ChevronDown } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';

export type Option = {
  value: string;
  label: string;
  description?: string;
};

interface MultiSelectProps {
  options: Option[];
  selected: string[];
  onChange: (selected: string[]) => void;
  placeholder?: string;
  className?: string;
  emptyText?: string;
  disabled?: boolean;
  onSelectAll?: () => void;
  onClearAll?: () => void;
  multiple?: boolean;
  maxDisplayCount?: number;
  searchPlaceholder?: string;
  maxHeight?: string;
  showSelectCount?: boolean;
  selectCountLabel?: string;
  // 分页加载相关
  loading?: boolean;
  hasMore?: boolean;
  onLoadMore?: () => void;
  onSearch?: (keyword: string) => void;
}

export function MultiSelect({
  options,
  selected,
  onChange,
  placeholder = '请选择...',
  className,
  emptyText = '没有找到选项',
  disabled = false,
  onSelectAll,
  onClearAll,
  multiple = true,
  maxDisplayCount = 3,
  searchPlaceholder = '搜索选项...',
  maxHeight = '300px',
  showSelectCount = true,
  selectCountLabel = '项',
  loading = false,
  hasMore = false,
  onLoadMore,
  onSearch,
}: MultiSelectProps) {
  const [open, setOpen] = React.useState(false);
  const [searchValue, setSearchValue] = React.useState('');
  const scrollRef = React.useRef<HTMLDivElement>(null);
  const searchTimeoutRef = React.useRef<NodeJS.Timeout>();

  // 搜索处理（带防抖）
  const handleSearchChange = React.useCallback((value: string) => {
    setSearchValue(value);
    
    if (onSearch) {
      // 清除之前的定时器
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
      // 设置新的防抖定时器
      searchTimeoutRef.current = setTimeout(() => {
        onSearch(value);
      }, 300);
    }
  }, [onSearch]);

  // 清理定时器
  React.useEffect(() => {
    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
    };
  }, []);

  // 滚动加载监听
  React.useEffect(() => {
    const scrollElement = scrollRef.current;
    if (!scrollElement || !onLoadMore || !hasMore || loading) return;

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = scrollElement;
      // 当滚动到底部附近时触发加载更多
      if (scrollHeight - scrollTop - clientHeight < 50) {
        onLoadMore();
      }
    };

    scrollElement.addEventListener('scroll', handleScroll);
    return () => scrollElement.removeEventListener('scroll', handleScroll);
  }, [onLoadMore, hasMore, loading]);

  // 过滤选项（如果没有提供onSearch，则使用本地过滤）
  const filteredOptions = React.useMemo(() => {
    if (onSearch) {
      // 如果提供了onSearch回调，由外部控制过滤
      return options;
    }
    if (!searchValue.trim()) return options;
    return options.filter(
      (option) =>
        option.label.toLowerCase().includes(searchValue.toLowerCase()) ||
        option.value.toLowerCase().includes(searchValue.toLowerCase())
    );
  }, [options, searchValue, onSearch]);

  // 获取选中项的标签
  const selectedLabels = React.useMemo(() => {
    return selected.map((value) => {
      const option = options.find((opt) => opt.value === value);
      return option ? option.label : value;
    });
  }, [selected, options]);

  // 处理选择
  const handleToggle = (value: string) => {
    if (multiple) {
      // 多选模式
      if (selected.includes(value)) {
        onChange(selected.filter((item) => item !== value));
      } else {
        onChange([...selected, value]);
      }
    } else {
      // 单选模式
      onChange([value]);
      setOpen(false);
    }
  };

  // 全选
  const handleSelectAll = () => {
    if (searchValue.trim()) {
      const filteredValues = filteredOptions.map((opt) => opt.value);
      const newSelected = [...new Set([...selected, ...filteredValues])];
      onChange(newSelected);
    } else {
      if (onSelectAll) {
        onSelectAll();
      } else {
        onChange(options.map((opt) => opt.value));
      }
    }
  };

  // 清空
  const handleClearAll = () => {
    if (searchValue.trim()) {
      const filteredValues = filteredOptions.map((opt) => opt.value);
      const newSelected = selected.filter((val) => !filteredValues.includes(val));
      onChange(newSelected);
    } else {
      if (onClearAll) {
        onClearAll();
      } else {
        onChange([]);
      }
    }
  };

  // 移除单个选中项
  const handleRemove = (value: string, e: React.MouseEvent) => {
    e.stopPropagation();
    onChange(selected.filter((item) => item !== value));
  };

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            'min-h-10 h-auto w-full justify-between border-[#D1D5DB] text-[14px] hover:border-[#9CA3AF] focus:border-[#22C55E] focus:ring-1 focus:ring-[#22C55E]',
            className
          )}
        >
          <div className="flex flex-wrap gap-1 flex-1 min-w-0">
            {selected.length === 0 ? (
              <span className="text-[#9CA3AF]">{placeholder}</span>
            ) : showSelectCount ? (
              <span className="text-[#374151] font-medium">
                共选择 {selected.length} 个{selectCountLabel}
              </span>
            ) : (
              <div className="flex flex-wrap gap-1">
                {selectedLabels.slice(0, maxDisplayCount).map((label, index) => (
                  <Badge
                    key={selected[index]}
                    variant="secondary"
                    className="px-2 py-0.5 text-xs bg-[#F3F4F6] text-[#374151] hover:bg-[#E5E7EB]"
                  >
                    {label}
                    {multiple && (
                      <button
                        className="ml-1 hover:text-[#EF4444]"
                        onClick={(e) => handleRemove(selected[index], e)}
                      >
                        <X className="h-3 w-3" />
                      </button>
                    )}
                  </Badge>
                ))}
                {selected.length > maxDisplayCount && (
                  <span className="text-[#6B7280] text-xs font-medium py-1">
                    +{selected.length - maxDisplayCount} 更多
                  </span>
                )}
              </div>
            )}
          </div>
          <ChevronDown className="h-4 w-4 shrink-0 opacity-50 ml-2" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent 
        className="w-[var(--radix-dropdown-menu-trigger-width)] p-0"
        align="start"
        style={{ maxHeight }}
      >
        {/* 搜索框 */}
        <div className="p-2 border-b border-[#E5E7EB]">
          <Input
            placeholder={searchPlaceholder}
            value={searchValue}
            onChange={(e) => handleSearchChange(e.target.value)}
            className="h-8 text-sm"
          />
        </div>

        {/* 操作按钮 */}
        {multiple && (
          <div className="flex justify-between items-center px-2 py-1.5 border-b border-[#F3F4F6] bg-[#F9FAFB]">
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                handleSelectAll();
              }}
              className="text-xs h-6 px-2 text-[#22C55E] hover:text-[#16A34A] hover:bg-[#F0FDF4]"
              disabled={filteredOptions.length === 0}
            >
              {searchValue.trim() ? '全选搜索结果' : '全选'}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                handleClearAll();
              }}
              className="text-xs h-6 px-2 text-[#EF4444] hover:text-[#DC2626] hover:bg-[#FEF2F2]"
              disabled={selected.length === 0}
            >
              {searchValue.trim() ? '清空搜索结果' : '清空'}
            </Button>
          </div>
        )}

        {/* 搜索结果提示 */}
        {searchValue.trim() && filteredOptions.length > 0 && (
          <div className="px-3 py-2 text-xs text-[#6B7280] bg-[#F9FAFB]">
            找到 {filteredOptions.length} 个匹配项
          </div>
        )}

        {/* 选项列表 */}
        <div ref={scrollRef} className="overflow-y-auto" style={{ maxHeight: `calc(${maxHeight} - 120px)` }}>
          {filteredOptions.length === 0 && !loading ? (
            <div className="py-6 text-center text-sm text-[#6B7280]">
              {emptyText}
            </div>
          ) : (
            <>
              {filteredOptions.map((option) => {
                const isSelected = selected.includes(option.value);
                return (
                  <DropdownMenuItem
                    key={option.value}
                    onSelect={(e) => {
                      e.preventDefault();
                    }}
                    onClick={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      handleToggle(option.value);
                    }}
                    className={cn(
                      'cursor-pointer px-3 py-2.5 focus:bg-[#F0FDF4] focus:text-[#374151]',
                      isSelected && 'bg-[#F0FDF4]'
                    )}
                  >
                    <div className="flex items-center gap-3 w-full">
                      {/* 复选框/单选框 */}
                      {multiple ? (
                        <div
                          className={cn(
                            'flex h-4 w-4 items-center justify-center rounded border-2 transition-all',
                            isSelected
                              ? 'bg-[#22C55E] border-[#22C55E]'
                              : 'bg-white border-[#D1D5DB] hover:border-[#22C55E]'
                          )}
                        >
                          {isSelected && <Check className="h-3 w-3 text-white stroke-[3]" />}
                        </div>
                      ) : (
                        <div
                          className={cn(
                            'flex h-4 w-4 items-center justify-center rounded-full border-2 transition-all',
                            isSelected
                              ? 'border-[#22C55E]'
                              : 'border-[#D1D5DB] hover:border-[#22C55E]'
                          )}
                        >
                          {isSelected && (
                            <div className="h-2 w-2 rounded-full bg-[#22C55E]" />
                          )}
                        </div>
                      )}
                      {/* 选项文本 */}
                      <span
                        className={cn(
                          'flex-1 text-sm',
                          isSelected ? 'text-[#374151] font-medium' : 'text-[#6B7280]'
                        )}
                      >
                        {option.label}
                      </span>
                      {/* 选中指示器 */}
                      {isSelected && multiple && (
                        <div className="w-2 h-2 rounded-full bg-[#22C55E]" />
                      )}
                    </div>
                  </DropdownMenuItem>
                );
              })}
              {/* 加载指示器 */}
              {loading && (
                <div className="py-3 text-center">
                  <div className="inline-flex items-center gap-2 text-sm text-[#6B7280]">
                    <div className="w-4 h-4 border-2 border-[#22C55E] border-t-transparent rounded-full animate-spin"></div>
                    <span>加载中...</span>
                  </div>
                </div>
              )}
              {/* 加载更多提示 */}
              {!loading && hasMore && (
                <div className="py-2 text-center text-xs text-[#9CA3AF]">
                  向下滚动加载更多
                </div>
              )}
            </>
          )}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
