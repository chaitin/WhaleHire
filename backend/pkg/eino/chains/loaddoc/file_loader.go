package loaddoc

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
)

// FileLoader 文件加载器（基于Eino官方实现）
type FileLoader struct {
	loader document.Loader
}

// NewFileLoader 创建文件加载器
func NewFileLoader() *FileLoader {
	// 初始化Eino官方文件加载器
	loader, err := file.NewFileLoader(context.Background(), &file.FileLoaderConfig{
		UseNameAsID: true, // 使用文件名作为文档ID
		// Parser 可以根据需要设置自定义解析器，默认使用扩展名解析器
	})
	if err != nil {
		// 如果初始化失败，使用默认配置
		loader, _ = file.NewFileLoader(context.Background(), nil)
	}

	return &FileLoader{
		loader: loader,
	}
}

// LoadFile 加载单个文件
func (fl *FileLoader) LoadFile(ctx context.Context, filePath string) (*Document, error) {
	// 使用Eino官方加载器加载文档
	docs, err := fl.loader.Load(ctx, document.Source{
		URI: filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load file %s: %w", filePath, err)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents loaded from file: %s", filePath)
	}

	// 转换为我们的Document格式
	einoDoc := docs[0]

	// 创建文档
	doc := &Document{
		ID:      einoDoc.ID,
		Content: einoDoc.Content,
		Source:  filePath,
		Type:    DocumentTypeFile,
		Metadata: map[string]any{
			"file_name": filepath.Base(filePath),
			"file_ext":  filepath.Ext(filePath),
		},
	}

	// 从Eino文档的元数据中提取信息
	if einoDoc.MetaData != nil {
		for key, value := range einoDoc.MetaData {
			// 使用官方的元数据键名 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/document/loader_local_file/" index="0">0</mcreference>
			switch key {
			case file.MetaKeyFileName:
				doc.Metadata["file_name"] = value
			case file.MetaKeyExtension:
				doc.Metadata["file_ext"] = value
			case file.MetaKeySource:
				doc.Metadata["source"] = value
			default:
				doc.Metadata[key] = value
			}
		}
	}

	return doc, nil
}

// LoadFiles 批量加载文件
func (fl *FileLoader) LoadFiles(ctx context.Context, filePaths []string) ([]*Document, []LoadError) {
	var documents []*Document
	var errors []LoadError

	for _, filePath := range filePaths {
		select {
		case <-ctx.Done():
			errors = append(errors, LoadError{
				Source: filePath,
				Error:  "context cancelled",
			})
			continue
		default:
		}

		doc, err := fl.LoadFile(ctx, filePath)
		if err != nil {
			errors = append(errors, LoadError{
				Source: filePath,
				Error:  err.Error(),
			})
			continue
		}

		documents = append(documents, doc)
	}

	return documents, errors
}

// getSupportedExtensions 获取支持的文件扩展名
func (fl *FileLoader) getSupportedExtensions() []string {
	// 根据 Eino 官方文档，默认支持文本文件 <mcreference link="https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/document/loader_local_file/" index="0">0</mcreference>
	return []string{".txt", ".md", ".json", ".yaml", ".yml", ".csv", ".log"}
}

// IsSupportedFile 检查文件是否支持
func (fl *FileLoader) IsSupportedFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, supportedExt := range fl.getSupportedExtensions() {
		if ext == supportedExt {
			return true
		}
	}
	return false
}
