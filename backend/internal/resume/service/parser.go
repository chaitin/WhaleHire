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
	"github.com/chaitin/WhaleHire/backend/db"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/docparser"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/resumeparser"
	resumeparsergraph "github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/resumeparser"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
	"github.com/cloudwego/eino/compose"
	"github.com/google/uuid"
)

type ParserService struct {
	modelFactory        *models.ModelFactory
	config              *config.Config
	logger              *slog.Logger
	documentParser      *docparser.DocumentParserService
	resumeRepo          domain.ResumeRepo
	resumeParseRunnable compose.Runnable[*resumeparser.ResumeParseInput, *resumeparser.ResumeParseResult]
	resumeParseGraph    *resumeparsergraph.ResumeParseGraph
}

// NewParserService 创建解析服务
func NewParserService(config *config.Config, logger *slog.Logger, resumeRepo domain.ResumeRepo) (domain.ParserService, error) {
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

	// 创建并初始化简历解析图
	ctx := context.Background()
	chatModel, err := factory.GetModel(ctx, models.ModelTypeOpenAI, config.GeneralAgent.LLM.ModelName)
	if err != nil {
		logger.Error("failed to get model during initialization", "error", err)
		// 返回一个没有 runnable 的服务，在 ParseResume 时会报错
		return nil, err
	}

	// 创建简历解析图
	graph, err := resumeparsergraph.NewResumeParseGraph(ctx, config, chatModel, logger)
	if err != nil {
		logger.Error("failed to create resume parser graph during initialization", "error", err)
		// 返回一个没有 runnable 的服务，在 ParseResume 时会报错
		return nil, err
	}

	// 编译图
	runnable, err := graph.Compile(ctx)
	if err != nil {
		logger.Error("failed to compile resume parser graph during initialization", "error", err)
		// 返回一个没有 runnable 的服务，在 ParseResume 时会报错
		return nil, err
	}

	return &ParserService{
		modelFactory:        factory,
		config:              config,
		logger:              logger,
		documentParser:      documentParser,
		resumeRepo:          resumeRepo,
		resumeParseRunnable: runnable,
		resumeParseGraph:    graph,
	}, nil
}

// downloadFile 下载文件到临时目录
func (s *ParserService) downloadFile(ctx context.Context, fileURL string, fileType string) (string, error) {
	// 解析原始URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %w", err)
	}

	// 使用MinIO配置的endpoint替换URL中的host
	minioEndpoint := s.config.S3.Endpoint
	if !strings.HasPrefix(minioEndpoint, "http://") && !strings.HasPrefix(minioEndpoint, "https://") {
		minioEndpoint = "http://" + minioEndpoint
	}

	// 构建新的URL，使用MinIO的endpoint但保持原有的路径
	actualURL := minioEndpoint + parsedURL.Path

	// 创建HTTP请求
	req, createErr := http.NewRequestWithContext(ctx, "GET", actualURL, nil)
	if createErr != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", createErr)
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
	parsedURL, err = url.Parse(fileURL)
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
func (s *ParserService) ParseResume(ctx context.Context, resumeID string, fileURL string, fileType string) (*domain.ParsedResumeData, error) {
	// 检查 runnable 是否已初始化
	if s.resumeParseRunnable == nil {
		s.logger.Error("resume parse runnable not initialized", "resumeID", resumeID)
		return nil, fmt.Errorf("简历解析服务未正确初始化")
	}

	var resumeContent string

	// 如果配置了文档解析服务，则下载并解析文件内容
	if s.documentParser != nil {
		// 下载文件到临时目录
		tempFilePath, downloadErr := s.downloadFile(ctx, fileURL, fileType)
		if downloadErr != nil {
			s.logger.Error("下载文件失败", "error", downloadErr, "fileURL", fileURL)
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

		// 保存docparser的解析结果到数据库
		if err := s.saveDocumentParseResult(ctx, resumeID, parseResult); err != nil {
			s.logger.Error("保存文档解析结果失败", "error", err, "resumeID", resumeID)
			// 这里不返回错误，因为主要的解析流程应该继续
		}
	} else {
		// 如果没有配置文档解析服务，直接返回错误
		s.logger.Error("文档解析服务未配置，无法解析简历文件", "fileURL", fileURL)
		return nil, fmt.Errorf("文档解析服务未配置，无法解析简历文件")
	}

	// 准备输入
	input := &resumeparser.ResumeParseInput{
		Resume: resumeContent,
	}

	// 执行解析 - 使用预编译的 runnable
	result, err := s.resumeParseRunnable.Invoke(ctx, input)
	if err != nil {
		s.logger.Error("failed to parse resume", "error", err, "fileURL", fileURL)
		return nil, fmt.Errorf("解析简历失败: %w", err)
	}

	s.logger.Info("resume parsed successfully", "fileURL", fileURL, "name", result.BasicInfo.Name)
	return result, nil
}

// saveDocumentParseResult 保存文档解析结果到数据库
func (s *ParserService) saveDocumentParseResult(ctx context.Context, resumeID string, parseResult *docparser.ParseDocumentResult) error {
	// 将docparser结果转换为数据库实体
	docParse := &db.ResumeDocumentParse{
		ResumeID: uuid.MustParse(resumeID),
		FileID:   parseResult.FileID,
		Content:  parseResult.Content,
		FileType: parseResult.FileType,
		Filename: parseResult.Filename,
		Title:    parseResult.Title,
		UploadAt: parseResult.UploadAt,
		Status:   "success",
	}

	// 保存到数据库
	_, err := s.resumeRepo.CreateDocumentParse(ctx, docParse)
	if err != nil {
		return fmt.Errorf("创建文档解析记录失败: %w", err)
	}

	s.logger.Info("文档解析结果已保存", "resumeID", resumeID, "fileID", parseResult.FileID)
	return nil
}
