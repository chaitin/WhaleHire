package screening

import (
	"context"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/compose"
)

// InvokeWithCollector 结合输出收集器执行图，便于一次性注入回调选项
func InvokeWithCollector(
	ctx context.Context,
	runnable compose.Runnable[*domain.MatchInput, *domain.JobResumeMatch],
	input *domain.MatchInput,
	collector *AgentOutputCollector,
	extra ...compose.Option,
) (*domain.JobResumeMatch, error) {
	options := make([]compose.Option, 0, len(extra))
	options = append(options, extra...)
	if collector != nil {
		options = append(options, collector.ComposeOptions()...)
	}
	return runnable.Invoke(ctx, input, options...)
}
