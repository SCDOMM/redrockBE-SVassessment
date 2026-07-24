package rag

import (
	"Main/utils"
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func NewEmbeddingClient() openai.Client {
	return openai.NewClient(
		option.WithAPIKey(utils.GetEmbedApiKeyConfig()),
		option.WithBaseURL(utils.GetEmbedUrl()),
	)
}
func EmbeddingHandler(ctx context.Context, client *openai.Client, texts []string) ([][]float32, error) {
	resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: utils.GetEmbedModel(),
		Input: openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: texts},
	})
	if err != nil {
		return nil, fmt.Errorf("rag:embedding 调用失败: %w", err)
	}
	vectors := make([][]float32, len(resp.Data))
	fmt.Print(vectors)
	for i, item := range resp.Data {
		vec64 := item.Embedding
		vec32 := make([]float32, len(vec64))
		for j, v := range vec64 {
			vec32[j] = float32(v)
		}
		vectors[i] = vec32
	}
	return vectors, nil
}
