//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ptonlix/whalehire/backend/pkg/eino/chains/loaddoc"
)

func main() {
	loaddoc.Init()
	ctx := context.Background()

	// 示例1: 使用并行批量加载链（批量模式）
	// fmt.Println("=== 并行批量加载链示例（批量模式）===")
	// parallelExample(ctx)

	fmt.Println("\n=== 流式响应示例 ===")
	// 示例2: 使用流式响应
	streamExample(ctx)

	// fmt.Println("\n=== 实时处理示例 ===")
	// // 示例3: 实时处理每个文档
	// realTimeExample(ctx)
}

// streamExample 流式响应示例
func streamExample(ctx context.Context) {
	// 创建并行批量加载链（最多5个并发worker）
	chain, err := loaddoc.NewParallelBatchLoadChain(ctx, 30*time.Second, 5)
	if err != nil {
		log.Fatalf("Failed to create parallel batch load chain: %v", err)
	}

	// 编译链
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("Failed to compile chain: %v", err)
	}

	// 准备输入数据
	input := &loaddoc.BatchLoadInput{
		// FilePaths: []string{
		// 	"./README.md",
		// 	"./go.mod",
		// },
		URLs: []string{
			"https://www.cloudwego.io/zh/docs/eino",
			"https://www.cloudwego.io/zh/docs/eino/core_modules/chain_and_graph_orchestration/chain_graph_introduction/#graph",
		},
	}

	// 使用流式处理
	fmt.Println("开始流式处理文档...")

	// 执行流式加载
	result, err := runnable.Stream(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute stream load: %v", err)
	}

	// 实时处理流式结果
	docCount := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Println("处理被取消")
			return
		default:
			// 从流中读取结果
			docs, err := result.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					fmt.Printf("流式处理完成，共处理 %d 个文档\n", docCount)
					return
				}
				log.Printf("流式处理错误: %v", err)
				continue
			}

			// 处理接收到的文档
			for _, doc := range docs {
				docCount++
				fmt.Printf("📄 实时接收文档 %d:\n", docCount)
				fmt.Printf("  ID: %s\n", doc.ID)
				fmt.Printf("  内容长度: %d 字符\n", len(doc.Content))
				if doc.MetaData != nil {
					if source, ok := doc.MetaData["source"]; ok {
						fmt.Printf("  来源: %v\n", source)
					}
				}
				fmt.Println("  ✅ 处理完成")
				fmt.Println()
			}
		}
	}
}

func parallelExample(ctx context.Context) {
	// 创建并行批量加载链（最多10个并发worker）
	chain, err := loaddoc.NewParallelBatchLoadChain(ctx, 30*time.Second, 10)
	if err != nil {
		log.Fatalf("Failed to create parallel batch load chain: %v", err)
	}

	// 编译链
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("Failed to compile chain: %v", err)
	}

	// 准备更多的输入数据来测试并行处理
	input := &loaddoc.BatchLoadInput{
		// FilePaths: []string{
		// 	"./example1.txt",
		// 	"./example2.txt",
		// 	"./README.md",
		// 	"./config.yaml",
		// },
		URLs: []string{
			"https://www.cloudwego.io/zh/docs/eino",
			"https://www.cloudwego.io/zh/docs/eino/core_modules/chain_and_graph_orchestration/chain_graph_introduction/#graph",
		},
	}

	// 记录开始时间
	startTime := time.Now()

	// 执行并行批量加载
	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute parallel batch load: %v", err)
	}

	// 记录结束时间
	duration := time.Since(startTime)

	// 输出结果
	fmt.Printf("并行处理完成，耗时: %v\n", duration)
	fmt.Printf("成功加载 %d 个文档\n", len(result))

	// 按类型统计文档
	fileCount := 0
	urlCount := 0
	for _, doc := range result {
		fmt.Printf("文档ID: %s, 内容长度: %d\n", doc.ID, len(doc.Content))
		if doc.MetaData != nil {
			if docType, ok := doc.MetaData["type"]; ok {
				switch docType {
				case "file":
					fileCount++
				case "url":
					urlCount++
				}
			}
		}
	}

	fmt.Printf("文件文档: %d 个\n", fileCount)
	fmt.Printf("URL文档: %d 个\n", urlCount)
}

// realTimeExample 实时处理示例
func realTimeExample(ctx context.Context) {
	// 创建并行批量加载链（最多3个并发worker，便于观察实时效果）
	chain, err := loaddoc.NewParallelBatchLoadChain(ctx, 30*time.Second, 3)
	if err != nil {
		log.Fatalf("Failed to create parallel batch load chain: %v", err)
	}

	// 编译链
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("Failed to compile chain: %v", err)
	}

	// 准备输入数据（包含一些可能较慢的URL来展示实时效果）
	input := &loaddoc.BatchLoadInput{
		// FilePaths: []string{
		// 	"./README.md",
		// 	"./go.mod",
		// 	"./go.sum",
		// },
		URLs: []string{
			"https://httpbin.org/delay/1", // 延迟1秒
			"https://httpbin.org/delay/2", // 延迟2秒
			"https://httpbin.org/json",
			"https://httpbin.org/uuid",
		},
	}

	fmt.Println("🚀 开始实时处理，每个文档处理完成后立即显示...")
	fmt.Println("⏱️  注意观察文档的到达时间差异")
	fmt.Println()

	// 使用流式处理来实现实时效果
	result, err := runnable.Stream(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute stream load: %v", err)
	}

	// 实时处理每个文档
	docCount := 0
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("❌ 处理被取消")
			return
		default:
			// 从流中读取结果
			docs, err := result.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					totalTime := time.Since(startTime)
					fmt.Printf("✅ 实时处理完成！\n")
					fmt.Printf("📊 总计处理 %d 个文档，总耗时: %v\n", docCount, totalTime)
					return
				}
				log.Printf("⚠️  流式处理错误: %v", err)
				continue
			}

			// 处理接收到的文档
			for _, doc := range docs {
				docCount++
				currentTime := time.Since(startTime)

				fmt.Printf("📄 [%v] 实时接收文档 #%d\n", currentTime.Round(time.Millisecond), docCount)
				fmt.Printf("   🆔 ID: %s\n", doc.ID)
				fmt.Printf("   📏 内容长度: %d 字符\n", len(doc.Content))

				// 显示来源信息
				if doc.MetaData != nil {
					if source, ok := doc.MetaData["source"]; ok {
						fmt.Printf("   🔗 来源: %v\n", source)
					}
					if fileName, ok := doc.MetaData["file_name"]; ok {
						fmt.Printf("   📁 文件名: %v\n", fileName)
					}
				}

				// 显示内容预览（前100个字符）
				preview := doc.Content
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				fmt.Printf("   📝 内容预览: %s\n", preview)
				fmt.Printf("   ⚡ 处理完成 (耗时: %v)\n", currentTime.Round(time.Millisecond))
				fmt.Println("   " + strings.Repeat("-", 50))
				fmt.Println()
			}
		}
	}
}
