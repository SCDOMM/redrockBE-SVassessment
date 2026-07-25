package model

import (
	"time"

	"github.com/openai/openai-go/v3"
)

type ChatModel struct {
	ChatId    int64                                    `gorm:"primary_key"`
	RepoUrl   string                                   `gorm:"size:255;not null;index"`
	History   []openai.ChatCompletionMessageParamUnion `gorm:"serializer:json;type:longtext"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `gorm:"index"`
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
