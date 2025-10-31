import { useState, useEffect, useCallback } from 'react';
import {
  Plus,
  Search,
  Edit,
  Trash2,
  BarChart3,
  Users,
  FileText,
  ChevronLeft,
  ChevronRight,
  X,
  RefreshCw,
  Eye,
  Copy,
} from 'lucide-react';
import { Button } from '@/ui/button';
import { Input } from '@/ui/input';
import { cn } from '@/lib/utils';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/ui/select';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/ui/dialog';

import { Label } from '@/ui/label';
import { Textarea } from '@/ui/textarea';
import { MultiSelect } from '@/ui/multi-select';
import { toast } from '@/ui/toast';
import {
  listJobProfiles,
  createJobProfile,
  updateJobProfile,
  deleteJobProfile,
  duplicateJobProfile,
  listJobSkillMeta,
  getJobProfile,
} from '@/services/job-profile';
import { listDepartments } from '@/services/department';
import type {
  JobProfile,
  JobProfileDetail,
  CreateJobProfileReq,
  UpdateJobProfileReq,
  JobSkillMeta,
  JobSkillInput,
  JobResponsibilityInput,
  JobIndustryRequirementInput,
  JobEducationRequirementInput,
  JobExperienceRequirementInput,
} from '@/types/job-profile';
import type { Department } from '@/services/department';
import { AddSkillModal } from '@/features/resume/components/add-skill-modal';
import { ConfirmDialog } from '@/ui/confirm-dialog';
import { JobProfilePreview } from '@/features/job/components/job-profile-preview';
import { AIJobEditor } from '@/features/job/components/ai-job-editor';
import { CreateJobModal } from '@/features/job/components/create-job-modal';

export function JobProfilePage() {
  const [jobProfiles, setJobProfiles] = useState<JobProfile[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [departmentFilter, setDepartmentFilter] = useState<string>('all');
  const [searchKeyword, setSearchKeyword] = useState<string>('');
  const [activeSearchKeyword, setActiveSearchKeyword] = useState<string>(''); // 当前生效的搜索关键词
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [totalCount, setTotalCount] = useState(0); // API返回的总数

  // 参数转换映射函数
  const convertEducationRequirement = (education: string): string => {
    const educationMap: Record<string, string> = {
      不限: 'unlimited',
      大专: 'junior_college',
      本科: 'bachelor',
      硕士: 'master',
      博士: 'doctor',
    };
    return educationMap[education] || education;
  };

  const convertExperienceRequirement = (experience: string): string => {
    const experienceMap: Record<string, string> = {
      不限: 'unlimited',
      应届毕业生: 'fresh_graduate',
      '1年以下': 'under_one_year',
      '1-3年': 'one_to_three_years',
      '3-5年': 'three_to_five_years',
      '5-10年': 'five_to_ten_years',
      '10年以上': 'over_ten_years',
    };
    return experienceMap[experience] || experience;
  };

  const convertWorkType = (workType: string): string => {
    const workTypeMap: Record<string, string> = {
      全职: 'full_time',
      兼职: 'part_time',
      实习: 'internship',
      外包: 'outsourcing',
    };
    return workTypeMap[workType] || workType;
  };

  // 获取岗位画像列表
  const fetchJobProfiles = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await listJobProfiles({
        page: currentPage,
        page_size: pageSize,
        keyword: activeSearchKeyword || undefined,
        department_id:
          departmentFilter !== 'all' ? departmentFilter : undefined,
      });
      setJobProfiles(response.items);
      // 设置总数，API直接返回total_count
      setTotalCount(response.total_count || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取岗位画像列表失败');
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, activeSearchKeyword, departmentFilter]);

  // 页面加载时获取数据
  useEffect(() => {
    fetchJobProfiles();
  }, [currentPage, pageSize, fetchJobProfiles]);

  // 页面加载时获取部门数据
  useEffect(() => {
    fetchDepartments();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 新建岗位弹窗状态管理
  const [isCreateJobModalOpen, setIsCreateJobModalOpen] = useState(false);

  // 编辑岗位弹窗状态管理
  const [isEditJobModalOpen, setIsEditJobModalOpen] = useState(false);
  const [editingJob, setEditingJob] = useState<JobProfile | null>(null);
  const [editJobLoading, setEditJobLoading] = useState(false);
  const [editJobError, setEditJobError] = useState<string | null>(null);

  // 删除岗位确认弹窗状态管理
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deletingJob, setDeletingJob] = useState<JobProfile | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  // 删除错误提示弹窗状态
  const [deleteErrorOpen, setDeleteErrorOpen] = useState(false);
  const [deleteErrorMessage, setDeleteErrorMessage] = useState<string>('');

  // 复制岗位确认弹窗状态管理
  const [copyConfirmOpen, setCopyConfirmOpen] = useState(false);
  const [copyingJob, setCopyingJob] = useState<JobProfile | null>(null);
  const [isCopying, setIsCopying] = useState(false);

  // 预览岗位画像弹窗状态
  const [isPreviewModalOpen, setIsPreviewModalOpen] = useState(false);
  const [previewingJob, setPreviewingJob] = useState<JobProfileDetail | null>(
    null
  );
  const [previewLoading, setPreviewLoading] = useState(false);

  // 技能相关状态管理
  const [isAddSkillModalOpen, setIsAddSkillModalOpen] = useState(false);

  // 部门相关状态管理
  const [availableDepartments, setAvailableDepartments] = useState<
    Department[]
  >([]);
  const [departmentsLoading, setDepartmentsLoading] = useState(false);
  const [departmentsError, setDepartmentsError] = useState<string | null>(null);
  const [departmentsPage, setDepartmentsPage] = useState(1);
  const [departmentsPageSize] = useState(20);
  const [departmentsHasMore, setDepartmentsHasMore] = useState(false);

  const [editMode, setEditMode] = useState<'manual' | 'ai'>('manual');
  const [jobFormData, setJobFormData] = useState({
    name: '',
    location: '',
    department: '',
    workType: '',
    education: '',
    salaryMin: '', // 最低薪资
    salaryMax: '', // 最高薪资
    experience: '',
    industryRequirements: [{ industry: '', companyName: '' }], // 行业要求和公司要求数组
    responsibilities: [''], // 改为数组类型，默认包含一个空字符串
  });

  // 技能管理状态 - 保存完整的技能记录信息
  const [selectedRequiredSkills, setSelectedRequiredSkills] = useState<
    JobSkillInput[]
  >([]);
  const [selectedOptionalSkills, setSelectedOptionalSkills] = useState<
    JobSkillInput[]
  >([]);
  // 为了兼容MultiSelect组件，保留技能ID数组
  const [selectedRequiredSkillIds, setSelectedRequiredSkillIds] = useState<
    string[]
  >([]);
  const [selectedOptionalSkillIds, setSelectedOptionalSkillIds] = useState<
    string[]
  >([]);
  const [availableSkills, setAvailableSkills] = useState<JobSkillMeta[]>([]);

  // 技能加载状态
  const [skillsLoading, setSkillsLoading] = useState(false);
  const [skillsHasMore, setSkillsHasMore] = useState(false);
  const [skillsNextToken, setSkillsNextToken] = useState<string | undefined>();
  const [skillsKeyword, setSkillsKeyword] = useState<string>('');

  // 统计数据（使用API返回的总数）
  const totalJobs = totalCount; // 所有岗位的总数
  const publishedJobs = jobProfiles.filter(
    (job) => job.status === 'published'
  ).length;
  const draftJobs = jobProfiles.filter((job) => job.status === 'draft').length;

  // 筛选后的数据（API已处理搜索和department，这里只需要处理状态筛选）
  // 注意：由于API不支持status参数，这里的筛选只作用于当前页的数据
  const filteredJobs = jobProfiles.filter((job) => {
    const statusMatch = statusFilter === 'all' || job.status === statusFilter;
    return statusMatch;
  });

  // 分页数据 - 使用API返回的分页信息
  const totalPages = Math.ceil(totalCount / pageSize);
  const startIndex = (currentPage - 1) * pageSize + 1;
  const endIndex = Math.min(currentPage * pageSize, totalCount);
  const currentJobs = filteredJobs; // 显示当前页筛选后的结果

  // 处理搜索
  const handleSearch = () => {
    setCurrentPage(1); // 搜索时重置到第一页
    setActiveSearchKeyword(searchKeyword); // 设置生效的搜索关键词
  };

  // 处理分页变化
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  // 处理每页条数变化
  const handlePageSizeChange = (newPageSize: number) => {
    setPageSize(newPageSize);
    setCurrentPage(1); // 重置到第一页
  };

  // 获取技能列表（初始加载或搜索）
  const fetchSkills = async (keyword?: string, reset = true) => {
    try {
      setSkillsLoading(true);
      const response = await listJobSkillMeta({
        size: 20,
        keyword: keyword || undefined,
      });

      if (reset) {
        setAvailableSkills(response.items || []);
      } else {
        setAvailableSkills((prev) => [...prev, ...(response.items || [])]);
      }

      setSkillsHasMore(!!response.page_info?.next_token);
      setSkillsNextToken(response.page_info?.next_token);
    } catch (err) {
      console.error('获取技能列表失败:', err);
      setAvailableSkills([]);
    } finally {
      setSkillsLoading(false);
    }
  };

  // 加载更多技能
  const loadMoreSkills = async () => {
    if (!skillsNextToken || skillsLoading) return;

    try {
      setSkillsLoading(true);
      const response = await listJobSkillMeta({
        size: 20,
        next_token: skillsNextToken,
        keyword: skillsKeyword || undefined,
      });

      setAvailableSkills((prev) => [...prev, ...(response.items || [])]);
      setSkillsHasMore(!!response.page_info?.next_token);
      setSkillsNextToken(response.page_info?.next_token);
    } catch (err) {
      console.error('加载更多技能失败:', err);
    } finally {
      setSkillsLoading(false);
    }
  };

  // 搜索技能
  const handleSkillSearch = async (keyword: string) => {
    setSkillsKeyword(keyword);
    await fetchSkills(keyword, true);
  };

  // 处理技能添加成功
  const handleSkillAdded = (newSkill: JobSkillMeta) => {
    // 将新技能添加到可用技能列表中
    setAvailableSkills((prev) => [...prev, newSkill]);
  };

  // 处理必填技能选择变化
  const handleRequiredSkillsChange = (selectedSkillIds: string[]) => {
    setSelectedRequiredSkillIds(selectedSkillIds);

    // 更新完整的技能记录信息
    const updatedSkills: JobSkillInput[] = selectedSkillIds.map((skillId) => {
      // 查找是否已存在的技能记录
      const existingSkill = selectedRequiredSkills.find(
        (skill) => skill.skill_id === skillId
      );
      if (existingSkill) {
        return existingSkill; // 保留现有的技能记录ID
      }

      // 新添加的技能，不包含技能记录ID
      return {
        skill_id: skillId,
        type: 'required' as const,
        weight: 1,
      };
    });

    setSelectedRequiredSkills(updatedSkills);
  };

  // 处理选填技能选择变化
  const handleOptionalSkillsChange = (selectedSkillIds: string[]) => {
    setSelectedOptionalSkillIds(selectedSkillIds);

    // 更新完整的技能记录信息
    const updatedSkills: JobSkillInput[] = selectedSkillIds.map((skillId) => {
      // 查找是否已存在的技能记录
      const existingSkill = selectedOptionalSkills.find(
        (skill) => skill.skill_id === skillId
      );
      if (existingSkill) {
        return existingSkill; // 保留现有的技能记录ID
      }

      // 新添加的技能，不包含技能记录ID
      return {
        skill_id: skillId,
        type: 'bonus' as const,
        weight: 1,
      };
    });

    setSelectedOptionalSkills(updatedSkills);
  };

  // 获取部门列表（分页）
  const fetchDepartments = async (page = 1) => {
    try {
      setDepartmentsLoading(true);
      setDepartmentsError(null);
      const response = await listDepartments({
        page,
        size: departmentsPageSize,
      });
      const items = response.items || [];
      setAvailableDepartments(items);
      setDepartmentsPage(page);
      // 依据has_next_page或next_token判断是否还有更多
      const hasMore = !!response.has_next_page || !!response.next_token;
      setDepartmentsHasMore(hasMore);
    } catch (err) {
      console.error('获取部门列表失败:', err);
      setDepartmentsError(
        err instanceof Error ? err.message : '获取部门列表失败'
      );
      setAvailableDepartments([]);
      setDepartmentsHasMore(false);
    } finally {
      setDepartmentsLoading(false);
    }
  };

  // 加载更多部门
  const loadMoreDepartments = async () => {
    if (departmentsLoading || !departmentsHasMore) return;
    try {
      setDepartmentsLoading(true);
      const nextPage = departmentsPage + 1;
      const response = await listDepartments({
        page: nextPage,
        size: departmentsPageSize,
      });
      const items = response.items || [];
      setAvailableDepartments((prev) => {
        // 去重追加
        const existingIds = new Set(prev.map((d) => d.id));
        const merged = [...prev];
        for (const item of items) {
          if (!existingIds.has(item.id)) merged.push(item);
        }
        return merged;
      });
      setDepartmentsPage(nextPage);
      const hasMore = !!response.has_next_page || !!response.next_token;
      setDepartmentsHasMore(hasMore);
    } catch (err) {
      console.error('加载更多部门失败:', err);
      setDepartmentsHasMore(false);
    } finally {
      setDepartmentsLoading(false);
    }
  };

  // 重试获取部门列表
  const retryFetchDepartments = async () => {
    await fetchDepartments();
  };

  // 处理跳转到平台配置页面（新标签页打开）
  const handleNavigateToPlatformConfig = () => {
    window.open('/platform-config?openModal=true', '_blank');
  };

  // 处理跳转到平台配置页面的技能配置弹窗（新标签页打开）
  const handleNavigateToSkillConfig = () => {
    window.open('/platform-config?openSkillModal=true', '_blank');
  };

  // 打开新建岗位弹窗
  const handleOpenCreateJobModal = () => {
    // 重置表单数据，确保新增岗位时所有字段都为空
    setJobFormData({
      name: '',
      location: '',
      department: '',
      workType: '',
      education: '',
      salaryMin: '',
      salaryMax: '',
      experience: '',
      industryRequirements: [{ industry: '', companyName: '' }],
      responsibilities: [''],
    });
    setEditMode('manual');
    // 清除技能相关状态
    setSelectedRequiredSkills([]);
    setSelectedOptionalSkills([]);
    setSelectedRequiredSkillIds([]);
    setSelectedOptionalSkillIds([]);

    // 先打开弹窗，确保用户能看到界面
    setIsCreateJobModalOpen(true);
    // 然后异步获取数据，不阻塞弹窗显示
    fetchSkills();
    // 重置部门分页并拉取第一页
    setDepartmentsPage(1);
    setAvailableDepartments([]);
    setDepartmentsHasMore(false);
    fetchDepartments(1);
  };

  // 关闭新建岗位弹窗
  const handleCloseCreateJobModal = () => {
    setIsCreateJobModalOpen(false);
    // 重置表单数据
    setJobFormData({
      name: '',
      location: '',
      department: '',
      workType: '',
      education: '',
      salaryMin: '', // 最低薪资
      salaryMax: '', // 最高薪资
      experience: '',
      industryRequirements: [{ industry: '', companyName: '' }],
      responsibilities: [''], // 重置为包含一个空字符串的数组
    });
    setEditMode('manual');
    // 清除技能相关状态
    setSelectedRequiredSkills([]);
    setSelectedOptionalSkills([]);
    setSelectedRequiredSkillIds([]);
    setSelectedOptionalSkillIds([]);
  };

  // 处理表单数据变化
  const handleFormDataChange = (field: string, value: string) => {
    setJobFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  // 处理行业要求变化
  const handleIndustryRequirementChange = (
    index: number,
    field: 'industry' | 'companyName',
    value: string
  ) => {
    setJobFormData((prev) => ({
      ...prev,
      industryRequirements: prev.industryRequirements.map((req, i) =>
        i === index ? { ...req, [field]: value } : req
      ),
    }));
  };

  // 添加新的行业要求
  const addIndustryRequirement = () => {
    setJobFormData((prev) => ({
      ...prev,
      industryRequirements: [
        ...prev.industryRequirements,
        { industry: '', companyName: '' },
      ],
    }));
  };

  // 删除行业要求
  const removeIndustryRequirement = (index: number) => {
    if (index > 0) {
      // 保留第一个，不允许删除
      setJobFormData((prev) => ({
        ...prev,
        industryRequirements: prev.industryRequirements.filter(
          (_, i) => i !== index
        ),
      }));
    }
  };

  // 处理岗位职责数组变化
  const handleResponsibilityChange = (index: number, value: string) => {
    setJobFormData((prev) => ({
      ...prev,
      responsibilities: prev.responsibilities.map((item, i) =>
        i === index ? value : item
      ),
    }));
  };

  // 添加新的岗位职责
  const addResponsibility = () => {
    setJobFormData((prev) => ({
      ...prev,
      responsibilities: [...prev.responsibilities, ''],
    }));
  };

  // 删除岗位职责
  const removeResponsibility = (index: number) => {
    if (index === 0) return; // 不允许删除第一个
    setJobFormData((prev) => ({
      ...prev,
      responsibilities: prev.responsibilities.filter((_, i) => i !== index),
    }));
  };

  // 处理保存草稿
  const handleSaveDraft = async () => {
    try {
      setLoading(true);

      // 准备技能数据
      const skillInputs: JobSkillInput[] = [
        ...selectedRequiredSkillIds.map((skillId) => ({
          skill_id: skillId,
          type: 'required' as const,
        })),
        ...selectedOptionalSkillIds.map((skillId) => ({
          skill_id: skillId,
          type: 'bonus' as const,
        })),
      ];

      // 准备职责数据
      const responsibilityInputs: JobResponsibilityInput[] =
        jobFormData.responsibilities
          .filter((r) => r.trim())
          .map((responsibility, index) => ({
            responsibility,
            sort_order: index + 1,
          }));

      // 准备行业要求数据
      const industryInputs: JobIndustryRequirementInput[] =
        jobFormData.industryRequirements
          .filter((item) => item.industry.trim())
          .map((item) => ({
            industry: item.industry.trim(),
            company_name: item.companyName.trim() || undefined,
          }));

      // 准备学历要求数据
      const educationInputs: JobEducationRequirementInput[] =
        jobFormData.education
          ? [
              {
                education_type: convertEducationRequirement(
                  jobFormData.education
                ),
              },
            ]
          : [];

      // 准备工作经验要求数据
      const experienceInputs: JobExperienceRequirementInput[] =
        jobFormData.experience
          ? [
              {
                experience_type: convertExperienceRequirement(
                  jobFormData.experience
                ),
              },
            ]
          : [];

      const createData: CreateJobProfileReq = {
        name: jobFormData.name,
        department_id: jobFormData.department,
        work_type: convertWorkType(jobFormData.workType),
        description: jobFormData.industryRequirements[0]?.companyName || '', // 第一个公司要求作为描述
        location: jobFormData.location,
        salary_min: jobFormData.salaryMin
          ? parseInt(jobFormData.salaryMin)
          : undefined,
        salary_max: jobFormData.salaryMax
          ? parseInt(jobFormData.salaryMax)
          : undefined,
        status: 'draft',
        education_requirements: educationInputs,
        experience_requirements: experienceInputs,
        industry_requirements: industryInputs,
        responsibilities: responsibilityInputs,
        skills: skillInputs,
      };

      await createJobProfile(createData);
      handleCloseCreateJobModal();
      fetchJobProfiles(); // 重新获取列表
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存草稿失败');
    } finally {
      setLoading(false);
    }
  };

  // 处理保存并发布
  const handleSaveAndPublish = async () => {
    try {
      setLoading(true);

      // 准备技能数据
      const skillInputs: JobSkillInput[] = [
        ...selectedRequiredSkillIds.map((skillId) => ({
          skill_id: skillId,
          type: 'required' as const,
        })),
        ...selectedOptionalSkillIds.map((skillId) => ({
          skill_id: skillId,
          type: 'bonus' as const,
        })),
      ];

      // 准备职责数据
      const responsibilityInputs: JobResponsibilityInput[] =
        jobFormData.responsibilities
          .filter((r) => r.trim())
          .map((responsibility, index) => ({
            responsibility,
            sort_order: index + 1,
          }));

      // 准备行业要求数据
      const industryInputs: JobIndustryRequirementInput[] =
        jobFormData.industryRequirements
          .filter((item) => item.industry.trim())
          .map((item) => ({
            industry: item.industry.trim(),
            company_name: item.companyName.trim() || undefined,
          }));

      // 准备学历要求数据
      const educationInputs: JobEducationRequirementInput[] =
        jobFormData.education
          ? [
              {
                education_type: convertEducationRequirement(
                  jobFormData.education
                ),
              },
            ]
          : [];

      // 准备工作经验要求数据
      const experienceInputs: JobExperienceRequirementInput[] =
        jobFormData.experience
          ? [
              {
                experience_type: convertExperienceRequirement(
                  jobFormData.experience
                ),
              },
            ]
          : [];

      const createData: CreateJobProfileReq = {
        name: jobFormData.name,
        department_id: jobFormData.department,
        work_type: convertWorkType(jobFormData.workType),
        description: jobFormData.industryRequirements[0]?.companyName || '', // 第一个公司要求作为描述
        location: jobFormData.location,
        salary_min: jobFormData.salaryMin
          ? parseInt(jobFormData.salaryMin)
          : undefined,
        salary_max: jobFormData.salaryMax
          ? parseInt(jobFormData.salaryMax)
          : undefined,
        status: 'published',
        education_requirements: educationInputs,
        experience_requirements: experienceInputs,
        industry_requirements: industryInputs,
        responsibilities: responsibilityInputs,
        skills: skillInputs,
      };

      await createJobProfile(createData);
      handleCloseCreateJobModal();
      fetchJobProfiles(); // 重新获取列表
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存并发布失败');
    } finally {
      setLoading(false);
    }
  };

  // 处理编辑按钮点击
  const handleEditJob = async (job: JobProfile) => {
    try {
      setEditJobLoading(true);
      setEditJobError(null);

      // 调用API获取岗位详情
      const jobDetail = await getJobProfile(job.id);

      // 设置编辑的岗位数据
      setEditingJob(jobDetail);

      // 打开新建/编辑岗位弹窗(会自动根据数据判断模式)
      setIsCreateJobModalOpen(true);
    } catch (err) {
      console.error('获取岗位详情失败:', err);
      setEditJobError(err instanceof Error ? err.message : '获取岗位详情失败');
      toast.error('获取岗位详情失败,请重试');
    } finally {
      setEditJobLoading(false);
    }
  };

  // 关闭编辑弹窗
  const handleCloseEditJobModal = () => {
    setIsEditJobModalOpen(false);
    setEditingJob(null);
    // 重置表单数据
    setJobFormData({
      name: '',
      location: '',
      department: '',
      workType: '',
      education: '',
      salaryMin: '',
      salaryMax: '',
      experience: '',
      industryRequirements: [{ industry: '', companyName: '' }],
      responsibilities: [''],
    });
    setEditMode('manual');
    // 清除技能相关状态
    setSelectedRequiredSkillIds([]);
    setSelectedOptionalSkillIds([]);
    // 清除编辑状态
    setEditJobLoading(false);
    setEditJobError(null);
  };

  // 处理更新岗位
  const handleUpdateJob = async () => {
    if (!editingJob) return;

    try {
      setLoading(true);

      // 准备技能数据 - 使用完整的技能记录信息，包含技能记录ID
      const skillInputs: JobSkillInput[] = [
        ...selectedRequiredSkills,
        ...selectedOptionalSkills,
      ];

      // 准备行业要求数据
      const industryRequirements = jobFormData.industryRequirements
        .filter((req) => req.industry.trim() || req.companyName.trim())
        .map((req) => ({
          industry: req.industry,
          company_name: req.companyName,
        }));

      // 准备岗位职责数据
      const responsibilities = jobFormData.responsibilities
        .filter((r) => r.trim())
        .map((responsibility) => ({ responsibility }));

      const updateData: UpdateJobProfileReq = {
        id: editingJob.id,
        name: jobFormData.name,
        department_id: jobFormData.department,
        description: jobFormData.responsibilities
          .filter((r) => r.trim())
          .join('\n'),
        location: jobFormData.location,
        salary_min: jobFormData.salaryMin
          ? parseInt(jobFormData.salaryMin)
          : undefined,
        salary_max: jobFormData.salaryMax
          ? parseInt(jobFormData.salaryMax)
          : undefined,
        work_type: convertWorkType(jobFormData.workType),
        education_requirements: jobFormData.education
          ? [
              {
                education_type: convertEducationRequirement(
                  jobFormData.education
                ),
              },
            ]
          : [],
        experience_requirements: jobFormData.experience
          ? [
              {
                experience_type: convertExperienceRequirement(
                  jobFormData.experience
                ),
              },
            ]
          : [],
        industry_requirements: industryRequirements,
        responsibilities: responsibilities,
        skills: skillInputs,
      };
      await updateJobProfile(editingJob.id, updateData);
      handleCloseEditJobModal();
      fetchJobProfiles(); // 重新获取列表
    } catch (err) {
      setError(err instanceof Error ? err.message : '更新岗位失败');
    } finally {
      setLoading(false);
    }
  };

  // 处理预览岗位按钮点击
  const handlePreviewJob = async (job: JobProfile) => {
    try {
      setPreviewLoading(true);
      // 调用API获取岗位详情
      const jobDetail = await getJobProfile(job.id);
      setPreviewingJob(jobDetail);
      setIsPreviewModalOpen(true);
    } catch (err) {
      console.error('获取岗位详情失败:', err);
      setError(err instanceof Error ? err.message : '获取岗位详情失败');
    } finally {
      setPreviewLoading(false);
    }
  };

  // 处理删除岗位按钮点击
  const handleDeleteJob = (job: JobProfile) => {
    setDeletingJob(job);
    setDeleteConfirmOpen(true);
  };

  // 处理复制岗位按钮点击
  const handleCopyJob = (job: JobProfile) => {
    setCopyingJob(job);
    setCopyConfirmOpen(true);
  };

  // 确认复制岗位
  const handleCopyConfirm = async () => {
    if (!copyingJob) return;

    try {
      setIsCopying(true);
      await duplicateJobProfile(copyingJob.id);
      toast.success(`岗位"${copyingJob.name}"复制成功`);
      setCopyConfirmOpen(false);
      setCopyingJob(null);
      fetchJobProfiles(); // 重新获取列表以显示复制的岗位
    } catch (err: unknown) {
      console.error('复制岗位失败:', err);
      const errorMessage =
        err instanceof Error ? err.message : '复制岗位失败，请重试';
      toast.error(errorMessage);
    } finally {
      setIsCopying(false);
    }
  };

  // 确认删除岗位
  const handleDeleteConfirm = async () => {
    if (!deletingJob) return;

    try {
      setIsDeleting(true);
      await deleteJobProfile(deletingJob.id);
      setDeleteConfirmOpen(false);
      setDeletingJob(null);
      fetchJobProfiles(); // 重新获取列表
    } catch (err: unknown) {
      // 关闭删除确认弹窗
      setDeleteConfirmOpen(false);

      // 类型守卫：检查错误对象结构
      const isErrorWithData = (
        error: unknown
      ): error is {
        data?: { code?: number };
        code?: number;
        message?: string;
      } => {
        return typeof error === 'object' && error !== null;
      };

      // 调试信息：打印错误结构
      if (isErrorWithData(err)) {
        console.log('Delete error:', {
          err,
          code: err.code,
          dataCode: err.data?.code,
          message: err.message,
          data: err.data,
        });

        // 检查是否是岗位已关联简历的错误（错误码 40002）
        // 业务错误码通常在 err.data.code 中
        const businessErrorCode = err.data?.code;
        if (businessErrorCode === 40002) {
          setDeleteErrorMessage(
            '当前岗位已关联简历无法删除，请取消关联后重试。'
          );
          setDeleteErrorOpen(true);
        } else {
          // 其他错误使用通用错误提示
          setError(err instanceof Error ? err.message : '删除岗位失败');
        }
      } else {
        // 未知错误结构
        console.log('Delete error:', err);
        setError(err instanceof Error ? err.message : '删除岗位失败');
      }
    } finally {
      setIsDeleting(false);
    }
  };

  // 生成页码数组
  const generatePageNumbers = () => {
    const pages = [];
    const maxVisiblePages = 5;
    const halfVisible = Math.floor(maxVisiblePages / 2);

    let startPage = Math.max(1, currentPage - halfVisible);
    let endPage = Math.min(totalPages, currentPage + halfVisible);

    // 调整起始页和结束页，确保显示足够的页码
    if (endPage - startPage + 1 < maxVisiblePages) {
      if (startPage === 1) {
        endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);
      } else if (endPage === totalPages) {
        startPage = Math.max(1, endPage - maxVisiblePages + 1);
      }
    }

    for (let i = startPage; i <= endPage; i++) {
      pages.push(i);
    }

    return pages;
  };

  // 获取状态标签样式
  const getStatusBadge = (status: 'published' | 'draft') => {
    if (status === 'published') {
      return (
        <span className="inline-flex items-center rounded-full bg-green-50 px-2 py-1 text-xs font-medium text-green-700 border border-green-200">
          发布
        </span>
      );
    } else {
      return (
        <span className="inline-flex items-center rounded-full bg-yellow-50 px-2 py-1 text-xs font-medium text-yellow-700 border border-yellow-200">
          草稿
        </span>
      );
    }
  };

  return (
    <div className="flex h-full flex-col gap-6 px-6 pb-6 pt-6 page-content">
      {/* 页面标题和操作区域 */}
      <div className="flex items-center justify-between rounded-xl border border-[#E5E7EB] bg-white px-6 py-5 shadow-sm transition-all duration-200 hover:shadow-md">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">岗位画像</h1>
          <p className="text-sm text-muted-foreground">
            管理和创建岗位画像，提升招聘效率
          </p>
        </div>
        <Button
          className="gap-2 rounded-lg px-4 py-2 shadow-sm btn-primary text-white"
          style={{
            background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
          }}
          onClick={handleOpenCreateJobModal}
          disabled={loading}
        >
          <Plus className="h-4 w-4" />
          新建岗位
        </Button>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-4">
          <div className="flex items-center">
            <div className="text-sm text-red-700">{error}</div>
            <Button
              variant="ghost"
              size="sm"
              className="ml-auto text-red-700 hover:text-red-900"
              onClick={() => setError(null)}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}

      {/* 统计卡片区域 */}
      <div className="grid grid-cols-3 gap-6">
        {/* 总岗位数 */}
        <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm card-hover">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">总岗位数</p>
              <p className="text-2xl font-bold text-blue-600 mt-1">
                {totalJobs}
              </p>
            </div>
            <div className="rounded-lg bg-blue-50 p-3 transition-transform duration-200 hover:scale-110">
              <BarChart3 className="h-5 w-5 text-blue-600" />
            </div>
          </div>
        </div>

        {/* 已发布 */}
        <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm card-hover">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">已发布</p>
              <p className="text-2xl font-bold text-green-600 mt-1">
                {publishedJobs}
              </p>
            </div>
            <div className="rounded-lg bg-green-50 p-3 transition-transform duration-200 hover:scale-110">
              <Users className="h-5 w-5 text-green-600" />
            </div>
          </div>
        </div>

        {/* 草稿状态 */}
        <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm card-hover">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">草稿</p>
              <p className="text-2xl font-bold text-yellow-600 mt-1">
                {draftJobs}
              </p>
            </div>
            <div className="rounded-lg bg-yellow-50 p-3 transition-transform duration-200 hover:scale-110">
              <FileText className="h-5 w-5 text-yellow-600" />
            </div>
          </div>
        </div>
      </div>

      {/* 筛选和搜索区域 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            {/* 状态筛选 */}
            <Select
              value={statusFilter}
              onValueChange={(value) => {
                setStatusFilter(value);
                setCurrentPage(1); // 筛选时重置到第一页
              }}
            >
              <SelectTrigger className="w-32">
                <SelectValue placeholder="所有状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">所有状态</SelectItem>
                <SelectItem value="published">已发布</SelectItem>
                <SelectItem value="draft">草稿</SelectItem>
              </SelectContent>
            </Select>

            {/* 部门筛选 */}
            <Select
              value={departmentFilter}
              onValueChange={(value) => {
                setDepartmentFilter(value);
                setCurrentPage(1); // 筛选时重置到第一页
              }}
            >
              <SelectTrigger className="w-32">
                <SelectValue placeholder="所属部门" />
              </SelectTrigger>
              <SelectContent
                onLoadMore={loadMoreDepartments}
                hasMore={departmentsHasMore}
                loading={departmentsLoading}
              >
                <SelectItem value="all">所属部门</SelectItem>
                {availableDepartments.map((department) => (
                  <SelectItem key={department.id} value={department.id}>
                    {department.name}
                  </SelectItem>
                ))}
                {departmentsLoading && (
                  <SelectItem value="__loading_more__" disabled>
                    加载中...
                  </SelectItem>
                )}
              </SelectContent>
            </Select>
          </div>

          {/* 搜索区域 */}
          <div className="flex items-center gap-3">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
              <Input
                placeholder="搜索岗位名称或相关信息"
                value={searchKeyword}
                onChange={(e) => setSearchKeyword(e.target.value)}
                className="pl-10 w-64"
                onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
              />
            </div>
            <Button
              onClick={handleSearch}
              className="gap-2 text-white"
              style={{
                background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
              }}
            >
              <Search className="h-4 w-4" />
              搜索
            </Button>
          </div>
        </div>
      </div>

      {/* 岗位列表表格 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white shadow-sm overflow-hidden">
        {/* 表格标题 */}
        <div className="border-b border-[#E5E7EB] px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">
            岗位列表 ({totalJobs})
          </h3>
        </div>

        {/* 表格内容 */}
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  岗位名称
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  所属部门
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  工作地点
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  状态
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  创建时间
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  创建人
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {loading ? (
                <tr>
                  <td
                    colSpan={6}
                    className="px-6 py-12 text-center text-sm text-gray-500"
                  >
                    <div className="flex items-center justify-center">
                      <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
                      <span className="ml-2">加载中...</span>
                    </div>
                  </td>
                </tr>
              ) : error ? (
                <tr>
                  <td
                    colSpan={6}
                    className="px-6 py-12 text-center text-sm text-red-500"
                  >
                    <div className="flex flex-col items-center">
                      <span>{error}</span>
                      <Button
                        variant="outline"
                        size="sm"
                        className="mt-2"
                        onClick={fetchJobProfiles}
                      >
                        重试
                      </Button>
                    </div>
                  </td>
                </tr>
              ) : currentJobs.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-16">
                    <div className="flex flex-col items-center justify-center space-y-3">
                      <div className="text-gray-400">
                        <svg
                          className="h-12 w-12 mx-auto"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={1}
                            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                          />
                        </svg>
                      </div>
                      <div className="text-sm text-gray-500">
                        暂无岗位画像数据
                      </div>
                      <div className="text-xs text-gray-400">
                        请尝试调整搜索条件或创建新的岗位画像
                      </div>
                    </div>
                  </td>
                </tr>
              ) : (
                currentJobs.map((job) => (
                  <tr key={job.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {job.name}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {job.department}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {job.location}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getStatusBadge(
                        job.status === 'published' || job.status === 'draft'
                          ? job.status
                          : 'draft'
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(job.created_at * 1000).toLocaleDateString(
                        'zh-CN'
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {job.creator_name || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-gray-400 hover:text-blue-600"
                          onClick={() => handlePreviewJob(job)}
                          disabled={previewLoading}
                          title="预览岗位"
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-gray-400 hover:text-gray-600"
                          onClick={() => handleEditJob(job)}
                          title="编辑岗位"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-gray-400 hover:text-blue-600"
                          onClick={() => handleCopyJob(job)}
                          title="复制岗位"
                        >
                          <Copy className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-gray-400 hover:text-red-600"
                          onClick={() => handleDeleteJob(job)}
                          title="删除岗位"
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

        {/* 分页区域 */}
        <div className="border-t border-[#E5E7EB] px-6 py-6">
          <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex items-center gap-4">
              <div className="text-sm text-[#6B7280]">
                {statusFilter === 'all' ? (
                  <>
                    显示第{' '}
                    <span className="text-[#111827] font-medium">
                      {startIndex}
                    </span>{' '}
                    -{' '}
                    <span className="text-[#111827] font-medium">
                      {endIndex}
                    </span>{' '}
                    条，共{' '}
                    <span className="text-[#111827] font-medium">
                      {totalCount}
                    </span>{' '}
                    条结果
                  </>
                ) : (
                  <>
                    当前页显示{' '}
                    <span className="text-[#111827] font-medium">
                      {filteredJobs.length}
                    </span>{' '}
                    条筛选结果（共{' '}
                    <span className="text-[#111827] font-medium">
                      {totalCount}
                    </span>{' '}
                    条数据）
                  </>
                )}
              </div>

              {/* 每页条数选择器 */}
              <div className="flex items-center gap-2">
                <span className="text-sm text-[#6B7280]">每页显示</span>
                <Select
                  value={pageSize.toString()}
                  onValueChange={(value) =>
                    handlePageSizeChange(parseInt(value))
                  }
                >
                  <SelectTrigger className="w-20 h-8">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="10">10</SelectItem>
                    <SelectItem value="20">20</SelectItem>
                    <SelectItem value="50">50</SelectItem>
                  </SelectContent>
                </Select>
                <span className="text-sm text-[#6B7280]">条</span>
              </div>
            </div>

            {/* 分页按钮 - 只在有数据时显示 */}
            {filteredJobs.length > 0 && (
              <div className="flex flex-wrap items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(currentPage - 1)}
                  disabled={currentPage === 1}
                  className={cn(
                    'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                    currentPage === 1 && 'opacity-50'
                  )}
                >
                  <ChevronLeft className="h-4 w-4 text-[#6B7280]" />
                </Button>

                {generatePageNumbers().map((page) => (
                  <Button
                    key={page}
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange(page)}
                    className={cn(
                      'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0 text-sm font-normal text-[#374151]',
                      page === currentPage && 'border-[#7bb8ff] text-white'
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
                  onClick={() => handlePageChange(currentPage + 1)}
                  disabled={currentPage >= totalPages}
                  className={cn(
                    'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                    currentPage >= totalPages && 'opacity-50'
                  )}
                >
                  <ChevronRight className="h-4 w-4 text-[#6B7280]" />
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 新建岗位弹窗 - 使用新的CreateJobModal组件 */}
      <CreateJobModal
        open={isCreateJobModalOpen}
        onOpenChange={(open) => {
          setIsCreateJobModalOpen(open);
          if (!open) {
            // 关闭时清除编辑状态
            setEditingJob(null);
          }
        }}
        onSuccess={fetchJobProfiles}
        editingJob={editingJob}
      />

      {/* 旧的弹窗已被替换,下面的代码可以删除 */}
      {false && (
        <Dialog
          open={isCreateJobModalOpen}
          onOpenChange={setIsCreateJobModalOpen}
        >
          <DialogContent
            className={cn(
              'max-h-[90vh] overflow-y-auto p-0 gap-0',
              editMode === 'ai' ? 'max-w-[1100px]' : 'max-w-[800px]'
            )}
            onInteractOutside={(e) => e.preventDefault()}
          >
            {/* 弹窗标题和关闭按钮 */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-[#E5E7EB]">
              <DialogHeader className="space-y-0">
                <DialogTitle className="text-[18px] font-semibold text-[#1F2937]">
                  新建岗位
                </DialogTitle>
              </DialogHeader>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleCloseCreateJobModal}
                className="h-6 w-6 p-0 text-[#9CA3AF] hover:text-[#6B7280]"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            <div className="px-6 py-6 space-y-6">
              {/* 选择编辑模式 */}
              <div className="space-y-4">
                <div className="flex items-center gap-2">
                  <div
                    className="w-4 h-4 rounded-sm flex items-center justify-center"
                    style={{
                      background:
                        'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                    }}
                  >
                    <div className="w-2 h-2 bg-white rounded-sm"></div>
                  </div>
                  <h3 className="text-[16px] font-medium text-[#1F2937]">
                    选择编辑模式
                  </h3>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div
                    className={cn(
                      'flex items-start gap-3 p-4 border rounded-lg cursor-pointer transition-colors',
                      editMode === 'manual'
                        ? 'border-[#7bb8ff] bg-[#F0FDF4]'
                        : 'border-[#E5E7EB] bg-white hover:bg-[#F9FAFB]'
                    )}
                    onClick={() => setEditMode('manual')}
                  >
                    <div className="flex items-center justify-center w-4 h-4 mt-0.5">
                      <div
                        className={cn(
                          'w-4 h-4 rounded-full border-2 flex items-center justify-center',
                          editMode === 'manual'
                            ? 'border-[#7bb8ff]'
                            : 'border-[#D1D5DB]'
                        )}
                      >
                        {editMode === 'manual' && (
                          <div
                            className="w-2 h-2 rounded-full"
                            style={{
                              background:
                                'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                            }}
                          ></div>
                        )}
                      </div>
                    </div>
                    <div className="flex-1">
                      <div className="text-[14px] font-medium text-[#1F2937] mb-1">
                        手动编辑模式
                      </div>
                      <div className="text-[12px] text-[#6B7280]">
                        完全自定义编辑，精细控制
                      </div>
                    </div>
                  </div>

                  <div
                    className={cn(
                      'flex items-start gap-3 p-4 border rounded-lg cursor-pointer transition-colors',
                      editMode === 'ai'
                        ? 'border-[#7bb8ff] bg-[#F0FDF4]'
                        : 'border-[#E5E7EB] bg-white hover:bg-[#F9FAFB]'
                    )}
                    onClick={() => setEditMode('ai')}
                  >
                    <div className="flex items-center justify-center w-4 h-4 mt-0.5">
                      <div
                        className={cn(
                          'w-4 h-4 rounded-full border-2 flex items-center justify-center',
                          editMode === 'ai'
                            ? 'border-[#7bb8ff]'
                            : 'border-[#D1D5DB]'
                        )}
                      >
                        {editMode === 'ai' && (
                          <div
                            className="w-2 h-2 rounded-full"
                            style={{
                              background:
                                'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                            }}
                          ></div>
                        )}
                      </div>
                    </div>
                    <div className="flex-1">
                      <div className="text-[14px] font-medium text-[#1F2937] mb-1">
                        AI编辑模式
                      </div>
                      <div className="text-[12px] text-[#6B7280]">
                        智能生成岗位内容，快速高效
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* AI编辑模式 */}
              {editMode === 'ai' && (
                <AIJobEditor
                  onClose={handleCloseCreateJobModal}
                  onSave={(content) => {
                    // TODO: 处理AI生成的内容并保存
                    console.log('AI生成的内容:', content);
                  }}
                />
              )}

              {/* 手动编辑模式 - 基本信息表单 */}
              {editMode === 'manual' && (
                <>
                  <div className="space-y-4">
                    <div className="flex items-center gap-2">
                      <div className="w-4 h-4 bg-[#3B82F6] rounded-sm flex items-center justify-center">
                        <div className="w-2 h-2 bg-white rounded-sm"></div>
                      </div>
                      <h3 className="text-[16px] font-medium text-[#1F2937]">
                        基本信息
                      </h3>
                      <span className="text-[12px] text-[#6B7280]">
                        (<span className="text-[#EF4444]">*</span>为必填项)
                      </span>
                    </div>

                    {/* 第一行：岗位名称、工作地点 */}
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label
                          htmlFor="jobName"
                          className="text-[14px] font-medium text-[#374151]"
                        >
                          岗位名称 <span className="text-[#EF4444]">*</span>
                        </Label>
                        <Input
                          id="jobName"
                          placeholder="例如：高级前端开发工程师"
                          value={jobFormData.name}
                          onChange={(e) =>
                            handleFormDataChange('name', e.target.value)
                          }
                          className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label
                          htmlFor="location"
                          className="text-[14px] font-medium text-[#374151]"
                        >
                          工作地点 <span className="text-[#EF4444]">*</span>
                        </Label>
                        <Input
                          id="location"
                          placeholder="例如：北京市朝阳区"
                          value={jobFormData.location}
                          onChange={(e) =>
                            handleFormDataChange('location', e.target.value)
                          }
                          className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
                        />
                      </div>
                    </div>

                    {/* 第二行：所属部门、工作性质 */}
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <div className="flex items-center justify-between min-h-[24px]">
                          <Label
                            htmlFor="department"
                            className="text-[14px] font-medium text-[#374151]"
                          >
                            所属部门 <span className="text-[#EF4444]">*</span>
                          </Label>
                          <div className="flex items-center gap-2">
                            {departmentsError && (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={retryFetchDepartments}
                                disabled={departmentsLoading}
                                className="h-6 w-6 p-0 text-[#6B7280] hover:text-[#374151]"
                                title="重试加载部门"
                              >
                                <RefreshCw
                                  className={`h-3 w-3 ${departmentsLoading ? 'animate-spin' : ''}`}
                                />
                              </Button>
                            )}
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={handleNavigateToPlatformConfig}
                              className="h-8 px-3 text-[12px] text-[#7bb8ff] border-[#7bb8ff] hover:bg-[#F0FDF4] hover:text-[#7bb8ff] hover:border-[#7bb8ff]"
                            >
                              + 新增所属部门
                            </Button>
                          </div>
                        </div>
                        <Select
                          value={jobFormData.department}
                          onValueChange={(value) =>
                            handleFormDataChange('department', value)
                          }
                        >
                          <SelectTrigger className="h-10 border-[#D1D5DB] text-[14px]">
                            <SelectValue placeholder="请选择所属部门" />
                          </SelectTrigger>
                          <SelectContent
                            onLoadMore={loadMoreDepartments}
                            hasMore={departmentsHasMore}
                            loading={departmentsLoading}
                          >
                            {departmentsLoading &&
                            availableDepartments.length === 0 ? (
                              <SelectItem value="__loading__" disabled>
                                <div className="flex items-center gap-2">
                                  <RefreshCw className="h-3 w-3 animate-spin" />
                                  加载中...
                                </div>
                              </SelectItem>
                            ) : departmentsError ? (
                              <div className="p-2">
                                <div className="text-[12px] text-[#EF4444] mb-2">
                                  加载失败
                                </div>
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={retryFetchDepartments}
                                  className="h-6 text-[11px] px-2"
                                >
                                  <RefreshCw className="h-3 w-3 mr-1" />
                                  重试
                                </Button>
                              </div>
                            ) : availableDepartments.length > 0 ? (
                              <>
                                {availableDepartments.map((department) => (
                                  <SelectItem
                                    key={department.id}
                                    value={department.id}
                                  >
                                    {department.name}
                                  </SelectItem>
                                ))}
                                {departmentsLoading && (
                                  <SelectItem value="__loading_more__" disabled>
                                    加载中...
                                  </SelectItem>
                                )}
                              </>
                            ) : (
                              <SelectItem value="__empty__" disabled>
                                暂无部门选项
                              </SelectItem>
                            )}
                          </SelectContent>
                        </Select>
                      </div>
                      <div className="space-y-2">
                        <div className="flex items-center justify-between min-h-[24px]">
                          <Label
                            htmlFor="workType"
                            className="text-[14px] font-medium text-[#374151]"
                          >
                            工作性质 <span className="text-[#EF4444]">*</span>
                          </Label>
                          {/* 占位div保持对齐，与所属部门标题行高度一致 */}
                          <div className="flex items-center gap-2 h-8">
                            {/* 空占位，确保与所属部门按钮区域高度一致 */}
                          </div>
                        </div>
                        <Select
                          value={jobFormData.workType}
                          onValueChange={(value) =>
                            handleFormDataChange('workType', value)
                          }
                        >
                          <SelectTrigger className="h-10 border-[#D1D5DB] text-[14px]">
                            <SelectValue placeholder="请选择工作性质" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="全职">全职</SelectItem>
                            <SelectItem value="兼职">兼职</SelectItem>
                            <SelectItem value="实习">实习</SelectItem>
                            <SelectItem value="外包">外包</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </div>

                    {/* 第三行：学历要求、薪资范围 */}
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label
                          htmlFor="education"
                          className="text-[14px] font-medium text-[#374151]"
                        >
                          学历要求 <span className="text-[#EF4444]">*</span>
                        </Label>
                        <Select
                          value={jobFormData.education}
                          onValueChange={(value) =>
                            handleFormDataChange('education', value)
                          }
                        >
                          <SelectTrigger className="h-10 border-[#D1D5DB] text-[14px]">
                            <SelectValue placeholder="请选择学历要求" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="不限">不限</SelectItem>
                            <SelectItem value="大专">大专</SelectItem>
                            <SelectItem value="本科">本科</SelectItem>
                            <SelectItem value="硕士">硕士</SelectItem>
                            <SelectItem value="博士">博士</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                      <div className="space-y-2">
                        <Label className="text-[14px] font-medium text-[#374151]">
                          薪资范围 <span className="text-[#EF4444]">*</span>
                        </Label>
                        <div className="grid grid-cols-2 gap-2">
                          <Input
                            id="salaryMin"
                            placeholder="最低薪资"
                            value={jobFormData.salaryMin}
                            onChange={(e) =>
                              handleFormDataChange('salaryMin', e.target.value)
                            }
                            className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
                          />
                          <Input
                            id="salaryMax"
                            placeholder="最高薪资"
                            value={jobFormData.salaryMax}
                            onChange={(e) =>
                              handleFormDataChange('salaryMax', e.target.value)
                            }
                            className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
                          />
                        </div>
                      </div>
                    </div>

                    {/* 第四行：工作经验 */}
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label
                          htmlFor="experience"
                          className="text-[14px] font-medium text-[#374151]"
                        >
                          工作经验 <span className="text-[#EF4444]">*</span>
                        </Label>
                        <Select
                          value={jobFormData.experience}
                          onValueChange={(value) =>
                            handleFormDataChange('experience', value)
                          }
                        >
                          <SelectTrigger className="h-10 border-[#D1D5DB] text-[14px]">
                            <SelectValue placeholder="请选择工作经验" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="不限">不限</SelectItem>
                            <SelectItem value="应届毕业生">
                              应届毕业生
                            </SelectItem>
                            <SelectItem value="1年以下">1年以下</SelectItem>
                            <SelectItem value="1-3年">1-3年</SelectItem>
                            <SelectItem value="3-5年">3-5年</SelectItem>
                            <SelectItem value="5-10年">5-10年</SelectItem>
                            <SelectItem value="10年以上">10年以上</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                      <div></div>
                    </div>

                    {/* 行业要求和公司要求 */}
                    <div className="space-y-3">
                      <div>
                        <Label className="text-[14px] font-medium text-[#374151]">
                          行业要求与公司要求
                        </Label>
                      </div>

                      {jobFormData.industryRequirements.map(
                        (requirement, index) => (
                          <div key={index} className="flex gap-2 items-center">
                            <Input
                              placeholder="如：互联网"
                              value={requirement.industry}
                              onChange={(e) =>
                                handleIndustryRequirementChange(
                                  index,
                                  'industry',
                                  e.target.value
                                )
                              }
                              className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF] flex-1"
                            />
                            <Input
                              placeholder="如：阿里巴巴"
                              value={requirement.companyName}
                              onChange={(e) =>
                                handleIndustryRequirementChange(
                                  index,
                                  'companyName',
                                  e.target.value
                                )
                              }
                              className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF] flex-1"
                            />
                            {/* 新增按钮 - 只在最后一行显示 */}
                            {index ===
                              jobFormData.industryRequirements.length - 1 && (
                              <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={addIndustryRequirement}
                                className="h-8 w-8 p-0 border-[#7bb8ff] text-[#7bb8ff] hover:bg-[#F0FDF4] hover:border-[#7bb8ff] hover:text-[#7bb8ff]"
                              >
                                <Plus className="h-3.5 w-3.5" />
                              </Button>
                            )}
                            {/* 删除按钮 - 从第二行开始显示 */}
                            {index > 0 && (
                              <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={() => removeIndustryRequirement(index)}
                                className="h-8 w-8 p-0 border-[#D1D5DB] hover:bg-[#FEF2F2] hover:border-[#EF4444] text-[#EF4444]"
                              >
                                <X className="h-3.5 w-3.5" />
                              </Button>
                            )}
                          </div>
                        )
                      )}
                    </div>

                    {/* 工作技能 */}
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Label className="text-[14px] font-medium text-[#374151]">
                            工作技能
                          </Label>
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={handleNavigateToSkillConfig}
                            className="h-6 px-2 text-[12px] text-[#7bb8ff] border-[#7bb8ff] hover:bg-[#F0FDF4] hover:text-[#7bb8ff] hover:border-[#7bb8ff]"
                          >
                            <Plus className="h-3 w-3 mr-1" />
                            增加
                          </Button>
                        </div>
                        <div></div>
                      </div>

                      <div className="grid grid-cols-2 gap-4">
                        {/* 必填技能 */}
                        <div className="space-y-2">
                          <Label className="text-[14px] font-medium text-[#374151]">
                            必填技能 <span className="text-[#EF4444]">*</span>
                          </Label>
                          <MultiSelect
                            options={availableSkills.map((skill) => ({
                              value: skill.id,
                              label: skill.name,
                            }))}
                            selected={selectedRequiredSkillIds}
                            onChange={handleRequiredSkillsChange}
                            placeholder="请选择必填技能"
                            multiple={true}
                            onSelectAll={() => {
                              const allSkillIds = availableSkills.map(
                                (skill) => skill.id
                              );
                              setSelectedRequiredSkillIds(allSkillIds);
                            }}
                            onClearAll={() => {
                              setSelectedRequiredSkillIds([]);
                            }}
                            selectCountLabel="技能"
                            loading={skillsLoading}
                            hasMore={skillsHasMore}
                            onLoadMore={loadMoreSkills}
                            onSearch={handleSkillSearch}
                          />
                        </div>

                        {/* 选填技能 */}
                        <div className="space-y-2">
                          <Label className="text-[14px] font-medium text-[#374151]">
                            选填技能
                          </Label>
                          <MultiSelect
                            options={availableSkills.map((skill) => ({
                              value: skill.id,
                              label: skill.name,
                            }))}
                            selected={selectedOptionalSkillIds}
                            onChange={handleOptionalSkillsChange}
                            placeholder="请选择选填技能"
                            multiple={true}
                            onSelectAll={() => {
                              const allSkillIds = availableSkills.map(
                                (skill) => skill.id
                              );
                              setSelectedOptionalSkillIds(allSkillIds);
                            }}
                            onClearAll={() => {
                              setSelectedOptionalSkillIds([]);
                            }}
                            selectCountLabel="技能"
                            loading={skillsLoading}
                            hasMore={skillsHasMore}
                            onLoadMore={loadMoreSkills}
                            onSearch={handleSkillSearch}
                          />
                        </div>
                      </div>
                    </div>

                    {/* 岗位职责与要求 */}
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <Label className="text-[14px] font-medium text-[#374151]">
                          岗位职责与要求{' '}
                          <span className="text-[#EF4444]">*</span>
                        </Label>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={addResponsibility}
                          className="h-8 px-3 text-[12px] text-[#7bb8ff] border-[#7bb8ff] hover:bg-[#F0FDF4] hover:text-[#7bb8ff] hover:border-[#7bb8ff]"
                        >
                          <Plus className="h-3 w-3 mr-1" />
                          添加职责
                        </Button>
                      </div>
                      <div className="space-y-3">
                        {jobFormData.responsibilities.map(
                          (responsibility, index) => (
                            <div key={index} className="flex gap-2">
                              <Textarea
                                placeholder={`请描述第${index + 1}项岗位职责与要求`}
                                value={responsibility}
                                onChange={(e) =>
                                  handleResponsibilityChange(
                                    index,
                                    e.target.value
                                  )
                                }
                                className="min-h-[80px] resize-none border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF] flex-1"
                              />
                              {index > 0 && (
                                <Button
                                  type="button"
                                  variant="outline"
                                  size="sm"
                                  onClick={() => removeResponsibility(index)}
                                  className="h-8 w-8 p-0 border-[#D1D5DB] hover:bg-[#FEF2F2] hover:border-[#EF4444] text-[#EF4444] mt-0"
                                  title="删除此项职责"
                                >
                                  <X className="h-3.5 w-3.5" />
                                </Button>
                              )}
                            </div>
                          )
                        )}
                      </div>
                    </div>
                  </div>
                </>
              )}
            </div>
            {/* 底部操作按钮 - 仅在手动编辑模式下显示 */}
            {editMode === 'manual' && (
              <div className="flex justify-end gap-3 px-6 py-4 border-t border-[#E5E7EB] bg-[#F9FAFB]">
                <Button
                  variant="outline"
                  onClick={handleCloseCreateJobModal}
                  className="h-10 px-6 border-[#D1D5DB] text-[#374151] hover:bg-[#F3F4F6]"
                >
                  取消
                </Button>
                <Button
                  variant="outline"
                  onClick={handleSaveDraft}
                  disabled={loading}
                  className="h-10 px-6 border-[#D1D5DB] text-[#374151] hover:bg-[#F3F4F6]"
                >
                  {loading ? '保存中...' : '保存草稿'}
                </Button>
                <Button
                  onClick={handleSaveAndPublish}
                  disabled={loading}
                  className="h-10 px-6 text-white"
                  style={{
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }}
                >
                  {loading ? '发布中...' : '保存并发布'}
                </Button>
              </div>
            )}
          </DialogContent>
        </Dialog>
      )}

      {/* 编辑岗位弹窗 */}
      <Dialog open={isEditJobModalOpen} onOpenChange={setIsEditJobModalOpen}>
        <DialogContent
          className="max-w-[800px] max-h-[90vh] overflow-y-auto p-0 gap-0"
          onInteractOutside={(e) => e.preventDefault()}
        >
          {/* 弹窗标题和关闭按钮 */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-[#E5E7EB]">
            <DialogHeader className="space-y-0">
              <DialogTitle className="text-[18px] font-semibold text-[#1F2937]">
                编辑岗位
              </DialogTitle>
            </DialogHeader>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleCloseEditJobModal}
              className="h-6 w-6 p-0 text-[#9CA3AF] hover:text-[#6B7280]"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>

          {/* 加载状态 */}
          {editJobLoading && (
            <div className="flex items-center justify-center py-8">
              <RefreshCw className="h-6 w-6 animate-spin text-blue-500 mr-2" />
              <span className="text-gray-600">正在加载岗位详情...</span>
            </div>
          )}

          {/* 错误状态 */}
          {editJobError && (
            <div className="bg-red-50 border border-red-200 rounded-md p-4 mx-6 mb-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <X className="h-5 w-5 text-red-400" />
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800">加载失败</h3>
                  <div className="mt-2 text-sm text-red-700">
                    {editJobError}
                  </div>
                </div>
              </div>
            </div>
          )}

          <div className="px-6 py-6 space-y-6">
            {/* 基本信息 */}
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <div
                  className="w-4 h-4 rounded-sm flex items-center justify-center"
                  style={{
                    background:
                      'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                  }}
                >
                  <div className="w-2 h-2 bg-white rounded-sm"></div>
                </div>
                <h3 className="text-[16px] font-medium text-[#1F2937]">
                  基本信息
                </h3>
                <span className="text-[12px] text-[#6B7280]">
                  (<span className="text-[#EF4444]">*</span>为必填项)
                </span>
              </div>

              {/* 第一行：岗位名称、工作地点 */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label
                    htmlFor="jobName"
                    className="text-[14px] font-medium text-[#374151]"
                  >
                    岗位名称 <span className="text-[#EF4444]">*</span>
                  </Label>
                  <Input
                    id="jobName"
                    placeholder="例如：高级前端开发工程师"
                    value={jobFormData.name}
                    onChange={(e) =>
                      handleFormDataChange('name', e.target.value)
                    }
                    className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
                  />
                </div>
                <div className="space-y-2">
                  <Label
                    htmlFor="location"
                    className="text-[14px] font-medium text-[#374151]"
                  >
                    工作地点 <span className="text-[#EF4444]">*</span>
                  </Label>
                  <Input
                    id="location"
                    placeholder="例如：北京市朝阳区"
                    value={jobFormData.location}
                    onChange={(e) =>
                      handleFormDataChange('location', e.target.value)
                    }
                    className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
                  />
                </div>
              </div>

              {/* 第二行：所属部门、工作性质 */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between min-h-[24px]">
                    <Label
                      htmlFor="department"
                      className="text-[14px] font-medium text-[#374151]"
                    >
                      所属部门 <span className="text-[#EF4444]">*</span>
                    </Label>
                    <div className="flex items-center gap-2">
                      {departmentsError && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={retryFetchDepartments}
                          disabled={departmentsLoading}
                          className="h-6 w-6 p-0 text-[#6B7280] hover:text-[#374151]"
                          title="重试加载部门"
                        >
                          <RefreshCw
                            className={`h-3 w-3 ${departmentsLoading ? 'animate-spin' : ''}`}
                          />
                        </Button>
                      )}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleNavigateToPlatformConfig}
                        className="h-8 px-3 text-[12px] text-[#7bb8ff] border-[#7bb8ff] hover:bg-[#F0FDF4] hover:text-[#7bb8ff] hover:border-[#7bb8ff]"
                      >
                        + 新增所属部门
                      </Button>
                    </div>
                  </div>
                  <Select
                    value={jobFormData.department}
                    onValueChange={(value) =>
                      handleFormDataChange('department', value)
                    }
                  >
                    <SelectTrigger className="h-10 border-[#D1D5DB] text-[14px]">
                      <SelectValue placeholder="请选择所属部门" />
                    </SelectTrigger>
                    <SelectContent
                      onLoadMore={loadMoreDepartments}
                      hasMore={departmentsHasMore}
                      loading={departmentsLoading}
                    >
                      {departmentsLoading &&
                      availableDepartments.length === 0 ? (
                        <SelectItem value="__loading__" disabled>
                          <div className="flex items-center gap-2">
                            <RefreshCw className="h-3 w-3 animate-spin" />
                            加载中...
                          </div>
                        </SelectItem>
                      ) : departmentsError ? (
                        <SelectItem value="__error__" disabled>
                          <div className="flex items-center gap-2 text-red-600">
                            <X className="h-3 w-3" />
                            加载失败
                          </div>
                        </SelectItem>
                      ) : (
                        <>
                          {availableDepartments.map((department) => (
                            <SelectItem
                              key={department.id}
                              value={department.id}
                            >
                              {department.name}
                            </SelectItem>
                          ))}
                          {departmentsLoading && (
                            <SelectItem value="__loading_more__" disabled>
                              加载中...
                            </SelectItem>
                          )}
                        </>
                      )}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label
                    htmlFor="workType"
                    className="text-[14px] font-medium text-[#374151]"
                  >
                    工作性质 <span className="text-[#EF4444]">*</span>
                  </Label>
                  <Select
                    value={jobFormData.workType}
                    onValueChange={(value) =>
                      handleFormDataChange('workType', value)
                    }
                  >
                    <SelectTrigger className="h-10 border-[#D1D5DB] text-[14px]">
                      <SelectValue placeholder="请选择工作性质" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="全职">全职</SelectItem>
                      <SelectItem value="兼职">兼职</SelectItem>
                      <SelectItem value="实习">实习</SelectItem>
                      <SelectItem value="外包">外包</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {/* 第三行：学历要求、薪资范围 */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label
                    htmlFor="education"
                    className="text-[14px] font-medium text-[#374151]"
                  >
                    学历要求 <span className="text-[#EF4444]">*</span>
                  </Label>
                  <Select
                    value={jobFormData.education}
                    onValueChange={(value) =>
                      handleFormDataChange('education', value)
                    }
                  >
                    <SelectTrigger className="h-10 border-[#D1D5DB] text-[14px]">
                      <SelectValue placeholder="请选择学历要求" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="不限">不限</SelectItem>
                      <SelectItem value="大专">大专</SelectItem>
                      <SelectItem value="本科">本科</SelectItem>
                      <SelectItem value="硕士">硕士</SelectItem>
                      <SelectItem value="博士">博士</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label className="text-[14px] font-medium text-[#374151]">
                    薪资范围
                  </Label>
                  <div className="flex gap-2 items-center">
                    <Input
                      id="salaryMin"
                      placeholder="最低薪资"
                      value={jobFormData.salaryMin}
                      onChange={(e) =>
                        handleFormDataChange('salaryMin', e.target.value)
                      }
                      className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
                    />
                    <span className="text-[#6B7280] text-[14px]">-</span>
                    <Input
                      id="salaryMax"
                      placeholder="最高薪资"
                      value={jobFormData.salaryMax}
                      onChange={(e) =>
                        handleFormDataChange('salaryMax', e.target.value)
                      }
                      className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
                    />
                  </div>
                </div>
              </div>

              {/* 第四行：工作经验 */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label
                    htmlFor="experience"
                    className="text-[14px] font-medium text-[#374151]"
                  >
                    工作经验 <span className="text-[#EF4444]">*</span>
                  </Label>
                  <Select
                    value={jobFormData.experience}
                    onValueChange={(value) =>
                      handleFormDataChange('experience', value)
                    }
                  >
                    <SelectTrigger className="h-10 border-[#D1D5DB] text-[14px]">
                      <SelectValue placeholder="请选择工作经验" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="不限">不限</SelectItem>
                      <SelectItem value="应届毕业生">应届毕业生</SelectItem>
                      <SelectItem value="1年以下">1年以下</SelectItem>
                      <SelectItem value="1-3年">1-3年</SelectItem>
                      <SelectItem value="3-5年">3-5年</SelectItem>
                      <SelectItem value="5-10年">5-10年</SelectItem>
                      <SelectItem value="10年以上">10年以上</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div></div>
              </div>

              {/* 行业要求和公司要求 */}
              <div className="space-y-3">
                <div>
                  <Label className="text-[14px] font-medium text-[#374151]">
                    行业要求与公司要求
                  </Label>
                </div>

                {jobFormData.industryRequirements.map((requirement, index) => (
                  <div key={index} className="flex gap-2 items-center">
                    <Input
                      placeholder="如：互联网"
                      value={requirement.industry}
                      onChange={(e) =>
                        handleIndustryRequirementChange(
                          index,
                          'industry',
                          e.target.value
                        )
                      }
                      className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF] flex-1"
                    />
                    <Input
                      placeholder="如：阿里巴巴"
                      value={requirement.companyName}
                      onChange={(e) =>
                        handleIndustryRequirementChange(
                          index,
                          'companyName',
                          e.target.value
                        )
                      }
                      className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF] flex-1"
                    />
                    {/* 新增按钮 - 只在最后一行显示 */}
                    {index === jobFormData.industryRequirements.length - 1 && (
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={addIndustryRequirement}
                        className="h-8 w-8 p-0 border-[#7bb8ff] text-[#7bb8ff] hover:bg-[#F0FDF4] hover:border-[#7bb8ff] hover:text-[#7bb8ff]"
                      >
                        <Plus className="h-3.5 w-3.5" />
                      </Button>
                    )}
                    {/* 删除按钮 - 从第二行开始显示 */}
                    {index > 0 && (
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => removeIndustryRequirement(index)}
                        className="h-8 w-8 p-0 border-[#D1D5DB] hover:bg-[#FEF2F2] hover:border-[#EF4444] text-[#EF4444]"
                      >
                        <X className="h-3.5 w-3.5" />
                      </Button>
                    )}
                  </div>
                ))}
              </div>

              {/* 工作技能 */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div
                      className="w-4 h-4 rounded-sm flex items-center justify-center"
                      style={{
                        background:
                          'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
                      }}
                    >
                      <div className="w-2 h-2 bg-white rounded-sm"></div>
                    </div>
                    <h3 className="text-[16px] font-medium text-[#1F2937]">
                      工作技能
                    </h3>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={handleNavigateToSkillConfig}
                      className="h-6 px-2 text-[12px] text-[#7bb8ff] border-[#7bb8ff] hover:bg-[#F0FDF4] hover:text-[#7bb8ff] hover:border-[#7bb8ff]"
                    >
                      <Plus className="h-3 w-3 mr-1" />
                      增加
                    </Button>
                  </div>
                  <div></div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  {/* 必填技能 */}
                  <div className="space-y-2">
                    <Label className="text-[14px] font-medium text-[#374151]">
                      必填技能 <span className="text-[#EF4444]">*</span>
                    </Label>
                    <MultiSelect
                      options={availableSkills.map((skill) => ({
                        value: skill.id,
                        label: skill.name,
                      }))}
                      selected={selectedRequiredSkillIds}
                      onChange={handleRequiredSkillsChange}
                      placeholder="请选择必填技能"
                      multiple={true}
                      onSelectAll={() => {
                        const allSkillIds = availableSkills.map(
                          (skill) => skill.id
                        );
                        handleRequiredSkillsChange(allSkillIds);
                      }}
                      onClearAll={() => {
                        handleRequiredSkillsChange([]);
                      }}
                      selectCountLabel="技能"
                      loading={skillsLoading}
                      hasMore={skillsHasMore}
                      onLoadMore={loadMoreSkills}
                      onSearch={handleSkillSearch}
                    />
                  </div>

                  {/* 选填技能 */}
                  <div className="space-y-2">
                    <Label className="text-[14px] font-medium text-[#374151]">
                      选填技能
                    </Label>
                    <MultiSelect
                      options={availableSkills.map((skill) => ({
                        value: skill.id,
                        label: skill.name,
                      }))}
                      selected={selectedOptionalSkillIds}
                      onChange={handleOptionalSkillsChange}
                      placeholder="请选择选填技能"
                      multiple={true}
                      onSelectAll={() => {
                        const allSkillIds = availableSkills.map(
                          (skill) => skill.id
                        );
                        handleOptionalSkillsChange(allSkillIds);
                      }}
                      onClearAll={() => {
                        handleOptionalSkillsChange([]);
                      }}
                      selectCountLabel="技能"
                      loading={skillsLoading}
                      hasMore={skillsHasMore}
                      onLoadMore={loadMoreSkills}
                      onSearch={handleSkillSearch}
                    />
                  </div>
                </div>
              </div>

              {/* 岗位职责与要求 */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label className="text-[14px] font-medium text-[#374151]">
                    岗位职责与要求 <span className="text-[#EF4444]">*</span>
                  </Label>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={addResponsibility}
                    className="h-8 px-3 text-[12px] text-[#7bb8ff] border-[#7bb8ff] hover:bg-[#F0FDF4] hover:text-[#7bb8ff] hover:border-[#7bb8ff]"
                  >
                    <Plus className="h-3 w-3 mr-1" />
                    添加职责
                  </Button>
                </div>
                <div className="space-y-3">
                  {jobFormData.responsibilities.map((responsibility, index) => (
                    <div key={index} className="flex gap-2">
                      <Textarea
                        placeholder={`请描述第${index + 1}项岗位职责与要求`}
                        value={responsibility}
                        onChange={(e) =>
                          handleResponsibilityChange(index, e.target.value)
                        }
                        className="min-h-[80px] resize-none border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF] flex-1"
                      />
                      {index > 0 && (
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => removeResponsibility(index)}
                          className="h-8 w-8 p-0 border-[#D1D5DB] hover:bg-[#FEF2F2] hover:border-[#EF4444] text-[#EF4444] mt-0"
                          title="删除此项职责"
                        >
                          <X className="h-3.5 w-3.5" />
                        </Button>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* 底部操作按钮 */}
          <div className="flex justify-end gap-3 px-6 py-4 border-t border-[#E5E7EB] bg-[#F9FAFB]">
            <Button
              variant="outline"
              onClick={handleCloseEditJobModal}
              className="h-10 px-6 border-[#D1D5DB] text-[#374151] hover:bg-[#F3F4F6]"
            >
              取消
            </Button>
            <Button
              onClick={handleUpdateJob}
              disabled={loading}
              className="h-10 px-6 text-white"
              style={{
                background: 'linear-gradient(135deg, #7bb8ff 0%, #3F3663 100%)',
              }}
            >
              {loading ? '更新中...' : '保存更新'}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* 添加技能弹窗 */}
      <AddSkillModal
        isOpen={isAddSkillModalOpen}
        onClose={() => setIsAddSkillModalOpen(false)}
        onSkillAdded={handleSkillAdded}
      />

      {/* 复制确认对话框 */}
      <ConfirmDialog
        open={copyConfirmOpen}
        onOpenChange={setCopyConfirmOpen}
        onConfirm={handleCopyConfirm}
        title="复制岗位"
        description={`确定要复制岗位"${copyingJob?.name}"吗？`}
        confirmText="确定"
        cancelText="取消"
        loading={isCopying}
      />

      {/* 删除确认对话框 */}
      <ConfirmDialog
        open={deleteConfirmOpen}
        onOpenChange={setDeleteConfirmOpen}
        onConfirm={handleDeleteConfirm}
        title="删除岗位"
        description={`确定要删除岗位"${deletingJob?.name}"吗？此操作不可撤销。`}
        confirmText="删除"
        cancelText="取消"
        variant="destructive"
        loading={isDeleting}
      />

      {/* 删除错误提示对话框 */}
      <Dialog open={deleteErrorOpen} onOpenChange={setDeleteErrorOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="mb-1 font-medium leading-none tracking-tight text-destructive flex items-center gap-2">
              <X className="h-4 w-4" />
              删除失败
            </DialogTitle>
          </DialogHeader>
          <div className="py-4">
            <div className="text-sm text-destructive [&_p]:leading-relaxed">
              {deleteErrorMessage}
            </div>
          </div>
          <div className="flex justify-end">
            <Button
              onClick={() => {
                setDeleteErrorOpen(false);
                setDeletingJob(null);
              }}
              className="bg-destructive hover:bg-destructive/90 text-destructive-foreground"
            >
              我知道了
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* 岗位画像预览对话框 */}
      <JobProfilePreview
        open={isPreviewModalOpen}
        onOpenChange={setIsPreviewModalOpen}
        jobProfile={previewingJob}
      />
    </div>
  );
}

// 默认导出JobProfilePage组件
export default JobProfilePage;
