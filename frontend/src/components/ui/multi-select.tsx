import * as React from 'react';
import { Check, ChevronsUpDown } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';

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
  multiple?: boolean; // 是否支持多选，默认为true
  maxDisplayCount?: number; // 最大显示选中项数量，超过则显示"已选择X项"
  searchPlaceholder?: string; // 搜索框占位符
  maxHeight?: string; // 下拉列表最大高度
  showSelectCount?: boolean; // 是否显示选中数量
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
  multiple = true, // 默认为多选模式
  maxDisplayCount = 3, // 默认最多显示3个选中项
  searchPlaceholder = '搜索选项...',
  maxHeight = '300px',
  showSelectCount = true,
}: MultiSelectProps) {
  const [open, setOpen] = React.useState(false);
  const [searchValue, setSearchValue] = React.useState('');

  // 过滤选项 - 使用useMemo优化性能
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

  const handleSelect = React.useCallback(
    (value: string) => {
      if (multiple) {
        // 多选模式
        if (selected.includes(value)) {
          onChange(selected.filter((item) => item !== value));
        } else {
          onChange([...selected, value]);
        }
      } else {
        // 单选模式 - 直接选择新项，不允许取消选择
        onChange([value]); // 只选择当前项
        setOpen(false); // 单选模式下选择后自动关闭
      }
    },
    [multiple, selected, onChange]
  );

  // 重置搜索值当弹窗关闭时
  React.useEffect(() => {
    if (!open) {
      setSearchValue('');
    }
  }, [open]);

  // 全选功能
  const handleSelectAll = () => {
    if (searchValue.trim()) {
      // 如果有搜索，则全选搜索结果
      const filteredValues = filteredOptions.map((opt) => opt.value);
      const newSelected = [...new Set([...selected, ...filteredValues])];
      onChange(newSelected);
    } else {
      // 否则全选所有选项
      if (onSelectAll) {
        onSelectAll();
      } else {
        onChange(options.map((opt) => opt.value));
      }
    }
  };

  // 清空功能
  const handleClearAll = () => {
    if (searchValue.trim()) {
      // 如果有搜索，则清空搜索结果中的选中项
      const filteredValues = filteredOptions.map((opt) => opt.value);
      const newSelected = selected.filter(
        (val) => !filteredValues.includes(val)
      );
      onChange(newSelected);
    } else {
      // 否则清空所有选中项
      if (onClearAll) {
        onClearAll();
      } else {
        onChange([]);
      }
    }
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn(
            'min-h-10 h-auto w-full justify-between border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF] hover:border-[#9CA3AF] focus:border-[#22C55E] focus:ring-1 focus:ring-[#22C55E]',
            className
          )}
          disabled={disabled}
        >
          <div className="flex flex-wrap gap-1 flex-1 min-w-0">
            {selected.length === 0 ? (
              <span className="text-[#9CA3AF]">{placeholder}</span>
            ) : multiple ? (
              showSelectCount && selected.length > maxDisplayCount ? (
                <span className="text-[#374151] font-medium">
                  已选择 {selected.length} 项
                </span>
              ) : showSelectCount ? (
                <span className="text-[#374151] font-medium">
                  共选择 {selected.length} 个技能
                </span>
              ) : (
                <div className="flex flex-wrap gap-1 max-w-full">
                  {selectedLabels
                    .slice(0, maxDisplayCount)
                    .map((label, index) => (
                      <span
                        key={selected[index]}
                        className="inline-flex items-center px-2 py-1 rounded-md bg-[#F3F4F6] text-[#374151] text-xs max-w-[120px] truncate border"
                        title={label}
                      >
                        {label}
                      </span>
                    ))}
                  {selected.length > maxDisplayCount && (
                    <span className="text-[#6B7280] text-xs font-medium">
                      +{selected.length - maxDisplayCount} 更多
                    </span>
                  )}
                </div>
              )
            ) : (
              <span
                className="text-[#374151] truncate font-medium"
                title={selectedLabels[0]}
              >
                {selectedLabels[0] || selected[0]}
              </span>
            )}
          </div>
          <ChevronsUpDown className="h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="w-full p-0 shadow-lg border-[#E5E7EB]"
        align="start"
        style={{ maxHeight }}
      >
        <Command shouldFilter={false}>
          <CommandInput
            placeholder={searchPlaceholder}
            value={searchValue}
            onValueChange={setSearchValue}
            className="border-0 border-b border-[#E5E7EB] rounded-none focus:ring-0"
          />
          <CommandList style={{ maxHeight: `calc(${maxHeight} - 60px)` }}>
            <CommandEmpty className="py-6 text-center text-[#6B7280]">
              {emptyText}
            </CommandEmpty>
            {searchValue.trim() && filteredOptions.length > 0 && (
              <div className="px-3 py-2 text-xs text-[#6B7280] border-b border-[#F3F4F6] bg-[#F9FAFB]">
                找到 {filteredOptions.length} 个匹配项
              </div>
            )}
            <CommandGroup>
              {multiple && (
                <div className="flex justify-between items-center px-3 py-2 border-b border-[#F3F4F6] bg-[#F9FAFB]">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.preventDefault();
                      handleSelectAll();
                    }}
                    className="text-xs h-7 px-2 text-[#22C55E] hover:text-[#16A34A] hover:bg-[#F0FDF4]"
                    disabled={filteredOptions.length === 0}
                  >
                    {searchValue.trim() ? '全选搜索结果' : '全选'}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.preventDefault();
                      handleClearAll();
                    }}
                    className="text-xs h-7 px-2 text-[#EF4444] hover:text-[#DC2626] hover:bg-[#FEF2F2]"
                    disabled={selected.length === 0}
                  >
                    {searchValue.trim() ? '清空搜索结果' : '清空'}
                  </Button>
                </div>
              )}
              {filteredOptions.map((option) => {
                const isSelected = selected.includes(option.value);
                return (
                  <CommandItem
                    key={option.value}
                    value={option.value}
                    onSelect={() => {
                      handleSelect(option.value);
                    }}
                    className="cursor-pointer hover:bg-[#F9FAFB] px-3 py-2.5"
                  >
                    <div className="flex items-center gap-3 w-full">
                      {/* 改进的复选框样式 */}
                      <div
                        className={cn(
                          'flex h-4 w-4 items-center justify-center rounded border-2 transition-all duration-200',
                          isSelected
                            ? 'bg-[#22C55E] border-[#22C55E] shadow-sm'
                            : 'bg-white border-[#D1D5DB] hover:border-[#9CA3AF]'
                        )}
                      >
                        {isSelected && (
                          <Check className="h-3 w-3 text-white stroke-[3]" />
                        )}
                      </div>
                      {/* 选项文本 */}
                      <span
                        className={cn(
                          'flex-1 text-sm transition-colors duration-200',
                          isSelected
                            ? 'text-[#374151] font-medium'
                            : 'text-[#6B7280] hover:text-[#374151]'
                        )}
                      >
                        {option.label}
                      </span>
                      {/* 选中状态指示器 */}
                      {isSelected && (
                        <div className="w-2 h-2 rounded-full bg-[#22C55E]" />
                      )}
                    </div>
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
