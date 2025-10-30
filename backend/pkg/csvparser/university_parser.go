package csvparser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/jszwec/csvutil"
)

// UniversityCSVRecord 表示CSV中的一行大学数据，使用csvutil标签进行映射
type UniversityCSVRecord struct {
	RankQS             string `csv:"rank_qs"`
	NameEn             string `csv:"name_en"`
	NameCn             string `csv:"name_cn"`
	Location           string `csv:"location"`
	OverallScore       string `csv:"overall_score"`
	IsDoubleFirstClass string `csv:"is_double_first_class"`
	Is985              string `csv:"is_985"`
	Is211              string `csv:"is_211"`
	IsQSTop100         string `csv:"is_qs_top100"`
	Alias              string `csv:"alias"`
}

// UniversityCSVParser CSV解析器
type UniversityCSVParser struct{}

// NewUniversityCSVParser 创建新的CSV解析器
func NewUniversityCSVParser() *UniversityCSVParser {
	return &UniversityCSVParser{}
}

// ParseCSV 解析CSV数据，使用csvutil进行高效解析
func (p *UniversityCSVParser) ParseCSV(reader io.Reader) ([]*domain.CreateUniversityReq, error) {
	// 读取所有数据到字节切片
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV data: %w", err)
	}

	// 使用csvutil直接解析CSV数据到结构体切片
	var records []UniversityCSVRecord
	if err := csvutil.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("failed to decode CSV: %w", err)
	}

	// 转换CSV记录为CreateUniversityReq
	universities := make([]*domain.CreateUniversityReq, 0, len(records))
	for i, record := range records {
		university, err := p.convertRecord(record, i+2) // +2 因为第一行是表头，第二行开始是数据
		if err != nil {
			return nil, fmt.Errorf("failed to convert record at line %d: %w", i+2, err)
		}

		if university != nil {
			universities = append(universities, university)
		}
	}

	return universities, nil
}

// convertRecord 将CSV记录转换为CreateUniversityReq
func (p *UniversityCSVParser) convertRecord(record UniversityCSVRecord, lineNumber int) (*domain.CreateUniversityReq, error) {
	// 跳过空记录
	if p.isEmptyCSVRecord(record) {
		return nil, nil
	}

	// 解析排名
	var rankQS *int
	if record.RankQS != "" && record.RankQS != "N/A" {
		rank, err := strconv.Atoi(record.RankQS)
		if err != nil {
			return nil, fmt.Errorf("invalid rank_qs value '%s': %w", record.RankQS, err)
		}
		rankQS = &rank
	}

	// 解析总分
	var overallScore *float64
	if record.OverallScore != "" && record.OverallScore != "N/A" {
		score, err := strconv.ParseFloat(record.OverallScore, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid overall_score value '%s': %w", record.OverallScore, err)
		}
		overallScore = &score
	}

	// 解析布尔值字段
	isDoubleFirstClass := p.parseBooleanField(record.IsDoubleFirstClass)
	isProject985 := p.parseBooleanField(record.Is985)
	isProject211 := p.parseBooleanField(record.Is211)
	isQsTop100 := p.parseBooleanField(record.IsQSTop100)

	// 处理字符串字段
	var nameEn *string
	if record.NameEn != "" {
		nameEn = &record.NameEn
	}

	var alias *string
	if record.Alias != "" {
		alias = &record.Alias
	}

	return &domain.CreateUniversityReq{
		RankQs:             rankQS,
		NameEn:             nameEn,
		NameCn:             record.NameCn,
		Country:            &record.Location, // 将Location映射为Country
		OverallScore:       overallScore,
		IsDoubleFirstClass: isDoubleFirstClass,
		IsProject985:       isProject985,
		IsProject211:       isProject211,
		IsQsTop100:         isQsTop100,
		Alias:              alias,
	}, nil
}

// parseBooleanField 解析布尔值字段
func (p *UniversityCSVParser) parseBooleanField(value string) *bool {
	if value == "" || value == "N/A" {
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "y":
		result := true
		return &result
	case "false", "0", "no", "n":
		result := false
		return &result
	default:
		return nil
	}
}

// isEmptyCSVRecord 检查CSV记录是否为空
func (p *UniversityCSVParser) isEmptyCSVRecord(record UniversityCSVRecord) bool {
	return record.NameEn == "" && record.NameCn == "" && record.Location == ""
}
