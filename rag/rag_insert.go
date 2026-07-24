package rag

import (
	"Main/utils"
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/openai/openai-go/v3"
)

type Pipeline struct {
	embedClient    *openai.Client
	milvusClient   *milvusclient.Client
	chunker        *utils.TextChunker
	collectionName string
}

func (p *Pipeline) Close() {
	p.milvusClient.Close(context.Background())
}
func repeat(s string, n int) []string {
	arr := make([]string, n)
	for i := range arr {
		arr[i] = s
	}
	return arr
}

func NewPipeline(collectionName string) (*Pipeline, error) {
	embedClient := NewEmbeddingClient()
	milvusClient, err := milvusclient.New(context.Background(), &milvusclient.ClientConfig{
		Address: "localhost:19530",
	})
	if err != nil {
		return nil, fmt.Errorf("rag:创建Pipeline错误！" + err.Error())
	}
	chunker := utils.NewTextChunker(1000, 200)
	return &Pipeline{
		embedClient:    &embedClient,
		milvusClient:   milvusClient,
		chunker:        chunker,
		collectionName: collectionName,
	}, nil
}
func (p *Pipeline) InsertDocument(ctx context.Context, docText string, resource string) error {
	chunks := p.chunker.SplitByParagraph(docText)
	// 批量喂AI向量化
	const batchLength = 100
	for i := 0; i < len(chunks); i += batchLength {
		//一次喂100条
		end := i + batchLength
		if end > len(chunks) {
			end = len(chunks)
		}
		batchChunks := chunks[i:end]
		vectors, err := EmbeddingHandler(ctx, p.embedClient, batchChunks)
		if err != nil {
			return fmt.Errorf("rag:第%d批Embedding失败:%w", i/batchLength, err)
		}
		// 构建列数据，id自增不用摁造
		columns := []column.Column{
			column.NewColumnVarChar("wiki_varchar", batchChunks),
			column.NewColumnVarChar("wiki_resource", repeat(resource, len(batchChunks))),
			column.NewColumnFloatVector("wiki_vector", int(utils.GetEmbedDim()), vectors),
		}
		result, err := p.milvusClient.Insert(ctx,
			milvusclient.NewColumnBasedInsertOption(p.collectionName, columns...),
		)
		if err != nil {
			return fmt.Errorf("rag:第%d批插入失败: %w", i/batchLength, err)
		}
		fmt.Printf("tag:成功插入%d条数据(第%d批),IDs:%v\n",
			result.InsertCount, i/batchLength+1, result.IDs)
	}
	return nil
}
