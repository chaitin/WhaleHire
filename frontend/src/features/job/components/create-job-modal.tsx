import { useState, useEffect, useCallback, Fragment, useRef } from 'react';
import {
  X,
  Settings,
  Trash2,
  FileText,
  Sparkles,
  FolderOpen,
  Eye,
  Wand2,
  Send,
  BookmarkPlus,
  LayoutList,
  MapPin,
  Briefcase,
  Building2,
  DollarSign,
  GraduationCap,
  Target,
  Check,
} from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/ui/dialog';
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
  updateJobProfile,
  parseJobProfile,
  listJobSkillMeta,
  polishPrompt,
  generateByPrompt,
} from '@/services/job-profile';
import { listDepartments } from '@/services/department';
import { toast } from '@/ui/toast';
import type { JobProfileDetail } from '@/types/job-profile';

interface CreateJobModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void; // 成功创建后的回调函数
  editingJob?: JobProfileDetail | null; // 编辑的岗位数据
}

export function CreateJobModal({
  open,
  onOpenChange,
  onSuccess,
  editingJob,
}: CreateJobModalProps) {
  // 编辑模式状态(会在useEffect中根据editingJob动态设置)
  const [editMode, setEditMode] = useState<'manual' | 'ai'>('manual');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isParsing, setIsParsing] = useState(false);

  // 追踪是否已识别过内容(用于控制编辑模式下职责区域的可编辑性)
  const [hasParsed, setHasParsed] = useState(false);

  // AI模式相关状态
  const [aiPrompt, setAiPrompt] = useState('');
  const [isPolishing, setIsPolishing] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [generatedProfiles, setGeneratedProfiles] = useState<
    JobProfileDetail[]
  >([]);
  const [selectedProfile, setSelectedProfile] =
    useState<JobProfileDetail | null>(null);
  const [displayMode, setDisplayMode] = useState<'markdown' | 'structured'>(
    'markdown'
  ); // 显示模式
  const [polishTips, setPolishTips] = useState<{
    responsibility_tips?: string[];
    requirement_tips?: string[];
    bonus_tips?: string[];
  }>({});

  // Markdown内容的ref
  const markdownContentRef = useRef<HTMLDivElement>(null);

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

  // 保存编辑时的各项ID
  const [editingIds, setEditingIds] = useState<{
    educationId?: string;
    experienceId?: string;
    industryId?: string;
  }>({});

  // 技能列表 - 存储完整的技能对象
  const [skills, setSkills] = useState<
    Array<{
      id?: string; // 编辑时保留技能id
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

  // AI模式部门选择弹窗状态
  const [isDepartmentSelectModalOpen, setIsDepartmentSelectModalOpen] =
    useState(false);

  // 岗位职责列表
  const [responsibilities, setResponsibilities] = useState<
    Array<{ content: string; id: string; dbId?: string }> // dbId用于保存数据库中的id
  >([{ content: '', id: '1' }]);

  // 重置表单
  const resetForm = useCallback(() => {
    // 重置手动编辑模式的表单数据
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
    setEditingIds({}); // 清空编辑ID
    setHasParsed(false); // 重置识别状态

    // 重置AI编辑模式的状态
    setEditMode('manual'); // 默认选中手动编辑模式
    setAiPrompt(''); // 清空AI提示词
    setIsPolishing(false);
    setIsGenerating(false);
    setGeneratedProfiles([]); // 清空生成的岗位列表
    setSelectedProfile(null); // 清空选中的岗位
    setDisplayMode('markdown'); // 重置显示模式为markdown
    setPolishTips({}); // 清空优化建议
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

  // 监听弹窗打开状态，打开时重置表单或填充编辑数据并加载部门
  useEffect(() => {
    if (open) {
      console.log('=== CreateJobModal打开 ===');
      console.log('editingJob:', editingJob);
      console.log(
        'editingJob.description_markdown:',
        editingJob?.description_markdown
      );

      if (editingJob) {
        // 编辑模式:根据数据判断编辑模式并填充数据
        // 判断依据:
        // 1. 如果有description_markdown字段且不为空,肯定是AI模式
        // 2. 如果有responsibilities数组为空且description中包含markdown格式,认为是AI模式
        // 3. 如果有responsibilities数组不为空,则是手动模式
        const hasDescriptionMarkdown =
          editingJob.description_markdown &&
          editingJob.description_markdown.trim().length > 0;

        // 检查是否有手动填写的结构化数据
        const hasStructuredData =
          (editingJob.responsibilities &&
            editingJob.responsibilities.length > 0) ||
          (editingJob.skills && editingJob.skills.length > 0);

        // 只有当有description_markdown或者(没有结构化数据且description包含特定格式)时才认为是AI模式
        const descriptionContent = editingJob.description || '';
        const hasMarkdownFormat =
          descriptionContent.includes('岗位职责：') ||
          descriptionContent.includes('任职要求：') ||
          descriptionContent.includes('加分项：');

        const isAIMode =
          hasDescriptionMarkdown || (!hasStructuredData && hasMarkdownFormat);
        console.log('>>> hasDescriptionMarkdown:', hasDescriptionMarkdown);
        console.log('>>> hasStructuredData:', hasStructuredData);
        console.log('>>> hasMarkdownFormat:', hasMarkdownFormat);
        console.log('>>> isAIMode:', isAIMode);

        if (isAIMode) {
          // AI模式
          console.log('>>> 设置为AI编辑模式');
          setEditMode('ai');
          setSelectedProfile({
            ...editingJob,
            description:
              editingJob.description_markdown || editingJob.description || '',
            description_markdown:
              editingJob.description_markdown || editingJob.description || '',
          });
          // 如果有生成的数据,添加到概览中
          if (editingJob.name) {
            setGeneratedProfiles([editingJob]);
          }
        } else {
          // 手动模式
          console.log('>>> 设置为手动编辑模式');
          setEditMode('manual');
          // 编辑手动模式时,默认设置为未识别状态,需要重新识别
          setHasParsed(false);
          // 填充手动模式表单数据
          setFormData({
            description: editingJob.description || '',
            jobName: editingJob.name || '',
            educationRequirement:
              editingJob.education_requirements?.[0]?.education_type || '',
            workType: editingJob.work_type || '',
            department: editingJob.department_id || '',
            location: editingJob.location || '',
            salaryMin: editingJob.salary_min?.toString() || '',
            salaryMax: editingJob.salary_max?.toString() || '',
            industryRequirement:
              editingJob.industry_requirements?.[0]?.industry || '',
            companyRequirement:
              editingJob.industry_requirements?.[0]?.company_name || '',
            workExperience:
              editingJob.experience_requirements?.[0]?.experience_type || '',
          });
          // 保存各项的ID
          setEditingIds({
            educationId: editingJob.education_requirements?.[0]?.id,
            experienceId: editingJob.experience_requirements?.[0]?.id,
            industryId: editingJob.industry_requirements?.[0]?.id,
          });
          // 填充职责
          const respList = editingJob.responsibilities?.map((r) => ({
            content: r.responsibility,
            id: r.id, // 保留原始id用于更新
            dbId: r.id, // 单独保存数据库id
          })) || [{ content: '', id: 'resp-0' }];
          setResponsibilities(respList);
          // 填充技能
          const jobSkills =
            editingJob.skills?.map((s) => ({
              id: s.id, // 保留技能id用于更新
              skill_id: s.skill_id || '',
              skill_name: s.skill_name || s.skill || '', // 优先使用skill_name,否则使用skill
              type: (s.type || 'required') as 'required' | 'bonus',
            })) || [];
          setSkills(jobSkills);
        }
      } else {
        // 新建模式:重置表单
        console.log('>>> 新建模式,重置表单');
        resetForm();
      }
      fetchDepartments();
    }
  }, [open, editingJob, resetForm, fetchDepartments]);

  // 更新Markdown内容显示(当selectedProfile变化时强制更新DOM)
  useEffect(() => {
    if (
      markdownContentRef.current &&
      selectedProfile &&
      displayMode === 'markdown'
    ) {
      const content =
        selectedProfile.description_markdown ||
        selectedProfile.description ||
        '暂无岗位描述内容';

      // 将Markdown文本转换为HTML,为特定标题添加样式
      const formattedContent = content
        .replace(
          /岗位职责：/g,
          '<span style="font-weight: bold; font-size: 16px;">岗位职责：</span>'
        )
        .replace(
          /任职要求：/g,
          '<span style="font-weight: bold; font-size: 16px;">任职要求：</span>'
        )
        .replace(
          /加分项：/g,
          '<span style="font-weight: bold; font-size: 16px;">加分项：</span>'
        )
        .replace(/\n/g, '<br/>');

      // 使用innerHTML来渲染带样式的HTML
      markdownContentRef.current.innerHTML = formattedContent;
    }
  }, [selectedProfile, displayMode]);

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
        // 解析成功后,标记为已识别,允许编辑职责区域
        setHasParsed(true);
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

  // 处理优化提示词
  const handlePolish = async () => {
    if (!aiPrompt.trim()) {
      toast.error('请先输入岗位要求');
      return;
    }

    setIsPolishing(true);
    try {
      const result = await polishPrompt(aiPrompt);
      if (result && result.polished_prompt) {
        setAiPrompt(result.polished_prompt);
        // 保存优化建议
        setPolishTips({
          responsibility_tips: result.responsibility_tips,
          requirement_tips: result.requirement_tips,
          bonus_tips: result.bonus_tips,
        });
        toast.success('提示词优化成功');
      }
    } catch (error) {
      toast.error(
        '优化失败: ' + (error instanceof Error ? error.message : '请稍后重试')
      );
    } finally {
      setIsPolishing(false);
    }
  };

  // 处理生成岗位画像
  const handleGenerate = async () => {
    if (!aiPrompt.trim()) {
      toast.error('请先输入岗位要求');
      return;
    }

    setIsGenerating(true);
    try {
      const result = await generateByPrompt(aiPrompt);

      // 调试日志
      console.log('=== 生成岗位画像 ===');
      console.log('API返回结果:', result);

      if (result) {
        // 生成一个临时ID(如果后端没有返回ID)
        const profileWithId = {
          ...result,
          id: result.id || `temp-${Date.now()}`,
        };

        console.log('处理后的profile:', profileWithId);
        console.log(
          'description_markdown:',
          profileWithId.description_markdown
        );
        console.log('description:', profileWithId.description);

        // 重置显示模式为markdown
        setDisplayMode('markdown');
        // 直接设置为选中状态(不自动添加到概览,等待用户点击暂存)
        setSelectedProfile(profileWithId);
        toast.success('岗位画像生成成功');
      }
    } catch (error) {
      toast.error(
        '生成失败: ' + (error instanceof Error ? error.message : '请稍后重试')
      );
    } finally {
      setIsGenerating(false);
    }
  };

  // 处理表单提交 - 调用创建接口
  const handleSubmit = async (status: 'draft' | 'published') => {
    // AI模式和手动模式的不同验证和数据处理
    if (editMode === 'ai') {
      // AI模式验证
      if (!selectedProfile) {
        toast.error('请先生成岗位画像');
        return;
      }
      if (!selectedProfile.name?.trim()) {
        toast.error('请输入岗位名称');
        return;
      }
      if (!selectedProfile.department_id) {
        toast.error('请选择岗位所属部门');
        return;
      }
    } else {
      // 手动模式验证
      if (!formData.jobName.trim()) {
        toast.error('请输入岗位名称');
        return;
      }
      if (!formData.department) {
        toast.error('请选择所属部门');
        return;
      }
    }

    setIsSubmitting(true);
    try {
      // 根据编辑模式构建请求数据
      let requestData: {
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
      };

      if (editMode === 'ai' && selectedProfile) {
        // AI模式：使用selectedProfile的数据
        requestData = {
          name: selectedProfile.name || '',
          department_id: selectedProfile.department_id || '',
          description: selectedProfile.description_markdown || '', // AI模式使用description_markdown
          location: selectedProfile.location || '',
          work_type: selectedProfile.work_type || '',
          status: status,
        };

        // 添加薪资范围
        if (selectedProfile.salary_min) {
          requestData.salary_min = selectedProfile.salary_min;
        }
        if (selectedProfile.salary_max) {
          requestData.salary_max = selectedProfile.salary_max;
        }

        // 添加学历要求
        if (
          selectedProfile.education_requirements &&
          selectedProfile.education_requirements.length > 0
        ) {
          requestData.education_requirements =
            selectedProfile.education_requirements.map((edu) => ({
              ...(edu.id && { id: edu.id }), // 编辑时保留id
              education_type: edu.education_type,
            }));
        }

        // 添加工作经验要求
        if (
          selectedProfile.experience_requirements &&
          selectedProfile.experience_requirements.length > 0
        ) {
          requestData.experience_requirements =
            selectedProfile.experience_requirements.map((exp) => ({
              ...(exp.id && { id: exp.id }), // 编辑时保留id
              experience_type: exp.experience_type,
            }));
        }

        // 添加行业要求
        if (
          selectedProfile.industry_requirements &&
          selectedProfile.industry_requirements.length > 0
        ) {
          requestData.industry_requirements =
            selectedProfile.industry_requirements.map((req) => ({
              ...(req.id && { id: req.id }), // 编辑时保留id
              industry: req.industry || '',
              company_name: req.company_name || '',
            }));
        }

        // 添加技能要求
        if (selectedProfile.skills && selectedProfile.skills.length > 0) {
          requestData.skills = selectedProfile.skills.map((skill) => ({
            ...(skill.id && { id: skill.id }), // 编辑时保留id
            skill_id: skill.skill_id || '',
            skill_name:
              'skill_name' in skill && skill.skill_name
                ? skill.skill_name
                : skill.skill || '',
            type: skill.type as 'required' | 'bonus',
          }));
        }

        // 添加岗位职责
        if (
          selectedProfile.responsibilities &&
          selectedProfile.responsibilities.length > 0
        ) {
          requestData.responsibilities = selectedProfile.responsibilities.map(
            (resp) => ({
              ...(typeof resp !== 'string' && resp.id && { id: resp.id }), // 编辑时保留id
              responsibility:
                typeof resp === 'string' ? resp : resp.responsibility,
            })
          );
        }
      } else {
        // 手动模式：使用formData的数据
        requestData = {
          name: formData.jobName,
          department_id: formData.department,
          description: formData.description, // 手动模式使用description
          location: formData.location,
          work_type: formData.workType,
          status: status,
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
              ...(editingIds.educationId && { id: editingIds.educationId }), // 编辑时保留id
              education_type: formData.educationRequirement,
            },
          ];
        }

        // 添加工作经验要求
        if (formData.workExperience) {
          requestData.experience_requirements = [
            {
              ...(editingIds.experienceId && { id: editingIds.experienceId }), // 编辑时保留id
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
              ...(editingIds.industryId && { id: editingIds.industryId }), // 编辑时保留id
              industry: formData.industryRequirement,
              company_name: formData.companyRequirement,
            });
          } else if (formData.industryRequirement) {
            requestData.industry_requirements.push({
              ...(editingIds.industryId && { id: editingIds.industryId }), // 编辑时保留id
              industry: formData.industryRequirement,
              company_name: '', // 提供默认值
            });
          } else if (formData.companyRequirement) {
            requestData.industry_requirements.push({
              ...(editingIds.industryId && { id: editingIds.industryId }), // 编辑时保留id
              industry: '', // 提供默认值
              company_name: formData.companyRequirement,
            });
          }
        }

        // 添加技能要求（按照API格式传递完整的技能对象）
        if (skills.length > 0) {
          requestData.skills = skills.map((skill) => ({
            ...(skill.id && { id: skill.id }), // 编辑时保留id
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
            ...(r.dbId && { id: r.dbId }), // 编辑时保留id
            responsibility: r.content,
          }));
        }
      }

      // 判断是创建还是更新
      if (editingJob) {
        // 更新操作
        await updateJobProfile(editingJob.id, {
          id: editingJob.id,
          ...requestData,
        });
        toast.success('岗位画像已更新');
      } else {
        // 创建操作
        await createJobProfile(requestData);
        // 根据状态显示不同的成功消息
        if (status === 'draft') {
          toast.success('岗位画像已保存为草稿');
        } else {
          toast.success('岗位画像已发布');
        }
      }

      // 调用成功回调刷新列表
      if (onSuccess) {
        onSuccess();
      }

      // 关闭弹窗并重置表单
      onOpenChange(false);
      resetForm();
    } catch (error) {
      console.error(editingJob ? '更新失败:' : '创建失败:', error);
      const errorMessage =
        error instanceof Error ? error.message : '请稍后重试';
      toast.error(`${editingJob ? '更新' : '创建'}失败: ${errorMessage}`);
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
                {editingJob ? '编辑岗位' : '新建岗位'}
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
                              <SelectItem value="no-department" disabled>
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
                        disabled={!!editingJob && !hasParsed}
                        className={cn(
                          'h-7 px-2 text-[13px] flex items-center gap-1 border rounded',
                          !!editingJob && !hasParsed
                            ? 'text-gray-400 bg-gray-100 border-gray-200 cursor-not-allowed'
                            : 'text-gray-900 border-gray-300 hover:bg-gray-100'
                        )}
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
                            disabled={!!editingJob && !hasParsed}
                            className={cn(
                              'w-full min-h-[100px] text-[14px] border rounded-lg p-3 pr-12 focus:outline-none',
                              !!editingJob && !hasParsed
                                ? 'bg-gray-50 border-gray-200 text-gray-400 cursor-not-allowed'
                                : 'bg-white border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-100'
                            )}
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
              <div className="space-y-6">
                {/* 小鲸互动区域 */}
                <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                  <div className="flex items-center gap-2 mb-4">
                    <Sparkles className="h-5 w-5 text-gray-700" />
                    <h3 className="text-[16px] font-medium text-gray-900">
                      小鲸互动
                    </h3>
                  </div>
                  <div className="relative">
                    <textarea
                      value={aiPrompt}
                      onChange={(e) => setAiPrompt(e.target.value)}
                      placeholder="请输入您想创建的岗位要求,例如:'我需要一个高级前端开发工程师,5年以上经验,精通React和Vue...'"
                      className="w-full min-h-[120px] text-[14px] border border-gray-300 rounded-lg p-4 pb-14 focus:outline-none focus:border-[#7bb8ff] focus:ring-2 focus:ring-[#7bb8ff]/20 bg-white resize-vertical"
                    />
                    {/* 右下角按钮组 */}
                    <div className="absolute bottom-3 right-3 flex items-center gap-1.5">
                      {/* 优化按钮 - 魔法概念,只显示icon */}
                      <button
                        onClick={handlePolish}
                        disabled={isPolishing || !aiPrompt.trim()}
                        className="flex items-center justify-center w-6 h-6 rounded-md border border-gray-300 bg-white hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        title={isPolishing ? '优化中...' : '优化提示词'}
                      >
                        <Wand2 className="w-3 h-3 text-purple-600" />
                      </button>
                      {/* 生成按钮 - 渐变色,只显示icon */}
                      <button
                        onClick={handleGenerate}
                        disabled={isGenerating || !aiPrompt.trim()}
                        className="flex items-center justify-center w-7 h-7 rounded-md text-white transition-all hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed"
                        style={{
                          background:
                            isGenerating || !aiPrompt.trim()
                              ? '#d1d5db'
                              : 'linear-gradient(135deg, #7bb8ff 0%, #a78bfa 100%)',
                        }}
                        onMouseEnter={(e) => {
                          if (!isGenerating && aiPrompt.trim()) {
                            e.currentTarget.style.background =
                              'linear-gradient(135deg, #6aa8ee 0%, #9575e8 100%)';
                          }
                        }}
                        onMouseLeave={(e) => {
                          if (!isGenerating && aiPrompt.trim()) {
                            e.currentTarget.style.background =
                              'linear-gradient(135deg, #7bb8ff 0%, #a78bfa 100%)';
                          }
                        }}
                        title={isGenerating ? '生成中...' : '生成岗位'}
                      >
                        <Send className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>

                  {/* 优化建议区域 */}
                  {polishTips.responsibility_tips?.length ||
                  polishTips.requirement_tips?.length ||
                  polishTips.bonus_tips?.length ? (
                    <div className="mt-4 space-y-3">
                      {polishTips.responsibility_tips &&
                        polishTips.responsibility_tips.length > 0 && (
                          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                            <h4 className="text-[13px] font-medium text-blue-900 mb-2">
                              岗位职责建议:
                            </h4>
                            <ul className="space-y-1">
                              {polishTips.responsibility_tips.map(
                                (tip, idx) => (
                                  <li
                                    key={idx}
                                    className="text-[12px] text-blue-700 pl-4 relative before:content-['•'] before:absolute before:left-0"
                                  >
                                    {tip}
                                  </li>
                                )
                              )}
                            </ul>
                          </div>
                        )}
                      {polishTips.requirement_tips &&
                        polishTips.requirement_tips.length > 0 && (
                          <div className="bg-purple-50 border border-purple-200 rounded-lg p-3">
                            <h4 className="text-[13px] font-medium text-purple-900 mb-2">
                              任职要求建议:
                            </h4>
                            <ul className="space-y-1">
                              {polishTips.requirement_tips.map((tip, idx) => (
                                <li
                                  key={idx}
                                  className="text-[12px] text-purple-700 pl-4 relative before:content-['•'] before:absolute before:left-0"
                                >
                                  {tip}
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}
                      {polishTips.bonus_tips &&
                        polishTips.bonus_tips.length > 0 && (
                          <div className="bg-green-50 border border-green-200 rounded-lg p-3">
                            <h4 className="text-[13px] font-medium text-green-900 mb-2">
                              加分项建议:
                            </h4>
                            <ul className="space-y-1">
                              {polishTips.bonus_tips.map((tip, idx) => (
                                <li
                                  key={idx}
                                  className="text-[12px] text-green-700 pl-4 relative before:content-['•'] before:absolute before:left-0"
                                >
                                  {tip}
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}
                    </div>
                  ) : null}
                </div>

                {/* 下方两栏布局 - 缩小左侧宽度 */}
                <div className="grid grid-cols-[320px_1fr] gap-6">
                  {/* 左侧：生成岗位概览 */}
                  <div className="bg-white rounded-lg shadow-sm border border-gray-200">
                    {/* 标题栏 */}
                    <div className="px-5 py-5 border-b border-gray-200 flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <FolderOpen className="h-5 w-5 text-gray-700" />
                        <h3 className="text-[16px] font-medium text-gray-900">
                          生成岗位概览
                        </h3>
                      </div>
                      <button className="text-gray-400 hover:text-gray-600">
                        <Settings className="w-4 h-4" />
                      </button>
                    </div>

                    {/* 已生成的岗位列表 */}
                    <div className="p-5 space-y-4">
                      {generatedProfiles.length === 0 ? (
                        <div className="text-center py-12 text-gray-400">
                          <p className="text-[14px]">暂无生成记录</p>
                          <p className="text-[12px] mt-1">
                            点击"生成"按钮创建岗位画像
                          </p>
                        </div>
                      ) : (
                        generatedProfiles.map((profile, index) => {
                          // 获取岗位职责的第一条,并截取前15个字
                          const firstResponsibility =
                            profile.responsibilities &&
                            profile.responsibilities.length > 0
                              ? typeof profile.responsibilities[0] === 'string'
                                ? profile.responsibilities[0]
                                : profile.responsibilities[0].responsibility
                              : '';
                          const displayText =
                            firstResponsibility.length > 15
                              ? firstResponsibility.substring(0, 15) + '...'
                              : firstResponsibility || '暂无描述';

                          return (
                            <div
                              key={profile.id || index}
                              onClick={() => {
                                setDisplayMode('markdown'); // 切换岗位时重置为markdown模式
                                setSelectedProfile(profile);
                              }}
                              className={cn(
                                'border rounded-lg p-4 cursor-pointer transition-all',
                                selectedProfile?.id === profile.id
                                  ? 'bg-blue-50'
                                  : 'border-gray-200 hover:border-gray-300'
                              )}
                              style={
                                selectedProfile?.id === profile.id
                                  ? { borderColor: '#7bb8ff' }
                                  : undefined
                              }
                            >
                              <div className="flex items-center justify-between mb-2">
                                <div className="flex items-center gap-2 flex-1 min-w-0">
                                  <div
                                    className={cn(
                                      'w-4 h-4 rounded border flex-shrink-0 flex items-center justify-center',
                                      selectedProfile?.id === profile.id
                                        ? ''
                                        : ''
                                    )}
                                    style={
                                      selectedProfile?.id === profile.id
                                        ? {
                                            background:
                                              'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                                            borderColor: '#7bb8ff',
                                          }
                                        : { borderColor: '#7bb8ff' }
                                    }
                                  >
                                    {selectedProfile?.id === profile.id && (
                                      <Check className="w-3 h-3 text-white" />
                                    )}
                                  </div>
                                  <h4
                                    className="text-[14px] font-medium truncate"
                                    style={{ color: '#7bb8ff' }}
                                  >
                                    {profile.name || '未命名岗位'}
                                  </h4>
                                </div>
                                <span
                                  className="px-2 py-0.5 text-[12px] rounded-full flex-shrink-0 ml-2"
                                  style={{
                                    backgroundColor: '#e6f4ff',
                                    color: '#7bb8ff',
                                  }}
                                >
                                  待保存
                                </span>
                              </div>
                              <p className="text-[13px] text-gray-600 truncate">
                                {displayText}
                              </p>
                            </div>
                          );
                        })
                      )}
                    </div>
                  </div>

                  {/* 右侧：生成岗位 */}
                  <div className="bg-white rounded-lg shadow-sm border border-gray-200">
                    {/* 标题栏 */}
                    <div className="px-6 py-5 border-b border-gray-200 flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Eye className="h-5 w-5 text-gray-700" />
                        <h3 className="text-[16px] font-medium text-gray-900">
                          生成岗位
                        </h3>
                      </div>
                      <div className="flex gap-2">
                        {/* 暂存按钮 - 使用书签加号图标 */}
                        <button
                          onClick={() => {
                            if (selectedProfile) {
                              const existingIndex = generatedProfiles.findIndex(
                                (p) => p.id === selectedProfile.id
                              );

                              if (existingIndex !== -1) {
                                // 如果已存在,更新该岗位
                                const updatedProfiles = [...generatedProfiles];
                                updatedProfiles[existingIndex] =
                                  selectedProfile;
                                setGeneratedProfiles(updatedProfiles);
                                toast.success('已更新暂存内容');
                              } else {
                                // 如果不存在,添加到列表开头
                                setGeneratedProfiles([
                                  selectedProfile,
                                  ...generatedProfiles,
                                ]);
                                toast.success('已暂存到生成岗位概览');
                              }
                            }
                          }}
                          disabled={!selectedProfile}
                          className="p-2 hover:bg-blue-50 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                          title="暂存到概览"
                        >
                          <BookmarkPlus className="w-4 h-4 text-[#7bb8ff] hover:text-[#6aa8ee]" />
                        </button>
                        {/* 切换显示模式按钮 - 使用结构化列表图标 */}
                        <button
                          onClick={() =>
                            setDisplayMode(
                              displayMode === 'markdown'
                                ? 'structured'
                                : 'markdown'
                            )
                          }
                          disabled={!selectedProfile}
                          className={cn(
                            'p-2 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-colors',
                            displayMode === 'structured'
                              ? 'bg-blue-50'
                              : 'hover:bg-blue-50'
                          )}
                          title={
                            displayMode === 'markdown'
                              ? '结构化视图'
                              : 'Markdown视图'
                          }
                        >
                          <LayoutList
                            className={cn(
                              'w-4 h-4 transition-colors',
                              displayMode === 'structured'
                                ? 'text-[#7bb8ff]'
                                : 'text-gray-700'
                            )}
                          />
                        </button>
                      </div>
                    </div>

                    {/* 预览内容 - 平铺展示,可直接编辑 */}
                    <div className="p-6">
                      {!selectedProfile ? (
                        <div className="text-center py-20 text-gray-400">
                          <Eye className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                          <p className="text-[14px]">请先生成岗位画像</p>
                          <p className="text-[12px] mt-1">
                            生成后将在此处显示预览
                          </p>
                        </div>
                      ) : displayMode === 'markdown' ? (
                        /* Markdown 模式 */
                        <div className="space-y-5" key={selectedProfile.id}>
                          {/* 标题和部门 */}
                          <div className="pb-4 border-b border-gray-200">
                            <div className="flex items-center justify-between">
                              <h2
                                key={`title-${selectedProfile.id}`}
                                contentEditable
                                suppressContentEditableWarning
                                onBlur={(e) =>
                                  setSelectedProfile({
                                    ...selectedProfile,
                                    name: e.currentTarget.textContent || '',
                                  })
                                }
                                className="text-[18px] font-semibold text-gray-900 outline-none focus:bg-blue-50 px-2 py-1 rounded"
                              >
                                {selectedProfile.name || '岗位名称'}
                              </h2>

                              {/* 部门显示/选择 - 右对齐 */}
                              {selectedProfile.department ? (
                                // 已选择部门,显示部门名称+icon
                                <div
                                  className="flex items-center gap-1.5 px-3 py-1.5 bg-blue-50 text-blue-700 rounded-lg text-[13px] cursor-pointer hover:bg-blue-100 transition-colors"
                                  onClick={() =>
                                    setIsDepartmentSelectModalOpen(true)
                                  }
                                  title="点击修改所属部门"
                                >
                                  <Building2 className="w-3.5 h-3.5" />
                                  <span>{selectedProfile.department}</span>
                                </div>
                              ) : (
                                // 未选择部门,显示图标按钮
                                <button
                                  onClick={() =>
                                    setIsDepartmentSelectModalOpen(true)
                                  }
                                  className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                                  title="选择所属部门"
                                >
                                  <Building2 className="w-5 h-5 text-gray-400 hover:text-gray-600" />
                                </button>
                              )}
                            </div>
                          </div>

                          {/* Markdown内容 */}
                          <div
                            ref={markdownContentRef}
                            contentEditable
                            suppressContentEditableWarning
                            onBlur={(e) => {
                              // 从innerHTML获取HTML,然后转换为纯文本,保留换行
                              const html = e.currentTarget.innerHTML;

                              // 先将块级元素和br标签转换为换行符
                              let text = html
                                // 处理各种换行标签
                                .replace(/<br\s*\/?>/gi, '\n')
                                .replace(/<div>/gi, '\n')
                                .replace(/<\/div>/gi, '')
                                .replace(/<p>/gi, '\n')
                                .replace(/<\/p>/gi, '')
                                // 处理span标签(保留内容,移除标签)
                                .replace(/<span[^>]*>/gi, '')
                                .replace(/<\/span>/gi, '');

                              // 移除其他所有HTML标签
                              const tempDiv = document.createElement('div');
                              tempDiv.innerHTML = text;
                              text =
                                tempDiv.textContent || tempDiv.innerText || '';

                              // 清理首尾的换行,但保留内部的所有换行
                              text = text.trim();
                              // 将连续的3个以上换行替换为2个(保留段落间距)
                              text = text.replace(/\n{3,}/g, '\n\n');

                              setSelectedProfile({
                                ...selectedProfile,
                                description_markdown: text,
                              });
                            }}
                            className="text-[14px] text-gray-700 leading-relaxed whitespace-pre-wrap outline-none focus:bg-blue-50 p-2 rounded min-h-[200px]"
                          >
                            {selectedProfile.description_markdown ||
                              selectedProfile.description ||
                              '暂无岗位描述内容'}
                          </div>
                        </div>
                      ) : (
                        /* 结构化模式 */
                        <div
                          className="space-y-6"
                          key={`structured-${selectedProfile.id}`}
                        >
                          {/* 岗位标题 */}
                          <div className="pb-4 border-b border-gray-200">
                            <h2
                              key={`structured-title-${selectedProfile.id}`}
                              contentEditable
                              suppressContentEditableWarning
                              onBlur={(e) =>
                                setSelectedProfile({
                                  ...selectedProfile,
                                  name: e.currentTarget.textContent || '',
                                })
                              }
                              className="text-[20px] font-bold text-gray-900 outline-none focus:bg-blue-50 px-2 py-1 rounded"
                            >
                              {selectedProfile.name || '岗位名称'}
                            </h2>
                          </div>

                          {/* 基本信息卡片 */}
                          <div
                            className="rounded-xl p-5 border"
                            style={{
                              backgroundColor: '#e6f4ff',
                              borderColor: '#7bb8ff',
                            }}
                          >
                            <div className="flex flex-wrap gap-x-6 gap-y-3 text-[14px]">
                              {/* 工作地点 */}
                              <div className="flex items-center gap-2 min-w-[140px]">
                                <MapPin
                                  className="w-4 h-4 flex-shrink-0"
                                  style={{ color: '#7bb8ff' }}
                                />
                                <span className="text-gray-700 font-medium">
                                  {selectedProfile.location || '暂无要求'}
                                </span>
                              </div>

                              {/* 工作经验 */}
                              <div className="flex items-center gap-2 min-w-[140px]">
                                <Briefcase
                                  className="w-4 h-4 flex-shrink-0"
                                  style={{ color: '#7bb8ff' }}
                                />
                                <span className="text-gray-700 font-medium">
                                  {selectedProfile.experience_requirements &&
                                  selectedProfile.experience_requirements
                                    .length > 0
                                    ? selectedProfile.experience_requirements[0]
                                        .experience_type === 'unlimited'
                                      ? '经验不限'
                                      : selectedProfile
                                            .experience_requirements[0]
                                            .experience_type ===
                                          'fresh_graduate'
                                        ? '应届毕业生'
                                        : selectedProfile
                                              .experience_requirements[0]
                                              .experience_type ===
                                            'under_one_year'
                                          ? '1年以下'
                                          : selectedProfile
                                                .experience_requirements[0]
                                                .experience_type ===
                                              'one_to_three_years'
                                            ? '1-3年经验'
                                            : selectedProfile
                                                  .experience_requirements[0]
                                                  .experience_type ===
                                                'three_to_five_years'
                                              ? '3-5年经验'
                                              : selectedProfile
                                                    .experience_requirements[0]
                                                    .experience_type ===
                                                  'five_to_ten_years'
                                                ? '5-10年经验'
                                                : selectedProfile
                                                      .experience_requirements[0]
                                                      .experience_type ===
                                                    'over_ten_years'
                                                  ? '10年以上'
                                                  : '暂无要求'
                                    : '暂无要求'}
                                </span>
                              </div>

                              {/* 工作性质 */}
                              <div className="flex items-center gap-2 min-w-[140px]">
                                <Briefcase
                                  className="w-4 h-4 flex-shrink-0"
                                  style={{ color: '#7bb8ff' }}
                                />
                                <span className="text-gray-700 font-medium">
                                  {selectedProfile.work_type
                                    ? selectedProfile.work_type === 'full_time'
                                      ? '全职'
                                      : selectedProfile.work_type ===
                                          'part_time'
                                        ? '兼职'
                                        : selectedProfile.work_type ===
                                            'internship'
                                          ? '实习'
                                          : selectedProfile.work_type ===
                                              'outsourcing'
                                            ? '外包'
                                            : '暂无要求'
                                    : '暂无要求'}
                                </span>
                              </div>

                              {/* 教育经历要求 */}
                              <div className="flex items-center gap-2 min-w-[140px]">
                                <GraduationCap
                                  className="w-4 h-4 flex-shrink-0"
                                  style={{ color: '#7bb8ff' }}
                                />
                                <span className="text-gray-700 font-medium">
                                  {selectedProfile.education_requirements &&
                                  selectedProfile.education_requirements
                                    .length > 0
                                    ? selectedProfile.education_requirements
                                        .map((edu) =>
                                          edu.education_type === 'unlimited'
                                            ? '学历不限'
                                            : edu.education_type ===
                                                'junior_college'
                                              ? '大专'
                                              : edu.education_type ===
                                                  'bachelor'
                                                ? '本科'
                                                : edu.education_type ===
                                                    'master'
                                                  ? '硕士'
                                                  : edu.education_type ===
                                                      'doctor'
                                                    ? '博士'
                                                    : ''
                                        )
                                        .filter(Boolean)
                                        .join('、') || '暂无要求'
                                    : '暂无要求'}
                                </span>
                              </div>

                              {/* 行业背景要求 */}
                              <div className="flex items-center gap-2 min-w-[140px]">
                                <Building2
                                  className="w-4 h-4 flex-shrink-0"
                                  style={{ color: '#7bb8ff' }}
                                />
                                <span className="text-gray-700 font-medium">
                                  {selectedProfile.industry_requirements &&
                                  selectedProfile.industry_requirements.length >
                                    0
                                    ? selectedProfile.industry_requirements
                                        .map((req) => {
                                          const parts = [];
                                          if (req.industry)
                                            parts.push(req.industry);
                                          if (req.company_name)
                                            parts.push(req.company_name);
                                          return parts.join('·');
                                        })
                                        .filter(Boolean)
                                        .join('、') || '暂无要求'
                                    : '暂无要求'}
                                </span>
                              </div>

                              {/* 薪资范围 */}
                              {(selectedProfile.salary_min ||
                                selectedProfile.salary_max) && (
                                <div className="flex items-center gap-2 min-w-[140px]">
                                  <DollarSign
                                    className="w-4 h-4 flex-shrink-0"
                                    style={{ color: '#7bb8ff' }}
                                  />
                                  <span className="text-gray-700 font-medium">
                                    {selectedProfile.salary_min &&
                                    selectedProfile.salary_max
                                      ? `${selectedProfile.salary_min}-${selectedProfile.salary_max}K/月`
                                      : selectedProfile.salary_min
                                        ? `${selectedProfile.salary_min}K以上/月`
                                        : `${selectedProfile.salary_max}K以下/月`}
                                  </span>
                                </div>
                              )}
                            </div>
                          </div>

                          {/* 职责要求 */}
                          {selectedProfile.responsibilities &&
                            selectedProfile.responsibilities.length > 0 && (
                              <div className="bg-white rounded-lg p-5 border border-gray-200">
                                <div className="flex items-center gap-2 mb-4">
                                  <div
                                    className="p-2 rounded-lg"
                                    style={{ backgroundColor: '#e6f4ff' }}
                                  >
                                    <Target
                                      className="w-5 h-5"
                                      style={{ color: '#7bb8ff' }}
                                    />
                                  </div>
                                  <h3 className="text-[16px] font-semibold text-gray-900">
                                    职责要求
                                  </h3>
                                </div>
                                <ul className="space-y-2.5">
                                  {selectedProfile.responsibilities.map(
                                    (resp, idx) => (
                                      <li
                                        key={idx}
                                        className="flex items-start gap-2 text-[14px] text-gray-700"
                                      >
                                        <span
                                          className="inline-block w-1.5 h-1.5 rounded-full mt-2 flex-shrink-0"
                                          style={{
                                            background:
                                              'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                                          }}
                                        ></span>
                                        <span>
                                          {typeof resp === 'string'
                                            ? resp
                                            : resp.responsibility}
                                        </span>
                                      </li>
                                    )
                                  )}
                                </ul>
                              </div>
                            )}

                          {/* 技能要求 */}
                          {selectedProfile.skills &&
                            selectedProfile.skills.length > 0 && (
                              <div className="bg-white rounded-lg p-5 border border-gray-200">
                                <div className="flex items-center gap-2 mb-4">
                                  <div className="p-2 bg-green-100 rounded-lg">
                                    <Target className="w-5 h-5 text-green-600" />
                                  </div>
                                  <h3 className="text-[16px] font-semibold text-gray-900">
                                    技能要求
                                  </h3>
                                </div>
                                <div className="flex flex-wrap gap-2.5">
                                  {selectedProfile.skills.map((skill, idx) => {
                                    // 优先使用skill_name(AI生成时的字段),否则使用skill
                                    let skillName =
                                      'skill_name' in skill && skill.skill_name
                                        ? skill.skill_name
                                        : skill.skill;

                                    // 移除技能名称中的 * 号
                                    if (skillName) {
                                      skillName = skillName
                                        .replace(/\*/g, '')
                                        .trim();
                                    }

                                    // 根据技能类型使用不同的背景色
                                    const isRequired =
                                      skill.type === 'required';

                                    return (
                                      <span
                                        key={idx}
                                        className={`inline-flex items-center px-3 py-1.5 rounded-lg text-[14px] border transition-colors ${
                                          isRequired
                                            ? 'text-gray-800'
                                            : 'bg-gray-100 text-gray-700 border-gray-300'
                                        }`}
                                        style={
                                          isRequired
                                            ? {
                                                backgroundColor: '#e6f4ff',
                                                borderColor: '#7bb8ff',
                                              }
                                            : undefined
                                        }
                                      >
                                        {skillName}
                                      </span>
                                    );
                                  })}
                                </div>
                                <div className="mt-3 text-[12px] text-gray-500 flex items-center gap-4">
                                  <div className="flex items-center gap-1.5">
                                    <span
                                      className="w-3 h-3 rounded border"
                                      style={{
                                        backgroundColor: '#e6f4ff',
                                        borderColor: '#7bb8ff',
                                      }}
                                    ></span>
                                    <span>必须技能</span>
                                  </div>
                                  <div className="flex items-center gap-1.5">
                                    <span className="w-3 h-3 rounded bg-gray-100 border border-gray-300"></span>
                                    <span>加分技能</span>
                                  </div>
                                </div>
                              </div>
                            )}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* 底部按钮 */}
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
                background: isSubmitting
                  ? '#a0d1ff'
                  : 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
              }}
            >
              {isSubmitting ? '发布中...' : '保存发布'}
            </Button>
          </div>
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
            <DialogDescription className="sr-only">
              选择该岗位需要的工作技能
            </DialogDescription>
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
                  background:
                    selectedSkillIds.length === 0
                      ? '#d1d5db'
                      : 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                }}
                onMouseEnter={(e) =>
                  selectedSkillIds.length > 0 &&
                  (e.currentTarget.style.background = '#6aa8ee')
                }
                onMouseLeave={(e) =>
                  selectedSkillIds.length > 0 &&
                  (e.currentTarget.style.background =
                    'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)')
                }
              >
                确定
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* AI模式 - 部门选择弹窗 */}
      <Dialog
        open={isDepartmentSelectModalOpen}
        onOpenChange={setIsDepartmentSelectModalOpen}
      >
        <DialogContent
          className="max-w-md"
          onInteractOutside={(e) => e.preventDefault()}
        >
          <DialogHeader>
            <DialogTitle className="text-[16px] font-medium text-gray-900">
              选择所属部门
            </DialogTitle>
            <DialogDescription className="sr-only">
              为该岗位选择所属部门
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* 部门选择下拉框 */}
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
                value={selectedProfile?.department_id || ''}
                onValueChange={(value) => {
                  if (selectedProfile) {
                    // 找到选中的部门对象
                    const selectedDept = departments.find(
                      (dept) => dept.id === value
                    );
                    setSelectedProfile({
                      ...selectedProfile,
                      department_id: value,
                      department: selectedDept?.name || '',
                    });
                  }
                }}
                disabled={isLoadingDepartments}
              >
                <SelectTrigger className="text-[14px] border-gray-300 focus:border-blue-500 rounded-lg bg-white">
                  <SelectValue
                    placeholder={isLoadingDepartments ? '加载中...' : '请选择'}
                  />
                </SelectTrigger>
                <SelectContent
                  position="popper"
                  side="bottom"
                  align="start"
                  className="max-h-[300px]"
                >
                  {departments.length === 0 ? (
                    <SelectItem value="no-department" disabled>
                      暂无部门
                    </SelectItem>
                  ) : (
                    departments.map((dept) => (
                      <SelectItem key={dept.id} value={dept.id}>
                        {dept.name}
                      </SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
            </div>

            {/* 底部按钮 */}
            <div className="flex items-center justify-end gap-3 pt-2">
              <Button
                variant="outline"
                onClick={() => setIsDepartmentSelectModalOpen(false)}
                className="text-[14px]"
              >
                取消
              </Button>
              <Button
                onClick={() => {
                  setIsDepartmentSelectModalOpen(false);
                  if (selectedProfile?.department) {
                    toast.success('部门选择成功');
                  }
                }}
                className="text-[14px] text-white"
                style={{
                  backgroundColor: '#2563EB',
                }}
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
