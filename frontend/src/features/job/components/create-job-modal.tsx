import { useState, useEffect, useCallback, Fragment } from 'react';
import { X, Settings, Trash2, FileText } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/ui/dialog';
import { Button } from '@/ui/button';
import { Input } from '@/ui/input';
import { Label } from '@/ui/label';
import { cn } from '@/lib/utils';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/ui/select';
import {
  createJobProfile,
  parseJobProfile,
  listJobSkillMeta,
} from '@/services/job-profile';
import { listDepartments } from '@/services/department';
import { toast } from '@/ui/toast';

interface CreateJobModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void; // 成功创建后的回调函数
}

export function CreateJobModal({
  open,
  onOpenChange,
  onSuccess,
}: CreateJobModalProps) {
  const [editMode, setEditMode] = useState<'manual' | 'ai'>('manual');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isParsing, setIsParsing] = useState(false);
  const [formData, setFormData] = useState({
    description: '',
    jobName: '',
    educationRequirement: '',
    workType: '',
    department: '',
    location: '',
    salaryMin: '',
    salaryMax: '',
    industryRequirement: '',
    companyRequirement: '',
    workExperience: '',
  });

  // 技能列表 - 存储完整的技能对象
  const [skills, setSkills] = useState<
    Array<{
      skill_id: string;
      skill_name: string;
      type: 'required' | 'bonus';
    }>
  >([]);

  // 部门列表
  const [departments, setDepartments] = useState<
    Array<{ id: string; name: string }>
  >([]);
  const [isLoadingDepartments, setIsLoadingDepartments] = useState(false);

  // 技能选择弹窗状态
  const [isSkillSelectModalOpen, setIsSkillSelectModalOpen] = useState(false);
  const [availableSkills, setAvailableSkills] = useState<
    Array<{ id: string; name: string }>
  >([]);
  const [selectedSkillIds, setSelectedSkillIds] = useState<string[]>([]); // 改为数组支持多选
  const [isLoadingSkills, setIsLoadingSkills] = useState(false);

  // 岗位职责列表
  const [responsibilities, setResponsibilities] = useState<
    Array<{ content: string; id: string }>
  >([{ content: '', id: '1' }]);

  // 重置表单
  const resetForm = useCallback(() => {
    setFormData({
      description: '',
      jobName: '',
      educationRequirement: '',
      workType: '',
      department: '',
      location: '',
      salaryMin: '',
      salaryMax: '',
      industryRequirement: '',
      companyRequirement: '',
      workExperience: '',
    });
    setSkills([]);
    setResponsibilities([{ content: '', id: '1' }]);
  }, []);

  // 获取部门列表
  const fetchDepartments = useCallback(async () => {
    setIsLoadingDepartments(true);
    try {
      const result = await listDepartments({ page: 1, size: 100 });
      if (result && result.items) {
        setDepartments(
          result.items.map((dept) => ({
            id: dept.id,
            name: dept.name,
          }))
        );
      }
    } catch (error) {
      console.error('获取部门列表失败:', error);
      toast.error('获取部门列表失败');
    } finally {
      setIsLoadingDepartments(false);
    }
  }, []);

  // 监听弹窗打开状态，打开时重置表单并加载部门
  useEffect(() => {
    if (open) {
      resetForm();
      fetchDepartments();
    }
  }, [open, resetForm, fetchDepartments]);

  // 处理新增部门按钮点击
  const handleAddDepartment = () => {
    window.open('/platform-config?openModal=true', '_blank');
  };

  // 获取技能列表
  const fetchSkills = async () => {
    setIsLoadingSkills(true);
    try {
      const result = await listJobSkillMeta({ page: 1, size: 100 });
      if (result && result.items) {
        setAvailableSkills(
          result.items.map((skill) => ({
            id: skill.id,
            name: skill.name,
          }))
        );
      }
    } catch (error) {
      console.error('获取技能列表失败:', error);
      toast.error('获取技能列表失败');
    } finally {
      setIsLoadingSkills(false);
    }
  };

  // 处理新增技能按钮点击 - 打开技能选择弹窗
  const handleAddSkill = () => {
    setIsSkillSelectModalOpen(true);
    fetchSkills();
  };

  // 处理技能选择确认 - 支持多选
  const handleConfirmSkill = () => {
    if (selectedSkillIds.length === 0) {
      toast.error('请至少选择一个技能');
      return;
    }

    // 过滤出新技能（未添加的）
    const newSkills: Array<{
      skill_id: string;
      skill_name: string;
      type: 'required' | 'bonus';
    }> = [];

    for (const id of selectedSkillIds) {
      // 检查是否已添加
      if (skills.some((s) => s.skill_id === id)) {
        continue;
      }

      // 查找技能
      const skill = availableSkills.find((s) => s.id === id);
      if (skill) {
        newSkills.push({
          skill_id: skill.id,
          skill_name: skill.name,
          type: 'required', // 默认为必需技能
        });
      }
    }

    if (newSkills.length === 0) {
      toast.error('所选技能已全部添加');
      return;
    }

    // 添加技能到列表
    setSkills([...skills, ...newSkills]);

    // 显示成功消息
    toast.success(`成功添加 ${newSkills.length} 个技能`);

    // 关闭弹窗并重置选择
    setIsSkillSelectModalOpen(false);
    setSelectedSkillIds([]);
  };

  // 打开技能配置页面
  const handleOpenSkillConfig = () => {
    window.open('/platform-config?openSkillModal=true', '_blank');
  };

  // 添加岗位职责
  const addResponsibility = () => {
    setResponsibilities([
      ...responsibilities,
      { content: '', id: Date.now().toString() },
    ]);
  };

  // 删除岗位职责
  const removeResponsibility = (id: string) => {
    if (responsibilities.length > 1) {
      setResponsibilities(responsibilities.filter((r) => r.id !== id));
    }
  };

  // 更新岗位职责内容
  const updateResponsibility = (id: string, content: string) => {
    setResponsibilities(
      responsibilities.map((r) => (r.id === id ? { ...r, content } : r))
    );
  };

  // 处理识别按钮点击 - 调用解析接口
  const handleParse = async () => {
    if (!formData.description.trim()) {
      toast.error('请先输入岗位描述内容');
      return;
    }

    setIsParsing(true);
    try {
      const result = await parseJobProfile(formData.description);

      // 用解析结果填充表单
      if (result) {
        // 映射学历要求
        let educationReq = formData.educationRequirement;
        if (
          result.education_requirements &&
          result.education_requirements.length > 0
        ) {
          const eduType = result.education_requirements[0].education_type;
          // 映射后端返回的学历类型到前端的值
          const eduMap: Record<string, string> = {
            unlimited: 'unlimited',
            junior_college: 'junior_college',
            bachelor: 'bachelor',
            master: 'master',
            doctor: 'doctor',
          };
          educationReq = eduMap[eduType] || educationReq;
        }

        // 映射工作经验
        let workExp = formData.workExperience;
        if (
          result.experience_requirements &&
          result.experience_requirements.length > 0
        ) {
          const expType = result.experience_requirements[0].experience_type;
          // 直接使用后端返回的experience_type
          const expMap: Record<string, string> = {
            unlimited: 'unlimited',
            fresh_graduate: 'fresh_graduate',
            under_one_year: 'under_one_year',
            one_to_three_years: 'one_to_three_years',
            three_to_five_years: 'three_to_five_years',
            five_to_ten_years: 'five_to_ten_years',
            over_ten_years: 'over_ten_years',
          };
          workExp = expMap[expType] || workExp;
        }

        // 映射工作性质
        let workType = formData.workType;
        if (result.work_type) {
          const typeMap: Record<string, string> = {
            full_time: 'full_time',
            part_time: 'part_time',
            internship: 'internship',
            outsourcing: 'outsourcing',
          };
          workType = typeMap[result.work_type] || workType;
        }

        // 解析行业与公司要求
        let industryReq = formData.industryRequirement;
        let companyReq = formData.companyRequirement;
        if (
          result.industry_requirements &&
          Array.isArray(result.industry_requirements)
        ) {
          // 从返回的行业要求中提取行业和公司名称
          result.industry_requirements.forEach(
            (req: { industry?: string; company_name?: string }) => {
              if (req.industry && !industryReq) {
                industryReq = req.industry;
              }
              if (req.company_name && !companyReq) {
                companyReq = req.company_name;
              }
            }
          );
        }

        // 更新表单数据
        setFormData({
          ...formData,
          jobName: result.name || formData.jobName,
          location: result.location || formData.location,
          salaryMin: result.salary_min
            ? result.salary_min.toString()
            : formData.salaryMin,
          salaryMax: result.salary_max
            ? result.salary_max.toString()
            : formData.salaryMax,
          educationRequirement: educationReq,
          workExperience: workExp,
          workType: workType,
          industryRequirement: industryReq,
          companyRequirement: companyReq,
        });

        // 如果有技能数据，更新技能列表（保存完整的技能对象）
        if (result.skills && Array.isArray(result.skills)) {
          const skillList = result.skills.map(
            (s: {
              skill_id?: string;
              skill_name?: string;
              name?: string;
              type?: string;
            }) => ({
              skill_id: s.skill_id || '',
              skill_name: s.skill_name || s.name || '',
              type: (s.type === 'bonus' ? 'bonus' : 'required') as
                | 'required'
                | 'bonus',
            })
          );
          setSkills(skillList);
        }

        // 如果有职责数据，更新职责列表
        if (result.responsibilities && Array.isArray(result.responsibilities)) {
          const respList = result.responsibilities.map(
            (
              r: { responsibility?: string; description?: string } | string,
              index: number
            ) => ({
              id: (index + 1).toString(),
              content:
                typeof r === 'string'
                  ? r
                  : r.responsibility || r.description || '',
            })
          );
          setResponsibilities(
            respList.length > 0 ? respList : [{ content: '', id: '1' }]
          );
        }

        toast.success('岗位描述解析完成');
      }
    } catch (error) {
      console.error('解析失败:', error);
      const errorMessage =
        error instanceof Error ? error.message : '请稍后重试';
      toast.error(`解析失败: ${errorMessage}`);
    } finally {
      setIsParsing(false);
    }
  };

  // 处理表单提交 - 调用创建接口
  const handleSubmit = async (status: 'draft' | 'published') => {
    // 基本验证
    if (!formData.jobName.trim()) {
      toast.error('请输入岗位名称');
      return;
    }

    if (!formData.department) {
      toast.error('请选择所属部门');
      return;
    }

    setIsSubmitting(true);
    try {
      // 构建请求数据
      const requestData: {
        name: string;
        department_id: string;
        description: string;
        location: string;
        work_type: string;
        status: string;
        salary_min?: number;
        salary_max?: number;
        education_requirements?: Array<{ education_type: string }>;
        experience_requirements?: Array<{ experience_type: string }>;
        industry_requirements?: Array<{
          industry: string;
          company_name: string;
        }>;
        skills?: Array<{
          skill_id: string;
          skill_name: string;
          type: 'required' | 'bonus';
        }>;
        responsibilities?: Array<{ responsibility: string }>;
      } = {
        name: formData.jobName,
        department_id: formData.department,
        description: formData.description,
        location: formData.location,
        work_type: formData.workType,
        status: status, // 添加状态字段
      };

      // 添加薪资范围
      if (formData.salaryMin) {
        requestData.salary_min = parseFloat(formData.salaryMin);
      }
      if (formData.salaryMax) {
        requestData.salary_max = parseFloat(formData.salaryMax);
      }

      // 添加学历要求
      if (formData.educationRequirement) {
        requestData.education_requirements = [
          {
            education_type: formData.educationRequirement,
          },
        ];
      }

      // 添加工作经验要求
      if (formData.workExperience) {
        requestData.experience_requirements = [
          {
            experience_type: formData.workExperience,
          },
        ];
      }

      // 添加行业要求
      if (formData.industryRequirement || formData.companyRequirement) {
        requestData.industry_requirements = [];
        // 如果两者都有，合并为一个对象
        if (formData.industryRequirement && formData.companyRequirement) {
          requestData.industry_requirements.push({
            industry: formData.industryRequirement,
            company_name: formData.companyRequirement,
          });
        } else if (formData.industryRequirement) {
          requestData.industry_requirements.push({
            industry: formData.industryRequirement,
            company_name: '', // 提供默认值
          });
        } else if (formData.companyRequirement) {
          requestData.industry_requirements.push({
            industry: '', // 提供默认值
            company_name: formData.companyRequirement,
          });
        }
      }

      // 添加技能要求（按照API格式传递完整的技能对象）
      if (skills.length > 0) {
        requestData.skills = skills.map((skill) => ({
          skill_id: skill.skill_id,
          skill_name: skill.skill_name,
          type: skill.type,
        }));
      }

      // 添加岗位职责
      const validResponsibilities = responsibilities.filter((r) =>
        r.content.trim()
      );
      if (validResponsibilities.length > 0) {
        requestData.responsibilities = validResponsibilities.map((r) => ({
          responsibility: r.content,
        }));
      }

      await createJobProfile(requestData);

      // 根据状态显示不同的成功消息
      if (status === 'draft') {
        toast.success('岗位画像已保存为草稿');
      } else {
        toast.success('岗位画像已发布');
      }

      // 调用成功回调刷新列表
      if (onSuccess) {
        onSuccess();
      }

      // 关闭弹窗并重置表单
      onOpenChange(false);
      resetForm();
    } catch (error) {
      console.error('创建失败:', error);
      const errorMessage =
        error instanceof Error ? error.message : '请稍后重试';
      toast.error(`创建失败: ${errorMessage}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Fragment>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-[920px] max-h-[90vh] overflow-hidden p-0 gap-0">
          {/* 顶部标题栏 */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-white">
            <DialogHeader className="space-y-0">
              <DialogTitle className="text-[18px] font-semibold text-gray-900">
                新建岗位
              </DialogTitle>
            </DialogHeader>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onOpenChange(false)}
              className="h-8 w-8 p-0 rounded-full hover:bg-gray-100"
            >
              <X className="h-5 w-5 text-gray-500" />
            </Button>
          </div>

          {/* 内容区域 */}
          <div className="overflow-y-auto max-h-[calc(90vh-140px)] bg-gray-50 px-6 py-6">
            {/* 选择编辑模式区域 */}
            <div className="mb-6">
              <div className="flex items-center gap-2 mb-4">
                <Settings className="h-5 w-5 text-gray-700" />
                <h3 className="text-[16px] font-medium text-gray-900">
                  选择编辑模式
                </h3>
              </div>

              <div className="grid grid-cols-2 gap-4">
                {/* 手动编辑模式 */}
                <div
                  onClick={() => setEditMode('manual')}
                  className={cn(
                    'relative flex items-start gap-3 p-4 rounded-lg border-2 cursor-pointer transition-all',
                    editMode === 'manual'
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 bg-white hover:border-gray-300'
                  )}
                >
                  <div className="flex items-center justify-center mt-0.5">
                    <div
                      className={cn(
                        'w-5 h-5 rounded-full border-2 flex items-center justify-center transition-colors',
                        editMode === 'manual'
                          ? 'border-blue-500 bg-white'
                          : 'border-gray-300 bg-white'
                      )}
                    >
                      {editMode === 'manual' && (
                        <div className="w-3 h-3 rounded-full bg-blue-500"></div>
                      )}
                    </div>
                  </div>
                  <div className="flex-1">
                    <div className="text-[15px] font-medium text-gray-900 mb-1">
                      手动编辑模式
                    </div>
                    <div className="text-[13px] text-gray-500">
                      完全自定义编辑，精细控制
                    </div>
                  </div>
                </div>

                {/* AI编辑模式 */}
                <div
                  onClick={() => setEditMode('ai')}
                  className={cn(
                    'relative flex items-start gap-3 p-4 rounded-lg border-2 cursor-pointer transition-all',
                    editMode === 'ai'
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 bg-white hover:border-gray-300'
                  )}
                >
                  <div className="flex items-center justify-center mt-0.5">
                    <div
                      className={cn(
                        'w-5 h-5 rounded-full border-2 flex items-center justify-center transition-colors',
                        editMode === 'ai'
                          ? 'border-blue-500 bg-white'
                          : 'border-gray-300 bg-white'
                      )}
                    >
                      {editMode === 'ai' && (
                        <div className="w-3 h-3 rounded-full bg-blue-500"></div>
                      )}
                    </div>
                  </div>
                  <div className="flex-1">
                    <div className="text-[15px] font-medium text-gray-900 mb-1">
                      AI编辑模式
                    </div>
                    <div className="text-[13px] text-gray-500">
                      智能生成岗位内容，快速高效
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* 基本信息区域 */}
            {editMode === 'manual' && (
              <div>
                {/* 基本信息标题 */}
                <div className="flex items-center gap-2 mb-5">
                  <FileText className="h-5 w-5 text-gray-700" />
                  <h3 className="text-[16px] font-medium text-gray-900">
                    基本信息
                  </h3>
                </div>

                <div className="space-y-5">
                  {/* 描述内容 - 识别按钮在框体外右上角 */}
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-[14px] font-medium text-gray-700">
                        描述内容:
                      </Label>
                      {/* 识别按钮 - 框体外右上角 */}
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={handleParse}
                        disabled={isParsing || !formData.description.trim()}
                        className="h-7 px-3 text-[13px] text-gray-900 border border-gray-300 hover:bg-gray-50 rounded bg-white disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {isParsing ? '识别中...' : '识别'}
                      </Button>
                    </div>
                    <div className="relative">
                      <textarea
                        placeholder="请在这里输入该岗位的职责说明，此后你同样可以简单的说明你的职责说明"
                        value={formData.description}
                        onChange={(e) =>
                          setFormData({
                            ...formData,
                            description: e.target.value,
                          })
                        }
                        className="w-full min-h-[120px] text-[14px] border border-gray-300 rounded-lg p-3 pl-3 pr-3 pt-3 pb-8 focus:outline-none bg-white text-left align-top"
                        style={{
                          resize: 'vertical',
                          textAlign: 'left',
                          verticalAlign: 'top',
                        }}
                        onFocus={(e) => {
                          e.currentTarget.style.borderColor = '#7bb8ff';
                          e.currentTarget.style.boxShadow =
                            '0 0 0 3px rgba(123, 184, 255, 0.1)';
                        }}
                        onBlur={(e) => {
                          e.currentTarget.style.borderColor = '#d1d5db';
                          e.currentTarget.style.boxShadow = 'none';
                        }}
                      />
                      {/* 字符计数 - 框体内右下角 */}
                      <div className="absolute bottom-2 right-3 text-[12px] text-gray-400">
                        {formData.description.length}/4000
                      </div>
                    </div>
                  </div>

                  {/* 第一行：岗位名称、学历要求 */}
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label className="text-[14px] font-medium text-gray-700">
                        岗位名称:
                      </Label>
                      <Input
                        placeholder="请输入岗位名称，如产品经理"
                        value={formData.jobName}
                        onChange={(e) =>
                          setFormData({ ...formData, jobName: e.target.value })
                        }
                        className="text-[14px] border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-100 rounded-lg bg-white"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label className="text-[14px] font-medium text-gray-700">
                        学历要求:
                      </Label>
                      <Select
                        value={formData.educationRequirement}
                        onValueChange={(value) =>
                          setFormData({
                            ...formData,
                            educationRequirement: value,
                          })
                        }
                      >
                        <SelectTrigger className="text-[14px] border-gray-300 focus:border-blue-500 rounded-lg bg-white">
                          <SelectValue placeholder="请选择" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="unlimited">不限</SelectItem>
                          <SelectItem value="junior_college">大专</SelectItem>
                          <SelectItem value="bachelor">本科</SelectItem>
                          <SelectItem value="master">硕士</SelectItem>
                          <SelectItem value="doctor">博士</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  {/* 第二行：工作性质、所属部门(新增按钮字体黑色,带边框) */}
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label className="text-[14px] font-medium text-gray-700">
                        工作性质:
                      </Label>
                      <Select
                        value={formData.workType}
                        onValueChange={(value) =>
                          setFormData({ ...formData, workType: value })
                        }
                      >
                        <SelectTrigger className="text-[14px] border-gray-300 focus:border-blue-500 rounded-lg bg-white">
                          <SelectValue placeholder="请选择" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="full_time">全职</SelectItem>
                          <SelectItem value="part_time">兼职</SelectItem>
                          <SelectItem value="internship">实习</SelectItem>
                          <SelectItem value="outsourcing">外包</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <Label className="text-[14px] font-medium text-gray-700">
                          所属部门:
                        </Label>
                        {/* 新增按钮 - 黑色字体,带边框 */}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={handleAddDepartment}
                          className="h-7 px-2 text-[13px] text-gray-900 hover:bg-gray-100 border border-gray-300 rounded"
                        >
                          + 新增
                        </Button>
                      </div>
                      <Select
                        value={formData.department}
                        onValueChange={(value) =>
                          setFormData({ ...formData, department: value })
                        }
                        disabled={isLoadingDepartments}
                      >
                        <SelectTrigger className="text-[14px] border-gray-300 focus:border-blue-500 rounded-lg bg-white">
                          <SelectValue
                            placeholder={
                              isLoadingDepartments ? '加载中...' : '请选择'
                            }
                          />
                        </SelectTrigger>
                        <SelectContent>
                          {departments.map((dept) => (
                            <SelectItem key={dept.id} value={dept.id}>
                              {dept.name}
                            </SelectItem>
                          ))}
                          {departments.length === 0 &&
                            !isLoadingDepartments && (
                              <SelectItem value="" disabled>
                                暂无部门
                              </SelectItem>
                            )}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  {/* 第三行：工作地点、薪资范围 */}
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label className="text-[14px] font-medium text-gray-700">
                        工作地点:
                      </Label>
                      <Input
                        placeholder="如：北京，输入多个用，隔开"
                        value={formData.location}
                        onChange={(e) =>
                          setFormData({ ...formData, location: e.target.value })
                        }
                        className="text-[14px] border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-100 rounded-lg bg-white"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label className="text-[14px] font-medium text-gray-700">
                        薪资范围:
                      </Label>
                      <div className="flex items-center gap-3">
                        <div className="relative flex-1">
                          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-[14px] text-gray-400">
                            如：
                          </span>
                          <Input
                            placeholder="10000元/月"
                            value={formData.salaryMin}
                            onChange={(e) =>
                              setFormData({
                                ...formData,
                                salaryMin: e.target.value,
                              })
                            }
                            className="pl-12 text-[14px] border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-100 rounded-lg bg-white"
                          />
                        </div>
                        <span className="text-gray-500 font-medium">—</span>
                        <Input
                          placeholder="15000元/月"
                          value={formData.salaryMax}
                          onChange={(e) =>
                            setFormData({
                              ...formData,
                              salaryMax: e.target.value,
                            })
                          }
                          className="flex-1 text-[14px] border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-100 rounded-lg bg-white"
                        />
                      </div>
                    </div>
                  </div>

                  {/* 第四行：行业与公司要求、工作经验 - 在同一行 */}
                  <div className="grid grid-cols-2 gap-4">
                    {/* 行业与公司要求 */}
                    <div className="space-y-2">
                      <Label className="text-[14px] font-medium text-gray-700">
                        行业与公司要求:
                      </Label>
                      <div className="flex gap-2">
                        <Input
                          placeholder="如：互联网"
                          value={formData.industryRequirement}
                          onChange={(e) =>
                            setFormData({
                              ...formData,
                              industryRequirement: e.target.value,
                            })
                          }
                          className="flex-1 text-[14px] border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-100 rounded-lg bg-white"
                        />
                        <Input
                          placeholder="如：互联网"
                          value={formData.companyRequirement}
                          onChange={(e) =>
                            setFormData({
                              ...formData,
                              companyRequirement: e.target.value,
                            })
                          }
                          className="flex-1 text-[14px] border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-100 rounded-lg bg-white"
                        />
                      </div>
                    </div>

                    {/* 工作经验 */}
                    <div className="space-y-2">
                      <Label className="text-[14px] font-medium text-gray-700">
                        工作经验:
                      </Label>
                      <Select
                        value={formData.workExperience}
                        onValueChange={(value) =>
                          setFormData({ ...formData, workExperience: value })
                        }
                      >
                        <SelectTrigger className="text-[14px] border-gray-300 focus:border-blue-500 rounded-lg bg-white">
                          <SelectValue placeholder="请选择" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="unlimited">不限</SelectItem>
                          <SelectItem value="fresh_graduate">
                            应届毕业生
                          </SelectItem>
                          <SelectItem value="under_one_year">
                            1年以下
                          </SelectItem>
                          <SelectItem value="one_to_three_years">
                            1-3年
                          </SelectItem>
                          <SelectItem value="three_to_five_years">
                            3-5年
                          </SelectItem>
                          <SelectItem value="five_to_ten_years">
                            5-10年
                          </SelectItem>
                          <SelectItem value="over_ten_years">
                            10年以上
                          </SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  {/* 工作技能 - 标签展示 */}
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-[14px] font-medium text-gray-700">
                        工作技能:
                      </Label>
                      {/* 新增按钮 */}
                      <button
                        onClick={handleAddSkill}
                        className="h-7 px-2 text-[13px] text-gray-900 hover:bg-gray-100 flex items-center gap-1 border border-gray-300 rounded"
                      >
                        + 新增
                      </button>
                    </div>

                    {/* 标签展示区域 */}
                    <div className="flex flex-wrap gap-2 min-h-[40px] items-center">
                      {skills.map((skill, index) => (
                        <div
                          key={index}
                          className="inline-flex items-center gap-1 px-2 py-1 bg-gray-100 border border-gray-300 rounded text-[13px] text-gray-700"
                        >
                          <span className="text-gray-700">
                            {skill.skill_name}
                          </span>
                          <button
                            onClick={() =>
                              setSkills(skills.filter((_, i) => i !== index))
                            }
                            className="text-gray-400 hover:text-red-500 transition-colors"
                            title={`删除技能: ${skill.skill_name}`}
                          >
                            <X className="h-3 w-3" />
                          </button>
                        </div>
                      ))}
                      {skills.length === 0 && (
                        <div className="text-[13px] text-gray-400">
                          暂无工作技能
                        </div>
                      )}
                    </div>

                    {/* 底部说明文字 */}
                    <div className="text-[12px] text-gray-400">
                      说明：工作技能是通过职位描述中的信息进行自动识别，支持手动添加
                    </div>
                  </div>

                  {/* 岗位职责与要求 - 支持多个框体,黑色按钮字体 */}
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-[14px] font-medium text-gray-700">
                        岗位职责与要求:
                      </Label>
                      {/* 新增按钮 - 黑色字体,带边框 */}
                      <button
                        onClick={addResponsibility}
                        className="h-7 px-2 text-[13px] text-gray-900 hover:bg-gray-100 flex items-center gap-1 border border-gray-300 rounded"
                      >
                        + 新增
                      </button>
                    </div>

                    {/* 职责列表 */}
                    <div className="space-y-3">
                      {responsibilities.map((resp, index) => (
                        <div key={resp.id} className="relative">
                          <textarea
                            placeholder="请描述该岗位的职责与要求"
                            value={resp.content}
                            onChange={(e) =>
                              updateResponsibility(resp.id, e.target.value)
                            }
                            className="w-full min-h-[100px] text-[14px] border border-gray-300 rounded-lg p-3 pr-12 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100 bg-white"
                            style={{ resize: 'vertical' }}
                          />
                          {/* 删除按钮 - 只在第二个及以后的框体显示 */}
                          {index > 0 && (
                            <button
                              onClick={() => removeResponsibility(resp.id)}
                              className="absolute top-3 right-3 p-1.5 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded"
                              title="删除此项"
                            >
                              <Trash2 className="h-4 w-4" />
                            </button>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* AI模式内容区域 */}
            {editMode === 'ai' && (
              <div className="bg-white rounded-lg border border-gray-200 p-6">
                <div className="text-center py-12">
                  <Settings className="h-12 w-12 text-blue-400 mx-auto mb-4" />
                  <p className="text-gray-500 text-[14px]">
                    AI编辑模式正在开发中...
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* 底部按钮 */}
          {editMode === 'manual' && (
            <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-white">
              <Button
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={isSubmitting}
                className="h-10 px-6 text-[14px] border-gray-300 text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                取消
              </Button>
              <Button
                variant="outline"
                onClick={() => handleSubmit('draft')}
                disabled={isSubmitting}
                className="h-10 px-6 text-[14px] border-gray-300 text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isSubmitting ? '保存中...' : '保存草稿'}
              </Button>
              <Button
                onClick={() => handleSubmit('published')}
                disabled={isSubmitting}
                className="h-10 px-6 text-[14px] text-white disabled:opacity-50 disabled:cursor-not-allowed"
                style={{
                  backgroundColor: isSubmitting ? '#a0d1ff' : '#7bb8ff',
                }}
                onMouseEnter={(e) =>
                  !isSubmitting &&
                  (e.currentTarget.style.backgroundColor = '#6aa8ee')
                }
                onMouseLeave={(e) =>
                  !isSubmitting &&
                  (e.currentTarget.style.backgroundColor = '#7bb8ff')
                }
              >
                {isSubmitting ? '发布中...' : '保存发布'}
              </Button>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* 技能选择弹窗 */}
      <Dialog
        open={isSkillSelectModalOpen}
        onOpenChange={setIsSkillSelectModalOpen}
      >
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <div className="flex items-center justify-between">
              <DialogTitle className="text-[18px] font-medium text-gray-900">
                工作技能
              </DialogTitle>
              {/* 新增技能按钮 - 右上角 */}
              <Button
                variant="outline"
                onClick={handleOpenSkillConfig}
                className="h-7 px-2 text-[13px] text-gray-900 border-gray-300"
              >
                + 新增技能
              </Button>
            </div>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* 技能选择 - 多选复选框列表 */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label className="text-[14px] font-medium text-gray-700">
                  选择技能:
                </Label>
                {/* 全选/清空按钮 */}
                {availableSkills.length > 0 && (
                  <div className="flex gap-2">
                    <button
                      onClick={() =>
                        setSelectedSkillIds(availableSkills.map((s) => s.id))
                      }
                      className="text-[13px] hover:underline"
                      style={{ color: '#7bb8ff' }}
                      onMouseEnter={(e) =>
                        (e.currentTarget.style.color = '#6aa8ee')
                      }
                      onMouseLeave={(e) =>
                        (e.currentTarget.style.color = '#7bb8ff')
                      }
                    >
                      全选
                    </button>
                    <span className="text-gray-300">|</span>
                    <button
                      onClick={() => setSelectedSkillIds([])}
                      className="text-[13px] hover:underline"
                      style={{ color: '#7bb8ff' }}
                      onMouseEnter={(e) =>
                        (e.currentTarget.style.color = '#6aa8ee')
                      }
                      onMouseLeave={(e) =>
                        (e.currentTarget.style.color = '#7bb8ff')
                      }
                    >
                      清空
                    </button>
                  </div>
                )}
              </div>
              <div className="border border-gray-300 rounded-lg p-3 max-h-[300px] overflow-y-auto bg-white">
                {isLoadingSkills ? (
                  <div className="text-[14px] text-gray-500 text-center py-4">
                    加载中...
                  </div>
                ) : availableSkills.length === 0 ? (
                  <div className="text-[14px] text-gray-500 text-center py-4">
                    暂无可选技能
                  </div>
                ) : (
                  <div className="space-y-2">
                    {availableSkills.map((skill) => (
                      <label
                        key={skill.id}
                        className="flex items-center gap-2 cursor-pointer hover:bg-gray-50 p-2 rounded"
                      >
                        <input
                          type="checkbox"
                          checked={selectedSkillIds.includes(skill.id)}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setSelectedSkillIds([
                                ...selectedSkillIds,
                                skill.id,
                              ]);
                            } else {
                              setSelectedSkillIds(
                                selectedSkillIds.filter((id) => id !== skill.id)
                              );
                            }
                          }}
                          className="w-4 h-4 border-gray-300 rounded"
                          style={{
                            accentColor: '#7bb8ff',
                          }}
                        />
                        <span className="text-[14px] text-gray-700">
                          {skill.name}
                        </span>
                      </label>
                    ))}
                  </div>
                )}
              </div>
              {selectedSkillIds.length > 0 && (
                <div className="text-[13px] text-gray-500">
                  已选择 {selectedSkillIds.length} 个技能
                </div>
              )}
            </div>

            {/* 底部按钮区域 */}
            <div className="flex items-center justify-end gap-3 pt-4">
              <Button
                variant="outline"
                onClick={() => {
                  setIsSkillSelectModalOpen(false);
                  setSelectedSkillIds([]);
                }}
                className="text-[14px]"
              >
                取消
              </Button>
              <Button
                onClick={handleConfirmSkill}
                disabled={selectedSkillIds.length === 0}
                className="text-[14px] text-white disabled:opacity-50"
                style={{
                  backgroundColor:
                    selectedSkillIds.length === 0 ? '#d1d5db' : '#7bb8ff',
                }}
                onMouseEnter={(e) =>
                  selectedSkillIds.length > 0 &&
                  (e.currentTarget.style.backgroundColor = '#6aa8ee')
                }
                onMouseLeave={(e) =>
                  selectedSkillIds.length > 0 &&
                  (e.currentTarget.style.backgroundColor = '#7bb8ff')
                }
              >
                确定
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </Fragment>
  );
}
