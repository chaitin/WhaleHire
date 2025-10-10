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
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  listJobProfiles,
  createJobProfile,
  listJobSkillMeta,
} from '@/services/job-profile';
import type {
  JobProfile,
  CreateJobProfileReq,
  JobSkillMeta,
} from '@/types/job-profile';
import { AddSkillModal } from '@/components/modals/add-skill-modal';

export function JobProfilePage() {
  const [jobProfiles, setJobProfiles] = useState<JobProfile[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [departmentFilter, setDepartmentFilter] = useState<string>('all');
  const [searchKeyword, setSearchKeyword] = useState<string>('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // 获取岗位画像列表
  const fetchJobProfiles = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await listJobProfiles({
        page: currentPage,
        page_size: pageSize,
        keyword: searchKeyword || undefined,
        department_id:
          departmentFilter !== 'all' ? departmentFilter : undefined,
      });
      setJobProfiles(response.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取岗位画像列表失败');
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, searchKeyword, departmentFilter]);

  // 页面加载时获取数据
  useEffect(() => {
    fetchJobProfiles();
  }, [currentPage, pageSize, fetchJobProfiles]);

  // 新建岗位弹窗状态管理
  const [isCreateJobModalOpen, setIsCreateJobModalOpen] = useState(false);

  // 技能相关状态管理
  const [availableSkills, setAvailableSkills] = useState<JobSkillMeta[]>([]);
  const [isAddSkillModalOpen, setIsAddSkillModalOpen] = useState(false);
  const [skillsLoading, setSkillsLoading] = useState(false);
  const [skillsError, setSkillsError] = useState<string | null>(null);

  // 部门相关状态管理
  const [availableDepartments, setAvailableDepartments] = useState<string[]>([
    '技术部',
    '产品部',
    '设计部',
    '数据部',
    '市场部',
    '运营部',
  ]);
  const [departmentsLoading, setDepartmentsLoading] = useState(false);
  const [departmentsError, setDepartmentsError] = useState<string | null>(null);

  const [editMode, setEditMode] = useState<'manual' | 'ai'>('manual');
  const [jobFormData, setJobFormData] = useState({
    name: '',
    location: '',
    department: '',
    workType: '',
    education: '',
    salaryRange: '',
    experience: '',
    industry: '',
    skills: '',
    responsibilities: [''], // 改为数组类型，默认包含一个空字符串
  });

  // 统计数据
  const totalJobs = jobProfiles.length;
  const publishedJobs = jobProfiles.filter(
    (job) => job.status === 'published'
  ).length;
  const draftJobs = jobProfiles.filter((job) => job.status === 'draft').length;

  // 筛选后的数据
  const filteredJobs = jobProfiles.filter((job) => {
    const statusMatch = statusFilter === 'all' || job.status === statusFilter;
    const departmentMatch =
      departmentFilter === 'all' || job.department === departmentFilter;
    const keywordMatch =
      !searchKeyword ||
      job.name.toLowerCase().includes(searchKeyword.toLowerCase()) ||
      job.department.toLowerCase().includes(searchKeyword.toLowerCase()) ||
      (job.location &&
        job.location.toLowerCase().includes(searchKeyword.toLowerCase())) ||
      (job.creator_name &&
        job.creator_name.toLowerCase().includes(searchKeyword.toLowerCase()));

    return statusMatch && departmentMatch && keywordMatch;
  });

  // 分页数据
  const totalPages = Math.ceil(filteredJobs.length / pageSize);
  const startIndex = (currentPage - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const currentJobs = filteredJobs.slice(startIndex, endIndex);

  // 处理搜索
  const handleSearch = () => {
    setCurrentPage(1); // 搜索时重置到第一页
    fetchJobProfiles(); // 重新获取数据
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

  // 获取技能列表
  const fetchSkills = async () => {
    try {
      setSkillsLoading(true);
      setSkillsError(null);
      const response = await listJobSkillMeta();
      setAvailableSkills(response.skills || []);
    } catch (err) {
      console.error('获取技能列表失败:', err);
      setSkillsError(err instanceof Error ? err.message : '获取技能列表失败');
      // 即使失败也设置为空数组，确保下拉框能正常显示
      setAvailableSkills([]);
    } finally {
      setSkillsLoading(false);
    }
  };

  // 重试获取技能列表
  const retryFetchSkills = async () => {
    await fetchSkills();
  };

  // 处理技能添加成功
  const handleSkillAdded = (newSkill: JobSkillMeta) => {
    setAvailableSkills((prev) => [...prev, newSkill]);
    // 清除错误状态
    setSkillsError(null);
  };

  // 获取部门列表（目前使用硬编码数据，可以后续扩展为API调用）
  const fetchDepartments = async () => {
    try {
      setDepartmentsLoading(true);
      setDepartmentsError(null);
      // 模拟API调用延迟
      await new Promise((resolve) => setTimeout(resolve, 100));
      // 这里可以替换为真实的API调用
      // const response = await listDepartments();
      // setAvailableDepartments(response.departments || []);
      setAvailableDepartments([
        '技术部',
        '产品部',
        '设计部',
        '数据部',
        '市场部',
        '运营部',
      ]);
    } catch (err) {
      console.error('获取部门列表失败:', err);
      setDepartmentsError(
        err instanceof Error ? err.message : '获取部门列表失败'
      );
      // 即使失败也设置为空数组，确保下拉框能正常显示
      setAvailableDepartments([]);
    } finally {
      setDepartmentsLoading(false);
    }
  };

  // 重试获取部门列表
  const retryFetchDepartments = async () => {
    await fetchDepartments();
  };

  // 打开新建岗位弹窗
  const handleOpenCreateJobModal = () => {
    // 先打开弹窗，确保用户能看到界面
    setIsCreateJobModalOpen(true);
    // 然后异步获取数据，不阻塞弹窗显示
    fetchSkills();
    fetchDepartments();
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
      salaryRange: '',
      experience: '',
      industry: '',
      skills: '',
      responsibilities: [''], // 重置为包含一个空字符串的数组
    });
    setEditMode('manual');
  };

  // 处理表单数据变化
  const handleFormDataChange = (field: string, value: string) => {
    setJobFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
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
      const createData: CreateJobProfileReq = {
        name: jobFormData.name,
        department_id: jobFormData.department,
        description: jobFormData.responsibilities
          .filter((r) => r.trim())
          .join('\n'), // 将数组合并为字符串
        location: jobFormData.location,
        status: 'draft',
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
      const createData: CreateJobProfileReq = {
        name: jobFormData.name,
        department_id: jobFormData.department,
        description: jobFormData.responsibilities
          .filter((r) => r.trim())
          .join('\n'), // 将数组合并为字符串
        location: jobFormData.location,
        status: 'published',
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
    <div className="flex h-full flex-col gap-6 px-6 pb-6 pt-6">
      {/* 页面标题和操作区域 */}
      <div className="flex items-center justify-between rounded-xl border border-[#E5E7EB] bg-white px-6 py-5 shadow-sm">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">岗位画像</h1>
          <p className="text-sm text-muted-foreground">
            管理和创建岗位画像，提升招聘效率
          </p>
        </div>
        <Button
          className="gap-2 rounded-lg px-4 py-2 shadow-sm"
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
        <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">总岗位数</p>
              <p className="text-2xl font-bold text-blue-600 mt-1">
                {totalJobs}
              </p>
            </div>
            <div className="rounded-lg bg-blue-50 p-3">
              <BarChart3 className="h-5 w-5 text-blue-600" />
            </div>
          </div>
        </div>

        {/* 已发布 */}
        <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">已发布</p>
              <p className="text-2xl font-bold text-green-600 mt-1">
                {publishedJobs}
              </p>
            </div>
            <div className="rounded-lg bg-green-50 p-3">
              <Users className="h-5 w-5 text-green-600" />
            </div>
          </div>
        </div>

        {/* 草稿状态 */}
        <div className="rounded-xl border border-[#E5E7EB] bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">草稿状态</p>
              <p className="text-2xl font-bold text-yellow-600 mt-1">
                {draftJobs}
              </p>
            </div>
            <div className="rounded-lg bg-yellow-50 p-3">
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
            <Select value={statusFilter} onValueChange={setStatusFilter}>
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
              onValueChange={setDepartmentFilter}
            >
              <SelectTrigger className="w-32">
                <SelectValue placeholder="所属部门" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">所属部门</SelectItem>
                <SelectItem value="技术部">技术部</SelectItem>
                <SelectItem value="产品部">产品部</SelectItem>
                <SelectItem value="设计部">设计部</SelectItem>
                <SelectItem value="数据部">数据部</SelectItem>
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
            <Button onClick={handleSearch} className="gap-2">
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
            岗位列表 ({filteredJobs.length})
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
                    <div className="flex flex-col items-center justify-center text-center">
                      {/* 空数据图标 */}
                      <div className="w-20 h-20 mb-6 rounded-full bg-gray-100 flex items-center justify-center">
                        <FileText className="w-10 h-10 text-gray-400" />
                      </div>

                      {/* 主要提示文字 */}
                      <h3 className="text-lg font-medium text-gray-900 mb-2">
                        暂无岗位画像数据
                      </h3>

                      {/* 描述文字 */}
                      {(searchKeyword || departmentFilter !== 'all') && (
                        <p className="text-sm text-gray-500 mb-6 max-w-sm">
                          没有找到符合条件的岗位画像，请尝试调整搜索条件
                        </p>
                      )}

                      {/* 搜索状态下的清除按钮 */}
                      {(searchKeyword || departmentFilter !== 'all') && (
                        <Button
                          variant="outline"
                          onClick={() => {
                            setSearchKeyword('');
                            setDepartmentFilter('all');
                            setCurrentPage(1);
                            fetchJobProfiles();
                          }}
                          className="border-gray-300 text-gray-700 hover:bg-gray-50"
                        >
                          <X className="w-4 h-4 mr-2" />
                          清除筛选条件
                        </Button>
                      )}
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
                          className="h-8 w-8 p-0 text-gray-400 hover:text-gray-600"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0 text-gray-400 hover:text-red-600"
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
                显示
                <span className="text-[#6B7280]">
                  {filteredJobs.length > 0 ? startIndex + 1 : 0}
                </span>{' '}
                到{' '}
                <span className="text-[#6B7280]">
                  {filteredJobs.length > 0
                    ? Math.min(endIndex, filteredJobs.length)
                    : 0}
                </span>{' '}
                条，共{' '}
                <span className="text-[#6B7280]">{filteredJobs.length}</span>{' '}
                条结果
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
                      page === currentPage &&
                        'border-[#10B981] bg-[#10B981] text-white'
                    )}
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

      {/* 新建岗位弹窗 */}
      <Dialog
        open={isCreateJobModalOpen}
        onOpenChange={setIsCreateJobModalOpen}
      >
        <DialogContent className="max-w-[800px] max-h-[90vh] overflow-y-auto p-0 gap-0">
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
                <div className="w-4 h-4 bg-[#22C55E] rounded-sm flex items-center justify-center">
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
                      ? 'border-[#22C55E] bg-[#F0FDF4]'
                      : 'border-[#E5E7EB] bg-white hover:bg-[#F9FAFB]'
                  )}
                  onClick={() => setEditMode('manual')}
                >
                  <div className="flex items-center justify-center w-4 h-4 mt-0.5">
                    <div
                      className={cn(
                        'w-4 h-4 rounded-full border-2 flex items-center justify-center',
                        editMode === 'manual'
                          ? 'border-[#22C55E]'
                          : 'border-[#D1D5DB]'
                      )}
                    >
                      {editMode === 'manual' && (
                        <div className="w-2 h-2 bg-[#22C55E] rounded-full"></div>
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
                    'flex items-start gap-3 p-4 border rounded-lg transition-colors cursor-not-allowed opacity-50',
                    'border-[#E5E7EB] bg-[#F9FAFB]'
                  )}
                >
                  <div className="flex items-center justify-center w-4 h-4 mt-0.5">
                    <div className="w-4 h-4 rounded-full border-2 border-[#D1D5DB] flex items-center justify-center"></div>
                  </div>
                  <div className="flex-1">
                    <div className="text-[14px] font-medium text-[#9CA3AF] mb-1">
                      AI编辑模式
                    </div>
                    <div className="text-[12px] text-[#9CA3AF]">
                      智能生成岗位内容，快速高效
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* 基本信息表单 */}
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
                        className="h-8 px-3 text-[12px] text-[#22C55E] border-[#22C55E] hover:bg-[#F0FDF4] hover:text-[#16A34A] hover:border-[#16A34A]"
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
                    <SelectContent>
                      {departmentsLoading ? (
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
                        availableDepartments.map((department) => (
                          <SelectItem key={department} value={department}>
                            {department}
                          </SelectItem>
                        ))
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
                  <Label
                    htmlFor="salaryRange"
                    className="text-[14px] font-medium text-[#374151]"
                  >
                    薪资范围 <span className="text-[#EF4444]">*</span>
                  </Label>
                  <Input
                    id="salaryRange"
                    placeholder="例：10000-15000元/月"
                    value={jobFormData.salaryRange}
                    onChange={(e) =>
                      handleFormDataChange('salaryRange', e.target.value)
                    }
                    className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
                  />
                </div>
              </div>

              {/* 第四行：工作经验、行业要求 */}
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
                <div className="space-y-2">
                  <Label
                    htmlFor="industry"
                    className="text-[14px] font-medium text-[#374151]"
                  >
                    行业要求
                  </Label>
                  <Input
                    id="industry"
                    placeholder="例如：互联网、金融"
                    value={jobFormData.industry}
                    onChange={(e) =>
                      handleFormDataChange('industry', e.target.value)
                    }
                    className="h-10 border-[#D1D5DB] text-[14px] placeholder:text-[#9CA3AF]"
                  />
                </div>
              </div>

              {/* 工作技能 */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <Label className="text-[14px] font-medium text-[#374151]">
                    工作技能
                  </Label>
                  {skillsError && (
                    <div className="flex items-center gap-2">
                      <span className="text-[12px] text-[#EF4444]">
                        技能加载失败
                      </span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={retryFetchSkills}
                        disabled={skillsLoading}
                        className="h-6 w-6 p-0 text-[#6B7280] hover:text-[#374151]"
                        title="重试加载技能"
                      >
                        <RefreshCw
                          className={`h-3 w-3 ${skillsLoading ? 'animate-spin' : ''}`}
                        />
                      </Button>
                    </div>
                  )}
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label className="text-[14px] font-medium text-[#374151]">
                      必填技能 <span className="text-[#EF4444]">*</span>
                    </Label>
                    <div className="flex gap-2">
                      <Select value="" onValueChange={() => {}}>
                        <SelectTrigger className="h-10 border-[#D1D5DB] text-[14px]">
                          <SelectValue placeholder="请选择技能要求" />
                        </SelectTrigger>
                        <SelectContent>
                          {skillsLoading ? (
                            <SelectItem value="__loading__" disabled>
                              <div className="flex items-center gap-2">
                                <RefreshCw className="h-3 w-3 animate-spin" />
                                加载中...
                              </div>
                            </SelectItem>
                          ) : skillsError ? (
                            <div className="p-2">
                              <div className="text-[12px] text-[#EF4444] mb-2">
                                加载失败
                              </div>
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={retryFetchSkills}
                                className="h-6 text-[11px] px-2"
                              >
                                <RefreshCw className="h-3 w-3 mr-1" />
                                重试
                              </Button>
                            </div>
                          ) : availableSkills.length > 0 ? (
                            availableSkills.map((skill) => (
                              <SelectItem key={skill.id} value={skill.id}>
                                {skill.name}
                              </SelectItem>
                            ))
                          ) : (
                            <SelectItem value="__empty__" disabled>
                              暂无技能选项
                            </SelectItem>
                          )}
                        </SelectContent>
                      </Select>
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-10 w-10 p-0 border-[#D1D5DB]"
                        onClick={() => setIsAddSkillModalOpen(true)}
                        title="添加新技能"
                      >
                        <Plus className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label className="text-[14px] font-medium text-[#374151]">
                      选填技能
                    </Label>
                    <div className="flex gap-2">
                      <Select value="" onValueChange={() => {}}>
                        <SelectTrigger className="h-10 border-[#D1D5DB] text-[14px]">
                          <SelectValue placeholder="请选择技能要求" />
                        </SelectTrigger>
                        <SelectContent>
                          {skillsLoading ? (
                            <SelectItem value="__loading_optional__" disabled>
                              <div className="flex items-center gap-2">
                                <RefreshCw className="h-3 w-3 animate-spin" />
                                加载中...
                              </div>
                            </SelectItem>
                          ) : skillsError ? (
                            <div className="p-2">
                              <div className="text-[12px] text-[#EF4444] mb-2">
                                加载失败
                              </div>
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={retryFetchSkills}
                                className="h-6 text-[11px] px-2"
                              >
                                <RefreshCw className="h-3 w-3 mr-1" />
                                重试
                              </Button>
                            </div>
                          ) : availableSkills.length > 0 ? (
                            availableSkills.map((skill) => (
                              <SelectItem key={skill.id} value={skill.id}>
                                {skill.name}
                              </SelectItem>
                            ))
                          ) : (
                            <SelectItem value="__empty_optional__" disabled>
                              暂无技能选项
                            </SelectItem>
                          )}
                        </SelectContent>
                      </Select>
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-10 w-10 p-0 border-[#D1D5DB]"
                        onClick={() => setIsAddSkillModalOpen(true)}
                        title="添加新技能"
                      >
                        <Plus className="h-4 w-4" />
                      </Button>
                    </div>
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
                    className="h-8 px-3 text-[12px] text-[#22C55E] border-[#22C55E] hover:bg-[#F0FDF4] hover:text-[#16A34A] hover:border-[#16A34A]"
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
                          className="h-10 w-10 p-0 border-[#EF4444] text-[#EF4444] hover:bg-[#FEF2F2] hover:text-[#DC2626] hover:border-[#DC2626] mt-0"
                          title="删除此项职责"
                        >
                          <Trash2 className="h-4 w-4" />
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
              className="h-10 px-6 bg-[#22C55E] hover:bg-[#16A34A] text-white"
            >
              {loading ? '发布中...' : '保存并发布'}
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
    </div>
  );
}

// 默认导出JobProfilePage组件
export default JobProfilePage;
