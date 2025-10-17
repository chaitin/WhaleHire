package repo

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/db/screeningnoderun"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/entx"
)

// ScreeningNodeRunRepo 智能筛选节点运行记录仓储实现
type ScreeningNodeRunRepo struct {
	db *db.Client
}

// NewScreeningNodeRunRepo 构建仓储
func NewScreeningNodeRunRepo(client *db.Client) domain.ScreeningNodeRunRepo {
	return &ScreeningNodeRunRepo{db: client}
}

// Create 创建节点运行记录
func (r *ScreeningNodeRunRepo) Create(ctx context.Context, req *domain.CreateNodeRunRepoReq) (*db.ScreeningNodeRun, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	builder := r.db.ScreeningNodeRun.Create().
		SetTaskID(req.TaskID).
		SetTaskResumeID(req.TaskResumeID).
		SetNodeKey(req.NodeKey).
		SetStatus(string(req.Status)).
		SetAttemptNo(req.AttemptNo)

	// 设置可选字段
	if req.TraceID != nil {
		builder = builder.SetTraceID(*req.TraceID)
	}
	if req.AgentVersion != nil {
		builder = builder.SetAgentVersion(*req.AgentVersion)
	}
	if req.ModelName != nil {
		builder = builder.SetModelName(*req.ModelName)
	}
	if req.ModelProvider != nil {
		builder = builder.SetModelProvider(*req.ModelProvider)
	}
	if req.LLMParams != nil {
		builder = builder.SetLlmParams(req.LLMParams)
	}
	if req.InputPayload != nil {
		builder = builder.SetInputPayload(req.InputPayload)
	}
	if req.OutputPayload != nil {
		builder = builder.SetOutputPayload(req.OutputPayload)
	}

	entity, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create screening node run failed: %w", err)
	}
	return entity, nil
}

// Update 更新节点运行记录
func (r *ScreeningNodeRunRepo) Update(ctx context.Context, id uuid.UUID, fn func(tx *db.Tx, current *db.ScreeningNodeRun, updater *db.ScreeningNodeRunUpdateOne) error) (*db.ScreeningNodeRun, error) {
	// 在事务中执行更新操作
	var result *db.ScreeningNodeRun
	err := entx.WithTx(ctx, r.db, func(tx *db.Tx) error {
		// 获取当前记录
		current, err := tx.ScreeningNodeRun.Get(ctx, id)
		if err != nil {
			return fmt.Errorf("get current record failed: %w", err)
		}

		// 创建更新器
		updater := tx.ScreeningNodeRun.UpdateOneID(id)

		// 执行传入的更新函数
		if err := fn(tx, current, updater); err != nil {
			return fmt.Errorf("update function failed: %w", err)
		}

		// 保存更新
		updated, err := updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("save update failed: %w", err)
		}

		result = updated
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("update screening node run failed: %w", err)
	}

	return result, nil
}

// GetByTaskResumeAndNode 根据任务简历ID和节点键查询节点运行记录
func (r *ScreeningNodeRunRepo) GetByTaskResumeAndNode(ctx context.Context, taskResumeID uuid.UUID, nodeKey string, attemptNo int) (*db.ScreeningNodeRun, error) {
	entity, err := r.db.ScreeningNodeRun.Query().
		Where(
			screeningnoderun.TaskResumeID(taskResumeID),
			screeningnoderun.NodeKey(nodeKey),
			screeningnoderun.AttemptNo(attemptNo),
		).
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("get screening node run by task resume and node failed: %w", err)
	}
	return entity, nil
}

// GetByID 根据ID查询节点运行记录
func (r *ScreeningNodeRunRepo) GetByID(ctx context.Context, id uuid.UUID) (*db.ScreeningNodeRun, error) {
	entity, err := r.db.ScreeningNodeRun.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get screening node run by id failed: %w", err)
	}
	return entity, nil
}

// List 查询节点运行记录列表
func (r *ScreeningNodeRunRepo) List(ctx context.Context, req *domain.ListNodeRunsRepoReq) ([]*db.ScreeningNodeRun, *db.PageInfo, error) {
	query := r.db.ScreeningNodeRun.Query()

	// 应用过滤条件
	if req != nil && req.ListNodeRunsReq != nil {
		if req.TaskID != nil {
			query = query.Where(screeningnoderun.TaskID(*req.TaskID))
		}
		if req.TaskResumeID != nil {
			query = query.Where(screeningnoderun.TaskResumeID(*req.TaskResumeID))
		}
		if req.NodeKey != nil {
			query = query.Where(screeningnoderun.NodeKey(*req.NodeKey))
		}
		if req.Status != nil {
			query = query.Where(screeningnoderun.Status(string(*req.Status)))
		}
	}

	// 排序
	query = query.Order(screeningnoderun.ByCreatedAt(sql.OrderDesc()))

	// 分页
	page := 1
	pageSize := 20
	if req != nil && req.ListNodeRunsReq != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.Size > 0 {
			pageSize = req.Size
		}
	}

	items, pageInfo, err := query.Page(ctx, page, pageSize)
	if err != nil {
		return nil, nil, fmt.Errorf("list screening node runs failed: %w", err)
	}

	return items, pageInfo, nil
}

// GetByTaskResumeID 根据任务简历ID查询节点运行记录列表
func (r *ScreeningNodeRunRepo) GetByTaskResumeID(ctx context.Context, taskResumeID uuid.UUID) ([]*db.ScreeningNodeRun, error) {
	entities, err := r.db.ScreeningNodeRun.Query().
		Where(screeningnoderun.TaskResumeID(taskResumeID)).
		Order(screeningnoderun.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get screening node runs by task resume id failed: %w", err)
	}
	return entities, nil
}

// GetByTraceID 根据TraceID查询节点运行记录列表
func (r *ScreeningNodeRunRepo) GetByTraceID(ctx context.Context, traceID string) ([]*db.ScreeningNodeRun, error) {
	entities, err := r.db.ScreeningNodeRun.Query().
		Where(screeningnoderun.TraceID(traceID)).
		Order(screeningnoderun.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get screening node runs by trace id failed: %w", err)
	}
	return entities, nil
}

// DeleteByTaskID 根据任务ID删除节点运行记录
func (r *ScreeningNodeRunRepo) DeleteByTaskID(ctx context.Context, taskID uuid.UUID) error {
	_, err := r.db.ScreeningNodeRun.Delete().
		Where(screeningnoderun.TaskID(taskID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete screening node runs by task id failed: %w", err)
	}
	return nil
}
