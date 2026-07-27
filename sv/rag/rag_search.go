package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

type SearchResult struct {
	Score     float64 // 相似度分数
	Text      string  // wiki_varchar
	Resource  string  // wiki_resource
	FilePath  string
	Language  string
	StartLine int64
	EndLine   int64
}

// Search 调用前务必调用loadCollections函数
func (p *Pipeline) Search(ctx context.Context, queryText string, topK int, filterConfig string, repoUrl string) ([]SearchResult, error) {
	vectors, err := p.EmbeddingHandler(ctx, []string{queryText})
	if err != nil {
		return nil, fmt.Errorf("rag:fail to embed query %w", err)
	}
	queryVector := vectors[0]

	//设置过滤条件
	var filterParts []string
	if filterConfig != "" {
		filterParts = append(filterParts, "("+filterConfig+")") // 加括号防优先级错误
	}
	if repoUrl != "" {
		escaped := strings.ReplaceAll(repoUrl, "'", "\\'")
		filterParts = append(filterParts, fmt.Sprintf("wiki_resource == '%s'", escaped))
	}
	finalFilter := strings.Join(filterParts, " and ")

	searchOption := milvusclient.NewSearchOption(
		p.CollectionName,
		topK,
		[]entity.Vector{entity.FloatVector(queryVector)},
	).WithOutputFields("wiki_varchar",
		"wiki_resource",
		"file_path",
		"start_line",
		"end_line",
		"language",
	).WithANNSField("wiki_vector")

	if finalFilter != "" {
		searchOption = searchOption.WithFilter(finalFilter)
	}

	searchResult, err := p.MilvusClient.Search(ctx, searchOption)
	if err != nil {
		return nil, fmt.Errorf("rag:fail to search %w", err)
	}

	var results []SearchResult
	for _, res := range searchResult {
		varcharCol := res.GetColumn("wiki_varchar")
		resourceCol := res.GetColumn("wiki_resource")
		pathCol := res.GetColumn("file_path")
		startLineCol := res.GetColumn("start_line")
		endLineCol := res.GetColumn("end_line")
		languageCol := res.GetColumn("language")

		varcharColTyped, ok := varcharCol.(*column.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("rag: wiki_varchar mismatch")
		}
		resourceColTyped, ok := resourceCol.(*column.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("rag: wiki_resource mismatch")
		}
		pathColTyped, ok := pathCol.(*column.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("rag: file_path mismatch")
		}
		languageColTyped, ok := languageCol.(*column.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("rag: language mismatch")
		}
		startLineColTyped, ok := startLineCol.(*column.ColumnInt64)
		if !ok {
			return nil, fmt.Errorf("rag: start_line mismatch")
		}
		endLineColTyped, ok := endLineCol.(*column.ColumnInt64)
		if !ok {
			return nil, fmt.Errorf("rag: end_line mismatch")
		}

		texts := varcharColTyped.Data()
		resources := resourceColTyped.Data()
		path := pathColTyped.Data()
		language := languageColTyped.Data()
		startLine := startLineColTyped.Data()
		endLine := endLineColTyped.Data()

		for i := 0; i < len(texts); i++ {
			results = append(results, SearchResult{
				Score:     float64(res.Scores[i]),
				Text:      texts[i],
				Resource:  resources[i],
				FilePath:  path[i],
				Language:  language[i],
				StartLine: startLine[i],
				EndLine:   endLine[i],
			})
		}
	}
	return results, nil
}
