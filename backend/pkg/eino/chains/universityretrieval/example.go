//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/universityretrieval"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	_ "github.com/lib/pq"
)

func main() {
	query := flag.String("query", "", "要检索的高校名称，示例：清华大学")
	topK := flag.Int("topk", 3, "返回的候选数量")
	threshold := flag.Float64("threshold", 0.5, "相似度阈值（0~1，缺省不启用）")
	timeout := flag.Duration("timeout", 30*time.Second, "调用超时时间")
	flag.Parse()

	if strings.TrimSpace(*query) == "" {
		fmt.Fprintln(os.Stderr, "请通过 -query 指定高校名称")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	cfg, err := config.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化配置失败: %v\n", err)
		os.Exit(1)
	}

	// 创建嵌入模型
	embedder, _, err := embedding.NewEmbedding(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建嵌入模型失败: %v\n", err)
		os.Exit(1)
	}

	// 创建检索器，传入数据库连接和嵌入模型
	ret, err := universityretrieval.NewRetriever(ctx,
		universityretrieval.WithDataSource(cfg.Database.Master),
		universityretrieval.WithEmbedding(embedder, cfg.Embedding.Dimension),
		universityretrieval.WithTopK(cfg.Retriever.TopK),
		universityretrieval.WithDistanceThreshold(cfg.Retriever.DistanceThreshold),
		universityretrieval.WithExactMatch(true),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "构建高校召回器失败: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if closer, ok := ret.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}()

	normalized, err := universityretrieval.NormalizeQuery(*query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "高校名称校验失败: %v\n", err)
		os.Exit(1)
	}

	var opts []retriever.Option
	if *topK > 0 {
		opts = append(opts, retriever.WithTopK(*topK))
	}
	if !math.IsNaN(*threshold) {
		if *threshold < 0 || *threshold > 1 {
			fmt.Fprintln(os.Stderr, "threshold 需处于 0~1 区间")
			os.Exit(1)
		}
		opts = append(opts, retriever.WithScoreThreshold(*threshold))
	}

	fmt.Printf("开始检索高校：%s（TopK=%d）\n", normalized, *topK)
	docs, err := ret.Retrieve(ctx, normalized, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "召回失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("共召回 %d 条结果：\n", len(docs))
	for idx, doc := range docs {
		fmt.Printf("[%d] %s (ID=%s, Score=%.4f)\n", idx+1, doc.Content, doc.ID, doc.Score())
		if len(doc.MetaData) > 0 {
			meta, _ := json.MarshalIndent(doc.MetaData, "", "  ")
			fmt.Printf("    元数据: %s\n", string(meta))
		}
	}
}
