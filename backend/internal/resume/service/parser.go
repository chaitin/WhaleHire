package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/docparser"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/resumeparser"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
)

type ParserService struct {
	modelFactory   *models.ModelFactory
	config         *config.Config
	logger         *slog.Logger
	documentParser *docparser.DocumentParserService
}

// NewParserService 创建解析服务
func NewParserService(config *config.Config, logger *slog.Logger) domain.ParserService {
	factory := models.NewModelFactory()

	// 注册OpenAI模型
	openaiConfig := &models.OpenAIConfig{
		APIKey:         config.GeneralAgent.LLM.APIKey,
		BaseURL:        config.GeneralAgent.LLM.BaseURL,
		Model:          config.GeneralAgent.LLM.ModelName,
		ResponseFormat: "json_object",
	}
	openaiManager := models.NewOpenAIModelManager(openaiConfig)
	factory.Register(models.ModelTypeOpenAI, config.GeneralAgent.LLM.ModelName, openaiManager)

	// 创建文档解析服务
	var documentParser *docparser.DocumentParserService
	if config.DocumentParser.APIKey != "" {
		documentParser = docparser.NewDocumentParserServiceFromConfig(config)
	}

	return &ParserService{
		modelFactory:   factory,
		config:         config,
		logger:         logger,
		documentParser: documentParser,
	}
}

// downloadFile 下载文件到临时目录
func (s *ParserService) downloadFile(ctx context.Context, fileURL string, fileType string) (string, error) {
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载文件失败，状态码: %d", resp.StatusCode)
	}

	// 从URL中提取文件名，去除查询参数
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %w", err)
	}

	fileName := filepath.Base(parsedURL.Path)

	// 如果无法从URL获取文件名，使用fileType参数确定扩展名
	if fileName == "." || fileName == "/" || fileName == "" {
		// 根据fileType确定默认扩展名
		var defaultExt string
		switch fileType {
		case "pdf":
			defaultExt = ".pdf"
		case "doc":
			defaultExt = ".doc"
		case "docx":
			defaultExt = ".docx"
		default:
			defaultExt = ".pdf" // 默认为pdf
		}
		fileName = "resume_file" + defaultExt
	}

	// 确保文件有正确的扩展名
	ext := filepath.Ext(fileName)
	if ext == "" {
		// 如果文件名没有扩展名，根据fileType添加
		switch fileType {
		case "pdf":
			fileName += ".pdf"
		case "doc":
			fileName += ".doc"
		case "docx":
			fileName += ".docx"
		default:
			fileName += ".pdf"
		}
	}

	// 确保文件名不超过合理长度
	maxFileNameLen := 50
	if len(fileName) > maxFileNameLen {
		ext := filepath.Ext(fileName)
		baseName := strings.TrimSuffix(fileName, ext)
		if len(baseName) > maxFileNameLen-len(ext) {
			baseName = baseName[:maxFileNameLen-len(ext)]
		}
		fileName = baseName + ext
	}

	// 生成唯一的临时文件名
	tempDir := os.TempDir()
	timestamp := time.Now().Format("20060102150405")
	uniqueID := fmt.Sprintf("%x", md5.Sum([]byte(fileURL)))[:8]

	// 分离文件名和扩展名
	fileExt := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, fileExt)

	// 清理文件名中的特殊字符
	cleanBaseName := strings.ReplaceAll(baseName, " ", "_")
	cleanBaseName = strings.ReplaceAll(cleanBaseName, "/", "_")
	cleanBaseName = strings.ReplaceAll(cleanBaseName, "\\", "_")

	// 构建临时文件名：resume_baseName_timestamp_uniqueID.ext
	tempFile := filepath.Join(tempDir, fmt.Sprintf("resume_%s_%s_%s%s",
		cleanBaseName,
		timestamp,
		uniqueID,
		fileExt))

	// 创建文件
	out, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer out.Close()

	// 复制内容
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tempFile) // 清理失败的文件
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	s.logger.Info("文件下载成功", "url", fileURL, "tempFile", tempFile)
	return tempFile, nil
}

// ParseResume 解析简历内容
func (s *ParserService) ParseResume(ctx context.Context, fileURL string, fileType string) (*domain.ParsedResumeData, error) {
	// 获取模型
	chatModel, err := s.modelFactory.GetModel(ctx, models.ModelTypeOpenAI, s.config.GeneralAgent.LLM.ModelName)
	if err != nil {
		s.logger.Error("failed to get model", "error", err)
		return nil, fmt.Errorf("获取模型失败: %w", err)
	}

	// 创建简历解析链
	chain, err := resumeparser.NewResumeParserChain(ctx, chatModel)
	if err != nil {
		s.logger.Error("failed to create resume parser chain", "error", err)
		return nil, fmt.Errorf("创建解析链失败: %w", err)
	}

	// 编译链
	runnable, err := chain.Compile(ctx)
	if err != nil {
		s.logger.Error("failed to compile resume parser chain", "error", err)
		return nil, fmt.Errorf("编译解析链失败: %w", err)
	}

	var resumeContent string

	// 如果配置了文档解析服务，则下载并解析文件内容
	if s.documentParser != nil {
		// 下载文件到临时目录
		tempFilePath, downloadErr := s.downloadFile(ctx, fileURL, fileType)
		if downloadErr != nil {
			s.logger.Error("下载文件失败", "error", err, "fileURL", fileURL)
			return nil, fmt.Errorf("下载文件失败: %w", downloadErr)
		}

		// 确保清理临时文件
		defer func() {
			if removeErr := os.Remove(tempFilePath); removeErr != nil {
				s.logger.Warn("清理临时文件失败", "tempFile", tempFilePath, "error", removeErr)
			}
		}()

		// 使用docparser解析文件内容
		parseResult, parseErr := s.documentParser.ParseDocument(ctx, tempFilePath)
		if parseErr != nil {
			s.logger.Error("解析文档内容失败", "error", parseErr, "tempFile", tempFilePath)
			return nil, fmt.Errorf("解析文档内容失败: %w", parseErr)
		}

		resumeContent = parseResult.Content
		s.logger.Info("文档解析成功", "fileURL", fileURL, "contentLength", len(resumeContent))
	} else {
		// 如果没有配置文档解析服务，直接返回错误
		s.logger.Error("文档解析服务未配置，无法解析简历文件", "fileURL", fileURL)
		return nil, fmt.Errorf("文档解析服务未配置，无法解析简历文件")
	}
	// 准备输入
	input := &resumeparser.ResumeParseInput{
		Resume: resumeContent,
	}

	// 执行解析 - 现在直接返回 domain.ParsedResumeData，无需转换
	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		s.logger.Error("failed to parse resume", "error", err, "fileURL", fileURL)
		return nil, fmt.Errorf("解析简历失败: %w", err)
	}

	s.logger.Info("resume parsed successfully", "fileURL", fileURL, "name", result.BasicInfo.Name)
	return result, nil
}
