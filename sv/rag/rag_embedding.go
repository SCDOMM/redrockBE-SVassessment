package rag

import (
	"Main/utils"
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
)

func (p *Pipeline) EmbeddingHandler(ctx context.Context, texts []string) ([][]float32, error) {
	targetDim := utils.GetEmbedDim()
	resp, err := p.EmbedClient.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: utils.GetEmbedModel(),
		Input: openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: texts},
		//Dimensions: openai.Int(utils.GetEmbedDim()),
	})
	if err != nil {
		return nil, fmt.Errorf("rag:failed to embedding %w", err)
	}
	vectors := make([][]float32, len(resp.Data))
	for i, item := range resp.Data {
		vec64 := item.Embedding
		if len(vec64) != int(targetDim) {
			if len(vec64) > int(targetDim) {
				vec64 = vec64[:targetDim] // 截取前 targetDim 维
			} else {
				// 如果比预期短，补零（极少发生，但保险起见）
				padded := make([]float64, targetDim)
				copy(padded, vec64)
				vec64 = padded
			}
		}
		vec32 := make([]float32, len(vec64))
		for j, v := range vec64 {
			vec32[j] = float32(v)
		}
		vectors[i] = vec32
	}
	return vectors, nil
}
