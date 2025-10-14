# 简历岗位关联修复说明

## 问题描述

之前在简历详情模态框中编辑简历并保存岗位关联时，前端使用了错误的字段名 `job_ids` 发送到后端，导致保存后岗位关联无法正确更新。

## 修复方案

### 1. 类型定义修复

**文件**: `frontend/src/types/resume.ts`

修改 `ResumeUpdateParams` 类型，将 `job_ids` 字段更名为 `job_position_ids`，与后端接口保持一致：

```typescript
export interface ResumeUpdateParams {
  // ... 其他字段
  job_position_ids?: string[]; // 岗位ID列表（原 job_ids）
  job_applications?: JobApplicationAssociation[]; // 岗位申请详细信息
}
```

### 2. 简历详情模态框修复

**文件**: `frontend/src/components/modals/resume-preview-modal.tsx`

#### 2.1 初始化岗位选择状态

修改 `fetchResumeDetail` 函数，从 `job_positions` 数组中提取岗位ID进行初始化：

```typescript
const fetchResumeDetail = useCallback(async () => {
  if (!resumeId) return;

  setIsLoading(true);
  setError(null);
  try {
    const detail = await getResumeDetail(resumeId);
    setResumeDetail(detail);
    // 初始化已选岗位 - 从 job_positions 中提取岗位ID
    const jobIds = detail.job_positions?.map(jp => jp.job_position_id).filter(Boolean) || [];
    setSelectedJobIds(jobIds);
  } catch (err) {
    console.error('获取简历详情失败:', err);
    setError('获取简历详情失败，请稍后重试');
  } finally {
    setIsLoading(false);
  }
}, [resumeId]);
```

#### 2.2 保存时使用正确的字段名

修改 `handleSaveEdit` 函数，使用 `job_position_ids` 字段发送到后端：

```typescript
const handleSaveEdit = async () => {
  if (!editFormData || !resumeId) return;

  setIsSaving(true);
  try {
    // 将选中的岗位ID添加到更新参数中，使用 job_position_ids 字段
    const updateData = {
      ...editFormData,
      job_position_ids: selectedJobIds,
    };
    await updateResume(resumeId, updateData);
    await fetchResumeDetail(); // 重新获取数据
    setIsEditing(false);
    setEditFormData(null);
  } catch (err) {
    console.error('保存简历失败:', err);
    setError('保存简历失败，请稍后重试');
  } finally {
    setIsSaving(false);
  }
};
```

## 后端接口说明

### 更新简历接口

**端点**: `PUT /api/v1/resume/:id`

**请求参数**:
```json
{
  "name": "姓名",
  "email": "邮箱",
  "phone": "电话",
  // ... 其他字段
  "job_position_ids": ["岗位ID1", "岗位ID2"] // 岗位ID列表
}
```

### 获取简历详情接口

**端点**: `GET /api/v1/resume/:id`

**响应数据**:
```json
{
  "id": "简历ID",
  "name": "姓名",
  // ... 其他字段
  "job_positions": [
    {
      "id": "申请记录ID",
      "job_position_id": "岗位ID",
      "job_title": "岗位名称",
      "job_department": "岗位部门",
      "status": "申请状态",
      // ... 其他字段
    }
  ]
}
```

## 数据流程

1. **编辑简历岗位关联**：
   - 用户在简历详情模态框中打开编辑模式
   - 通过多选下拉框选择关联的岗位
   - 选中的岗位ID存储在 `selectedJobIds` 状态中

2. **保存岗位关联**：
   - 用户点击保存按钮
   - 前端将 `selectedJobIds` 作为 `job_position_ids` 字段发送到后端
   - 后端更新简历的岗位关联关系

3. **刷新显示**：
   - 保存成功后，重新调用 `getResumeDetail` 获取最新数据
   - 后端返回更新后的 `job_positions` 数组，包含岗位名称等详细信息
   - 前端在查看模式下显示岗位名称列表

## 测试验证

1. 打开简历详情模态框
2. 点击"编辑简历"按钮进入编辑模式
3. 在"关联岗位"下拉框中选择一个或多个岗位
4. 点击"保存"按钮
5. 验证：
   - 保存成功后，模态框切换回查看模式
   - 岗位名称正确显示在"关联岗位"区域
   - 鼠标悬停在岗位数量标签上，弹出框显示所有岗位名称

## 相关文件

- `frontend/src/types/resume.ts` - 简历类型定义
- `frontend/src/components/modals/resume-preview-modal.tsx` - 简历详情模态框
- `frontend/src/services/resume.ts` - 简历API服务
- `backend/domain/resume.go` - 后端简历领域模型
- `backend/internal/resume/usecase/resume.go` - 后端简历业务逻辑

## 注意事项

1. 前后端字段名必须保持一致，使用 `job_position_ids` 而非 `job_ids`
2. 保存后必须重新获取简历详情，以获取最新的岗位名称等信息
3. 初始化时从 `job_positions` 数组提取岗位ID，而非使用不存在的 `job_ids` 字段

