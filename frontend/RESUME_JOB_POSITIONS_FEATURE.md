# 简历管理岗位信息展示功能

## 功能概述

为简历管理系统添加了岗位信息展示和上传功能，包括：

1. **简历上传** - 上传时选择岗位，将岗位ID发送到接口
2. **简历列表** - 在简历信息后新增"岗位名称"列
3. **简历详情** - 在姓名下方展示岗位信息，使用浅绿色样式
4. **多岗位展示** - 支持展示多个岗位，hover 显示完整列表

## 实现细节

### 1. 简历上传功能

**文件：** 
- `frontend/src/services/resume.ts` - 上传接口
- `frontend/src/hooks/useResume.ts` - 上传Hook
- `frontend/src/components/modals/upload-resume-modal.tsx` - 上传模态框

#### 上传接口更新

```typescript
// 上传简历
export const uploadResume = async (
  file: File,
  jobPositionIds?: string[],  // 新增：岗位ID数组
  position?: string
): Promise<Resume> => {
  const formData = new FormData();
  formData.append('file', file);
  
  // 如果有岗位ID列表，用逗号分隔发送
  if (jobPositionIds && jobPositionIds.length > 0) {
    formData.append('job_position_ids', jobPositionIds.join(','));
  }
  
  // 保留旧的position参数以兼容
  if (position) {
    formData.append('position', position);
  }

  return apiPost<Resume>('/v1/resume/upload', formData);
};
```

#### 上传Hook更新

```typescript
export const useResumeUpload = () => {
  const uploadFile = useCallback(
    async (file: File, jobPositionIds?: string[], position?: string): Promise<Resume> => {
      // ... 上传逻辑
      const response = await uploadResume(file, jobPositionIds, position);
      return response;
    },
    []
  );
  // ...
};
```

#### 上传模态框更新

```typescript
const handleFileUpload = () => {
  // ...
  const resume = await uploadFile(file, selectedJobIds, position || undefined);
  // selectedJobIds 是用户选择的岗位ID数组
};
```

**关键要点：**
- 用户在上传简历时必须选择至少一个岗位
- 选中的岗位ID会以逗号分隔的字符串形式发送到 `/v1/resume/upload` 接口的 `job_position_ids` 参数
- 例如：选择3个岗位 `["id1", "id2", "id3"]` → 发送 `job_position_ids=id1,id2,id3`

### 2. TypeScript 类型定义更新

**文件：** `frontend/src/types/resume.ts`

#### 新增 JobApplication 接口

```typescript
// 简历岗位申请关联信息 - 根据backend domain.JobApplication定义
export interface JobApplication {
  id: string;
  resume_id: string;
  job_position_id: string;
  job_title: string; // 岗位标题
  job_department: string; // 岗位部门
  status: string; // 申请状态
  source: string; // 申请来源
  notes: string; // 备注信息
  applied_at?: string; // 申请时间
  created_at: number;
  updated_at: number;
}
```

#### 更新 Resume 接口

```typescript
export interface Resume {
  // ... 其他字段
  job_positions?: JobApplication[]; // 关联的岗位信息
}
```

### 3. 简历列表功能

**文件：** `frontend/src/components/resume/resume-table.tsx`

#### 新增岗位名称列

在表头中添加"岗位名称"列：

```typescript
<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
  岗位名称
</th>
```

#### 岗位展示逻辑

```typescript
const renderJobPositionsCell = (resume: Resume) => {
  const jobPositions = resume.job_positions || [];
  
  // 无岗位
  if (jobPositions.length === 0) {
    return <span className="text-gray-400">-</span>;
  }
  
  // 单个岗位
  if (jobPositions.length === 1) {
    return <span className="text-sm text-gray-900">{jobPositions[0].job_title}</span>;
  }
  
  // 多个岗位：显示"第一个岗位等X个岗位"，hover 显示完整列表
  const firstJobTitle = jobPositions[0].job_title;
  const remainingCount = jobPositions.length - 1;
  const allJobTitles = jobPositions.map(jp => jp.job_title).join('、');
  
  return (
    <div className="text-sm text-gray-900">
      {firstJobTitle}等
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="text-primary cursor-help underline decoration-dotted">
              {remainingCount}
            </span>
          </TooltipTrigger>
          <TooltipContent className="max-w-xs">
            <div className="text-sm">
              <div className="font-medium mb-1">所有岗位：</div>
              <div>{allJobTitles}</div>
            </div>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
      个岗位
    </div>
  );
};
```

### 4. 简历详情功能

**文件：** `frontend/src/components/modals/resume-preview-modal.tsx`

#### 岗位信息展示位置

在姓名下方，联系信息上方展示岗位信息：

```typescript
{/* 姓名 */}
<h1 className="text-3xl font-bold text-gray-900 mb-2">
  {resumeDetail.name}
</h1>

{/* 岗位信息 - 浅绿色样式 */}
{!isEditing && resumeDetail.job_positions && resumeDetail.job_positions.length > 0 && (
  <div className="flex items-center justify-center gap-2 text-sm mb-3">
    <Briefcase className="w-4 h-4 text-emerald-500" />
    {resumeDetail.job_positions.length === 1 ? (
      <span className="text-emerald-600">{resumeDetail.job_positions[0].job_title}</span>
    ) : (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="cursor-help text-emerald-600">
              {resumeDetail.job_positions[0].job_title}等
              <span className="font-medium underline decoration-dotted mx-1">
                {resumeDetail.job_positions.length - 1}
              </span>
              个岗位
            </span>
          </TooltipTrigger>
          <TooltipContent side="bottom" className="max-w-xs">
            <div className="space-y-1">
              <div className="font-semibold text-xs text-gray-500 mb-2">所有岗位：</div>
              {resumeDetail.job_positions.map((jobPos, idx) => (
                <div key={jobPos.id || idx} className="text-sm">
                  • {jobPos.job_title}
                </div>
              ))}
            </div>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )}
  </div>
)}

{/* 联系信息 */}
<div className="flex items-center justify-center gap-6 text-sm text-gray-600">
  ...
</div>
```

## 数据来源

### API 接口

#### 简历上传接口
- **路径：** `/v1/resume/upload`
- **方法：** POST (multipart/form-data)
- **参数：**
  - `file` - 简历文件
  - `job_position_ids` - 岗位ID列表，多个ID用逗号分隔（例如：`"id1,id2,id3"`）
  - `position` - （可选）岗位信息，兼容旧版本

#### 简历列表接口
- **路径：** `/v1/resume/list`
- **返回数据结构：** 
  ```json
  {
    "data": {
      "resumes": [
        {
          "id": "xxx",
          "name": "张三",
          "job_positions": [
            {
              "id": "xxx",
              "job_title": "前端工程师",
              "job_department": "技术部"
            }
          ]
        }
      ]
    }
  }
  ```
- **关键字段：** `data.resumes[].job_positions[].job_title`

#### 简历详情接口
- **路径：** `/v1/resume/{id}`
- **返回字段：** `job_positions` - JobApplication 数组
  - `job_title` - 从关联的岗位（job_position）获取
  - `job_department` - 从关联的部门获取

### 后端实现

根据后端 `domain.JobApplication` 定义：

```go
type JobApplication struct {
    ID            string     `json:"id"`
    ResumeID      string     `json:"resume_id"`
    JobPositionID string     `json:"job_position_id"`
    JobTitle      string     `json:"job_title"`      // 岗位标题
    JobDepartment string     `json:"job_department"` // 岗位部门
    Status        string     `json:"status"`
    Source        string     `json:"source"`
    Notes         string     `json:"notes"`
    AppliedAt     *time.Time `json:"applied_at"`
    CreatedAt     int64      `json:"created_at"`
    UpdatedAt     int64      `json:"updated_at"`
}
```

## UI/UX 设计

### 简历列表

1. **新增列位置：** 在"简历信息"列之后，"上传时间"列之前
2. **单个岗位：** 直接显示岗位名称
3. **多个岗位：** 显示 `{第一个岗位}等{N}个岗位`
4. **Hover 效果：** 
   - 数字 N 显示为主题色（绿色）
   - 添加下划线装饰（虚线样式）
   - 鼠标悬停显示所有岗位列表

### 简历详情模态框

1. **展示位置：** 姓名正下方，联系信息上方
2. **颜色样式：** 
   - 图标：`text-emerald-500` (浅绿色)
   - 文字：`text-emerald-600` (浅绿色)
   - 字号：比姓名小（text-sm）
3. **图标：** Briefcase（公文包）图标
4. **单个岗位：** 直接显示岗位名称
5. **多个岗位：** 
   - 显示格式：`{第一个岗位}等 {N} 个岗位`
   - Hover 显示所有岗位的详细列表

## 使用示例

### 简历上传场景

1. 用户点击"上传简历"按钮
2. 在上传模态框中：
   - **第一步**：选择至少一个岗位（必填），选择上传方式
   - **第二步**：预览简历解析结果
   - **第三步**：完成上传
3. 后端接收到的数据：
   ```
   POST /v1/resume/upload
   Content-Type: multipart/form-data
   
   file: [简历文件]
   job_position_ids: "岗位ID1,岗位ID2,岗位ID3"
   ```

### 简历列表场景

**数据源：** `/v1/resume/list` 接口返回的 `data.resumes[].job_positions[].job_title`

```
简历信息           岗位名称                    上传时间            上传人      状态
----------------------------------------------------------------------------------
张三              前端工程师                   2024-01-01 10:00    李四      待筛选
13800138000

王五              Java工程师等2个岗位          2024-01-02 11:00    赵六      待筛选
13900139000       (hover显示: Java工程师、Python工程师)
```

### 简历详情场景

```
                        张三
                🏢 前端工程师 (浅绿色)
            📧 zhang@example.com  📱 13800138000  📍 北京
            
            ──────────────────────────────────

                        王五
            🏢 Java工程师等 2 个岗位 (浅绿色)
            (hover "2" 显示: • Java工程师 • Python工程师)
        📧 wang@example.com  📱 13900139000  📍 上海
```

## 兼容性说明

1. **向后兼容：** 如果 `job_positions` 为空或不存在，显示 `-`
2. **数据格式：** 使用后端返回的 `JobApplication` 数组
3. **旧字段支持：** 保留 `job_ids` 和 `job_names` 字段供其他功能使用

## 测试要点

### 简历上传测试

1. ✅ 上传前必须选择至少一个岗位
2. ✅ 选择岗位后才能点击"上传文件"按钮
3. ✅ 多个岗位ID用逗号分隔发送到接口
4. ✅ 验证 `job_position_ids` 参数格式正确
5. ✅ 上传成功后简历关联到选中的岗位

### 简历列表测试

1. ✅ 无岗位时显示 `-`
2. ✅ 单个岗位时直接显示岗位名称（从 `job_positions[0].job_title` 获取）
3. ✅ 多个岗位时显示"第一个等N个岗位"
4. ✅ Hover 多个岗位时显示完整列表
5. ✅ 数字 N 显示为绿色带下划线
6. ✅ 确认数据源为 `/v1/resume/list` 接口的 `data.resumes[].job_positions[].job_title`

### 简历详情测试

1. ✅ 岗位信息显示在姓名下方
2. ✅ 使用浅绿色（emerald-500/600）
3. ✅ 字号小于姓名
4. ✅ 单个岗位时直接显示
5. ✅ 多个岗位时显示"等N个岗位"
6. ✅ Hover 显示所有岗位列表
7. ✅ 编辑模式下不显示岗位信息
8. ✅ 确认数据源为 `/v1/resume/{id}` 接口的 `job_positions[].job_title`

## 相关文件

### 前端文件
- ✅ `frontend/src/types/resume.ts` - 类型定义
- ✅ `frontend/src/services/resume.ts` - 上传接口服务
- ✅ `frontend/src/hooks/useResume.ts` - 上传Hook
- ✅ `frontend/src/components/modals/upload-resume-modal.tsx` - 上传模态框
- ✅ `frontend/src/components/resume/resume-table.tsx` - 简历列表
- ✅ `frontend/src/components/modals/resume-preview-modal.tsx` - 简历详情

### 后端文件
- 📄 `backend/domain/job_application.go` - 后端数据结构
- 📄 `backend/domain/resume.go` - 后端简历定义
- 📄 `backend/internal/resume/handler/v1/resume.go` - 上传接口处理

## 样式规范

### 简历列表
- 岗位文字：`text-sm text-gray-900`
- 数字高亮：`text-primary`（绿色主题）
- 下划线：`underline decoration-dotted`
- 空状态：`text-gray-400`

### 简历详情
- 图标颜色：`text-emerald-500`
- 文字颜色：`text-emerald-600`
- 字号：`text-sm`
- 间距：`mb-3`（姓名下方）
- Tooltip 位置：`side="bottom"`

