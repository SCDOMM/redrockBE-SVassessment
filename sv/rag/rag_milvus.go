package rag

import (
	"Main/utils"
	"context"
	"fmt"
	"time"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func (p *Pipeline) CreateCollection(ctx context.Context) error {
	schema := entity.NewSchema().WithDynamicFieldEnabled(true).
		WithField(entity.NewField().WithName("wiki_id").WithIsAutoID(true).WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName("wiki_vector").WithDataType(entity.FieldTypeFloatVector).WithDim(utils.GetEmbedDim())).
		WithField(entity.NewField().WithName("wiki_varchar").WithDataType(entity.FieldTypeVarChar).WithMaxLength(65535)).
		WithField(entity.NewField().WithName("wiki_resource").WithDataType(entity.FieldTypeVarChar).WithMaxLength(512)).
		WithField(entity.NewField().WithName("file_path").WithDataType(entity.FieldTypeVarChar).WithMaxLength(512)).
		WithField(entity.NewField().WithName("start_line").WithDataType(entity.FieldTypeInt64)).
		WithField(entity.NewField().WithName("end_line").WithDataType(entity.FieldTypeInt64)).
		WithField(entity.NewField().WithName("language").WithDataType(entity.FieldTypeVarChar).WithMaxLength(64))

	err := p.MilvusClient.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(p.CollectionName, schema))
	if err != nil {
		return fmt.Errorf("rag:fail to create collection %w", err)
	}
	return nil
}
func (p *Pipeline) CreateIndex(ctx context.Context, overwrite bool, autoIndex bool) error {
	if overwrite {
		err := p.MilvusClient.DropIndex(ctx, milvusclient.NewDropIndexOption(p.CollectionName, "wiki_vector"))
		if err != nil {
			return fmt.Errorf("rag:fail to override index %w", err)
		}
	}
	// 构建索引
	var idx index.Index
	if autoIndex {
		idx = index.NewAutoIndex(entity.COSINE)
	} else {
		idx = index.NewIvfFlatIndex(entity.COSINE, 128)
	}
	opt := milvusclient.NewCreateIndexOption(p.CollectionName, "wiki_vector", idx)
	task, err := p.MilvusClient.CreateIndex(ctx, opt)
	if err != nil {
		return fmt.Errorf("rag:fail to create index %w", err)
	}
	if err = task.Await(ctx); err != nil {
		return fmt.Errorf("rag:fail to wait index created %w", err)
	}
	return nil
}
func (p *Pipeline) DropCollection(ctx context.Context) error {
	err := p.MilvusClient.DropCollection(ctx, milvusclient.NewDropCollectionOption("wiki_docs"))
	return fmt.Errorf("rag:fail to drop collections %w", err)
}
func (p *Pipeline) LoadCollections(ctx context.Context) error {
	loadCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	loadTask, err := p.MilvusClient.LoadCollection(loadCtx,
		milvusclient.NewLoadCollectionOption(p.CollectionName))
	if err != nil {
		return fmt.Errorf("rag:fail to load collections %w", err)
	}
	return loadTask.Await(loadCtx)
}
