# 岗位画像与简历智能匹配设计文档

## 1. 概述

本文档描述了岗位画像（JobProfile）与简历（Resume）之间的智能匹配系统设计，包括数据模型对应关系、匹配算法和评分机制。

## 2. 数据模型分析

### 2.1 岗位画像数据结构 (JobProfileDetail)

```go
type JobProfileDetail struct {
    *JobProfile                                    // 基本信息
    Responsibilities       []*JobResponsibility    // 职责要求
    Skills                 []*JobSkill            // 技能要求
    EducationRequirements  []*JobEducationRequirement  // 教育要求
    ExperienceRequirements []*JobExperienceRequirement // 经验要求
    IndustryRequirements   []*JobIndustryRequirement   // 行业要求
}

// JobProfile 基本信息
type JobProfile struct {
    ID            string   `json:"id"`
    Name          string   `json:"name"`
    DepartmentID  string   `json:"department_id"`
    Department    string   `json:"department,omitempty"`
    WorkType      *string  `json:"work_type,omitempty"`      // 工作性质
    Location      *string  `json:"location,omitempty"`       // 工作地点
    SalaryMin     *float64 `json:"salary_min,omitempty"`     // 最低薪资
    SalaryMax     *float64 `json:"salary_max,omitempty"`     // 最高薪资
    Description   *string  `json:"description,omitempty"`    // 职位描述
    Status        string   `json:"status"`                   // 职位状态
}

// JobResponsibility 职责要求（注意：移除了SortOrder权重字段）
type JobResponsibility struct {
    ID             string `json:"id"`
    JobID          string `json:"job_id"`
    Responsibility string `json:"responsibility"`
}

// JobSkill 技能要求（注意：移除了Weight权重字段）
type JobSkill struct {
    ID      string `json:"id"`
    JobID   string `json:"job_id"`
    SkillID string `json:"skill_id"`
    Skill   string `json:"skill"`
    Type    string `json:"type"`    // required/bonus
}

// JobEducationRequirement 教育要求
type JobEducationRequirement struct {
    ID            string `json:"id"`
    JobID         string `json:"job_id"`
    EducationType string `json:"education_type"`  // unlimited/junior_college/bachelor/master/doctor
}

// JobExperienceRequirement 经验要求
type JobExperienceRequirement struct {
    ID             string `json:"id"`
    JobID          string `json:"job_id"`
    ExperienceType string `json:"experience_type"`  // unlimited/fresh_graduate/under_one_year/one_to_three_years/three_to_five_years/five_to_ten_years/over_ten_years
    MinYears       int    `json:"min_years"`        // 最低年限
    IdealYears     int    `json:"ideal_years"`      // 理想年限
}

// JobIndustryRequirement 行业要求
type JobIndustryRequirement struct {
    ID          string  `json:"id"`
    JobID       string  `json:"job_id"`
    Industry    string  `json:"industry"`           // 目标行业
    CompanyName *string `json:"company_name,omitempty"` // 目标公司（可选）
}
```

### 2.2 简历数据结构 (ResumeDetail)

```go
type ResumeDetail struct {
    *Resume                    // 基本信息
    Educations  []*ResumeEducation   // 教育经历
    Experiences []*ResumeExperience  // 工作经历
    Skills      []*ResumeSkill       // 技能列表
    Projects    []*ResumeProject     // 项目经验（新增）
    Logs        []*ResumeLog         // 操作日志
}

// Resume 基本信息
type Resume struct {
    ID               string       `json:"id"`
    Name             string       `json:"name"`
    Gender           string       `json:"gender"`
    Birthday         *time.Time   `json:"birthday,omitempty"`
    Email            string       `json:"email"`
    Phone            string       `json:"phone"`
    CurrentCity      string       `json:"current_city"`
    HighestEducation string       `json:"highest_education"`
    YearsExperience  float64      `json:"years_experience"`
    // 注意：缺少ExpectedSalaryMin/Max字段，需要在匹配时处理
}

// ResumeExperience 工作经历（新增Title字段）
type ResumeExperience struct {
    ID          string     `json:"id"`
    ResumeID    string     `json:"resume_id"`
    Company     string     `json:"company"`
    Position    string     `json:"position"`
    Title       string     `json:"title"`        // 新增：职位标题
    StartDate   *time.Time `json:"start_date,omitempty"`
    EndDate     *time.Time `json:"end_date,omitempty"`
    Description string     `json:"description"`
}

// ResumeProject 项目经验（新增结构）
type ResumeProject struct {
    ID               string     `json:"id"`
    ResumeID         string     `json:"resume_id"`
    Name             string     `json:"name"`
    Role             string     `json:"role"`
    Company          string     `json:"company"`
    StartDate        *time.Time `json:"start_date,omitempty"`
    EndDate          *time.Time `json:"end_date,omitempty"`
    Description      string     `json:"description"`
    Responsibilities string     `json:"responsibilities"`
    Achievements     string     `json:"achievements"`
    Technologies     string     `json:"technologies"`
    ProjectURL       string     `json:"project_url"`
    ProjectType      string     `json:"project_type"`
}
```

## 3. 字段对应关系与数据分发策略

### 3.1 基本信息匹配

| 岗位画像字段        | 简历字段    | 匹配类型     | 权重 | 说明                               |
| ------------------- | ----------- | ------------ | ---- | ---------------------------------- |
| Location            | CurrentCity | 地理位置匹配 | 10%  | 工作地点与当前城市匹配             |
| SalaryMin/SalaryMax | -           | 薪资期望匹配 | 15%  | 简历缺少期望薪资字段，需要特殊处理 |
| Department          | -           | 部门匹配     | 5%   | 可通过工作经历推断匹配度           |

**数据分发到 BasicInfoAgent**：

```go
type BasicInfoData struct {
    JobProfile *domain.JobProfileDetail `json:"job_profile"`
    Resume     *domain.ResumeDetail     `json:"resume"`
}
```

### 3.2 职责匹配

| 岗位画像                         | 简历                           | 匹配逻辑             |
| -------------------------------- | ------------------------------ | -------------------- |
| JobResponsibility.Responsibility | ResumeExperience.Description   | 职责描述语义匹配     |
| -                                | ResumeExperience.Position      | 职位标题相关性匹配   |
| -                                | ResumeExperience.Title         | 职位标题补充匹配     |
| -                                | ResumeProject.Responsibilities | 项目职责匹配（新增） |

**重要变化**：

- **移除了 SortOrder 权重**：所有职责等权重处理，或由 LLM 动态评估重要性
- **新增项目职责匹配**：ResumeProject.Responsibilities 可作为职责匹配的补充数据源
- **Title 字段增强**：ResumeExperience.Title 提供更精确的职位匹配

**数据分发到 ResponsibilityAgent**：

```go
type ResponsibilityData struct {
    JobResponsibilities []*domain.JobResponsibility `json:"job_responsibilities"`
    ResumeExperiences   []*domain.ResumeExperience  `json:"resume_experiences"`
    ResumeProjects      []*domain.ResumeProject     `json:"resume_projects"` // 新增
}
```

### 3.3 技能匹配

| 岗位画像                       | 简历                       | 匹配逻辑               |
| ------------------------------ | -------------------------- | ---------------------- |
| JobSkill.Skill                 | ResumeSkill.SkillName      | 技能名称精确/模糊匹配  |
| JobSkill.Type (required/bonus) | ResumeSkill.Level          | 必需技能 vs 技能熟练度 |
| -                              | ResumeProject.Technologies | 项目技术栈匹配（新增） |

**重要变化**：

- **移除了 Weight 权重**：改为基于 Type（required/bonus）的二分类权重
- **新增项目技术栈**：ResumeProject.Technologies 作为技能匹配的补充来源

**数据分发到 SkillAgent**：

```go
type SkillData struct {
    JobSkills       []*domain.JobSkill    `json:"job_skills"`
    ResumeSkills    []*domain.ResumeSkill `json:"resume_skills"`
    ResumeProjects  []*domain.ResumeProject `json:"resume_projects"` // 新增，用于提取技术栈
}
```

### 3.4 教育背景匹配

| 岗位画像                              | 简历                           | 匹配逻辑         |
| ------------------------------------- | ------------------------------ | ---------------- |
| JobEducationRequirement.ID            | -                              | 教育要求记录标识 |
| JobEducationRequirement.JobID         | -                              | 关联职位标识     |
| JobEducationRequirement.EducationType | ResumeEducation.Degree         | 学历层次匹配     |
| -                                     | ResumeEducation.School         | 学校声誉加权     |
| -                                     | ResumeEducation.Major          | 专业相关性匹配   |
| -                                     | ResumeEducation.UniversityType | 院校类型评估     |

**教育类型枚举值**：

- `unlimited`: 不限学历
- `junior_college`: 大专
- `bachelor`: 本科
- `master`: 硕士
- `doctor`: 博士

**数据分发到 EducationAgent**：

```go
type EducationData struct {
    JobEducationRequirements []*domain.JobEducationRequirement `json:"job_education_requirements"`
    ResumeEducations         []*domain.ResumeEducation         `json:"resume_educations"`
}
```

### 3.5 工作经验匹配

| 岗位画像                                | 简历                      | 匹配逻辑         |
| --------------------------------------- | ------------------------- | ---------------- |
| JobExperienceRequirement.ID             | -                         | 经验要求记录标识 |
| JobExperienceRequirement.JobID          | -                         | 关联职位标识     |
| JobExperienceRequirement.ExperienceType | -                         | 工作经验类型匹配 |
| JobExperienceRequirement.MinYears       | Resume.YearsExperience    | 最低年限匹配     |
| JobExperienceRequirement.IdealYears     | Resume.YearsExperience    | 理想年限匹配     |
| -                                       | ResumeExperience.Position | 职位相关性匹配   |
| -                                       | ResumeExperience.Title    | 职位标题匹配     |
| -                                       | ResumeExperience.Company  | 公司背景匹配     |

**经验类型枚举值**：

- `unlimited`: 不限经验
- `fresh_graduate`: 应届毕业生
- `under_one_year`: 1 年以下
- `one_to_three_years`: 1-3 年
- `three_to_five_years`: 3-5 年
- `five_to_ten_years`: 5-10 年
- `over_ten_years`: 10 年以上

**数据分发到 ExperienceAgent**：

```go
type ExperienceData struct {
    JobExperienceRequirements []*domain.JobExperienceRequirement `json:"job_experience_requirements"`
    ResumeExperiences         []*domain.ResumeExperience         `json:"resume_experiences"`
    ResumeYearsExperience     float64                            `json:"resume_years_experience"`
}
```

### 3.6 行业背景匹配

| 岗位画像                           | 简历                     | 匹配逻辑         |
| ---------------------------------- | ------------------------ | ---------------- |
| JobIndustryRequirement.ID          | -                        | 行业要求记录标识 |
| JobIndustryRequirement.JobID       | -                        | 关联职位标识     |
| JobIndustryRequirement.Industry    | ResumeExperience.Company | 行业分类匹配     |
| JobIndustryRequirement.CompanyName | ResumeExperience.Company | 特定公司匹配     |

**匹配策略说明**：

- **行业匹配**：通过公司名称推断所属行业，与目标行业进行语义匹配
- **公司匹配**：当指定特定公司时，进行精确或模糊匹配
- **权重分配**：特定公司匹配权重高于行业匹配

**数据分发到 IndustryAgent**：

```go
type IndustryData struct {
    JobIndustryRequirements []*domain.JobIndustryRequirement `json:"job_industry_requirements"`
    ResumeExperiences       []*domain.ResumeExperience       `json:"resume_experiences"`
}
```

## 4. 匹配评分模型

### 4.1 综合匹配度计算

```
总匹配度 = Σ(维度匹配度 × 维度权重)

维度权重分配：
- 技能匹配：35%
- 职责匹配：20%
- 工作经验：20%
- 教育背景：15%
- 行业背景：7%
- 基本信息：3%
```

### 4.2 各维度评分算法

#### 基本信息匹配评分

```go
type BasicMatchDetail struct {
    Score     float64            `json:"score"`
    SubScores map[string]float64 `json:"sub_scores"`
    Evidence  []string           `json:"evidence"`
    Notes     string             `json:"notes"`
}

// 计算公式（权重已调整）
LocationScore = 地点匹配度 (0-100)
SalaryScore = 0 // 简历缺少期望薪资字段，暂时设为0或通过其他方式推断
DepartmentScore = 部门匹配度 (0-100)

OverallScore = (LocationScore * 0.5) + (SalaryScore * 0.3) + (DepartmentScore * 0.2)
```

#### 技能匹配评分

```go
type SkillMatchDetail struct {
    Score         float64            `json:"score"`
    MatchedSkills []*MatchedSkill    `json:"matched_skills"`
    MissingSkills []*domain.JobSkill `json:"missing_skills"`
    ExtraSkills   []string           `json:"extra_skills"`
    LLMAnalysis   *SkillLLMAnalysis  `json:"llm_analysis"`
}

// 计算公式（移除Weight权重，改用Type权重）
RequiredSkillsScore = Σ(匹配的必需技能得分) / 必需技能总数 * 100
BonusSkillsScore = Σ(匹配的加分技能得分) / 加分技能总数 * 100
ProjectTechScore = 从项目技术栈中提取的技能匹配度

OverallScore = (RequiredSkillsScore * 0.7) + (BonusSkillsScore * 0.2) + (ProjectTechScore * 0.1)
```

#### 职责匹配评分

```go
type ResponsibilityMatchDetail struct {
    Score                     float64                     `json:"score"`
    MatchedResponsibilities   []*MatchedResponsibility    `json:"matched_responsibilities"`   // 匹配的职责
    UnmatchedResponsibilities []*domain.JobResponsibility `json:"unmatched_responsibilities"` // 未匹配的职责
    RelevantExperiences       []string                    `json:"relevant_experiences"`       // 相关工作经历ID
}

// 计算公式（移除SortOrder权重，等权重处理）
ExperienceMatchScore = 工作经历与职责的匹配度
ProjectMatchScore = 项目经历与职责的匹配度 // 新增

OverallScore = (ExperienceMatchScore * 0.8) + (ProjectMatchScore * 0.2)

// 大模型匹配算法增强
// 输入：职责要求 + 工作经历描述 + 项目职责描述
// 输出：结构化的匹配评分和分析
```

#### 经验匹配评分

```go
type ExperienceMatchDetail struct {
    Score           float64              `json:"score"`
    YearsMatch      *YearsMatchInfo      `json:"years_match"`
    PositionMatches []*PositionMatchInfo `json:"position_matches"`
    IndustryMatches []*IndustryMatchInfo `json:"industry_matches"`
}

// 经验类型枚举值
const (
    ExperienceTypeUnlimited        = "unlimited"         // 不限经验
    ExperienceTypeFreshGraduate    = "fresh_graduate"    // 应届毕业生
    ExperienceTypeUnderOneYear     = "under_one_year"    // 1年以下
    ExperienceTypeOneToThreeYears  = "one_to_three_years" // 1-3年
    ExperienceTypeThreeToFiveYears = "three_to_five_years" // 3-5年
    ExperienceTypeFiveToTenYears   = "five_to_ten_years"  // 5-10年
    ExperienceTypeOverTenYears     = "over_ten_years"     // 10年以上
)

// 经验匹配评分算法（基于JobExperienceRequirement）
YearsScore = 年限匹配度 (0-100)
PositionScore = 职位匹配度 (0-100)
IndustryScore = 行业匹配度 (0-100)

// 综合评分权重分配
OverallScore = (YearsScore * 0.5) + (PositionScore * 0.3) + (IndustryScore * 0.2)

// 年限评分逻辑（基于JobExperienceRequirement.ExperienceType）
switch JobExperienceRequirement.ExperienceType {
case "unlimited":         // 不限经验
    YearsScore = 100      // 任何经验都满足
case "fresh_graduate":    // 应届毕业生
    if Resume.YearsExperience <= 0.5: YearsScore = 100
    else if Resume.YearsExperience <= 1: YearsScore = 80
    else: YearsScore = 60
case "under_one_year":    // 1年以下
    if Resume.YearsExperience < 1: YearsScore = 100
    else if Resume.YearsExperience <= 2: YearsScore = 80
    else: YearsScore = 60
case "one_to_three_years": // 1-3年
    if Resume.YearsExperience >= 1 && Resume.YearsExperience <= 3: YearsScore = 100
    else if Resume.YearsExperience < 1: YearsScore = 70
    else if Resume.YearsExperience <= 5: YearsScore = 80
    else: YearsScore = 60
case "three_to_five_years": // 3-5年
    if Resume.YearsExperience >= 3 && Resume.YearsExperience <= 5: YearsScore = 100
    else if Resume.YearsExperience >= 2 && Resume.YearsExperience < 3: YearsScore = 80
    else if Resume.YearsExperience <= 7: YearsScore = 80
    else: YearsScore = 60
case "five_to_ten_years": // 5-10年
    if Resume.YearsExperience >= 5 && Resume.YearsExperience <= 10: YearsScore = 100
    else if Resume.YearsExperience >= 3 && Resume.YearsExperience < 5: YearsScore = 80
    else if Resume.YearsExperience > 10: YearsScore = 90
    else: YearsScore = 60
case "over_ten_years":    // 10年以上
    if Resume.YearsExperience >= 10: YearsScore = 100
    else if Resume.YearsExperience >= 8: YearsScore = 80
    else: YearsScore = 60
}

// 职位匹配逻辑（基于JobExperienceRequirement.Position）
ExactMatch: PositionScore = 100
SimilarMatch: PositionScore = 80  // 相似职位
RelatedMatch: PositionScore = 60  // 相关职位
NoMatch: PositionScore = 0

// 行业匹配逻辑（基于JobExperienceRequirement.Industry）
ExactMatch: IndustryScore = 100
RelatedMatch: IndustryScore = 80  // 相关行业
NoMatch: IndustryScore = 0
```

#### 教育匹配评分

```go
type EducationMatchDetail struct {
    Score         float64            `json:"score"`
    DegreeMatch   *DegreeMatchInfo   `json:"degree_match"`
    MajorMatches  []*MajorMatchInfo  `json:"major_matches"`
    SchoolMatches []*SchoolMatchInfo `json:"school_matches"`
}

// 教育类型枚举值
const (
    EducationTypeUnlimited      = "unlimited"       // 不限学历
    EducationTypeJuniorCollege  = "junior_college"  // 大专
    EducationTypeBachelor       = "bachelor"        // 本科
    EducationTypeMaster         = "master"          // 硕士
    EducationTypeDoctor         = "doctor"          // 博士
)

// 学历等级映射
博士 > 硕士 > 本科 > 大专

// 教育匹配评分算法（基于JobEducationRequirement.EducationType）
DegreeScore = 学历等级匹配度 (0-100)
MajorScore = 专业匹配度 (0-100)
SchoolScore = 院校匹配度 (0-100)

// 综合评分权重分配
OverallScore = (DegreeScore * 0.6) + (MajorScore * 0.3) + (SchoolScore * 0.1)

// 学历匹配逻辑（基于JobEducationRequirement.EducationType）
switch JobEducationRequirement.EducationType {
case "unlimited":       // 不限学历
    DegreeScore = 100   // 任何学历都满足
case "junior_college":  // 大专要求
    if Resume.Degree >= 大专: DegreeScore = 100
    else: DegreeScore = 60
case "bachelor":        // 本科要求
    if Resume.Degree >= 本科: DegreeScore = 100
    else if Resume.Degree == 大专: DegreeScore = 70
    else: DegreeScore = 40
case "master":          // 硕士要求
    if Resume.Degree >= 硕士: DegreeScore = 100
    else if Resume.Degree == 本科: DegreeScore = 80
    else if Resume.Degree == 大专: DegreeScore = 60
    else: DegreeScore = 40
case "doctor":          // 博士要求
    if Resume.Degree == 博士: DegreeScore = 100
    else if Resume.Degree == 硕士: DegreeScore = 85
    else if Resume.Degree == 本科: DegreeScore = 70
    else: DegreeScore = 50
}

// 专业匹配逻辑（基于JobEducationRequirement.Major）
ExactMatch: MajorScore = 100
RelatedMatch: MajorScore = 80  // 相关专业
NoMatch: MajorScore = 0

// 院校匹配逻辑（基于JobEducationRequirement.School）
if 指定院校匹配: SchoolScore = 100
else if 同等级院校: SchoolScore = 80
else: SchoolScore = 60
```

#### 行业匹配评分

```go
type IndustryMatchDetail struct {
    Score           float64              `json:"score"`
    IndustryMatches []*IndustryMatchInfo `json:"industry_matches"`
    CompanyMatches  []*CompanyMatchInfo  `json:"company_matches"`
}

// IndustryMatchInfo 行业匹配信息（已在ExperienceMatchDetail中定义，此处为行业维度专用）
type IndustryMatchInfo struct {
    JobRequirement     string             `json:"job_requirement"`     // 岗位要求的行业
    ResumeExperience   *ResumeExperience  `json:"resume_experience"`   // 匹配的简历经历
    MatchType          string             `json:"match_type"`          // 匹配类型: exact/related/similar/none
    Score              float64            `json:"score"`               // 行业匹配得分 (0-100)
    MatchReason        string             `json:"match_reason"`        // 匹配原因说明
}

// CompanyMatchInfo 公司匹配信息
type CompanyMatchInfo struct {
    JobRequirement     string             `json:"job_requirement"`     // 岗位要求的公司
    ResumeExperience   *ResumeExperience  `json:"resume_experience"`   // 匹配的简历经历
    MatchType          string             `json:"match_type"`          // 匹配类型: exact/same_group/competitor/same_industry/none
    Score              float64            `json:"score"`               // 公司匹配得分 (0-100)
    MatchReason        string             `json:"match_reason"`        // 匹配原因说明
}

// 行业匹配评分算法（基于JobIndustryRequirement）
IndustryScore = 行业匹配度 (0-100)
CompanyScore = 公司匹配度 (0-100)

// 综合评分计算（基于第三章节的权重分配）
OverallScore = (IndustryScore * 0.7) + (CompanyScore * 0.3)

// 行业匹配逻辑（基于JobIndustryRequirement.Industry）
ExactMatch: IndustryScore = 100      // 完全匹配
RelatedMatch: IndustryScore = 80     // 相关行业
SimilarMatch: IndustryScore = 60     // 相似行业
NoMatch: IndustryScore = 0           // 无匹配

// 公司匹配逻辑（基于JobIndustryRequirement.Company）
ExactMatch: CompanyScore = 100       // 完全匹配
SameGroupMatch: CompanyScore = 90    // 同集团公司
CompetitorMatch: CompanyScore = 80   // 竞争对手公司
SameIndustryMatch: CompanyScore = 60 // 同行业公司
NoMatch: CompanyScore = 0            // 无匹配

// 匹配策略说明
// 1. 行业匹配：基于简历工作经历中的行业信息与岗位要求的行业进行匹配
// 2. 公司匹配：基于简历工作经历中的公司信息与岗位要求的目标公司进行匹配
// 3. 权重分配：行业匹配占70%，公司匹配占30%，体现行业经验的重要性
```

## 5. 匹配结果数据模型

### 5.1 匹配结果结构

```go
// JobResumeMatch 岗位简历匹配结果
type JobResumeMatch struct {
    JobID               string                    `json:"job_id"`
    ResumeID            string                    `json:"resume_id"`
    OverallScore        float64                   `json:"overall_score"`         // 综合匹配度 (0-100)
    SkillMatch          *SkillMatchDetail         `json:"skill_match"`           // 技能匹配详情
    ResponsibilityMatch *ResponsibilityMatchDetail `json:"responsibility_match"` // 职责匹配详情
    ExperienceMatch     *ExperienceMatchDetail    `json:"experience_match"`      // 经验匹配详情
    EducationMatch      *EducationMatchDetail     `json:"education_match"`       // 教育匹配详情
    IndustryMatch       *IndustryMatchDetail      `json:"industry_match"`        // 行业匹配详情
    BasicMatch          *BasicMatchDetail         `json:"basic_match"`           // 基本信息匹配
    MatchedAt           time.Time                 `json:"matched_at"`            // 匹配时间
    Recommendations     []string                  `json:"recommendations"`       // 匹配建议
}

// BasicMatchDetail 基本信息匹配详情
type BasicMatchDetail struct {
    Score           float64            `json:"score"`
    LocationMatch   *LocationMatchInfo `json:"location_match"`
    SalaryMatch     *SalaryMatchInfo   `json:"salary_match"`
    DepartmentMatch *DepartmentMatchInfo `json:"department_match"`
}

// LocationMatchInfo 地点匹配信息
type LocationMatchInfo struct {
    JobLocation      string  `json:"job_location"`       // 岗位工作地点
    ResumeLocation   string  `json:"resume_location"`    // 简历当前城市
    MatchType        string  `json:"match_type"`         // 匹配类型: exact/same_region/different/none
    Score            float64 `json:"score"`              // 地点匹配得分 (0-100)
    MatchReason      string  `json:"match_reason"`       // 匹配原因说明
}

// SalaryMatchInfo 薪资匹配信息
type SalaryMatchInfo struct {
    JobSalaryMin     float64 `json:"job_salary_min"`     // 岗位最低薪资
    JobSalaryMax     float64 `json:"job_salary_max"`     // 岗位最高薪资
    ExpectedMin      float64 `json:"expected_min"`       // 期望最低薪资（可能需要推断）
    ExpectedMax      float64 `json:"expected_max"`       // 期望最高薪资（可能需要推断）
    MatchType        string  `json:"match_type"`         // 匹配类型: within_range/above_range/below_range/unknown
    Score            float64 `json:"score"`              // 薪资匹配得分 (0-100)
    MatchReason      string  `json:"match_reason"`       // 匹配原因说明
}

// DepartmentMatchInfo 部门匹配信息
type DepartmentMatchInfo struct {
    JobDepartment    string  `json:"job_department"`     // 岗位部门
    InferredDept     string  `json:"inferred_dept"`      // 从工作经历推断的部门
    MatchType        string  `json:"match_type"`         // 匹配类型: exact/related/none
    Score            float64 `json:"score"`              // 部门匹配得分 (0-100)
    MatchReason      string  `json:"match_reason"`       // 匹配原因说明
}

// SkillMatchDetail 技能匹配详情
type SkillMatchDetail struct {
    Score           float64            `json:"score"`
    MatchedSkills   []*MatchedSkill    `json:"matched_skills"`    // 匹配的技能
    MissingSkills   []*JobSkill        `json:"missing_skills"`    // 缺失的技能
    ExtraSkills     []*ResumeSkill     `json:"extra_skills"`      // 额外的技能
    LLMAnalysis     *SkillLLMAnalysis  `json:"llm_analysis"`      // 大模型整体分析
}

// MatchedSkill 匹配的技能
type MatchedSkill struct {
    JobSkill       *JobSkill         `json:"job_skill"`
    ResumeSkill    *ResumeSkill      `json:"resume_skill"`
    MatchType      string            `json:"match_type"`      // exact, semantic, related
    LLMScore       float64           `json:"llm_score"`       // 大模型评分 (0-100)
    ProficiencyGap float64           `json:"proficiency_gap"` // 熟练度差距
    Score          float64           `json:"score"`           // 该技能得分
    LLMAnalysis    *SkillItemAnalysis `json:"llm_analysis"`   // 大模型分析详情
}

// SkillLLMAnalysis 技能整体大模型分析
type SkillLLMAnalysis struct {
    OverallMatch     float64  `json:"overall_match"`      // 整体匹配度 (0-100)
    TechnicalFit     float64  `json:"technical_fit"`      // 技术契合度
    LearningCurve    string   `json:"learning_curve"`     // 学习曲线评估 (low/medium/high)
    StrengthAreas    []string `json:"strength_areas"`     // 优势技能领域
    GapAreas         []string `json:"gap_areas"`          // 技能缺口领域
    Recommendations  []string `json:"recommendations"`    // 技能提升建议
    AnalysisDetail   string   `json:"analysis_detail"`    // 详细分析说明
}

// SkillItemAnalysis 单项技能大模型分析
type SkillItemAnalysis struct {
    MatchLevel      string  `json:"match_level"`       // 匹配等级: perfect/good/partial/none
    MatchPercentage float64 `json:"match_percentage"`  // 匹配百分比 (0-100)
    ProficiencyGap  string  `json:"proficiency_gap"`   // 熟练度差距: none/minor/moderate/major
    Transferability string  `json:"transferability"`   // 技能可迁移性: high/medium/low
    LearningEffort  string  `json:"learning_effort"`   // 学习难度: minimal/moderate/significant
    MatchReason     string  `json:"match_reason"`      // 匹配原因说明
}

// ResponsibilityMatchDetail 职责匹配详情
type ResponsibilityMatchDetail struct {
    Score                   float64                      `json:"score"`
    MatchedResponsibilities []*MatchedResponsibility     `json:"matched_responsibilities"` // 匹配的职责
    UnmatchedResponsibilities []*JobResponsibility       `json:"unmatched_responsibilities"` // 未匹配的职责
    RelevantExperiences     []*ResumeExperience         `json:"relevant_experiences"`     // 相关工作经历
}

// MatchedResponsibility 匹配的职责
type MatchedResponsibility struct {
    JobResponsibility   *JobResponsibility  `json:"job_responsibility"`
    ResumeExperience    *ResumeExperience   `json:"resume_experience"`
    LLMAnalysis         *LLMMatchAnalysis   `json:"llm_analysis"`        // 大模型分析结果
    MatchScore          float64             `json:"match_score"`         // 该职责匹配得分
    MatchReason         string              `json:"match_reason"`        // 匹配原因说明
}

// LLMMatchAnalysis 大模型匹配分析结果
type LLMMatchAnalysis struct {
    MatchLevel          string              `json:"match_level"`         // 匹配等级: excellent/good/fair/poor
    MatchPercentage     float64             `json:"match_percentage"`    // 匹配百分比 (0-100)
    StrengthPoints      []string            `json:"strength_points"`     // 匹配优势点
    WeakPoints          []string            `json:"weak_points"`         // 不足之处
    RecommendedActions  []string            `json:"recommended_actions"` // 建议改进措施
    AnalysisDetail      string              `json:"analysis_detail"`     // 详细分析说明
}

// ExperienceMatchDetail 经验匹配详情
type ExperienceMatchDetail struct {
    Score              float64                    `json:"score"`
    YearsMatch         *YearsMatchInfo           `json:"years_match"`
    PositionMatches    []*PositionMatchInfo      `json:"position_matches"`
    IndustryMatches    []*IndustryMatchInfo      `json:"industry_matches"`
}

// YearsMatchInfo 年限匹配信息
type YearsMatchInfo struct {
    RequiredType       string  `json:"required_type"`       // 要求的经验类型 (ExperienceType)
    ActualYears        float64 `json:"actual_years"`        // 实际工作年限
    MatchLevel         string  `json:"match_level"`         // 匹配等级: perfect/good/acceptable/poor
    Score              float64 `json:"score"`               // 年限匹配得分 (0-100)
    Gap                string  `json:"gap"`                 // 差距描述
}

// PositionMatchInfo 职位匹配信息
type PositionMatchInfo struct {
    JobRequirement     string             `json:"job_requirement"`     // 岗位要求的职位
    ResumeExperience   *ResumeExperience  `json:"resume_experience"`   // 匹配的简历经历
    MatchType          string             `json:"match_type"`          // 匹配类型: exact/similar/related/none
    Score              float64            `json:"score"`               // 职位匹配得分 (0-100)
    MatchReason        string             `json:"match_reason"`        // 匹配原因说明
}

// IndustryMatchInfo 行业匹配信息
type IndustryMatchInfo struct {
    JobRequirement     string             `json:"job_requirement"`     // 岗位要求的行业
    ResumeExperience   *ResumeExperience  `json:"resume_experience"`   // 匹配的简历经历
    MatchType          string             `json:"match_type"`          // 匹配类型: exact/related/none
    Score              float64            `json:"score"`               // 行业匹配得分 (0-100)
    MatchReason        string             `json:"match_reason"`        // 匹配原因说明
}

// EducationMatchDetail 教育匹配详情
type EducationMatchDetail struct {
    Score           float64                `json:"score"`
    DegreeMatch     *DegreeMatchInfo      `json:"degree_match"`
    MajorMatches    []*MajorMatchInfo     `json:"major_matches"`
    SchoolMatches   []*SchoolMatchInfo    `json:"school_matches"`
}

// DegreeMatchInfo 学历匹配信息
type DegreeMatchInfo struct {
    RequiredType       string             `json:"required_type"`       // 要求的学历类型 (EducationType)
    ActualDegree       string             `json:"actual_degree"`       // 实际学历
    MatchLevel         string             `json:"match_level"`         // 匹配等级: perfect/exceed/acceptable/insufficient
    Score              float64            `json:"score"`               // 学历匹配得分 (0-100)
    Gap                string             `json:"gap"`                 // 差距描述
}

// MajorMatchInfo 专业匹配信息
type MajorMatchInfo struct {
    JobRequirement     string             `json:"job_requirement"`     // 岗位要求的专业
    ResumeEducation    *ResumeEducation   `json:"resume_education"`    // 匹配的教育经历
    MatchType          string             `json:"match_type"`          // 匹配类型: exact/related/none
    Score              float64            `json:"score"`               // 专业匹配得分 (0-100)
    MatchReason        string             `json:"match_reason"`        // 匹配原因说明
}

// SchoolMatchInfo 院校匹配信息
type SchoolMatchInfo struct {
    JobRequirement     string             `json:"job_requirement"`     // 岗位要求的院校
    ResumeEducation    *ResumeEducation   `json:"resume_education"`    // 匹配的教育经历
    MatchType          string             `json:"match_type"`          // 匹配类型: exact/same_tier/acceptable/none
    Score              float64            `json:"score"`               // 院校匹配得分 (0-100)
    MatchReason        string             `json:"match_reason"`        // 匹配原因说明
}
```

### 5.2 匹配等级定义

```go
type MatchLevel string

const (
    MatchLevelExcellent MatchLevel = "excellent"  // 90-100分
    MatchLevelGood      MatchLevel = "good"       // 75-89分
    MatchLevelFair      MatchLevel = "fair"       // 60-74分
    MatchLevelPoor      MatchLevel = "poor"       // 0-59分
)

```

## 6. Prompt 模板设计

### 6.1 设计原则

基于对现有 Prompt 文件的分析，我们制定了以下统一的设计原则：

#### 6.1.1 结构化设计原则
- **角色定位明确**：每个 Prompt 都明确定义 AI 的专业角色
- **评分规则标准化**：统一使用 0-100 分制，确保评分一致性
- **输出格式规范**：严格的 JSON 格式输出，便于系统解析
- **权重分配透明**：明确各维度的权重分配和计算公式

#### 6.1.2 内容组织原则
- **系统提示词**：包含角色定义、评分规则、输出格式
- **用户提示词**：使用 Go Template 格式，支持动态数据注入
- **模板配置**：统一的配置结构，便于维护和扩展

### 6.2 通用模板结构

#### 6.2.1 基础模板组件

```go
// ChatTemplateConfig 聊天模板配置
type ChatTemplateConfig struct {
    FormatType schema.FormatType           // 模板格式类型
    Templates  []schema.MessagesTemplate   // 消息模板列表
}

// 标准模板创建函数
func NewXXXChatTemplate(ctx context.Context) (prompt.ChatTemplate, error) {
    config := &ChatTemplateConfig{
        FormatType: schema.GoTemplate,
        Templates: []schema.MessagesTemplate{
            schema.SystemMessage(XXXSystemPrompt),
            schema.UserMessage(XXXUserPrompt),
        },
    }
    ctp := prompt.FromMessages(config.FormatType, config.Templates...)
    return ctp, nil
}
```

#### 6.2.2 系统提示词模板

```
你是一个专业的招聘匹配分析师，负责评估候选人的[具体维度]与职位要求的匹配度。

请根据以下评分规则对候选人进行评估：

## 评分维度

### 1. [维度1名称] ([字段名])
- [评分等级1]：[分数范围]分
- [评分等级2]：[分数范围]分
- [评分等级3]：[分数范围]分
- [评分等级4]：[分数范围]分

### 2. [维度2名称] ([字段名])
- [评分规则描述]

## 评分权重
- [维度1]：[权重百分比]
- [维度2]：[权重百分比]

## 输出格式

请严格按照以下JSON格式输出结果：

{
  "score": 总分（0-100的浮点数）,
  "[维度字段]": {
    // 具体维度的评分结构
  }
}

注意：
1. 所有分数必须是0-100之间的数值
2. 总分应该是各维度分数的加权平均
3. [具体注意事项]
```

#### 6.2.3 用户提示词模板

```
请分析以下候选人[具体维度]与职位要求的匹配度：

{{.input}}

请根据系统提示中的评分规则，对候选人的[具体维度]进行全面评估并输出JSON格式的结果。
```

### 6.3 各维度 Prompt 设计

#### 6.3.1 基本信息匹配 (BasicInfo)

**设计特点**：
- 关注地理位置、薪资期望、部门匹配三个核心维度
- 处理简历缺少期望薪资字段的特殊情况
- 权重分配：地理位置50%、薪资30%、部门20%

**关键评分规则**：
```
地理位置匹配：
- 完全匹配（同城市）：90-100分
- 相近地区（同省份/相邻城市）：70-89分
- 较远地区（需要搬迁）：40-69分
- 完全不匹配（跨国/跨大区）：0-39分

薪资期望匹配：
- 期望薪资在预算范围内：90-100分
- 期望薪资略高于预算（10%以内）：70-89分
- 期望薪资明显高于预算（10-30%）：40-69分
- 期望薪资严重超出预算（30%以上）：0-39分
```

#### 6.3.2 技能匹配 (Skill)

**设计特点**：
- 支持精确匹配、语义匹配、相关匹配多种匹配类型
- 区分必需技能和加分技能的权重
- 包含熟练度差距评估和学习难度分析

**关键评分规则**：
```
技能匹配类型：
- 精确匹配 (exact)：95-100分
- 语义匹配 (semantic)：85-94分
- 相关匹配 (related)：70-84分
- 无匹配 (none)：0-39分

熟练度等级：
- Expert (专家)：5年以上深度经验
- Advanced (高级)：3-5年丰富经验
- Intermediate (中级)：1-3年实践经验
- Beginner (初级)：1年以下或理论知识
```

#### 6.3.3 工作经验匹配 (Experience)

**设计特点**：
- 基于工作年限、职位相关性、行业背景三个维度评估
- 支持多种经验类型枚举值匹配
- 权重分配：年限40%、职位相关性40%、行业背景20%

**关键评分规则**：
```
工作年限匹配：
- 超出要求年限50%以上：90-100分
- 满足要求年限：80-89分
- 略低于要求年限（80-99%）：60-79分
- 明显低于要求年限（50-79%）：30-59分
- 严重不足（50%以下）：0-29分

职位相关性匹配：
- 完全相同职位：90-100分
- 高度相关职位：70-89分
- 中等相关职位：50-69分
- 低相关性职位：30-49分
- 无相关性：0-29分
```

#### 6.3.4 职责匹配 (Responsibility)

**设计特点**：
- 评估候选人工作职责与职位要求的匹配度
- 支持管理、技术、业务、运营等不同类型职责
- 提供详细的匹配分析和改进建议

**关键评分规则**：
```
职责匹配度评估：
- 完全匹配：90-100分，有完全相同或更高级别的职责经验
- 高度匹配：75-89分，有高度相关的职责经验，能够快速适应
- 中等匹配：60-74分，有部分相关职责经验，需要一定学习时间
- 低度匹配：40-59分，有少量相关经验，需要较长学习时间
- 不匹配：0-39分，缺乏相关职责经验
```

#### 6.3.5 教育背景匹配 (Education)

**设计特点**：
- 基于学历等级、专业相关性、学校声誉三个维度
- 支持国内外院校评估体系
- 权重分配：学历40%、专业40%、学校声誉20%

**关键评分规则**：
```
学历等级对应关系：
- 博士 (PhD)：最高等级
- 硕士 (Master)：高等级
- 学士 (Bachelor)：中等级
- 专科 (Associate)：基础等级
- 高中及以下：最低等级

学历匹配规则：
- 完全匹配或超出要求：90-100分
- 低一个等级：70-89分
- 低两个等级：40-69分
- 低三个等级及以上：0-39分
```

#### 6.3.6 行业背景匹配 (Industry)

**设计特点**：
- 评估行业相关性、公司背景、行业深度三个维度
- 考虑行业发展趋势和转换难度
- 权重分配：行业相关性60%、公司背景25%、行业深度15%

**关键评分规则**：
```
行业相关性匹配：
- 完全相同行业：90-100分
- 高度相关行业（上下游、相似业务模式）：70-89分
- 中等相关行业（部分业务重叠）：50-69分
- 低相关性行业（技能可迁移）：30-49分
- 完全无关行业：0-29分

行业深度评估：
- 在目标行业工作年限：
  - 5年以上：90-100分
  - 3-5年：70-89分
  - 1-3年：50-69分
  - 1年以下：30-49分
  - 无相关经验：0-29分
```

#### 6.3.7 匹配结果聚合 (Aggregator)

**设计特点**：
- 综合各维度匹配结果，生成最终匹配报告
- 严格按照权重计算综合得分
- 提供具体可行的匹配建议

**权重分配**：
```
维度权重分配（总计100%）：
- 技能匹配 (SkillMatch): 30%
- 工作经验 (ExperienceMatch): 25%
- 职责匹配 (ResponsibilityMatch): 20%
- 基本信息 (BasicMatch): 15%
- 行业背景 (IndustryMatch): 7%
- 教育背景 (EducationMatch): 3%
```

**匹配等级定义**：
```
- 优秀匹配 (Excellent): 85-100分
- 良好匹配 (Good): 70-84分
- 一般匹配 (Fair): 55-69分
- 较差匹配 (Poor): 40-54分
- 不匹配 (No Match): 0-39分
```

### 6.4 输出格式规范

#### 6.4.1 通用输出结构

所有 Prompt 都遵循以下 JSON 输出格式规范：

```json
{
  "score": 0.0,                    // 总分（0-100浮点数）
  "sub_scores": {                  // 子维度分数
    "dimension1": 0.0,
    "dimension2": 0.0
  },
  "detailed_analysis": {           // 详细分析结果
    // 具体维度的分析数据
  },
  "reasoning": "评分理由说明"        // 评分依据和分析
}
```

#### 6.4.2 各维度特定输出格式

**基本信息匹配输出**：
```json
{
  "score": 85.5,
  "subScores": {
    "location": 90.0,
    "salary": 80.0,
    "department": 85.0
  },
  "reasoning": "详细的评分理由说明"
}
```

**技能匹配输出**：
```json
{
  "score": 78.5,
  "matched_skills": [
    {
      "job_skill_id": "skill_001",
      "resume_skill_id": "resume_skill_001",
      "match_type": "exact",
      "llm_score": 95.0,
      "proficiency_gap": 10.0,
      "score": 90.0,
      "llm_analysis": {
        "match_level": "perfect",
        "match_percentage": 95.0,
        "proficiency_gap": "minor",
        "transferability": "high",
        "learning_effort": "minimal",
        "match_reason": "技能完全匹配"
      }
    }
  ],
  "missing_skills": [
    {
      "id": "skill_002",
      "name": "React",
      "priority": "high",
      "category": "frontend"
    }
  ],
  "extra_skills": ["Vue.js", "Angular"],
  "llm_analysis": {
    "overall_match": 78.5,
    "technical_fit": 80.0,
    "learning_curve": "medium",
    "strength_areas": ["后端开发", "数据库设计"],
    "gap_areas": ["前端框架", "移动开发"],
    "recommendations": ["建议学习React框架", "加强前端技能"],
    "analysis_detail": "候选人在后端技能方面表现优秀..."
  }
}
```

### 6.5 最佳实践

#### 6.5.1 Prompt 编写规范

1. **角色定义清晰**：明确 AI 的专业身份和职责范围
2. **评分标准具体**：提供详细的分数区间和评判标准
3. **输出格式严格**：使用 JSON Schema 确保输出结构一致
4. **权重分配合理**：基于业务重要性设定各维度权重
5. **异常处理完善**：考虑数据缺失、格式错误等边界情况

#### 6.5.2 模板维护原则

1. **版本控制**：对 Prompt 变更进行版本管理
2. **A/B 测试**：通过对比测试验证 Prompt 效果
3. **持续优化**：基于实际使用效果不断改进
4. **文档同步**：确保代码和文档的一致性

#### 6.5.3 质量保证措施

1. **输出验证**：确保 JSON 格式正确且字段完整
2. **分数校验**：验证所有分数在 0-100 范围内
3. **逻辑一致性**：检查各维度评分的逻辑合理性
4. **边界测试**：测试极端情况下的 Prompt 表现

### 6.6 扩展性设计

#### 6.6.1 新维度添加

当需要添加新的匹配维度时，遵循以下步骤：

1. **定义评分规则**：制定该维度的具体评分标准
2. **设计输出格式**：定义 JSON 输出结构
3. **创建 Prompt 文件**：按照统一模板创建新的 Prompt
4. **更新聚合器**：在 Aggregator 中添加新维度的权重
5. **测试验证**：确保新维度与现有系统兼容

#### 6.6.2 多语言支持

为支持多语言环境，可以考虑：

1. **模板国际化**：将 Prompt 文本提取为可配置的资源文件
2. **动态加载**：根据用户语言偏好加载对应的 Prompt 模板
3. **输出本地化**：确保评分说明和建议符合本地化要求

#### 6.6.3 个性化定制

支持不同企业或行业的个性化需求：

1. **权重配置**：允许自定义各维度的权重分配
2. **评分标准调整**：支持调整评分区间和标准
3. **行业特化**：针对特定行业优化 Prompt 内容

通过以上设计，我们建立了一套完整、规范、可扩展的 Prompt 模板体系，确保智能匹配系统的准确性和一致性。
