package model

import (
	"time"

	"github.com/openai/openai-go/v3"
)

type ChatModel struct {
	ChatId    int64                                    `gorm:"primary_key;auto_increment"`
	RepoUrl   string                                   `gorm:"size:255;not null"`
	History   []openai.ChatCompletionMessageParamUnion `gorm:"type:text;"`
	CreatedAt time.Time                                `gorm:"type:timestamp;not null"`
	UpdatedAt time.Time                                `gorm:"type:timestamp;not null"`
	DeletedAt time.Time                                `gorm:"type:timestamp;not null"`
}
type ChatNewModel struct {
	RepoUrl  string `json:"repo_url"`
	Question string `json:"question"`
}
type ChatNewDTO struct {
	ChatId int64  `json:"chat_id"`
	Answer string `json:"answer"`
}

type ChatContinueModel struct {
	ChatId   int64  `json:"chat_id"`
	Question string `json:"question"`
}
type ChatContinueDTO struct {
	Answer string `json:"answer"`
}
type ChatDeleteModel struct {
	ChatId int64 `json:"chat_id"`
}
