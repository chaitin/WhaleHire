package intent

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

var systemPrompt = `
You are an AI intent classifier for a conversational Agent.  
Your job is to analyze the user's input and classify it into one intent category.  
The set of categories may grow over time, but you must always return exactly one category.

### Current Intent Categories
- SearchOnline: The user's input requires real-time, external, or updated information.  
- AnswerDirectly: The user's input can be answered directly with your internal knowledge.  

### Output Rules
- Always return a valid JSON object.  
- Do not include explanations or extra text.  
- JSON format must be strictly:

{{
  \"intent\": \"<CategoryName>\"
}}
`

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	config := &ChatTemplateConfig{
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(systemPrompt),
			schema.MessagesPlaceholder("history", false),
			schema.UserMessage("{query}"),
		},
	}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}
