package dao

import "Main/model"

func CreateNewChat(chatModel *model.ChatModel) error {
	result := dataBase.Model(&model.ChatModel{}).Create(chatModel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func GetChat(chatId int64) (*model.ChatModel, error) {
	chatModel := &model.ChatModel{}
	result := dataBase.Model(&model.ChatModel{}).Where("chat_id=?", chatId).Find(chatModel)
	if result.Error != nil {
		return nil, result.Error
	}
	return chatModel, nil
}
func UpdateChat(chat *model.ChatModel) error {
	result := dataBase.Model(&model.ChatModel{}).Where("chat_id=?", chat.ChatId).Updates(chat)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func DeleteChat(chatId int64) error {
	result := dataBase.Model(&model.ChatModel{}).Where("chat_id=?", chatId).Delete(&model.ChatModel{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
