package rag

import (
	"Main/utils"
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func (p *Pipeline) CreateCollection(ctx context.Context) {
	schema := entity.NewSchema().WithDynamicFieldEnabled(true).
		WithField(entity.NewField().WithName("wiki_id").WithIsAutoID(true).WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName("wiki_vector").WithDataType(entity.FieldTypeFloatVector).WithDim(utils.GetEmbedDim())).
		WithField(entity.NewField().WithName("wiki_varchar").WithDataType(entity.FieldTypeVarChar).WithMaxLength(65535)).
		WithField(entity.NewField().WithName("wiki_resource").WithDataType(entity.FieldTypeVarChar).WithMaxLength(256))

	err := p.milvusClient.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(p.collectionName, schema))
	if err != nil {
		fmt.Println("rag:创建milvus错误！" + err.Error())
	}
}
func (p *Pipeline) CreateIndex(ctx context.Context, overwrite bool, autoIndex bool) {
	if overwrite {
		err := p.milvusClient.DropIndex(ctx, milvusclient.NewDropIndexOption(p.collectionName, "wiki_vector"))
		if err != nil {
			fmt.Println("rag: 尝试删除旧索引: " + err.Error())
		}
	}
	var idx index.Index
	if autoIndex {
		idx = index.NewAutoIndex(entity.COSINE)
	} else {
		idx = index.NewIvfFlatIndex(entity.COSINE, 128)
	}
	opt := milvusclient.NewCreateIndexOption(p.collectionName, "wiki_vector", idx)
	task, err := p.milvusClient.CreateIndex(ctx, opt)
	if err != nil {
		fmt.Println("rag: 创建索引失败！" + err.Error())
		return
	}

	if err = task.Await(ctx); err != nil {
		fmt.Println("rag: 等待索引创建完成失败！" + err.Error())
		return
	}
	fmt.Printf("rag:索引创建成功！\n")
}
func (p *Pipeline) DropCollection(ctx context.Context) error {
	p.milvusClient, _ = milvusclient.New(ctx, &milvusclient.ClientConfig{Address: "localhost:19530"})
	err := p.milvusClient.DropCollection(ctx, milvusclient.NewDropCollectionOption("wiki_docs"))
	return err
}
