package consts

// ScreeningTaskStatus 筛选任务状态
type ScreeningTaskStatus string

const (
	ScreeningTaskStatusPending   ScreeningTaskStatus = "pending"   // 待处理
	ScreeningTaskStatusRunning   ScreeningTaskStatus = "running"   // 运行中
	ScreeningTaskStatusCompleted ScreeningTaskStatus = "completed" // 已完成
	ScreeningTaskStatusFailed    ScreeningTaskStatus = "failed"    // 失败
)

// ScreeningTaskResumeStatus 筛选任务简历状态
type ScreeningTaskResumeStatus string

const (
	ScreeningTaskResumeStatusPending   ScreeningTaskResumeStatus = "pending"   // 待处理
	ScreeningTaskResumeStatusRunning   ScreeningTaskResumeStatus = "running"   // 处理中
	ScreeningTaskResumeStatusCompleted ScreeningTaskResumeStatus = "completed" // 已完成
	ScreeningTaskResumeStatusFailed    ScreeningTaskResumeStatus = "failed"    // 失败
)

// MatchLevel 匹配等级
type MatchLevel string

const (
	MatchLevelExcellent MatchLevel = "excellent" // 优秀
	MatchLevelGood      MatchLevel = "good"      // 良好
	MatchLevelFair      MatchLevel = "fair"      // 一般
	MatchLevelPoor      MatchLevel = "poor"      // 较差
)

// Values 返回所有筛选任务状态值
func (ScreeningTaskStatus) Values() []ScreeningTaskStatus {
	return []ScreeningTaskStatus{
		ScreeningTaskStatusPending,
		ScreeningTaskStatusRunning,
		ScreeningTaskStatusCompleted,
		ScreeningTaskStatusFailed,
	}
}

// IsValid 检查筛选任务状态是否有效
func (s ScreeningTaskStatus) IsValid() bool {
	for _, v := range s.Values() {
		if s == v {
			return true
		}
	}
	return false
}

// Values 返回所有筛选任务简历状态值
func (ScreeningTaskResumeStatus) Values() []ScreeningTaskResumeStatus {
	return []ScreeningTaskResumeStatus{
		ScreeningTaskResumeStatusPending,
		ScreeningTaskResumeStatusRunning,
		ScreeningTaskResumeStatusCompleted,
		ScreeningTaskResumeStatusFailed,
	}
}

// IsValid 检查筛选任务简历状态是否有效
func (s ScreeningTaskResumeStatus) IsValid() bool {
	for _, v := range s.Values() {
		if s == v {
			return true
		}
	}
	return false
}

// Values 返回所有匹配等级值
func (MatchLevel) Values() []MatchLevel {
	return []MatchLevel{
		MatchLevelExcellent,
		MatchLevelGood,
		MatchLevelFair,
		MatchLevelPoor,
	}
}

// IsValid 检查匹配等级是否有效
func (m MatchLevel) IsValid() bool {
	for _, v := range m.Values() {
		if m == v {
			return true
		}
	}
	return false
}
