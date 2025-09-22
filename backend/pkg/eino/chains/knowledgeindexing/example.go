//go:build ignore
// +build ignore

package main

import (
	"context"
	"log"
	"time"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/knowledgeindexing"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/loaddoc"
	"github.com/chaitin/WhaleHire/backend/pkg/store"
	"github.com/cloudwego/eino/compose"
)

func main() {
	ctx := context.Background()
	// 获取配置
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}
	err = store.InitVectorIndex(ctx, cfg)
	if err != nil {
		panic(err)
	}

	g := compose.NewGraph[*loaddoc.BatchLoadInput, []string]()

	chainLoad, err := loaddoc.NewParallelBatchLoadChain(ctx, 30*time.Second, 5)
	if err != nil {
		log.Fatalf("Failed to create parallel batch load chain: %v", err)
	}

	g.AddGraphNode("loaddoc", chainLoad.GetChain())

	chainKnowledge, err := knowledgeindexing.NewKnowledgeIndexingChain(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create knowledge indexing chain: %v", err)
	}

	g.AddGraphNode("knowledgeindexing", chainKnowledge.GetChain())

	_ = g.AddEdge(compose.START, "loaddoc")
	_ = g.AddEdge("loaddoc", "knowledgeindexing")
	_ = g.AddEdge("knowledgeindexing", compose.END)

	r, err := g.Compile(ctx)
	if err != nil {
		log.Fatalf("Compile failed, err=%v", err)
		return
	}

	out, err := r.Invoke(ctx, &loaddoc.BatchLoadInput{
		URLs: []string{
			"https://www.cloudwego.io/zh/docs/eino",
			"https://www.cloudwego.io/zh/docs/eino/core_modules/chain_and_graph_orchestration/chain_graph_introduction/#graph",
		},
	})
	if err != nil {
		log.Fatalf("Invoke failed, err=%v", err)
		return
	}

	log.Printf("knowledgeindexing output: %v", out)
}
