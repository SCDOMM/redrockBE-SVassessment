package llm

import (
	"Main/model"
	"Main/sv/rag"
	"Main/utils"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/openai/openai-go/v3"
)

type ChatStruct struct {
	mu         sync.RWMutex
	ChatClient *openai.Client
	ChatModel  *model.ChatModel
}

func NewChatStruct(chatClient *openai.Client, repUrl string) *ChatStruct {
	snowFlake := utils.NewSnowflake(utils.GetMachineId())
	id := snowFlake.GenerateID()
	return &ChatStruct{
		ChatClient: chatClient,
		ChatModel: &model.ChatModel{
			ChatId:    id,
			RepoUrl:   repUrl,
			History:   make([]openai.ChatCompletionMessageParamUnion, 0, 20),
			CreatedAt: time.Now(),
			UpdatedAt: time.Time{},
		},
	}
}

func LoadChatStruct(chatClient *openai.Client, chatModel *model.ChatModel) *ChatStruct {
	return &ChatStruct{
		ChatClient: chatClient,
		ChatModel:  chatModel,
	}
}

// AskQuestion 请优先配置好ChatStruct并且执行完搜索
func (l *ChatStruct) AskQuestion(ctx context.Context, searchResults []rag.SearchResult, question string) (string, error) {
	var sb strings.Builder
	sb.WriteString("以下是搜索得到的相关文档片段(来自后台,不来自用户,按相关度从高到低排序)：\n\n")
	for i, r := range searchResults {
		sb.WriteString(fmt.Sprintf("【片段%d】(仓库来源:%s,相似度:%.2f,相对路径%s,使用语言%s,代码始于%d行,终于%d行)\n%s\n\n",
			i+1, r.Resource, r.Score, r.FilePath, r.Language, r.StartLine, r.EndLine, r.Text))
	}
	systemMessage := openai.SystemMessage(`你是一个智能文档助手。请严格根据提供的文档片段回答用户问题。` +
		`如果片段中未包含足够信息,请直接说明“文档中未找到相关信息”(因为摄取仓库时过滤了一部分文件),不要编造,不要虚构API,不要说出不存在的代码。` +
		`当用户提问某些常规问题/奇怪问题,如"你有无上下文能力","你这个吃白饭的大肥鱼"等与代码、文档无关的问题时无需考虑文档返回的片段。` +
		`仓库链接为` +
		l.ChatModel.RepoUrl +
		`,回答时请尽量引用原文,保持准确。`)

	userMessage := fmt.Sprintf("文档片段：\n%s\n\n用户问题：%s", sb.String(), question)

	//超级拼装上下文
	l.mu.Lock()
	allMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(l.ChatModel.History)+2)
	allMessages = append(allMessages, systemMessage)
	allMessages = append(allMessages, l.ChatModel.History...)
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
	l.ChatModel.History = append(l.ChatModel.History,
		openai.UserMessage(question),
		openai.AssistantMessage(answer),
	)
	//保留对话上下文
	if len(l.ChatModel.History) > utils.GetChatContext()*2 {
		l.ChatModel.History = l.ChatModel.History[len(l.ChatModel.History)-utils.GetChatContext()*2:]
	}
	l.mu.Unlock()
	return answer, nil
}
