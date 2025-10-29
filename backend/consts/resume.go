package consts

type UniversityType string

const (
	// 基础类型
	UniversityTypeOrdinary         UniversityType = "ordinary"           // 普通学校
	UniversityType211              UniversityType = "211"                // 211工程学校
	UniversityType985              UniversityType = "985"                // 985工程学校
	UniversityTypeDoubleFirstClass UniversityType = "double_first_class" // 双一流学校
	UniversityTypeQSTop100         UniversityType = "qs_top100"          // QS Top 100
)

// Values 返回所有高校类型值
func (UniversityType) Values() []UniversityType {
	return []UniversityType{
		UniversityTypeOrdinary,
		UniversityType211,
		UniversityType985,
		UniversityTypeDoubleFirstClass,
		UniversityTypeQSTop100,
	}
}

// IsValid 检查高校类型是否有效
func (u UniversityType) IsValid() bool {
	for _, v := range UniversityType("").Values() {
		if u == v {
			return true
		}
	}
	return false
}

// UniversityMatchSource 高校匹配来源
type UniversityMatchSource string

const (
	UniversityMatchSourceExact  UniversityMatchSource = "exact"  // 精确匹配
	UniversityMatchSourceVector UniversityMatchSource = "vector" // 向量匹配
	UniversityMatchSourceManual UniversityMatchSource = "manual" // 手动匹配
)

// Values 返回所有匹配来源值
func (UniversityMatchSource) Values() []UniversityMatchSource {
	return []UniversityMatchSource{
		UniversityMatchSourceExact,
		UniversityMatchSourceVector,
		UniversityMatchSourceManual,
	}
}

// IsValid 检查匹配来源是否有效
func (u UniversityMatchSource) IsValid() bool {
	for _, v := range UniversityMatchSource("").Values() {
		if u == v {
			return true
		}
	}
	return false
}

// ProjectType 项目类型
type ProjectType string

const (
	ProjectTypePersonal   ProjectType = "personal"   // 个人项目
	ProjectTypeTeam       ProjectType = "team"       // 团队项目
	ProjectTypeOpenSource ProjectType = "opensource" // 开源项目
	ProjectTypePaper      ProjectType = "paper"      // 论文发表
	ProjectTypeOther      ProjectType = "other"      // 其他类型
)

// Values 返回所有项目类型值
func (ProjectType) Values() []ProjectType {
	return []ProjectType{
		ProjectTypePersonal,
		ProjectTypeTeam,
		ProjectTypeOpenSource,
		ProjectTypePaper,
		ProjectTypeOther,
	}
}

// IsValid 检查项目类型是否有效
func (p ProjectType) IsValid() bool {
	for _, v := range ProjectType("").Values() {
		if p == v {
			return true
		}
	}
	return false
}

// ExperienceType 经验类型
type ExperienceType string

const (
	ExperienceTypeWork         ExperienceType = "work"         // 工作经历
	ExperienceTypeOrganization ExperienceType = "organization" // 社团组织
	ExperienceTypeVolunteer    ExperienceType = "volunteer"    // 志愿服务
	ExperienceTypeInternship   ExperienceType = "internship"   // 实习经历
)

// Values 返回所有经验类型值
func (ExperienceType) Values() []ExperienceType {
	return []ExperienceType{
		ExperienceTypeWork,
		ExperienceTypeOrganization,
		ExperienceTypeVolunteer,
		ExperienceTypeInternship,
	}
}

// IsValid 检查经验类型是否有效
func (e ExperienceType) IsValid() bool {
	for _, v := range ExperienceType("").Values() {
		if e == v {
			return true
		}
	}
	return false
}

const (
	// 批量上传相关Redis键值
	BatchUploadTaskKeyFmt = "batch_upload:task:%s"
	BatchUploadItemKeyFmt = "batch_upload:item:%s:%s" // taskID:itemID
)
