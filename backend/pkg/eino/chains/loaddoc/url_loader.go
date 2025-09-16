package loaddoc

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/document/loader/url"
	"github.com/cloudwego/eino/components/document"
)

// URLLoader 在线链接加载器（基于Eino官方实现）
type URLLoader struct {
	loader  document.Loader
	timeout time.Duration
}

// NewURLLoader 创建在线链接加载器
func NewURLLoader(timeout time.Duration) *URLLoader {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// 自定义HTTP客户端
	client := &http.Client{
		Timeout: timeout,
	}

	// 自定义请求构建器
	requestBuilder := func(ctx context.Context, src document.Source, opts ...document.LoaderOption) (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", src.URI, nil)
		if err != nil {
			return nil, err
		}
		// 添加自定义请求头
		req.Header.Set("User-Agent", "WhaleHire-DocumentLoader/1.0")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		return req, nil
	}

	// 初始化Eino官方URL加载器
	loader, err := url.NewLoader(context.Background(), &url.LoaderConfig{
		Client:         client,
		RequestBuilder: requestBuilder,
	})
	if err != nil {
		// 如果初始化失败，使用默认配置
		loader, _ = url.NewLoader(context.Background(), nil)
	}

	return &URLLoader{
		loader:  loader,
		timeout: timeout,
	}
}

// LoadURL 加载单个URL
func (ul *URLLoader) LoadURL(ctx context.Context, urlStr string) (*Document, error) {
	// 使用Eino官方加载器加载文档
	docs, err := ul.loader.Load(ctx, document.Source{
		URI: urlStr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load URL %s: %w", urlStr, err)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents loaded from URL: %s", urlStr)
	}

	// 转换为我们的Document格式
	einoDoc := docs[0]

	// 生成文档ID
	docID := ul.generateDocumentID(urlStr, einoDoc.Content)

	// 创建文档
	doc := &Document{
		ID:      docID,
		Content: einoDoc.Content,
		Source:  urlStr,
		Type:    DocumentTypeURL,
		Metadata: map[string]any{
			"url":            urlStr,
			"content_length": len(einoDoc.Content),
			"fetch_time":     time.Now(),
		},
	}

	// 从Eino文档的元数据中提取信息
	if einoDoc.MetaData != nil {
		for key, value := range einoDoc.MetaData {
			doc.Metadata[key] = value
		}
	}

	return doc, nil
}

// LoadURLs 批量加载URLs
func (ul *URLLoader) LoadURLs(ctx context.Context, urls []string) ([]*Document, []LoadError) {
	// 使用通道来收集结果
	docChan := make(chan *Document, len(urls))
	errChan := make(chan LoadError, len(urls))

	// 使用WaitGroup来等待所有协程完成
	var wg sync.WaitGroup

	// 启动协程处理每个URL
	for _, urlStr := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				errChan <- LoadError{
					Source: url,
					Error:  "context cancelled",
				}
				return
			default:
			}

			// 加载文档
			doc, err := ul.LoadURL(ctx, url)
			if err != nil {
				errChan <- LoadError{
					Source: url,
					Error:  err.Error(),
				}
				return
			}

			docChan <- doc
		}(urlStr)
	}

	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(docChan)
		close(errChan)
	}()

	// 收集结果
	var documents []*Document
	var errors []LoadError

	// 从通道中读取所有文档
	for doc := range docChan {
		documents = append(documents, doc)
	}

	// 从通道中读取所有错误
	for err := range errChan {
		errors = append(errors, err)
	}

	return documents, errors
}

// generateDocumentID 生成文档唯一ID
func (ul *URLLoader) generateDocumentID(source, content string) string {
	h := md5.New()
	_, _ = io.WriteString(h, source)
	_, _ = io.WriteString(h, content)
	return fmt.Sprintf("url_%x", h.Sum(nil))
}

// IsValidURL 检查URL是否有效（保持向后兼容）
func (ul *URLLoader) IsValidURL(urlStr string) bool {
	// 尝试创建一个document.Source来验证URL
	_, err := ul.loader.Load(context.Background(), document.Source{
		URI: urlStr,
	})
	return err == nil
}
