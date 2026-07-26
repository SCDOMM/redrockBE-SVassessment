package model

import (
	"time"

	"github.com/openai/openai-go/v3"
	"gorm.io/gorm"
)

type ChatModel struct {
	ChatId    int64                                    `gorm:"primary_key"`
	RepoUrl   string                                   `gorm:"size:255;not null;index"`
	History   []openai.ChatCompletionMessageParamUnion `gorm:"serializer:json;type:longtext"`
	CreatedAt time.Time                                `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time                                `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt                           `gorm:"index" json:"deleted_at"`
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
