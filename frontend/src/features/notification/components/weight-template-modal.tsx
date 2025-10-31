import { useState, useEffect, useCallback } from 'react';
import {
  X,
  Plus,
  Edit,
  Trash2,
  ChevronLeft,
  ChevronRight,
  Loader2,
  AlertCircle,
} from 'lucide-react';
import { Button } from '@/ui/button';
import { Input } from '@/ui/input';
import { Label } from '@/ui/label';
import { Textarea } from '@/ui/textarea';
import { Slider } from '@/ui/slider';
import { ConfirmDialog } from '@/ui/confirm-dialog';
import { cn } from '@/lib/utils';
import { toast } from '@/ui/toast';
import {
  listWeightTemplates,
  createWeightTemplate,
  updateWeightTemplate,
  deleteWeightTemplate,
} from '@/services/weight-template';
import type {
  WeightTemplate,
  CreateWeightTemplateReq,
  UpdateWeightTemplateReq,
} from '@/types/weight-template';

interface WeightTemplateModalProps {
  isOpen: boolean;
  onClose: () => void;
}

// 权重字段配置
const WEIGHT_FIELDS = [
  {
    key: 'basic',
    label: '基本信息权重',
    emoji: '👤',
    desc: '性别、年龄、户籍等基本信息的匹配权重',
  },
  {
    key: 'responsibility',
    label: '工作职责权重',
    emoji: '📋',
    desc: '候选人过往工作职责与岗位职责的匹配权重',
  },
  {
    key: 'skill',
    label: '工作技能权重',
    emoji: '⚙️',
    desc: '候选人技能与岗位要求技能的匹配权重',
  },
  {
    key: 'education',
    label: '教育经历权重',
    emoji: '🎓',
    desc: '学历、专业、院校与岗位要求的匹配权重',
  },
  {
    key: 'experience',
    label: '工作经历权重',
    emoji: '💼',
    desc: '工作年限、工作经验与岗位要求的匹配权重',
  },
  {
    key: 'industry',
    label: '行业背景权重',
    emoji: '🏢',
    desc: '候选人所在行业与目标岗位行业的匹配权重',
  },
] as const;

export function WeightTemplateModal({
  isOpen,
  onClose,
}: WeightTemplateModalProps) {
  // 列表数据状态
  const [templates, setTemplates] = useState<WeightTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const pageSize = 7;

  // 表单状态
  const [isAddMode, setIsAddMode] = useState(false);
  const [isEditMode, setIsEditMode] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<WeightTemplate | null>(
    null
  );
  const [formData, setFormData] = useState<{
    name: string;
    description: string;
    weights: {
      basic: number;
      responsibility: number;
      skill: number;
      education: number;
      experience: number;
      industry: number;
    };
  }>({
    name: '',
    description: '',
    weights: {
      basic: 3,
      responsibility: 20,
      skill: 35,
      education: 15,
      experience: 20,
      industry: 7,
    },
  });

  // 删除确认弹窗状态
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [templateToDelete, setTemplateToDelete] =
    useState<WeightTemplate | null>(null);
  const [deleting, setDeleting] = useState(false);

  // 加载模板列表
  const fetchTemplates = useCallback(
    async (page?: number) => {
      try {
        setLoading(true);
        setError(null);
        const targetPage = page || currentPage;

        const response = await listWeightTemplates({
          page: targetPage,
          size: pageSize,
        });

        const items = response.items || [];
        const total = response.page_info?.total_count || 0;
        const newTotalPages = Math.ceil(total / pageSize);

        setTemplates(items);
        setTotalCount(total);
        setTotalPages(newTotalPages);

        // 如果当前页超出总页数，调整到最后一页
        if (targetPage > newTotalPages && newTotalPages > 0) {
          setCurrentPage(newTotalPages);
        }
      } catch (err) {
        setError('加载权重模板列表失败，请重试');
        console.error('加载权重模板列表失败:', err);
      } finally {
        setLoading(false);
      }
    },
    [currentPage]
  );

  // 初始化加载数据
  useEffect(() => {
    if (isOpen) {
      fetchTemplates();
    }
  }, [isOpen, currentPage, fetchTemplates]);

  // 计算总权重
  const totalWeight =
    formData.weights.basic +
    formData.weights.responsibility +
    formData.weights.skill +
    formData.weights.education +
    formData.weights.experience +
    formData.weights.industry;

  const isWeightValid = totalWeight === 100;

  // 更新权重值
  const updateWeight = (
    field: keyof typeof formData.weights,
    value: number
  ) => {
    setFormData({
      ...formData,
      weights: {
        ...formData.weights,
        [field]: value,
      },
    });
  };

  // 打开新增模式
  const handleAdd = () => {
    setIsAddMode(true);
    setIsEditMode(false);
    setEditingTemplate(null);
    setFormData({
      name: '',
      description: '',
      weights: {
        basic: 3,
        responsibility: 20,
        skill: 35,
        education: 15,
        experience: 20,
        industry: 7,
      },
    });
  };

  // 打开编辑模式
  const handleEdit = (template: WeightTemplate) => {
    setIsEditMode(true);
    setIsAddMode(true);
    setEditingTemplate(template);
    setFormData({
      name: template.name,
      description: template.description || '',
      weights: {
        basic: Math.round((template.weights.basic || 0) * 100),
        responsibility: Math.round(
          (template.weights.responsibility || 0) * 100
        ),
        skill: Math.round((template.weights.skill || 0) * 100),
        education: Math.round((template.weights.education || 0) * 100),
        experience: Math.round((template.weights.experience || 0) * 100),
        industry: Math.round((template.weights.industry || 0) * 100),
      },
    });
  };

  // 取消新增/编辑
  const handleCancel = () => {
    setIsAddMode(false);
    setIsEditMode(false);
    setEditingTemplate(null);
    setFormData({
      name: '',
      description: '',
      weights: {
        basic: 3,
        responsibility: 20,
        skill: 35,
        education: 15,
        experience: 20,
        industry: 7,
      },
    });
  };

  // 保存模板
  const handleSave = async () => {
    if (!formData.name.trim()) {
      toast.error('请输入模板名称');
      return;
    }

    if (!isWeightValid) {
      toast.error('权重总和必须等于100%');
      return;
    }

    try {
      setLoading(true);

      const requestData = {
        name: formData.name.trim(),
        description: formData.description.trim(),
        weights: {
          basic: formData.weights.basic / 100,
          responsibility: formData.weights.responsibility / 100,
          skill: formData.weights.skill / 100,
          education: formData.weights.education / 100,
          experience: formData.weights.experience / 100,
          industry: formData.weights.industry / 100,
        },
      };

      if (isEditMode && editingTemplate) {
        await updateWeightTemplate(
          editingTemplate.id,
          requestData as UpdateWeightTemplateReq
        );
        toast.success('权重模板更新成功');
      } else {
        await createWeightTemplate(requestData as CreateWeightTemplateReq);
        toast.success('权重模板创建成功');
      }

      handleCancel();
      fetchTemplates();
    } catch (err) {
      console.error('保存权重模板失败:', err);
      toast.error(
        isEditMode ? '更新权重模板失败，请重试' : '创建权重模板失败，请重试'
      );
    } finally {
      setLoading(false);
    }
  };

  // 打开删除确认
  const handleDeleteClick = (template: WeightTemplate) => {
    setTemplateToDelete(template);
    setIsDeleteConfirmOpen(true);
  };

  // 确认删除
  const handleDeleteConfirm = async () => {
    if (!templateToDelete) return;

    try {
      setDeleting(true);
      await deleteWeightTemplate(templateToDelete.id);
      toast.success('权重模板删除成功');

      // 删除后检查当前页是否还有数据
      const remainingItems = totalCount - 1;
      const maxPage = Math.ceil(remainingItems / pageSize);
      const targetPage =
        currentPage > maxPage ? Math.max(1, maxPage) : currentPage;

      if (targetPage !== currentPage) {
        setCurrentPage(targetPage);
      }

      await fetchTemplates(targetPage);
      setIsDeleteConfirmOpen(false);
      setTemplateToDelete(null);
    } catch (err) {
      console.error('删除权重模板失败:', err);
      toast.error('删除权重模板失败，请重试');
    } finally {
      setDeleting(false);
    }
  };

  // 生成页码数组
  const generatePageNumbers = () => {
    const pages = [];
    const safeTotalPages = totalPages || 1;
    const safeCurrentPage = currentPage || 1;

    const start = Math.max(1, safeCurrentPage - 2);
    const end = Math.min(safeTotalPages, safeCurrentPage + 2);

    for (let i = start; i <= end; i++) {
      pages.push(i);
    }

    return pages;
  };

  // 格式化时间
  const formatDate = (dateString: string) => {
    try {
      return new Date(dateString).toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
      });
    } catch {
      return dateString;
    }
  };

  if (!isOpen) return null;

  return (
    <>
      <div
        className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 p-4"
        onClick={onClose}
      >
        <div
          className="bg-white rounded-lg shadow-xl w-[900px] max-h-[90vh] overflow-hidden flex flex-col"
          onClick={(e) => e.stopPropagation()}
        >
          {/* 弹窗头部 */}
          <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
            <h2 className="text-lg font-medium text-gray-900">
              智能匹配权重配置
            </h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
            >
              <X className="w-6 h-6" />
            </button>
          </div>

          {/* 弹窗内容 */}
          <div className="flex-1 overflow-y-auto p-6">
            {isAddMode ? (
              /* 新增/编辑表单 */
              <div className="space-y-6">
                <div className="space-y-4">
                  <div>
                    <Label
                      htmlFor="name"
                      className="text-sm font-medium text-gray-700"
                    >
                      模板名称 <span className="text-red-500">*</span>
                    </Label>
                    <Input
                      id="name"
                      value={formData.name}
                      onChange={(e) =>
                        setFormData({ ...formData, name: e.target.value })
                      }
                      placeholder="请输入模板名称（最多100个字符）"
                      maxLength={100}
                      className="mt-1"
                    />
                  </div>

                  <div>
                    <Label
                      htmlFor="description"
                      className="text-sm font-medium text-gray-700"
                    >
                      模板描述
                    </Label>
                    <Textarea
                      id="description"
                      value={formData.description}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          description: e.target.value,
                        })
                      }
                      placeholder="请输入模板描述（最多500个字符）"
                      maxLength={500}
                      rows={3}
                      className="mt-1"
                    />
                  </div>
                </div>

                {/* 权重配置 */}
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-semibold text-gray-900">
                      权重配置
                    </h3>
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-gray-600">总和:</span>
                      <span
                        className={cn(
                          'text-lg font-bold',
                          isWeightValid ? 'text-green-600' : 'text-red-600'
                        )}
                      >
                        {totalWeight}%
                      </span>
                    </div>
                  </div>

                  {!isWeightValid && (
                    <div className="p-3 bg-amber-50 border border-amber-200 rounded-md">
                      <p className="text-sm text-amber-700">
                        ⚠️ 权重总和必须等于100%，当前为{totalWeight}%
                      </p>
                    </div>
                  )}

                  {WEIGHT_FIELDS.map((field) => (
                    <div key={field.key} className="bg-gray-50 rounded-md p-4">
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2">
                          <span className="text-sm">{field.emoji}</span>
                          <span className="text-sm font-semibold text-gray-900">
                            {field.label}
                          </span>
                        </div>
                        <div className="flex items-center gap-1">
                          <Input
                            type="number"
                            min="0"
                            max="100"
                            value={
                              formData.weights[
                                field.key as keyof typeof formData.weights
                              ]
                            }
                            onChange={(e) =>
                              updateWeight(
                                field.key as keyof typeof formData.weights,
                                parseInt(e.target.value) || 0
                              )
                            }
                            className="w-[70px] h-[32px] px-2 text-center text-base font-bold text-primary"
                          />
                          <span className="text-base font-semibold text-primary">
                            %
                          </span>
                        </div>
                      </div>
                      <Slider
                        value={[
                          formData.weights[
                            field.key as keyof typeof formData.weights
                          ],
                        ]}
                        onValueChange={(value: number[]) =>
                          updateWeight(
                            field.key as keyof typeof formData.weights,
                            value[0]
                          )
                        }
                        min={0}
                        max={100}
                        step={1}
                        className="mb-2"
                      />
                      <p className="text-xs text-gray-500">{field.desc}</p>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              /* 列表视图 */
              <div>
                {/* 操作按钮 */}
                <div className="mb-4 flex justify-end">
                  <Button
                    onClick={handleAdd}
                    className="gap-2 rounded-lg px-4 py-2 shadow-sm text-white"
                    style={{
                      background:
                        'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                    }}
                  >
                    <Plus className="w-4 h-4" />
                    新增模板
                  </Button>
                </div>

                {/* 表格 */}
                <div className="border border-gray-200 rounded-lg overflow-hidden">
                  <table className="w-full">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          模板名称
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          描述
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          创建时间
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          操作
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {loading ? (
                        <tr>
                          <td colSpan={4} className="px-4 py-8 text-center">
                            <div className="flex items-center justify-center gap-2">
                              <Loader2 className="w-5 h-5 animate-spin text-gray-400" />
                              <span className="text-sm text-gray-500">
                                加载中...
                              </span>
                            </div>
                          </td>
                        </tr>
                      ) : error ? (
                        <tr>
                          <td colSpan={4} className="px-4 py-8 text-center">
                            <div className="flex flex-col items-center gap-2">
                              <AlertCircle className="w-8 h-8 text-red-500" />
                              <p className="text-sm text-red-500">{error}</p>
                              <button
                                onClick={() => fetchTemplates()}
                                className="text-sm text-blue-500 hover:text-blue-600"
                              >
                                重试
                              </button>
                            </div>
                          </td>
                        </tr>
                      ) : templates.length === 0 ? (
                        <tr>
                          <td
                            colSpan={4}
                            className="px-4 py-8 text-center text-sm text-gray-500"
                          >
                            暂无权重模板，点击"新增模板"创建
                          </td>
                        </tr>
                      ) : (
                        templates.map((template) => (
                          <tr key={template.id} className="hover:bg-gray-50">
                            <td className="px-4 py-4 text-sm font-medium text-gray-900">
                              {template.name}
                            </td>
                            <td className="px-4 py-4 text-sm text-gray-500">
                              {template.description || '-'}
                            </td>
                            <td className="px-4 py-4 text-sm text-gray-500">
                              {formatDate(template.created_at)}
                            </td>
                            <td className="px-4 py-4 text-sm font-medium">
                              <div className="flex items-center gap-2">
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="h-8 w-8 p-0 text-gray-400 hover:text-gray-600"
                                  onClick={() => handleEdit(template)}
                                  title="编辑"
                                >
                                  <Edit className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="h-8 w-8 p-0 text-gray-400 hover:text-red-600"
                                  onClick={() => handleDeleteClick(template)}
                                  title="删除"
                                >
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              </div>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>

                {/* 分页 */}
                {!loading && !error && templates.length > 0 && (
                  <div className="mt-4 flex items-center justify-between">
                    <div className="text-sm text-gray-700">
                      显示 {(currentPage - 1) * pageSize + 1} 到{' '}
                      {Math.min(currentPage * pageSize, totalCount)} 条，共{' '}
                      {totalCount} 条
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          setCurrentPage(Math.max(1, currentPage - 1))
                        }
                        disabled={currentPage === 1}
                        className="h-8 w-8 p-0"
                      >
                        <ChevronLeft className="h-4 w-4" />
                      </Button>
                      {generatePageNumbers().map((page) => (
                        <Button
                          key={page}
                          variant="outline"
                          size="sm"
                          onClick={() => setCurrentPage(page)}
                          className={cn(
                            'h-8 w-8 p-0',
                            page === currentPage &&
                              'text-white border-primary hover:text-white'
                          )}
                          style={
                            page === currentPage
                              ? {
                                  background:
                                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                                }
                              : undefined
                          }
                        >
                          {page}
                        </Button>
                      ))}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          setCurrentPage(Math.min(totalPages, currentPage + 1))
                        }
                        disabled={currentPage >= totalPages}
                        className="h-8 w-8 p-0"
                      >
                        <ChevronRight className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* 弹窗底部 */}
          <div className="flex justify-end gap-3 px-6 py-4 border-t border-gray-100 bg-gray-50">
            {isAddMode ? (
              <>
                <Button
                  variant="outline"
                  onClick={handleCancel}
                  className="px-6"
                >
                  取消
                </Button>
                <Button
                  onClick={handleSave}
                  disabled={loading || !isWeightValid}
                  className="px-6 text-white"
                  style={{
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }}
                >
                  {loading ? (
                    <>
                      <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                      保存中...
                    </>
                  ) : (
                    '保存'
                  )}
                </Button>
              </>
            ) : (
              <Button variant="outline" onClick={onClose} className="px-6">
                关闭
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* 删除确认对话框 */}
      <ConfirmDialog
        open={isDeleteConfirmOpen}
        onOpenChange={setIsDeleteConfirmOpen}
        onConfirm={handleDeleteConfirm}
        title="删除权重模板"
        description={`确定要删除权重模板"${templateToDelete?.name}"吗？此操作不可撤销。`}
        confirmText="删除"
        cancelText="取消"
        variant="destructive"
        loading={deleting}
      />
    </>
  );
}
