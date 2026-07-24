package rag

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

type SearchResult struct {
	Score    float64 // 相似度分数
	Text     string  // wiki_varchar
	Resource string  // wiki_resource
}

// Search 调用前无比调用loadCollections函数
func (p *Pipeline) Search(ctx context.Context, queryText string, topK int) ([]SearchResult, error) {
	vectors, err := p.EmbeddingHandler(ctx, []string{queryText})
	if err != nil {
		return nil, fmt.Errorf("rag:fail to embed query %w", err)
	}
	queryVector := vectors[0]
	searchOption := milvusclient.NewSearchOption(
		p.collectionName,
		topK,
		[]entity.Vector{entity.FloatVector(queryVector)},
	).WithOutputFields("wiki_varchar", "wiki_resource").
		WithANNSField("wiki_vector")
	searchResult, err := p.milvusClient.Search(ctx, searchOption)
	if err != nil {
		return nil, fmt.Errorf("rag:fail to search %w", err)
	}

	var results []SearchResult
	for _, res := range searchResult {
		varcharCol := res.GetColumn("wiki_varchar")
		resourceCol := res.GetColumn("wiki_resource")
		varcharColTyped, ok := varcharCol.(*column.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("rag: wiki_varchar mismatch")
		}
		resourceColTyped, ok := resourceCol.(*column.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("rag: wiki_resource mismatch")
		}
		texts := varcharColTyped.Data()
		resources := resourceColTyped.Data()
		for i := 0; i < len(texts); i++ {
			results = append(results, SearchResult{
				Score:    float64(res.Scores[i]),
				Text:     texts[i],
				Resource: resources[i],
			})
		}
	}
	return results, nil
}
