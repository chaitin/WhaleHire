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
}: MultiSelectProps) {
  const [open, setOpen] = React.useState(false);
  const [searchValue, setSearchValue] = React.useState('');

  // 过滤选项
  const filteredOptions = React.useMemo(() => {
    if (!searchValue.trim()) return options;
    return options.filter(
      (option) =>
        option.label.toLowerCase().includes(searchValue.toLowerCase()) ||
        option.value.toLowerCase().includes(searchValue.toLowerCase())
    );
  }, [options, searchValue]);

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
                共选择 {selected.length} 个技能
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
            onChange={(e) => setSearchValue(e.target.value)}
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
        <div className="overflow-y-auto" style={{ maxHeight: `calc(${maxHeight} - 120px)` }}>
          {filteredOptions.length === 0 ? (
            <div className="py-6 text-center text-sm text-[#6B7280]">
              {emptyText}
            </div>
          ) : (
            filteredOptions.map((option) => {
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
            })
          )}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
