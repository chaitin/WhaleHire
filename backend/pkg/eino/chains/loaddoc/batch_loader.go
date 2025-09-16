package loaddoc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ParallelBatchLoader 并行批量加载器（使用更细粒度的并行处理）
type ParallelBatchLoader struct {
	fileLoader *FileLoader
	urlLoader  *URLLoader
	maxWorkers int
	workerPool chan struct{}
}

// NewParallelBatchLoader 创建并行批量加载器
func NewParallelBatchLoader(urlTimeout time.Duration, maxWorkers int) *ParallelBatchLoader {
	if maxWorkers <= 0 {
		maxWorkers = 10 // 默认10个worker
	}

	return &ParallelBatchLoader{
		fileLoader: NewFileLoader(),
		urlLoader:  NewURLLoader(urlTimeout),
		maxWorkers: maxWorkers,
		workerPool: make(chan struct{}, maxWorkers),
	}
}

// ParallelBatchLoadChain 并行批量加载链
type ParallelBatchLoadChain struct {
	loader *ParallelBatchLoader
	chain  *compose.Chain[*BatchLoadInput, []*schema.Document]
}

// GetChain 获取处理链
func (pblc *ParallelBatchLoadChain) GetChain() *compose.Chain[*BatchLoadInput, []*schema.Document] {
	return pblc.chain
}

// Compile 编译链为可执行的Runnable
func (pblc *ParallelBatchLoadChain) Compile(ctx context.Context) (compose.Runnable[*BatchLoadInput, []*schema.Document], error) {
	runnable, err := pblc.chain.Compile(ctx)
	if err != nil {
		return nil, err
	}
	return runnable, nil
}

// parallelBatchLoadLambda 并行批量加载Lambda函数
func (pbl *ParallelBatchLoader) parallelBatchLoadLambda(ctx context.Context, input *BatchLoadInput, opts ...any) (*schema.StreamReader[[]*schema.Document], error) {
	// 使用 schema.Pipe 创建 StreamReader 和 StreamWriter
	reader, writer := schema.Pipe[[]*schema.Document](1)

	fmt.Println(input)

	go func() {
		defer writer.Close()

		var wg sync.WaitGroup

		// 处理文件的worker函数
		processFile := func(filePath string) {
			defer wg.Done()

			// 获取worker slot
			pbl.workerPool <- struct{}{}
			defer func() { <-pbl.workerPool }()

			doc, err := pbl.fileLoader.LoadFile(ctx, filePath)
			if err == nil {
				// 转换为 schema.Document
				schemaDoc := &schema.Document{
					ID:       doc.ID,
					Content:  doc.Content,
					MetaData: doc.Metadata,
				}
				// 立即发送单个文档结果
				writer.Send([]*schema.Document{schemaDoc}, nil)
			}
		}

		// 处理URL的worker函数
		processURL := func(urlStr string) {
			defer wg.Done()

			// 获取worker slot
			pbl.workerPool <- struct{}{}
			defer func() { <-pbl.workerPool }()

			doc, err := pbl.urlLoader.LoadURL(ctx, urlStr)
			if err == nil {
				// 转换为 schema.Document
				schemaDoc := &schema.Document{
					ID:       doc.ID,
					Content:  doc.Content,
					MetaData: doc.Metadata,
				}
				// 立即发送单个文档结果
				writer.Send([]*schema.Document{schemaDoc}, nil)
			}
		}

		// 启动文件处理goroutines（仅当FilePaths不为空时）
		if len(input.FilePaths) > 0 {
			for _, filePath := range input.FilePaths {
				wg.Add(1)
				go processFile(filePath)
			}
		}

		// 启动URL处理goroutines（仅当URLs不为空时）
		if len(input.URLs) > 0 {
			for _, urlStr := range input.URLs {
				wg.Add(1)
				go processURL(urlStr)
			}
		}

		// 等待所有worker完成
		wg.Wait()
	}()

	return reader, nil
}

// NewParallelBatchLoadChain 创建并行批量加载链
func NewParallelBatchLoadChain(ctx context.Context, urlTimeout time.Duration, maxWorkers int) (*ParallelBatchLoadChain, error) {
	loader := NewParallelBatchLoader(urlTimeout, maxWorkers)

	// 创建处理链
	chain := compose.NewChain[*BatchLoadInput, []*schema.Document]()
	chain.AppendLambda(
		compose.StreamableLambdaWithOption(loader.parallelBatchLoadLambda),
		compose.WithNodeName("parallel_batch_document_loader"),
	)

	return &ParallelBatchLoadChain{
		loader: loader,
		chain:  chain,
	}, nil
}
