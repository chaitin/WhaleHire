package knowledgeindexing

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
)

func NewIDGenerator(ctx context.Context, originalID string, index int) string {
	return fmt.Sprintf("%s_%d", originalID, index)
}

func NewDocumentTransformer(ctx context.Context) (tfr document.Transformer, err error) {
	// 初始化分割器
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   1000,
		OverlapSize: 200,
		Separators:  []string{"\n\n", "\n", "。", "！", "？"},
		KeepType:    recursive.KeepTypeEnd,
		IDGenerator: NewIDGenerator,
	})
	if err != nil {
		return nil, err
	}

	return splitter, nil
}
