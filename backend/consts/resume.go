package consts

type UniversityType string

const (
	UniversityTypeOrdinary UniversityType = "ordinary" // 普通学校
	UniversityType211      UniversityType = "211"      // 211工程学校
	UniversityType985      UniversityType = "985"      // 985工程学校
)

// UniversityBadge 高校标签类型
type UniversityBadge string

const (
	UniversityBadgeDoubleFirstClass UniversityBadge = "double_first_class" // 双一流
	UniversityBadge985              UniversityBadge = "project_985"        // 985工程
	UniversityBadge211              UniversityBadge = "project_211"        // 211工程
	UniversityBadgeQSTop100         UniversityBadge = "qs_top100"          // QS Top 100
)

// Values 返回所有高校标签值
func (UniversityBadge) Values() []UniversityBadge {
	return []UniversityBadge{
		UniversityBadgeDoubleFirstClass,
		UniversityBadge985,
		UniversityBadge211,
		UniversityBadgeQSTop100,
	}
}

// IsValid 检查高校标签是否有效
func (u UniversityBadge) IsValid() bool {
	for _, v := range UniversityBadge("").Values() {
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

const (
	// 批量上传相关Redis键值
	BatchUploadTaskKeyFmt = "batch_upload:task:%s"
	BatchUploadItemKeyFmt = "batch_upload:item:%s:%s" // taskID:itemID
)
