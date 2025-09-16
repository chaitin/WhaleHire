package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
)

func defaultDDGSearchConfig(ctx context.Context) (*duckduckgo.Config, error) {
	config := &duckduckgo.Config{}
	return config, nil
}

func NewDDGSearchTool(ctx context.Context, config *duckduckgo.Config) (tn tool.BaseTool, err error) {
	if config == nil {
		config, err = defaultDDGSearchConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	tn, err = duckduckgo.NewTextSearchTool(ctx, config)
	if err != nil {
		return nil, err
	}
	return tn, nil
}
