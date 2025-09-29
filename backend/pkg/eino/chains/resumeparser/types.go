package resumeparser

import (
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/schema"
)

type ResumeParseInput struct {
	Resume  string            `json:"resume"`
	History []*schema.Message `json:"history,omitempty"`
}

// 直接使用 domain.ParsedResumeData 作为解析结果
type ResumeParseResult = domain.ParsedResumeData
