import { useState, useEffect, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import {
  Settings,
  Building2,
  X,
  Plus,
  Edit,
  Trash2,
  ChevronLeft,
  ChevronRight,
  Loader2,
  AlertCircle,
  Wrench,
  FileText,
  Bell,
} from 'lucide-react';
import { Button } from '@/ui/button';
import { ConfirmDialog } from '@/ui/confirm-dialog';
import { cn } from '@/lib/utils';
import { ResumeCollectionModal } from '@/features/resume-collection/components/ResumeCollectionModal';
import { NotificationConfigModal } from '@/features/notification/components/notification-config-modal';
import {
  Department,
  listDepartments,
  createDepartment,
  updateDepartment,
  deleteDepartment,
  CreateDepartmentRequest,
  UpdateDepartmentRequest,
} from '@/services/department';
import {
  listJobSkillMeta,
  createJobSkillMeta,
  deleteJobSkillMeta,
} from '@/services/job-profile';
import type { JobSkillMeta } from '@/types/job-profile';

/**
 * 平台配置页面组件
 * 严格按照Figma设计图实现界面布局、色彩方案、字体样式、间距和尺寸比例
 */
export default function PlatformConfig() {
  const [searchParams] = useSearchParams();

  // 弹窗状态管理
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isAddDeptModalOpen, setIsAddDeptModalOpen] = useState(false);
  const [isEditDeptModalOpen, setIsEditDeptModalOpen] = useState(false);

  // 技能配置弹窗状态管理
  const [isSkillModalOpen, setIsSkillModalOpen] = useState(false);
  const [isAddSkillModalOpen, setIsAddSkillModalOpen] = useState(false);

  // 简历收集配置弹窗状态管理
  const [isResumeCollectionModalOpen, setIsResumeCollectionModalOpen] =
    useState(false);

  // 通知配置弹窗状态管理
  const [isNotificationConfigModalOpen, setIsNotificationConfigModalOpen] =
    useState(false);

  // 删除确认弹窗状态
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [departmentToDelete, setDepartmentToDelete] = useState<{
    id: string;
    name: string;
  } | null>(null);

  // 技能删除确认弹窗状态
  const [isSkillDeleteConfirmOpen, setIsSkillDeleteConfirmOpen] =
    useState(false);
  const [skillToDelete, setSkillToDelete] = useState<{
    id: string;
    name: string;
  } | null>(null);

  // 部门数据状态
  const [departments, setDepartments] = useState<Department[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0); // 添加总数据量状态
  const [pageSize] = useState(7);

  // 技能数据状态
  const [skills, setSkills] = useState<JobSkillMeta[]>([]);
  const [skillLoading, setSkillLoading] = useState(false);
  const [skillError, setSkillError] = useState<string | null>(null);
  const [skillCurrentPage, setSkillCurrentPage] = useState(1);
  const [skillTotalPages, setSkillTotalPages] = useState(1);
  const [skillTotalCount, setSkillTotalCount] = useState(0);
  const [skillPageSize] = useState(7);

  // 表单数据
  const [newDeptForm, setNewDeptForm] = useState<CreateDepartmentRequest>({
    name: '',
    description: '',
  });

  const [editDeptForm, setEditDeptForm] = useState<UpdateDepartmentRequest>({
    name: '',
    description: '',
  });

  const [editingDeptId, setEditingDeptId] = useState<string | null>(null);

  // 技能表单数据
  const [newSkillForm, setNewSkillForm] = useState<{ name: string }>({
    name: '',
  });

  // 操作状态
  const [submitting, setSubmitting] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [skillSubmitting, setSkillSubmitting] = useState(false);
  const [skillDeletingId, setSkillDeletingId] = useState<string | null>(null);

  // 获取部门列表
  const fetchDepartments = useCallback(
    async (page?: number) => {
      try {
        setLoading(true);
        setError(null);
        const targetPage = page || currentPage;
        console.log(`📋 获取部门列表 - 页码: ${targetPage}, 每页: ${pageSize}`);

        const response = await listDepartments({
          page: targetPage,
          size: pageSize,
        });

        const newTotalPages = Math.ceil(response.total_count / pageSize);

        console.log(
          `📋 获取到 ${response.items.length} 个部门，总计: ${response.total_count}，总页数: ${newTotalPages}`
        );
        console.log(
          '📋 部门列表数据:',
          response.items.map((d) => ({ id: d.id, name: d.name }))
        );

        setDepartments(response.items);
        setTotalCount(response.total_count);
        setTotalPages(newTotalPages);

        // 如果当前页超出了总页数，自动调整到最后一页
        if (targetPage > newTotalPages && newTotalPages > 0) {
          console.log(
            `⚠️ 当前页 ${targetPage} 超出总页数 ${newTotalPages}，自动调整到第 ${newTotalPages} 页`
          );
          setCurrentPage(newTotalPages);
          // 递归调用获取正确页面的数据
          await fetchDepartments(newTotalPages);
          return;
        }
      } catch (err) {
        setError('获取部门列表失败，请重试');
        console.error('获取部门列表失败:', err);
        // 出错时重置数据
        setDepartments([]);
        setTotalCount(0);
        setTotalPages(0);
      } finally {
        setLoading(false);
      }
    },
    [currentPage, pageSize]
  );

  // 打开弹窗
  const handleOpenModal = async () => {
    console.log('🔧 打开部门配置弹窗，重新获取最新数据...');
    // 重置分页状态到第一页
    setCurrentPage(1);
    // 清除之前的错误信息
    setError(null);
    // 打开弹窗
    setIsModalOpen(true);
    // 重新获取最新的部门数据，强制从第一页开始
    await fetchDepartments(1);
  };

  // 关闭弹窗
  const handleCloseModal = () => {
    setIsModalOpen(false);
  };

  // 打开添加部门弹窗
  const handleOpenAddDeptModal = () => {
    setIsAddDeptModalOpen(true);
  };

  // 关闭添加部门弹窗
  const handleCloseAddDeptModal = () => {
    setIsAddDeptModalOpen(false);
    // 重置表单数据
    setNewDeptForm({
      name: '',
      description: '',
    });
  };

  // 打开编辑部门弹窗
  const handleOpenEditDeptModal = (dept: Department) => {
    setEditingDeptId(dept.id);
    setEditDeptForm({
      name: dept.name,
      description: dept.description || '',
    });
    setIsEditDeptModalOpen(true);
  };

  // 关闭编辑部门弹窗
  const handleCloseEditDeptModal = () => {
    setIsEditDeptModalOpen(false);
    setEditingDeptId(null);
    setEditDeptForm({
      name: '',
      description: '',
    });
  };

  // 处理新增表单输入变化
  const handleNewFormChange = (
    field: keyof CreateDepartmentRequest,
    value: string
  ) => {
    setNewDeptForm((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  // 处理编辑表单输入变化
  const handleEditFormChange = (
    field: keyof UpdateDepartmentRequest,
    value: string
  ) => {
    setEditDeptForm((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  // 保存新部门
  const handleSaveNewDept = async () => {
    if (!newDeptForm.name.trim()) {
      setError('部门名称不能为空');
      return;
    }

    try {
      setSubmitting(true);
      setError(null);
      console.log('💾 创建新部门:', newDeptForm);
      await createDepartment(newDeptForm);
      handleCloseAddDeptModal();

      // 重置到第一页并刷新数据，确保新添加的部门能够显示
      console.log('🔄 新部门创建成功，重置到第一页并刷新数据');
      setCurrentPage(1);
      await fetchDepartments(1); // 强制从第一页获取最新数据
    } catch (err) {
      setError('创建部门失败，请重试');
      console.error('创建部门失败:', err);
    } finally {
      setSubmitting(false);
    }
  };

  // 保存编辑部门
  const handleSaveEditDept = async () => {
    if (!editDeptForm.name?.trim() || !editingDeptId) {
      setError('部门名称不能为空');
      return;
    }

    try {
      setSubmitting(true);
      setError(null);
      console.log('✏️ 更新部门:', editingDeptId, editDeptForm);
      await updateDepartment(editingDeptId, editDeptForm);
      handleCloseEditDeptModal();

      // 刷新当前页数据
      console.log('🔄 部门更新成功，刷新当前页数据');
      await fetchDepartments(currentPage);
    } catch (err) {
      setError('更新部门失败，请重试');
      console.error('更新部门失败:', err);
    } finally {
      setSubmitting(false);
    }
  };

  // 删除部门
  const handleDeleteDept = async (deptId: string) => {
    // 找到要删除的部门信息
    const dept = departments.find((d) => d.id === deptId);
    if (!dept) return;

    // 设置要删除的部门信息并显示确认弹窗
    setDepartmentToDelete({ id: deptId, name: dept.name });
    setIsDeleteConfirmOpen(true);
  };

  // 确认删除部门
  const handleConfirmDelete = async () => {
    if (!departmentToDelete) return;

    try {
      setDeletingId(departmentToDelete.id);
      setError(null);
      console.log('🗑️ 删除部门:', departmentToDelete);
      await deleteDepartment(departmentToDelete.id);

      // 删除后需要检查当前页是否还有数据，如果没有则回到上一页
      const remainingItems = totalCount - 1;
      const maxPage = Math.ceil(remainingItems / pageSize);
      const targetPage =
        currentPage > maxPage ? Math.max(1, maxPage) : currentPage;

      console.log(
        `🔄 部门删除成功，剩余 ${remainingItems} 条数据，当前页: ${currentPage}, 目标页: ${targetPage}`
      );

      if (targetPage !== currentPage) {
        setCurrentPage(targetPage);
      }

      await fetchDepartments(targetPage);
      setIsDeleteConfirmOpen(false);
      setDepartmentToDelete(null);
    } catch (err) {
      setError('删除部门失败，请重试');
      console.error('删除部门失败:', err);
    } finally {
      setDeletingId(null);
    }
  };

  // 页面切换
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  // 生成页码数组 - 与简历管理页面保持一致
  const generatePageNumbers = () => {
    const pages = [];
    const safeTotalPages = totalPages || 1;
    const safeCurrentPage = currentPage || 1;

    // 显示当前页前后2页
    const start = Math.max(1, safeCurrentPage - 2);
    const end = Math.min(safeTotalPages, safeCurrentPage + 2);

    for (let i = start; i <= end; i++) {
      pages.push(i);
    }

    return pages;
  };

  // ==================== 技能管理相关函数 ====================

  // 获取技能列表
  const fetchSkills = useCallback(
    async (page?: number) => {
      try {
        setSkillLoading(true);
        setSkillError(null);
        const targetPage = page || skillCurrentPage;
        console.log(
          `📋 获取技能列表 - 页码: ${targetPage}, 每页: ${skillPageSize}`
        );

        const response = await listJobSkillMeta({
          page: targetPage,
          size: skillPageSize,
        });
        const skills = response.items || [];
        const pageInfo = response.page_info;

        // 使用后端返回的分页信息
        const totalCount = pageInfo.total_count;
        const newTotalPages = Math.ceil(totalCount / skillPageSize);

        console.log(
          `📋 获取到 ${skills.length} 个技能，总计: ${totalCount}，总页数: ${newTotalPages}`
        );

        setSkills(skills);
        setSkillTotalCount(totalCount);
        setSkillTotalPages(newTotalPages);

        // 如果当前页超出了总页数，自动调整到最后一页
        if (targetPage > newTotalPages && newTotalPages > 0) {
          console.log(
            `⚠️ 当前页 ${targetPage} 超出总页数 ${newTotalPages}，自动调整到第 ${newTotalPages} 页`
          );
          setSkillCurrentPage(newTotalPages);
          await fetchSkills(newTotalPages);
          return;
        }
      } catch (err) {
        setSkillError('获取技能列表失败，请重试');
        console.error('获取技能列表失败:', err);
        setSkills([]);
        setSkillTotalCount(0);
        setSkillTotalPages(0);
      } finally {
        setSkillLoading(false);
      }
    },
    [skillCurrentPage, skillPageSize]
  );

  // 打开技能配置弹窗
  const handleOpenSkillModal = useCallback(async () => {
    console.log('🔧 打开技能配置弹窗，重新获取最新数据...');
    setSkillCurrentPage(1);
    setSkillError(null);
    setIsSkillModalOpen(true);
    await fetchSkills(1);
  }, [fetchSkills]);

  // 关闭技能配置弹窗
  const handleCloseSkillModal = () => {
    setIsSkillModalOpen(false);
  };

  // 打开添加技能弹窗
  const handleOpenAddSkillModal = () => {
    setIsAddSkillModalOpen(true);
    setNewSkillForm({ name: '' });
  };

  // 关闭添加技能弹窗
  const handleCloseAddSkillModal = () => {
    setIsAddSkillModalOpen(false);
    setNewSkillForm({ name: '' });
  };

  // 技能表单处理函数
  const handleNewSkillFormChange = (field: string, value: string) => {
    setNewSkillForm((prev) => ({ ...prev, [field]: value }));
  };

  // 保存新技能
  const handleSaveNewSkill = async () => {
    if (!newSkillForm.name.trim()) {
      setSkillError('技能名称不能为空');
      return;
    }

    try {
      setSkillSubmitting(true);
      setSkillError(null);
      console.log('💾 创建新技能:', newSkillForm);
      await createJobSkillMeta(newSkillForm);
      handleCloseAddSkillModal();

      console.log('🔄 新技能创建成功，重置到第一页并刷新数据');
      setSkillCurrentPage(1);
      await fetchSkills(1);
    } catch (err) {
      setSkillError('创建技能失败，请重试');
      console.error('创建技能失败:', err);
    } finally {
      setSkillSubmitting(false);
    }
  };

  // 删除技能
  const handleDeleteSkill = async (skillId: string) => {
    const skill = skills.find((s) => s.id === skillId);
    if (!skill) return;

    setSkillToDelete({ id: skillId, name: skill.name });
    setIsSkillDeleteConfirmOpen(true);
  };

  // 确认删除技能
  const handleConfirmDeleteSkill = async () => {
    if (!skillToDelete) return;

    try {
      setSkillDeletingId(skillToDelete.id);
      setSkillError(null);
      console.log('🗑️ 删除技能:', skillToDelete);
      await deleteJobSkillMeta(skillToDelete.id);

      const remainingItems = skillTotalCount - 1;
      const maxPage = Math.ceil(remainingItems / skillPageSize);
      const targetPage =
        skillCurrentPage > maxPage ? Math.max(1, maxPage) : skillCurrentPage;

      console.log(
        `🔄 技能删除成功，剩余 ${remainingItems} 条数据，当前页: ${skillCurrentPage}, 目标页: ${targetPage}`
      );

      if (targetPage !== skillCurrentPage) {
        setSkillCurrentPage(targetPage);
      }

      await fetchSkills(targetPage);
      setIsSkillDeleteConfirmOpen(false);
      setSkillToDelete(null);
    } catch (err) {
      setSkillError('删除技能失败，请重试');
      console.error('删除技能失败:', err);
    } finally {
      setSkillDeletingId(null);
    }
  };

  // 技能页面切换
  const handleSkillPageChange = (page: number) => {
    setSkillCurrentPage(page);
  };

  // 生成技能页码数组
  const generateSkillPageNumbers = () => {
    const pages = [];
    const safeTotalPages = skillTotalPages || 1;
    const safeCurrentPage = skillCurrentPage || 1;

    const start = Math.max(1, safeCurrentPage - 2);
    const end = Math.min(safeTotalPages, safeCurrentPage + 2);

    for (let i = start; i <= end; i++) {
      pages.push(i);
    }

    return pages;
  };

  // ==================== 简历收集配置相关函数 ====================

  // 打开简历收集配置弹窗
  const handleOpenResumeCollectionModal = () => {
    console.log('🔧 打开简历收集配置弹窗...');
    setIsResumeCollectionModalOpen(true);
  };

  // 关闭简历收集配置弹窗
  const handleCloseResumeCollectionModal = () => {
    setIsResumeCollectionModalOpen(false);
  };

  // 初始化数据加载
  useEffect(() => {
    fetchDepartments();
  }, [currentPage, fetchDepartments]);

  // 技能页面切换时获取数据
  useEffect(() => {
    if (isSkillModalOpen) {
      fetchSkills();
    }
  }, [skillCurrentPage, fetchSkills, isSkillModalOpen]);

  // 检测URL参数，自动打开弹窗
  useEffect(() => {
    const openModal = searchParams.get('openModal');
    const openSkillModal = searchParams.get('openSkillModal');
    const openResumeCollectionModal = searchParams.get(
      'openResumeCollectionModal'
    );

    if (openModal === 'true') {
      setIsModalOpen(true);
    }

    if (openSkillModal === 'true') {
      handleOpenSkillModal();
    }

    if (openResumeCollectionModal === 'true') {
      handleOpenResumeCollectionModal();
    }
  }, [searchParams, handleOpenSkillModal]);

  return (
    <div className="flex h-full flex-col gap-6 px-6 pb-6 pt-6 page-content">
      {/* 页面标题 */}
      <div className="flex items-center justify-between rounded-xl border border-[#E5E7EB] bg-white px-6 py-5 shadow-sm transition-all duration-200 hover:shadow-md">
        <h1 className="text-2xl font-semibold text-gray-900">基础功能配置</h1>
      </div>

      {/* 所属部门配置卡片 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white shadow-sm overflow-hidden card-hover">
        {/* 卡片头部 */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
          <div className="flex items-center gap-4">
            {/* 图标容器 */}
            <div className="flex items-center justify-center w-8 h-8 rounded-md bg-emerald-50">
              <Building2 className="w-4 h-4 text-emerald-500" />
            </div>

            {/* 标题和描述 */}
            <div className="flex flex-col">
              <h3 className="text-sm font-medium text-gray-900">
                所属部门配置
              </h3>
              <p className="text-xs text-gray-500 mt-0.5">
                管理岗位画像关联的部门信息
              </p>
            </div>
          </div>

          {/* 配置按钮 */}
          <Button
            className="gap-2 rounded-lg px-5 py-2 shadow-sm"
            onClick={handleOpenModal}
          >
            <Settings className="w-3.5 h-3.5" />
            配置
          </Button>
        </div>

        {/* 卡片底部状态信息 */}
        <div className="px-6 py-5">
          {loading ? (
            <div className="flex items-center gap-2">
              <Loader2 className="w-4 h-4 animate-spin text-gray-400" />
              <p className="text-sm text-gray-500">加载中...</p>
            </div>
          ) : error ? (
            <div className="flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-red-500" />
              <p className="text-sm text-red-500">{error}</p>
              <button
                onClick={() => fetchDepartments()}
                className="text-sm text-blue-500 hover:text-blue-600 ml-2"
              >
                重试
              </button>
            </div>
          ) : null}
        </div>
      </div>

      {/* 工作技能配置卡片 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white shadow-sm overflow-hidden card-hover">
        {/* 卡片头部 */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
          <div className="flex items-center gap-4">
            {/* 图标容器 */}
            <div className="flex items-center justify-center w-8 h-8 rounded-md bg-blue-50">
              <Wrench className="w-4 h-4 text-blue-500" />
            </div>

            {/* 标题和描述 */}
            <div className="flex flex-col">
              <h3 className="text-sm font-medium text-gray-900">
                工作技能配置
              </h3>
              <p className="text-xs text-gray-500 mt-0.5">
                管理岗位画像关联的技能信息
              </p>
            </div>
          </div>

          {/* 配置按钮 */}
          <Button
            className="gap-2 rounded-lg px-5 py-2 shadow-sm"
            onClick={handleOpenSkillModal}
          >
            <Settings className="w-3.5 h-3.5" />
            配置
          </Button>
        </div>

        {/* 卡片底部状态信息 */}
        <div className="px-6 py-5">
          {skillLoading ? (
            <div className="flex items-center gap-2">
              <Loader2 className="w-4 h-4 animate-spin text-gray-400" />
              <p className="text-sm text-gray-500">加载中...</p>
            </div>
          ) : skillError ? (
            <div className="flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-red-500" />
              <p className="text-sm text-red-500">{skillError}</p>
              <button
                onClick={() => fetchSkills()}
                className="text-sm text-blue-500 hover:text-blue-600 ml-2"
              >
                重试
              </button>
            </div>
          ) : null}
        </div>
      </div>

      {/* 简历收集配置卡片 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white shadow-sm overflow-hidden card-hover">
        {/* 卡片头部 */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
          <div className="flex items-center gap-4">
            {/* 图标容器 */}
            <div className="flex items-center justify-center w-8 h-8 rounded-md bg-purple-50">
              <FileText className="w-4 h-4 text-purple-500" />
            </div>

            {/* 标题和描述 */}
            <div className="flex flex-col">
              <h3 className="text-sm font-medium text-gray-900">
                简历收集配置
              </h3>
              <p className="text-xs text-gray-500 mt-0.5">
                管理简历收集渠道和链接
              </p>
            </div>
          </div>

          {/* 配置按钮 */}
          <Button
            className="gap-2 rounded-lg px-5 py-2 shadow-sm"
            onClick={handleOpenResumeCollectionModal}
          >
            <Settings className="w-3.5 h-3.5" />
            配置
          </Button>
        </div>

        {/* 卡片底部状态信息 */}
        <div className="px-6 py-5">
          {/* 此卡片目前无加载/错误状态，预留占位以统一布局 */}
        </div>
      </div>

      {/* 通知配置卡片 */}
      <div className="rounded-xl border border-[#E5E7EB] bg-white shadow-sm overflow-hidden card-hover">
        {/* 卡片头部 */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
          <div className="flex items-center gap-4">
            {/* 图标容器 */}
            <div className="flex items-center justify-center w-8 h-8 rounded-md bg-orange-50">
              <Bell className="w-4 h-4 text-orange-500" />
            </div>

            {/* 标题和描述 */}
            <div className="flex flex-col">
              <h3 className="text-sm font-medium text-gray-900">通知配置</h3>
              <p className="text-xs text-gray-500 mt-0.5">
                管理系统通知和告警配置
              </p>
            </div>
          </div>

          {/* 配置按钮 */}
          <Button
            className="gap-2 rounded-lg px-5 py-2 shadow-sm"
            onClick={() => setIsNotificationConfigModalOpen(true)}
          >
            <Settings className="w-3.5 h-3.5" />
            配置
          </Button>
        </div>

        {/* 卡片底部状态信息 */}
        <div className="px-6 py-5">
          {/* 此卡片目前无加载/错误状态，预留占位以统一布局 */}
        </div>
      </div>

      {/* 所属部门创建弹窗 - 严格按照Figma设计图实现 */}
      {isModalOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 p-4"
          onClick={handleCloseModal}
        >
          <div
            className="bg-white rounded-lg shadow-xl w-[800px] max-h-[90vh] overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            {/* 弹窗头部 */}
            <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
              <h2 className="text-lg font-medium text-gray-900">
                所属部门创建
              </h2>
              <button
                onClick={handleCloseModal}
                className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
              >
                <X className="w-6 h-6" />
              </button>
            </div>

            {/* 弹窗内容 */}
            <div className="p-6">
              {/* 错误提示 */}
              {error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 text-red-500" />
                  <span className="text-sm text-red-600">{error}</span>
                  <button
                    onClick={() => setError(null)}
                    className="ml-auto text-red-500 hover:text-red-600"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              )}

              {/* 添加部门按钮 */}
              <div className="mb-6">
                <Button
                  className="gap-2"
                  onClick={handleOpenAddDeptModal}
                  disabled={loading}
                >
                  <Plus className="w-4 h-4" />
                  添加部门
                </Button>
              </div>

              {/* 表格 */}
              <div className="border border-gray-200 rounded-lg overflow-hidden">
                {/* 表头 */}
                <div className="bg-gray-50 px-6 py-3 border-b border-gray-200">
                  <div className="flex items-center w-full">
                    <div className="flex items-center flex-1 min-w-0">
                      <span className="text-sm font-medium text-gray-700">
                        部门名称
                      </span>
                    </div>
                    <div className="flex items-center flex-1 min-w-0">
                      <span className="text-sm font-medium text-gray-700">
                        描述
                      </span>
                    </div>
                    <div className="flex items-center w-32">
                      <span className="text-sm font-medium text-gray-700">
                        创建时间
                      </span>
                    </div>
                    <div className="flex items-center justify-end w-20">
                      <span className="text-sm font-medium text-gray-700">
                        操作
                      </span>
                    </div>
                  </div>
                </div>

                {/* 表格内容 */}
                <div className="bg-white">
                  {loading ? (
                    <div className="flex items-center justify-center py-8">
                      <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
                      <span className="ml-2 text-sm text-gray-500">
                        加载中...
                      </span>
                    </div>
                  ) : departments.length === 0 ? (
                    <div className="flex items-center justify-center py-8">
                      <span className="text-sm text-gray-500">
                        暂无部门数据
                      </span>
                    </div>
                  ) : (
                    departments.map((dept, index) => (
                      <div
                        key={dept.id}
                        className={`flex items-center px-6 py-4 w-full ${index !== departments.length - 1 ? 'border-b border-gray-100' : ''}`}
                      >
                        {/* 部门名称 */}
                        <div className="flex items-center flex-1 min-w-0 pr-4">
                          <span className="text-sm text-gray-900 truncate">
                            {dept.name}
                          </span>
                        </div>

                        {/* 描述 */}
                        <div className="flex items-center flex-1 min-w-0 pr-4">
                          <span className="text-sm text-gray-500 truncate">
                            {dept.description || '-'}
                          </span>
                        </div>

                        {/* 创建时间 */}
                        <div className="flex items-center w-32 pr-4">
                          <span className="text-sm text-gray-500">
                            {new Date(
                              dept.created_at * 1000
                            ).toLocaleDateString()}
                          </span>
                        </div>

                        {/* 操作 */}
                        <div className="flex items-center justify-end gap-2 w-20">
                          <button
                            className="text-[#6B7280] hover:text-[#374151] transition-colors"
                            onClick={() => handleOpenEditDeptModal(dept)}
                            disabled={submitting || deletingId === dept.id}
                          >
                            <Edit className="w-4 h-4" />
                          </button>
                          <button
                            className="text-[#6B7280] hover:text-[#EF4444] transition-colors"
                            onClick={() => handleDeleteDept(dept.id)}
                            disabled={submitting || deletingId === dept.id}
                          >
                            {deletingId === dept.id ? (
                              <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                              <Trash2 className="w-4 h-4" />
                            )}
                          </button>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>

              {/* 分页区域 - 与简历管理页面保持一致 */}
              {!loading && (
                <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between mt-6">
                  <div className="flex items-center gap-4">
                    <div className="text-sm text-[#6B7280]">
                      显示{' '}
                      <span className="text-[#6B7280]">
                        {(totalCount || 0) > 0
                          ? ((currentPage || 1) - 1) * pageSize + 1
                          : 0}
                      </span>{' '}
                      到{' '}
                      <span className="text-[#6B7280]">
                        {(totalCount || 0) > 0
                          ? Math.min(
                              (currentPage || 1) * pageSize,
                              totalCount || 0
                            )
                          : 0}
                      </span>{' '}
                      条，共{' '}
                      <span className="text-[#6B7280]">{totalCount || 0}</span>{' '}
                      条结果
                    </div>
                  </div>

                  {/* 分页按钮 - 在有数据且多页时显示 */}
                  {(totalCount || 0) > 0 && (totalPages || 0) > 1 && (
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
                            page === currentPage &&
                              'border-[#7bb8ff] bg-[#7bb8ff] text-white'
                          )}
                        >
                          {page}
                        </Button>
                      ))}

                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handlePageChange(currentPage + 1)}
                        disabled={currentPage >= (totalPages || 0)}
                        className={cn(
                          'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                          currentPage >= (totalPages || 0) && 'opacity-50'
                        )}
                      >
                        <ChevronRight className="h-4 w-4 text-[#6B7280]" />
                      </Button>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 添加部门弹窗 */}
      {isAddDeptModalOpen && (
        <div
          className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1001]"
          onClick={handleCloseAddDeptModal}
        >
          <div
            className="bg-white rounded-lg shadow-xl w-[600px]"
            onClick={(e) => e.stopPropagation()}
          >
            {/* 弹窗头部 */}
            <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
              <h2 className="text-lg font-medium text-gray-900">添加部门</h2>
              <button
                onClick={handleCloseAddDeptModal}
                className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
              >
                <X className="w-6 h-6" />
              </button>
            </div>

            {/* 弹窗内容 */}
            <div className="p-6">
              {/* 部门名称输入框 */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  部门名称 <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={newDeptForm.name}
                  onChange={(e) => handleNewFormChange('name', e.target.value)}
                  placeholder="请输入部门名称，最多50个字符"
                  maxLength={50}
                  className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                />
              </div>

              {/* 部门描述 */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  部门描述
                </label>
                <textarea
                  value={newDeptForm.description}
                  onChange={(e) =>
                    handleNewFormChange('description', e.target.value)
                  }
                  placeholder="请输入部门描述，最多200个字符"
                  maxLength={200}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent resize-none"
                />
              </div>
            </div>

            {/* 弹窗底部 */}
            <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100">
              <Button
                variant="outline"
                className="px-6"
                onClick={handleCloseAddDeptModal}
              >
                取消
              </Button>
              <Button
                className="px-6"
                onClick={handleSaveNewDept}
                disabled={submitting}
              >
                {submitting ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin mr-2" />
                    保存中...
                  </>
                ) : (
                  '保存部门'
                )}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* 编辑部门弹窗 */}
      {isEditDeptModalOpen && (
        <div
          className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1001]"
          onClick={handleCloseEditDeptModal}
        >
          <div
            className="bg-white rounded-lg shadow-xl w-[600px]"
            onClick={(e) => e.stopPropagation()}
          >
            {/* 弹窗头部 */}
            <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
              <h2 className="text-lg font-medium text-gray-900">编辑部门</h2>
              <button
                onClick={handleCloseEditDeptModal}
                className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
              >
                <X className="w-6 h-6" />
              </button>
            </div>

            {/* 弹窗内容 */}
            <div className="p-6">
              {/* 部门名称 */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  部门名称 <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={editDeptForm.name}
                  onChange={(e) => handleEditFormChange('name', e.target.value)}
                  placeholder="请输入部门名称，最多50个字符"
                  maxLength={50}
                  className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                />
              </div>

              {/* 部门描述 */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  部门描述
                </label>
                <textarea
                  value={editDeptForm.description}
                  onChange={(e) =>
                    handleEditFormChange('description', e.target.value)
                  }
                  placeholder="请输入部门描述，最多200个字符"
                  maxLength={200}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent resize-none"
                />
              </div>
            </div>

            {/* 弹窗底部 */}
            <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100">
              <Button
                variant="outline"
                className="px-6"
                onClick={handleCloseEditDeptModal}
              >
                取消
              </Button>
              <Button
                className="px-6"
                onClick={handleSaveEditDept}
                disabled={submitting}
              >
                {submitting ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin mr-2" />
                    保存中...
                  </>
                ) : (
                  '保存修改'
                )}
              </Button>
            </div>
          </div>
        </div>
      )}
      {/* 技能配置弹窗 */}
      {isSkillModalOpen && (
        <div
          className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1000]"
          onClick={handleCloseSkillModal}
        >
          <div
            className="bg-white rounded-lg shadow-xl w-[800px] max-h-[90vh] overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            {/* 弹窗头部 */}
            <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
              <h2 className="text-lg font-medium text-gray-900">
                工作技能配置
              </h2>
              <button
                onClick={handleCloseSkillModal}
                className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
              >
                <X className="w-6 h-6" />
              </button>
            </div>

            {/* 弹窗内容 */}
            <div className="p-6 overflow-y-auto max-h-[calc(90vh-140px)]">
              {/* 错误提示 */}
              {skillError && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
                  <div className="flex items-center">
                    <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
                    <span className="text-sm text-red-700">{skillError}</span>
                  </div>
                </div>
              )}

              {/* 添加技能按钮 */}
              <div className="mb-6">
                <Button
                  className="gap-2"
                  onClick={handleOpenAddSkillModal}
                  disabled={skillSubmitting}
                >
                  <Plus className="w-4 h-4" />
                  添加技能
                </Button>
              </div>

              {/* 技能列表表格 */}
              <div className="border border-gray-200 rounded-lg overflow-hidden">
                {/* 表格头部 */}
                <div className="bg-gray-50 border-b border-gray-200">
                  <div className="flex items-center px-6 py-3 w-full">
                    <div className="flex items-center flex-1 min-w-0 pr-4">
                      <span className="text-sm font-medium text-gray-700">
                        技能名称
                      </span>
                    </div>
                    <div className="flex items-center w-32 pr-4">
                      <span className="text-sm font-medium text-gray-700">
                        创建时间
                      </span>
                    </div>
                    <div className="flex items-center justify-center w-20">
                      <span className="text-sm font-medium text-gray-700">
                        操作
                      </span>
                    </div>
                  </div>
                </div>

                {/* 表格内容 */}
                <div className="bg-white">
                  {skillLoading ? (
                    <div className="flex items-center justify-center py-8">
                      <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
                      <span className="ml-2 text-sm text-gray-500">
                        加载中...
                      </span>
                    </div>
                  ) : skills.length === 0 ? (
                    <div className="flex items-center justify-center py-8">
                      <span className="text-sm text-gray-500">
                        暂无技能数据
                      </span>
                    </div>
                  ) : (
                    skills.map((skill, index) => (
                      <div
                        key={skill.id}
                        className={`flex items-center px-6 py-4 w-full ${index !== skills.length - 1 ? 'border-b border-gray-100' : ''}`}
                      >
                        {/* 技能名称 */}
                        <div className="flex items-center flex-1 min-w-0 pr-4">
                          <span className="text-sm text-gray-900 truncate">
                            {skill.name}
                          </span>
                        </div>

                        {/* 创建时间 */}
                        <div className="flex items-center w-32 pr-4">
                          <span className="text-sm text-gray-500">
                            {new Date(
                              skill.created_at * 1000
                            ).toLocaleDateString()}
                          </span>
                        </div>

                        {/* 操作 */}
                        <div className="flex items-center justify-center gap-2 w-20">
                          <button
                            className="text-[#6B7280] hover:text-[#EF4444] transition-colors"
                            onClick={() => handleDeleteSkill(skill.id)}
                            disabled={
                              skillSubmitting || skillDeletingId === skill.id
                            }
                          >
                            {skillDeletingId === skill.id ? (
                              <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                              <Trash2 className="w-4 h-4" />
                            )}
                          </button>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>

              {/* 分页区域 */}
              {!skillLoading && (
                <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between mt-6">
                  <div className="flex items-center gap-4">
                    <div className="text-sm text-[#6B7280]">
                      显示{' '}
                      <span className="text-[#6B7280]">
                        {(skillTotalCount || 0) > 0
                          ? ((skillCurrentPage || 1) - 1) * skillPageSize + 1
                          : 0}
                      </span>{' '}
                      到{' '}
                      <span className="text-[#6B7280]">
                        {(skillTotalCount || 0) > 0
                          ? Math.min(
                              (skillCurrentPage || 1) * skillPageSize,
                              skillTotalCount || 0
                            )
                          : 0}
                      </span>{' '}
                      条，共{' '}
                      <span className="text-[#6B7280]">
                        {skillTotalCount || 0}
                      </span>{' '}
                      条结果
                    </div>
                  </div>

                  {/* 分页按钮 */}
                  {(skillTotalCount || 0) > 0 && (skillTotalPages || 0) > 1 && (
                    <div className="flex flex-wrap items-center gap-1">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          handleSkillPageChange(skillCurrentPage - 1)
                        }
                        disabled={skillCurrentPage === 1}
                        className={cn(
                          'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                          skillCurrentPage === 1 && 'opacity-50'
                        )}
                      >
                        <ChevronLeft className="h-4 w-4 text-[#6B7280]" />
                      </Button>

                      {generateSkillPageNumbers().map((page) => (
                        <Button
                          key={page}
                          variant="outline"
                          size="sm"
                          onClick={() => handleSkillPageChange(page)}
                          className={cn(
                            'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0 text-sm font-normal text-[#374151]',
                            page === skillCurrentPage &&
                              'border-[#7bb8ff] bg-[#7bb8ff] text-white'
                          )}
                        >
                          {page}
                        </Button>
                      ))}

                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          handleSkillPageChange(skillCurrentPage + 1)
                        }
                        disabled={skillCurrentPage >= (skillTotalPages || 0)}
                        className={cn(
                          'h-[34px] w-[34px] rounded border border-[#D1D5DB] bg-white p-0',
                          skillCurrentPage >= (skillTotalPages || 0) &&
                            'opacity-50'
                        )}
                      >
                        <ChevronRight className="h-4 w-4 text-[#6B7280]" />
                      </Button>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 添加技能弹窗 */}
      {isAddSkillModalOpen && (
        <div
          className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-[1001]"
          onClick={handleCloseAddSkillModal}
        >
          <div
            className="bg-white rounded-lg shadow-xl w-[600px]"
            onClick={(e) => e.stopPropagation()}
          >
            {/* 弹窗头部 */}
            <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
              <h2 className="text-lg font-medium text-gray-900">添加技能</h2>
              <button
                onClick={handleCloseAddSkillModal}
                className="text-gray-400 hover:text-gray-600 transition-colors w-6 h-6"
              >
                <X className="w-6 h-6" />
              </button>
            </div>

            {/* 弹窗内容 */}
            <div className="p-6">
              {/* 技能名称输入框 */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  技能名称 <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={newSkillForm.name}
                  onChange={(e) =>
                    handleNewSkillFormChange('name', e.target.value)
                  }
                  placeholder="请输入技能名称，最多50个字符"
                  maxLength={50}
                  className="w-full h-10 px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            {/* 弹窗底部 */}
            <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100">
              <Button
                variant="outline"
                className="px-6"
                onClick={handleCloseAddSkillModal}
              >
                取消
              </Button>
              <Button
                className="px-6"
                onClick={handleSaveNewSkill}
                disabled={skillSubmitting}
              >
                {skillSubmitting ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin mr-2" />
                    保存中...
                  </>
                ) : (
                  '确认'
                )}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* 删除确认弹窗 */}
      <ConfirmDialog
        open={isDeleteConfirmOpen}
        onOpenChange={setIsDeleteConfirmOpen}
        title="确认删除"
        description={
          departmentToDelete ? (
            <span>
              确定要删除 <strong>{departmentToDelete.name}</strong>{' '}
              部门吗？此操作无法撤销。
            </span>
          ) : (
            ''
          )
        }
        confirmText="删除"
        cancelText="取消"
        onConfirm={handleConfirmDelete}
        variant="destructive"
        loading={
          departmentToDelete ? deletingId === departmentToDelete.id : false
        }
      />

      {/* 技能删除确认弹窗 */}
      <ConfirmDialog
        open={isSkillDeleteConfirmOpen}
        onOpenChange={setIsSkillDeleteConfirmOpen}
        title="确认删除"
        description={
          skillToDelete ? (
            <span>
              确定要删除 <strong>{skillToDelete.name}</strong>{' '}
              技能吗？此操作无法撤销。
            </span>
          ) : (
            ''
          )
        }
        confirmText="删除"
        cancelText="取消"
        onConfirm={handleConfirmDeleteSkill}
        variant="destructive"
        loading={skillToDelete ? skillDeletingId === skillToDelete.id : false}
        zIndex="z-[1002]"
      />

      {/* 简历收集配置弹窗 - 使用新的模态框组件 */}
      <ResumeCollectionModal
        isOpen={isResumeCollectionModalOpen}
        onClose={handleCloseResumeCollectionModal}
      />

      {/* 通知配置弹窗 */}
      <NotificationConfigModal
        open={isNotificationConfigModalOpen}
        onClose={() => setIsNotificationConfigModalOpen(false)}
      />
    </div>
  );
}
