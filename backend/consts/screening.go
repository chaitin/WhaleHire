package consts

// ScreeningTaskStatus 筛选任务状态
type ScreeningTaskStatus string

const (
	ScreeningTaskStatusPending   ScreeningTaskStatus = "pending"   // 待处理
	ScreeningTaskStatusRunning   ScreeningTaskStatus = "running"   // 运行中
	ScreeningTaskStatusCompleted ScreeningTaskStatus = "completed" // 已完成
	ScreeningTaskStatusFailed    ScreeningTaskStatus = "failed"    // 失败
	ScreeningTaskStatusCancelled ScreeningTaskStatus = "cancelled" // 已取消
)

// ScreeningTaskResumeStatus 筛选任务简历状态
type ScreeningTaskResumeStatus string

const (
	ScreeningTaskResumeStatusPending   ScreeningTaskResumeStatus = "pending"   // 待处理
	ScreeningTaskResumeStatusRunning   ScreeningTaskResumeStatus = "running"   // 处理中
	ScreeningTaskResumeStatusCompleted ScreeningTaskResumeStatus = "completed" // 已完成
	ScreeningTaskResumeStatusFailed    ScreeningTaskResumeStatus = "failed"    // 失败
	ScreeningTaskResumeStatusCancelled ScreeningTaskResumeStatus = "cancelled" // 已取消
)

// MatchLevel 匹配等级
type MatchLevel string

/*
- **优秀匹配 (Excellent)**: 85-100分
  - 各维度表现均衡，核心技能完全匹配
  - 工作经验丰富且高度相关
  - 具备快速上手和创造价值的能力

- **良好匹配 (Good)**: 70-84分
  - 核心技能基本匹配，少量技能需要培养
  - 工作经验相关性较高
  - 经过短期适应可胜任岗位要求

- **一般匹配 (Fair)**: 55-69分
  - 部分技能匹配，存在明显技能缺口
  - 工作经验有一定相关性
  - 需要较长时间培训和适应

- **较差匹配 (Poor)**: 40-54分
  - 技能匹配度较低，缺乏核心技能
  - 工作经验相关性不高
  - 需要大量培训投入，风险较高

- **不匹配 (No Match)**: 0-39分
  - 技能严重不匹配，缺乏基本胜任能力
  - 工作经验与岗位要求差距巨大
  - 不建议录用
*/
const (
	MatchLevelExcellent MatchLevel = "excellent" // 优秀
	MatchLevelGood      MatchLevel = "good"      // 良好
	MatchLevelFair      MatchLevel = "fair"      // 一般
	MatchLevelPoor      MatchLevel = "poor"      // 较差
	MatchLevelNoMatch   MatchLevel = "no_match"  // 不匹配
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

// ScreeningNodeRunStatus 筛选节点运行状态
type ScreeningNodeRunStatus string

const (
	ScreeningNodeRunStatusPending   ScreeningNodeRunStatus = "pending"   // 待处理
	ScreeningNodeRunStatusRunning   ScreeningNodeRunStatus = "running"   // 运行中
	ScreeningNodeRunStatusCompleted ScreeningNodeRunStatus = "completed" // 已完成
	ScreeningNodeRunStatusFailed    ScreeningNodeRunStatus = "failed"    // 失败
)

// Values 返回所有筛选节点运行状态值
func (ScreeningNodeRunStatus) Values() []ScreeningNodeRunStatus {
	return []ScreeningNodeRunStatus{
		ScreeningNodeRunStatusPending,
		ScreeningNodeRunStatusRunning,
		ScreeningNodeRunStatusCompleted,
		ScreeningNodeRunStatusFailed,
	}
}

// IsValid 检查筛选节点运行状态是否有效
func (s ScreeningNodeRunStatus) IsValid() bool {
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
		ScreeningTaskResumeStatusCancelled,
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
