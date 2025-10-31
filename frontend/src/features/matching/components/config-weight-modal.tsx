import { useState, useEffect } from 'react';
import {
  X,
  Briefcase,
  User,
  Scale,
  Settings,
  BarChart3,
  ChevronDown,
  ChevronUp,
  RefreshCw,
  Sparkles,
  CheckCircle2,
} from 'lucide-react';
import { Dialog, DialogContent, DialogTitle } from '@/ui/dialog';
import { Button } from '@/ui/button';
import { Input } from '@/ui/input';
import { Slider } from '@/ui/slider';
import { cn } from '@/lib/utils';
import { previewWeights } from '@/services/screening';
import type { PreviewWeightsResp, WeightSchemeResp } from '@/types/screening';

interface ConfigWeightModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onNext: (weights: WeightConfig) => void;
  onPrevious: () => void;
  selectedJobId: string; // 第一步选择的岗位ID
}

export interface WeightConfig {
  basicInfo: number;
  responsibilities: number;
  skills: number;
  education: number;
  experience: number;
  industry: number;
}

// 流程步骤配置
const STEPS = [
  {
    id: 1,
    name: '选择岗位',
    description: '选择需要匹配的岗位',
    icon: Briefcase,
    active: false,
  },
  {
    id: 2,
    name: '选择简历',
    description: '选择需要匹配的简历',
    icon: User,
    active: false,
  },
  {
    id: 3,
    name: '权重配置',
    description: '配置匹配权重',
    icon: Scale,
    active: true,
  },
  {
    id: 4,
    name: '匹配处理',
    description: 'AI智能匹配分析',
    icon: Settings,
    active: false,
  },
  {
    id: 5,
    name: '匹配结果',
    description: '查看匹配报告',
    icon: BarChart3,
    active: false,
  },
];

export function ConfigWeightModal({
  open,
  onOpenChange,
  onNext,
  onPrevious,
  selectedJobId,
}: ConfigWeightModalProps) {
  // 权重配置状态（默认值）
  const [weights, setWeights] = useState<WeightConfig>({
    basicInfo: 3,
    responsibilities: 20,
    skills: 35,
    education: 15,
    experience: 20,
    industry: 7,
  });

  // 智能权重计算状态
  const [aiWeightsLoading, setAiWeightsLoading] = useState(false);
  const [aiWeightsData, setAiWeightsData] = useState<PreviewWeightsResp | null>(
    null
  );
  const [aiWeightsError, setAiWeightsError] = useState<string | null>(null);
  const [showAiWeightsDetail, setShowAiWeightsDetail] = useState(false);

  // Tab状态 - 用于展示三种权重方案
  const [activeTab, setActiveTab] = useState<
    'default' | 'fresh_graduate' | 'experienced'
  >('default');
  const [selectedSchemeType, setSelectedSchemeType] =
    useState<string>('default'); // 用户选择的方案类型

  // 高级配置展开状态
  const [advancedConfigExpanded, setAdvancedConfigExpanded] = useState(false);

  // 计算总权重
  const totalWeight =
    weights.basicInfo +
    weights.responsibilities +
    weights.skills +
    weights.education +
    weights.experience +
    weights.industry;

  // 权重是否有效
  const isWeightValid = totalWeight === 100;

  // 弹窗打开/关闭时重置状态
  useEffect(() => {
    // 弹窗关闭时重置状态
    if (!open) {
      setAiWeightsLoading(false);
      setAiWeightsData(null);
      setAiWeightsError(null);
      setShowAiWeightsDetail(false);
      setAdvancedConfigExpanded(false);
      setActiveTab('default');
      setSelectedSchemeType('default');
    }
  }, [open]);

  // 应用权重方案到权重配置
  const applyWeightScheme = (scheme: WeightSchemeResp) => {
    if (!scheme || !scheme.weights) return;

    const newWeights: WeightConfig = {
      basicInfo: Math.round((scheme.weights.basic || 0) * 100),
      responsibilities: Math.round((scheme.weights.responsibility || 0) * 100),
      skills: Math.round((scheme.weights.skill || 0) * 100),
      education: Math.round((scheme.weights.education || 0) * 100),
      experience: Math.round((scheme.weights.experience || 0) * 100),
      industry: Math.round((scheme.weights.industry || 0) * 100),
    };

    // 确保总和为100（处理四舍五入误差）
    const sum = Object.values(newWeights).reduce((a, b) => a + b, 0);
    if (sum !== 100) {
      const maxKey = Object.keys(newWeights).reduce((a, b) =>
        newWeights[a as keyof WeightConfig] >
        newWeights[b as keyof WeightConfig]
          ? a
          : b
      ) as keyof WeightConfig;
      newWeights[maxKey] += 100 - sum;
    }

    console.log('🤖 应用权重方案:', scheme.type, newWeights);
    setWeights(newWeights);
    setSelectedSchemeType(scheme.type);
  };

  // 调用智能权重计算API
  const fetchAiWeights = async () => {
    setAiWeightsLoading(true);
    setAiWeightsError(null);

    try {
      console.log('🤖 开始智能权重计算，岗位ID:', selectedJobId);
      const response = await previewWeights({
        job_position_id: selectedJobId,
      });

      console.log('🤖 智能权重计算成功:', response);
      setAiWeightsData(response);

      // 默认选择第一个方案（default）并应用权重
      if (response.weight_schemes && response.weight_schemes.length > 0) {
        const defaultScheme =
          response.weight_schemes.find((s) => s.type === 'default') ||
          response.weight_schemes[0];
        applyWeightScheme(defaultScheme);
        setActiveTab(
          defaultScheme.type as 'default' | 'fresh_graduate' | 'experienced'
        );
        setSelectedSchemeType(defaultScheme.type);
      }
    } catch (error) {
      console.error('❌ 智能权重计算失败:', error);
      setAiWeightsError(
        error instanceof Error ? error.message : '智能权重计算失败，请重试'
      );
    } finally {
      setAiWeightsLoading(false);
    }
  };

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

  // 渲染权重配置项
  const renderWeightItem = (
    emoji: string,
    label: string,
    field: keyof WeightConfig
  ) => (
    <div className="bg-white rounded-md px-3 py-2">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-1.5">
          <span className="text-xs">{emoji}</span>
          <span className="text-xs text-[#333333]">{label}</span>
        </div>
        <div className="flex items-center gap-1">
          <Input
            type="number"
            min="0"
            max="100"
            value={weights[field]}
            onChange={(e) => updateWeight(field, e.target.value)}
            className="w-[60px] h-[28px] px-2 text-center text-sm font-bold text-primary border border-[#E8E8E8] rounded"
          />
          <span className="text-sm font-semibold text-primary">%</span>
        </div>
      </div>
      <Slider
        value={[weights[field]]}
        onValueChange={(value: number[]) =>
          updateWeight(field, value[0].toString())
        }
        min={0}
        max={100}
        step={1}
      />
    </div>
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-w-[920px] p-0 gap-0 bg-white rounded-xl"
        onInteractOutside={(e) => e.preventDefault()}
      >
        <DialogTitle className="sr-only">创建新匹配任务 - 配置权重</DialogTitle>
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
              <div
                className="absolute top-7 left-0 right-0 h-0.5 bg-[#E8E8E8]"
                style={{ marginLeft: '28px', marginRight: '28px' }}
              />

              {/* 步骤项 */}
              {STEPS.map((step) => {
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
                        <IconComponent className={'h-5 w-5 text-[#999999]'} />
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
                    <span
                      className={cn(
                        'text-xs text-center mt-0.5',
                        step.active ? 'text-[#666666]' : 'text-[#999999]'
                      )}
                    >
                      {step.description}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>

          {/* 智能权重计算区域 */}
          <div className="bg-gradient-to-br from-blue-50 to-purple-50 rounded-lg p-5 mb-6 border-2 border-blue-100">
            <div className="flex items-start gap-3 mb-4">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center flex-shrink-0">
                <Sparkles className="h-5 w-5 text-white" />
              </div>
              <div className="flex-1">
                <h4 className="text-base font-semibold text-[#333333] mb-1">
                  AI智能权重推荐
                </h4>
                <p className="text-xs text-[#666666]">
                  基于岗位画像，由大模型智能计算最优权重配比
                </p>
              </div>
            </div>

            {/* 初始状态 - 等待用户点击开始计算 */}
            {!aiWeightsLoading && !aiWeightsError && !aiWeightsData && (
              <div className="bg-white rounded-md p-6 text-center">
                <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-gradient-to-br from-blue-100 to-purple-100 mb-3">
                  <Sparkles className="h-6 w-6 text-blue-600" />
                </div>
                <p className="text-sm font-medium text-[#333333] mb-2">
                  准备好开始智能权重计算了吗？
                </p>
                <p className="text-xs text-[#999999] mb-4">
                  AI将基于岗位画像为您推荐最优权重配比方案
                </p>
                <Button
                  onClick={fetchAiWeights}
                  className="text-white"
                  style={{
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }}
                >
                  <Sparkles className="h-4 w-4 mr-2" />
                  开始计算
                </Button>
              </div>
            )}

            {/* 加载状态 */}
            {aiWeightsLoading && (
              <div className="bg-white rounded-md p-6 text-center">
                <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-blue-100 mb-3 animate-pulse">
                  <RefreshCw className="h-6 w-6 text-blue-600 animate-spin" />
                </div>
                <p className="text-sm font-medium text-[#333333] mb-1">
                  智能权重计算中...
                </p>
                <p className="text-xs text-[#999999]">
                  AI正在分析岗位画像，为您推荐最优权重配比
                </p>
              </div>
            )}

            {/* 错误状态 */}
            {!aiWeightsLoading && aiWeightsError && (
              <div className="bg-white rounded-md p-6">
                <div className="flex items-center justify-center gap-2 mb-3">
                  <span className="text-2xl">⚠️</span>
                  <p className="text-sm font-medium text-red-600">计算失败</p>
                </div>
                <p className="text-xs text-[#666666] text-center mb-4">
                  {aiWeightsError}
                </p>
                <Button
                  onClick={fetchAiWeights}
                  variant="outline"
                  size="sm"
                  className="w-full"
                >
                  <RefreshCw className="h-4 w-4 mr-2" />
                  重新计算
                </Button>
              </div>
            )}

            {/* 成功状态 - 使用Tab展示三种方案 */}
            {!aiWeightsLoading &&
              !aiWeightsError &&
              aiWeightsData &&
              aiWeightsData.weight_schemes && (
                <div className="bg-white rounded-md p-5">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="h-5 w-5 text-green-600" />
                      <span className="text-sm font-semibold text-[#333333]">
                        智能权重已完成
                      </span>
                    </div>
                    <Button
                      onClick={fetchAiWeights}
                      variant="outline"
                      size="sm"
                      className="h-8 px-3 text-xs"
                    >
                      <RefreshCw className="h-3 w-3 mr-1" />
                      再次计算
                    </Button>
                  </div>

                  {/* Tab导航 */}
                  <div className="flex gap-2 mb-4 border-b border-gray-200">
                    {aiWeightsData.weight_schemes.map((scheme) => {
                      const tabConfig = {
                        default: { label: 'AI综合推荐', color: 'blue' },
                        fresh_graduate: { label: '推荐方案2', color: 'green' },
                        experienced: { label: '推荐方案3', color: 'purple' },
                      };
                      const config = tabConfig[
                        scheme.type as keyof typeof tabConfig
                      ] || { label: scheme.type, color: 'gray' };
                      const isActive = activeTab === scheme.type;
                      const isSelected = selectedSchemeType === scheme.type;

                      return (
                        <button
                          key={scheme.type}
                          onClick={() => {
                            setActiveTab(
                              scheme.type as
                                | 'default'
                                | 'fresh_graduate'
                                | 'experienced'
                            );
                            applyWeightScheme(scheme);
                          }}
                          className={cn(
                            'flex-1 px-3 py-2 text-sm font-medium transition-all relative',
                            isActive
                              ? 'text-blue-600 border-b-2 border-blue-600'
                              : 'text-gray-500 hover:text-gray-700'
                          )}
                        >
                          <div className="flex items-center justify-center gap-2">
                            {config.label}
                            {isSelected && (
                              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs bg-green-100 text-green-700 border border-green-300">
                                已选中
                              </span>
                            )}
                          </div>
                        </button>
                      );
                    })}
                  </div>

                  {/* Tab内容 */}
                  {aiWeightsData.weight_schemes.map((scheme) => {
                    if (activeTab !== scheme.type) return null;

                    const descriptions = {
                      default:
                        '该策略采用综合权重评估体系，全面平衡职责匹配度、技能契合度等多维指标，旨在实现科学、高效的人才甄选，广泛适用于各类常规招聘场景。',
                      fresh_graduate:
                        '此权重策略重点关注候选人的教育背景、学习与适应能力及技能潜质，是为应届生和初级岗位量身定做的招聘策略，着眼于对企业未来人才的长远投资。',
                      experienced:
                        '该权重策略高度聚焦候选人过往实践经验与职位核心职责的精准契合度，旨在通过这种高度相关的筛选机制，显著提升关键资深岗位的招聘成功率与组织人效。',
                    };

                    return (
                      <div key={scheme.type} className="space-y-4">
                        {/* 方案描述文案区域 */}
                        <div className="p-3 bg-gray-50 rounded-md border border-gray-200">
                          <p className="text-sm text-gray-700 leading-relaxed">
                            {descriptions[
                              scheme.type as keyof typeof descriptions
                            ] || '权重推荐方案'}
                          </p>
                        </div>

                        {/* 推理说明 */}
                        {scheme.rationale && scheme.rationale.length > 0 && (
                          <div className="p-3 bg-blue-50 rounded-md border border-blue-100">
                            <p className="text-xs font-medium text-[#333333] mb-2">
                              💡 权重配比说明
                            </p>
                            <ul className="space-y-1">
                              {scheme.rationale.map((reason, index) => (
                                <li
                                  key={index}
                                  className="text-xs text-[#666666] pl-3 relative before:content-['•'] before:absolute before:left-0"
                                >
                                  {reason}
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}

                        {/* 权重详情（可展开） */}
                        <div className="border border-[#E8E8E8] rounded-md">
                          <button
                            onClick={() =>
                              setShowAiWeightsDetail(!showAiWeightsDetail)
                            }
                            className="w-full flex items-center justify-between px-4 py-3 hover:bg-[#FAFAFA] transition-colors"
                          >
                            <span className="text-sm font-medium text-[#333333]">
                              查看各维度权重值
                            </span>
                            {showAiWeightsDetail ? (
                              <ChevronUp className="h-4 w-4 text-[#666666]" />
                            ) : (
                              <ChevronDown className="h-4 w-4 text-[#666666]" />
                            )}
                          </button>

                          {showAiWeightsDetail && (
                            <div className="px-4 pb-4 pt-2 space-y-3 border-t border-[#E8E8E8]">
                              {Object.entries(weights).map(([key, value]) => {
                                const labels: Record<
                                  keyof WeightConfig,
                                  string
                                > = {
                                  basicInfo: '基本信息',
                                  responsibilities: '工作职责',
                                  skills: '工作技能',
                                  education: '教育经历',
                                  experience: '工作经历',
                                  industry: '行业背景',
                                };
                                return (
                                  <div
                                    key={key}
                                    className="flex items-center justify-between"
                                  >
                                    <span className="text-xs text-[#666666]">
                                      {labels[key as keyof WeightConfig]}
                                    </span>
                                    <span className="text-sm font-semibold text-primary">
                                      {value}%
                                    </span>
                                  </div>
                                );
                              })}
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
          </div>

          {/* 高级配置区域 */}
          <div className="bg-[#FAFAFA] rounded-lg border border-[#E8E8E8]">
            <button
              onClick={() => setAdvancedConfigExpanded(!advancedConfigExpanded)}
              className="w-full flex items-center justify-between px-3 py-1.5 hover:bg-[#F5F5F5] transition-colors rounded-t-lg group"
            >
              <div className="flex items-center gap-1.5">
                <Settings className="h-3 w-3 text-[#999999] group-hover:text-[#666666]" />
                <span className="text-xs text-[#666666]">高级配置</span>
              </div>
              {advancedConfigExpanded ? (
                <ChevronUp className="h-3 w-3 text-[#999999]" />
              ) : (
                <ChevronDown className="h-3 w-3 text-[#999999]" />
              )}
            </button>

            {advancedConfigExpanded && (
              <div className="px-4 pb-3 space-y-2">
                {/* 权重总和提示 - 精简版 */}
                <div className="bg-white rounded-md px-3 py-2 border border-gray-200 flex items-center justify-between">
                  <span className="text-xs text-[#666666]">权重总和</span>
                  <span
                    className={cn(
                      'text-lg font-bold',
                      isWeightValid ? 'text-primary' : 'text-red-500'
                    )}
                  >
                    {totalWeight}%
                  </span>
                </div>

                {/* 警告提示 - 精简版 */}
                {!isWeightValid && (
                  <div className="rounded-md px-3 py-2 bg-amber-50 border border-amber-200">
                    <p className="text-xs text-amber-700">
                      ⚠️ 权重总和必须等于100%
                    </p>
                  </div>
                )}

                {/* 各权重配置项 */}
                {renderWeightItem('👤', '基本信息', 'basicInfo')}
                {renderWeightItem('📋', '工作职责', 'responsibilities')}
                {renderWeightItem('⚙️', '工作技能', 'skills')}
                {renderWeightItem('🎓', '教育经历', 'education')}
                {renderWeightItem('💼', '工作经历', 'experience')}
                {renderWeightItem('🏢', '行业背景', 'industry')}
              </div>
            )}
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
                'px-6 text-white',
                !isWeightValid && 'bg-gray-300 cursor-not-allowed'
              )}
              style={
                isWeightValid
                  ? {
                      background:
                        'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                    }
                  : undefined
              }
            >
              下一步
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
