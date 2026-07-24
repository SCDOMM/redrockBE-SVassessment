package rag

import (
	"Main/utils"
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
)

func (p *Pipeline) EmbeddingHandler(ctx context.Context, texts []string) ([][]float32, error) {
	resp, err := p.embedClient.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: utils.GetEmbedModel(),
		Input: openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: texts},
	})
	if err != nil {
		return nil, fmt.Errorf("rag:failed to embedding %w", err)
	}
	vectors := make([][]float32, len(resp.Data))
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
