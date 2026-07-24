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

func (p *Pipeline) Search(ctx context.Context, queryText string, topK int) ([]SearchResult, error) {
	vectors, err := p.EmbeddingHandler(ctx, []string{queryText})
	if err != nil {
		return nil, fmt.Errorf("rag: 查询向量化失败: %w", err)
	}
	queryVector := vectors[0]
	loadTask, err := p.milvusClient.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(p.collectionName))
	if err != nil {
		return nil, fmt.Errorf("rag: 加载集合失败: %w", err)
	}
	if err = loadTask.Await(ctx); err != nil {
		return nil, fmt.Errorf("rag: 等待集合加载完成失败: %w", err)
	}

	searchOption := milvusclient.NewSearchOption(
		p.collectionName,
		topK,
		[]entity.Vector{entity.FloatVector(queryVector)},
	).WithOutputFields("wiki_varchar", "wiki_resource").WithANNSField("wiki_vector")
	//WithAnnParam(annParam)

	fmt.Print("searchOption:", searchOption)
	searchOption = searchOption.WithSearchParam("metric_type", "COSINE")
	searchResult, err := p.milvusClient.Search(ctx, searchOption)
	if err != nil {
		return nil, fmt.Errorf("rag: 搜索失败: %w", err)
	}
	fmt.Print("searchResult:", searchResult)
	var results []SearchResult
	for _, res := range searchResult {
		varcharCol := res.GetColumn("wiki_varchar")
		resourceCol := res.GetColumn("wiki_resource")

		varcharColTyped, ok := varcharCol.(*column.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("rag: wiki_varchar 列类型不匹配")
		}
		resourceColTyped, ok := resourceCol.(*column.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("rag: wiki_resource 列类型不匹配")
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
