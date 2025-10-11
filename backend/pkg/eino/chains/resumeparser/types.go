package resumeparser

import (
	"github.com/chaitin/WhaleHire/backend/domain"
)

type ResumeParseInput struct {
	Resume string `json:"resume"`
}

// 直接使用 domain.ParsedResumeData 作为解析结果
type ResumeParseResult = domain.ParsedResumeData
