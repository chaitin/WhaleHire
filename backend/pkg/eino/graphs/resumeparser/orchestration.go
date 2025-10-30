package resumeparsergraph

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/chaitin/WhaleHire/backend/config"
	chainresume "github.com/chaitin/WhaleHire/backend/pkg/eino/chains/resumeparser"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	_ "github.com/lib/pq"
)

// ResumeParseGraph 组合简历解析 LLM 与字段增强节点的图。
type ResumeParseGraph struct {
	graph   *compose.Graph[*chainresume.ResumeParseInput, *chainresume.ResumeParseResult]
	eduNode *educationNode
}

// GetGraph 返回底层图结构。
func (g *ResumeParseGraph) GetGraph() *compose.Graph[*chainresume.ResumeParseInput, *chainresume.ResumeParseResult] {
	return g.graph
}

// Compile 将图编译为可执行对象。
func (g *ResumeParseGraph) Compile(ctx context.Context) (compose.Runnable[*chainresume.ResumeParseInput, *chainresume.ResumeParseResult], error) {
	return g.graph.Compile(ctx)
}

// NewResumeParseGraph 构建新的简历解析图。
func NewResumeParseGraph(ctx context.Context, cfg *config.Config, chatModel model.ToolCallingChatModel, logger *slog.Logger) (*ResumeParseGraph, error) {
	// 创建数据库连接
	db, err := sql.Open("postgres", cfg.Database.Master)
	if err != nil {
		return nil, fmt.Errorf("resume graph: 创建数据库连接失败: %w", err)
	}

	llmChain, err := chainresume.NewResumeParserChain(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("resume graph: 创建 LLM 链失败: %w", err)
	}

	dispatcher := NewDispatcher()
	aggregator := NewAggregator()

	eduNode, err := newEducationNode(ctx, db, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("resume graph: 创建教育增强节点失败: %w", err)
	}

	graph := compose.NewGraph[*chainresume.ResumeParseInput, *chainresume.ResumeParseResult]()

	if err := graph.AddGraphNode(nodeResumeLLM, llmChain.GetChain(), compose.WithNodeName(nodeResumeLLM)); err != nil {
		return nil, fmt.Errorf("resume graph: 注册 LLM 节点失败: %w", err)
	}

	if err := graph.AddLambdaNode(nodeDispatcher, compose.InvokableLambda(dispatcher.Process), compose.WithNodeName(nodeDispatcher)); err != nil {
		return nil, fmt.Errorf("resume graph: 注册分发节点失败: %w", err)
	}

	if err := graph.AddLambdaNode(nodeEducation, compose.InvokableLambda(eduNode.Process), compose.WithNodeName(nodeEducation)); err != nil {
		return nil, fmt.Errorf("resume graph: 注册教育增强节点失败: %w", err)
	}

	if err := graph.AddLambdaNode(nodeAggregator, compose.InvokableLambda(aggregator.Process), compose.WithNodeName(nodeAggregator)); err != nil {
		return nil, fmt.Errorf("resume graph: 注册聚合节点失败: %w", err)
	}

	// 按顺序串联各节点
	_ = graph.AddEdge(compose.START, nodeResumeLLM)
	_ = graph.AddEdge(nodeResumeLLM, nodeDispatcher)
	_ = graph.AddEdge(nodeDispatcher, nodeEducation)
	_ = graph.AddEdge(nodeEducation, nodeAggregator)
	_ = graph.AddEdge(nodeAggregator, compose.END)

	return &ResumeParseGraph{
		graph:   graph,
		eduNode: eduNode,
	}, nil
}
