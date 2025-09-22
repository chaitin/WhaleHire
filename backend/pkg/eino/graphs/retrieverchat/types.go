package retrieverchat

import "github.com/cloudwego/eino/schema"

// UserMessage 用户输入消息结构
type UserMessage struct {
	Input   string            `json:"input"`   // 用户查询内容
	History []*schema.Message `json:"history"` // 历史对话记录
}
