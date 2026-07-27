package rag

import (
	"Main/utils"
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Pipeline struct {
	ChatClient     *openai.Client
	EmbedClient    *openai.Client
	MilvusClient   *milvusclient.Client
	Chunker        *utils.TextChunker
	CollectionName string
}

func (p *Pipeline) Close() {
	p.MilvusClient.Close(context.Background())
}

func NewPipeline(collectionsName string) (*Pipeline, error) {
	chatClient := openai.NewClient(
		option.WithAPIKey(utils.GetChatApiKeyConfig()),
		option.WithBaseURL(utils.GetChatUrl()),
	)
	embedClient := openai.NewClient(
		option.WithAPIKey(utils.GetEmbedApiKeyConfig()),
		option.WithBaseURL(utils.GetEmbedUrl()),
	)
	milvusClient, err := milvusclient.New(context.Background(), &milvusclient.ClientConfig{
		Address: utils.GetMilvusHost() + ":" + utils.GetMilvusPort(),
	})
	if err != nil {
		return nil, fmt.Errorf("rag:fail to create pipeline" + err.Error())
	}
	chunker := utils.NewTextChunker(1000, 200, 50)
	return &Pipeline{
		ChatClient:     &chatClient,
		EmbedClient:    &embedClient,
		MilvusClient:   milvusClient,
		Chunker:        chunker,
		CollectionName: collectionsName,
	}, nil
}
