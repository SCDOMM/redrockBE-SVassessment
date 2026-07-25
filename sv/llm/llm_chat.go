package llm

import (
	"Main/sv/rag"
	"Main/utils"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/openai/openai-go/v3"
)

type ChatStruct struct {
	mu         sync.RWMutex
	ChatClient openai.Client
	History    []openai.ChatCompletionMessageParamUnion
}

func NewChatStruct(chatClient openai.Client) *ChatStruct {
	return &ChatStruct{
		ChatClient: chatClient,
		History:    make([]openai.ChatCompletionMessageParamUnion, 0, 20),
	}
}

// AskQuestion 请优先配置好ChatStruct并且执行完搜索
func (l *ChatStruct) AskQuestion(searchResults []rag.SearchResult, ctx context.Context, question string) (string, error) {
	var sb strings.Builder
	sb.WriteString("以下是相关文档片段(按相关度从高到低排序)：\n\n")
	for i, r := range searchResults {
		sb.WriteString(fmt.Sprintf("【片段%d】(来源:%s，相似度:%.2f)\n%s\n\n",
			i+1, r.Resource, r.Score, r.Text))
	}
	systemMessage := openai.SystemMessage(`你是一个智能文档助手。请严格根据提供的文档片段回答用户问题。` +
		`如果片段中未包含足够信息，请直接说明“文档中未找到相关信息”，不要编造，不要虚构API，不要说出不存在的代码。` +
		`回答时请尽量引用原文，保持准确。`)

	userMessage := fmt.Sprintf("文档片段：\n%s\n\n用户问题：%s", sb.String(), question)

	//超级拼装上下文
	l.mu.Lock()
	allMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(l.History)+2)
	allMessages = append(allMessages, systemMessage)
	allMessages = append(allMessages, l.History...)
	allMessages = append(allMessages,
		openai.UserMessage(userMessage),
	)
	l.mu.Unlock()

	resp, err := l.ChatClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       utils.GetChatModel(),
		Messages:    allMessages,
		Temperature: openai.Float(0.3),
		MaxTokens:   openai.Int(2048),
		TopP:        openai.Float(0.9),
	})

	if err != nil {
		return "", fmt.Errorf("rag:fail to use LLM: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("rag:LLM can not reply validly")
	}

	answer := resp.Choices[0].Message.Content
	l.mu.Lock()
	l.History = append(l.History,
		openai.UserMessage(question),
		openai.AssistantMessage(answer),
	)
	//保留20轮消息
	if len(l.History) > 40 {
		l.History = l.History[len(l.History)-40:]
	}
	l.mu.Unlock()
	return answer, nil
}
