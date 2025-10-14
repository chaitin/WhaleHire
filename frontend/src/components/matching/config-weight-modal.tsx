import { useState } from 'react';
import { X, Briefcase, User, Scale, Settings, BarChart3 } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogOverlay,
  DialogPortal,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Slider } from '@/components/ui/slider';
import { cn } from '@/lib/utils';

interface ConfigWeightModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onNext: (weights: WeightConfig) => void;
  onPrevious: () => void;
}

export interface WeightConfig {
  basicInfo: number;
  responsibilities: number;
  skills: number;
  education: number;
  experience: number;
  industry: number;
  custom: number;
  customDescription: string;
}

// 流程步骤配置
const STEPS = [
  { id: 1, name: '选择岗位', description: '选择需要匹配的岗位', icon: Briefcase, active: false },
  { id: 2, name: '选择简历', description: '选择需要匹配的简历', icon: User, active: false },
  { id: 3, name: '权重配置', description: '配置匹配权重', icon: Scale, active: true },
  { id: 4, name: '匹配处理', description: 'AI智能匹配分析', icon: Settings, active: false },
  { id: 5, name: '匹配结果', description: '查看匹配报告', icon: BarChart3, active: false },
];

export function ConfigWeightModal({
  open,
  onOpenChange,
  onNext,
  onPrevious,
}: ConfigWeightModalProps) {
  // 权重配置状态（默认值）
  const [weights, setWeights] = useState<WeightConfig>({
    basicInfo: 10,
    responsibilities: 25,
    skills: 25,
    education: 15,
    experience: 15,
    industry: 10,
    custom: 0,
    customDescription: '',
  });

  // 计算总权重（不包括自定义权重）
  const totalWeight = weights.basicInfo + weights.responsibilities + weights.skills + 
                      weights.education + weights.experience + weights.industry;

  // 权重是否有效（必须等于100%）
  const isWeightValid = totalWeight === 100;

  // 更新权重值
  const updateWeight = (field: keyof WeightConfig, value: string) => {
    const numValue = parseInt(value) || 0;
    setWeights({
      ...weights,
      [field]: numValue,
    });
  };

  // 处理下一步
  const handleNext = () => {
    if (!isWeightValid) {
      alert('权重总和必须等于100%');
      return;
    }
    onNext(weights);
  };

  // 处理上一步
  const handlePrevious = () => {
    onPrevious();
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
              <X className="h-4 w-4 text-[#666666]" />
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
                {STEPS.map((step) => {
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
                          <IconComponent className={cn(
                            "h-5 w-5",
                            step.active ? "text-[#10B981]" : "text-[#999999]"
                          )} />
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
                      <span className={cn(
                        "text-xs text-center mt-0.5",
                        step.active ? "text-[#666666]" : "text-[#999999]"
                      )}>
                        {step.description}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* 权重配置区域 */}
            <div className="bg-[#FAFAFA] rounded-lg p-5 space-y-3">
              {/* 基本信息权重 */}
              <div className="bg-white rounded-md p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-sm opacity-60">👤</span>
                    <span className="text-sm font-semibold text-[#333333]">基本信息权重</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Input
                      type="number"
                      min="0"
                      max="100"
                      value={weights.basicInfo}
                      onChange={(e) => updateWeight('basicInfo', e.target.value)}
                      className="w-[70px] h-[32px] px-2 text-center text-base font-bold text-primary border border-[#E8E8E8] rounded"
                    />
                    <span className="text-base font-semibold text-primary">%</span>
                  </div>
                </div>
                <Slider
                  value={[weights.basicInfo]}
                  onValueChange={(value: number[]) => updateWeight('basicInfo', value[0].toString())}
                  min={0}
                  max={100}
                  step={1}
                  className="mb-2"
                />
                <p className="text-xs text-[#999999]">性别、年龄、户籍等基本信息的匹配权重</p>
              </div>

              {/* 工作职责权重 */}
              <div className="bg-white rounded-md p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-sm opacity-60">📋</span>
                    <span className="text-sm font-semibold text-[#333333]">工作职责权重</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Input
                      type="number"
                      min="0"
                      max="100"
                      value={weights.responsibilities}
                      onChange={(e) => updateWeight('responsibilities', e.target.value)}
                      className="w-[70px] h-[32px] px-2 text-center text-base font-bold text-primary border border-[#E8E8E8] rounded"
                    />
                    <span className="text-base font-semibold text-primary">%</span>
                  </div>
                </div>
                <Slider
                  value={[weights.responsibilities]}
                  onValueChange={(value: number[]) => updateWeight('responsibilities', value[0].toString())}
                  min={0}
                  max={100}
                  step={1}
                  className="mb-2"
                />
                <p className="text-xs text-[#999999]">候选人过往工作职责与岗位职责的匹配权重</p>
              </div>

              {/* 工作技能权重 */}
              <div className="bg-white rounded-md p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-sm opacity-60">⚙️</span>
                    <span className="text-sm font-semibold text-[#333333]">工作技能权重</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Input
                      type="number"
                      min="0"
                      max="100"
                      value={weights.skills}
                      onChange={(e) => updateWeight('skills', e.target.value)}
                      className="w-[70px] h-[32px] px-2 text-center text-base font-bold text-primary border border-[#E8E8E8] rounded"
                    />
                    <span className="text-base font-semibold text-primary">%</span>
                  </div>
                </div>
                <Slider
                  value={[weights.skills]}
                  onValueChange={(value: number[]) => updateWeight('skills', value[0].toString())}
                  min={0}
                  max={100}
                  step={1}
                  className="mb-2"
                />
                <p className="text-xs text-[#999999]">候选人技能与岗位要求技能的匹配权重</p>
              </div>

              {/* 教育经历权重 */}
              <div className="bg-white rounded-md p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-sm opacity-60">🎓</span>
                    <span className="text-sm font-semibold text-[#333333]">教育经历权重</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Input
                      type="number"
                      min="0"
                      max="100"
                      value={weights.education}
                      onChange={(e) => updateWeight('education', e.target.value)}
                      className="w-[70px] h-[32px] px-2 text-center text-base font-bold text-primary border border-[#E8E8E8] rounded"
                    />
                    <span className="text-base font-semibold text-primary">%</span>
                  </div>
                </div>
                <Slider
                  value={[weights.education]}
                  onValueChange={(value: number[]) => updateWeight('education', value[0].toString())}
                  min={0}
                  max={100}
                  step={1}
                  className="mb-2"
                />
                <p className="text-xs text-[#999999]">学历、专业、院校与岗位要求的匹配权重</p>
              </div>

              {/* 工作经历权重 */}
              <div className="bg-white rounded-md p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-sm opacity-60">💼</span>
                    <span className="text-sm font-semibold text-[#333333]">工作经历权重</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Input
                      type="number"
                      min="0"
                      max="100"
                      value={weights.experience}
                      onChange={(e) => updateWeight('experience', e.target.value)}
                      className="w-[70px] h-[32px] px-2 text-center text-base font-bold text-primary border border-[#E8E8E8] rounded"
                    />
                    <span className="text-base font-semibold text-primary">%</span>
                  </div>
                </div>
                <Slider
                  value={[weights.experience]}
                  onValueChange={(value: number[]) => updateWeight('experience', value[0].toString())}
                  min={0}
                  max={100}
                  step={1}
                  className="mb-2"
                />
                <p className="text-xs text-[#999999]">工作年限、工作经验与岗位要求的匹配权重</p>
              </div>

              {/* 行业背景权重 */}
              <div className="bg-white rounded-md p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-sm opacity-60">🏢</span>
                    <span className="text-sm font-semibold text-[#333333]">行业背景权重</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Input
                      type="number"
                      min="0"
                      max="100"
                      value={weights.industry}
                      onChange={(e) => updateWeight('industry', e.target.value)}
                      className="w-[70px] h-[32px] px-2 text-center text-base font-bold text-primary border border-[#E8E8E8] rounded"
                    />
                    <span className="text-base font-semibold text-primary">%</span>
                  </div>
                </div>
                <Slider
                  value={[weights.industry]}
                  onValueChange={(value: number[]) => updateWeight('industry', value[0].toString())}
                  min={0}
                  max={100}
                  step={1}
                  className="mb-2"
                />
                <p className="text-xs text-[#999999]">候选人所在行业与目标岗位行业的匹配权重</p>
              </div>

              {/* 其他自定义权重 */}
              <div className="bg-white rounded-md p-4">
                <div className="flex items-center justify-between mb-2.5">
                  <div className="flex items-center gap-2">
                    <span className="text-sm opacity-60">⚙️</span>
                    <span className="text-sm font-semibold text-[#333333]">其他自定义权重</span>
                  </div>
                  <div className="flex items-center gap-0.5">
                    <span className="text-base font-semibold text-primary">{weights.custom}</span>
                    <span className="text-base font-semibold text-primary">%</span>
                  </div>
                </div>
                <div className="h-[6px] bg-gradient-to-r from-[#10B981] to-[#059669] rounded-full mb-1.5" style={{ width: `${(weights.custom / 100) * 129}px` }} />
                <p className="text-xs text-[#999999] mb-2">其他自定义匹配因素的权重（不计入100%总和）</p>
                
                <Textarea
                  placeholder="请输入自定义权重的说明（可选）..."
                  value={weights.customDescription}
                  onChange={(e) => setWeights({ ...weights, customDescription: e.target.value })}
                  className="min-h-[60px] text-xs text-[#BBBBBB] resize-none border-[#E0E0E0]"
                />
              </div>

              {/* 注意事项 */}
              <div className="bg-[#FFFBEB] border-l-[3px] border-[#F59E0B] rounded px-3 py-2.5">
                <p className="text-xs text-[#F59E0B] leading-relaxed">
                  💡 注意：除自定义权重外，其他6项权重总和必须等于100%。自定义权重默认为0%，不计入总和。
                </p>
              </div>

              {/* 权重总和显示 */}
              <div className="bg-white border-2 border-[#E2E8F0] rounded-md px-4 py-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-semibold text-[#333333]">基础权重总和（需等于100%）</span>
                  <span className={cn(
                    "text-xl font-bold",
                    isWeightValid ? "text-primary" : "text-red-500"
                  )}>
                    {totalWeight}%
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* 底部：操作按钮 */}
          <div className="border-t border-[#E8E8E8] px-6 py-4">
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
                disabled={!isWeightValid}
                className={cn(
                  "px-6 text-white",
                  isWeightValid 
                    ? "bg-primary hover:bg-primary/90" 
                    : "bg-gray-300 cursor-not-allowed"
                )}
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

