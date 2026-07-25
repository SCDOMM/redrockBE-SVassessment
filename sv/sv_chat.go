package sv

import (
	"Main/dao"
	"Main/model"
	"Main/sv/llm"
	"context"
	"fmt"
)

func NewChat(newModel model.ChatNewModel) (model.ChatNewDTO, error) {
	ctx := context.Background()
	err := pipeline.LoadCollections(ctx)
	if err != nil {
		return model.ChatNewDTO{}, err
	}

	result, err := pipeline.Search(ctx, newModel.Question, 5)
	if err != nil {
		return model.ChatNewDTO{}, err
	}

	chatStruct := llm.NewChatStruct(pipeline.ChatClient, newModel.RepoUrl)

	answer, err := chatStruct.AskQuestion(ctx, result, newModel.Question)
	if err != nil {
		return model.ChatNewDTO{}, err
	}
	//把新对话存入数据库
	err = dao.CreateNewChat(chatStruct.ChatModel)

	if err != nil {
		return model.ChatNewDTO{
			ChatId: chatStruct.ChatModel.ChatId,
			Answer: answer,
		}, err
	}

	return model.ChatNewDTO{
		ChatId: chatStruct.ChatModel.ChatId,
		Answer: answer,
	}, nil
}
func ContinueChat(continueModel model.ChatContinueModel) (model.ChatContinueDTO, error) {
	ctx := context.Background()

	chatModel, err := dao.GetChat(continueModel.ChatId)
	if err != nil {
		return model.ChatContinueDTO{}, err
	}

	err = pipeline.LoadCollections(ctx)
	result, err := pipeline.Search(ctx, continueModel.Question, 5)
	if err != nil {
		return model.ChatContinueDTO{}, err
	}

	chatStruct := llm.LoadChatStruct(pipeline.ChatClient, chatModel)

	answer, err := chatStruct.AskQuestion(ctx, result, continueModel.Question)

	//升级对话
	err = dao.UpdateChat(chatStruct.ChatModel)
	if err != nil {
		return model.ChatContinueDTO{}, fmt.Errorf("保存对话历史失败: %w", err)
	}

	return model.ChatContinueDTO{
		Answer: answer,
	}, err
}
func DeleteChat(deleteModel model.ChatDeleteModel) error {
	err := dao.DeleteChat(deleteModel.ChatId)
	if err != nil {
		return err
	}
	return nil
}
