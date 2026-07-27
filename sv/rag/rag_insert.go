package rag

import (
	"Main/utils"
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func repeat(s string, n int) []string {
	arr := make([]string, n)
	for i := range arr {
		arr[i] = s
	}
	return arr
}

// InsertDocument 插入文本(按行分割文本)
func (p *Pipeline) InsertDocument(ctx context.Context, docText string, resource string, filePath string, fileHash string, language string) error {
	chunks, startLines, endLines := p.Chunker.SplitByLines(docText)
	if len(chunks) == 0 {
		return fmt.Errorf("empty document")
	}
	const batchLength = 100

	for i := 0; i < len(chunks); i += batchLength {
		end := i + batchLength
		if end > len(chunks) {
			end = len(chunks)
		}
		batchChunks := chunks[i:end]
		batchStartLines := startLines[i:end]
		batchEndLines := endLines[i:end]

		vectors, err := p.EmbeddingHandler(ctx, batchChunks)
		if err != nil {
			return err
		}
		if len(vectors) != len(batchChunks) {
			return fmt.Errorf("embedding returned %d vectors for %d texts", len(vectors), len(batchChunks))
		}
		columns := []column.Column{
			column.NewColumnVarChar("wiki_varchar", batchChunks),
			column.NewColumnFloatVector("wiki_vector", int(utils.GetEmbedDim()), vectors),
			column.NewColumnVarChar("wiki_resource", repeat(resource, len(batchChunks))),
			column.NewColumnVarChar("file_path", repeat(filePath, len(batchChunks))),
			column.NewColumnVarChar("file_hash", repeat(fileHash, len(batchChunks))),
			column.NewColumnInt64("start_line", batchStartLines),
			column.NewColumnInt64("end_line", batchEndLines),
			column.NewColumnVarChar("language", repeat(language, len(batchChunks))),
		}
		_, err = p.MilvusClient.Insert(ctx,
			milvusclient.NewColumnBasedInsertOption(p.CollectionName, columns...),
		)
		if err != nil {
			return fmt.Errorf("rag: fail to insert batch %d: %w", i/batchLength, err)
		}
	}
	return nil
}

// SimpleInsertDocument 简单插入文档(按段落分割文本,测试用,实际生产不用)
func (p *Pipeline) SimpleInsertDocument(ctx context.Context, docText string, resource string) error {
	chunks := p.Chunker.SplitByParagraph(docText)
	// 批量喂AI向量化
	const batchLength = 100
	for i := 0; i < len(chunks); i += batchLength {
		//一次喂100条
		end := i + batchLength
		if end > len(chunks) {
			end = len(chunks)
		}
		batchChunks := chunks[i:end]
		vectors, err := p.EmbeddingHandler(ctx, batchChunks)
		if err != nil {
			return err
		}
		// 构建列数据，id自增不用摁造
		columns := []column.Column{
			column.NewColumnVarChar("wiki_varchar", batchChunks),
			column.NewColumnFloatVector("wiki_vector", int(utils.GetEmbedDim()), vectors),
			column.NewColumnVarChar("wiki_resource", repeat(resource, len(batchChunks))),
		}
		_, err = p.MilvusClient.Insert(ctx,
			milvusclient.NewColumnBasedInsertOption(p.CollectionName, columns...),
		)
		if err != nil {
			return fmt.Errorf("rag:fail to insert the %d batch %w", i/batchLength, err)
		}

	}
	task, err := p.MilvusClient.Flush(ctx, milvusclient.NewFlushOption(p.CollectionName))
	if err != nil {
		return fmt.Errorf("rag:fail to flush: %w", err)
	}
	if err = task.Await(ctx); err != nil {
		return fmt.Errorf("rag:flush over time %w", err)
	}
	return nil
}
