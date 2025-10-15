package taskmeta

import (
	"context"
	"errors"

	"github.com/chaitin/WhaleHire/backend/domain"
)

var (
	// ErrInvalidInput 无效输入错误
	ErrInvalidInput = errors.New("invalid input")
	// ErrMissingRequiredFields 缺少必需字段错误
	ErrMissingRequiredFields = errors.New("missing required fields")
)

// TaskMetaProcessor 任务元数据处理器
type TaskMetaProcessor struct{}

// NewTaskMetaProcessor 创建任务元数据处理器
func NewTaskMetaProcessor() *TaskMetaProcessor {
	return &TaskMetaProcessor{}
}

// NewLambdaTaskMeta 做数据转换，增强任务元数据
func NewLambdaTaskMeta(ctx context.Context, input *domain.TaskMetaData) (output *domain.TaskMetaData, err error) {
	// 增强任务元数据，添加时间戳和处理状态
	enhanced := &domain.TaskMetaData{
		JobID:       input.JobID,
		ResumeID:    input.ResumeID,
		MatchTaskID: input.MatchTaskID,
		// 可以在这里添加更多元数据字段，如处理时间戳等
	}

	return enhanced, nil
}

// ProcessTaskMeta 处理任务元数据，为各Agent提供统一的任务上下文
func (p *TaskMetaProcessor) ProcessTaskMeta(ctx context.Context, input *domain.TaskMetaData) (*domain.TaskMetaData, error) {
	// 验证输入数据
	if input == nil {
		return nil, ErrInvalidInput
	}

	if input.JobID == "" || input.ResumeID == "" {
		return nil, ErrMissingRequiredFields
	}

	// 创建增强的任务元数据
	enhanced := &domain.TaskMetaData{
		JobID:       input.JobID,
		ResumeID:    input.ResumeID,
		MatchTaskID: input.MatchTaskID,
	}

	return enhanced, nil
}
