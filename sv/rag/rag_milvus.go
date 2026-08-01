package rag

import (
	"Main/utils"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func (p *Pipeline) CreateCollection(ctx context.Context) error {
	exists, err := p.MilvusClient.HasCollection(ctx,
		milvusclient.NewHasCollectionOption(p.CollectionName))
	if err != nil {
		return fmt.Errorf("rag: check collection existence before create failed: %w", err)
	}
	if exists {
		// 集合已存在
		return nil
	}
	schema := entity.NewSchema().WithDynamicFieldEnabled(true).
		WithField(entity.NewField().WithName("wiki_id").WithIsAutoID(true).WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName("wiki_vector").WithDataType(entity.FieldTypeFloatVector).WithDim(utils.GetEmbedDim())).
		WithField(entity.NewField().WithName("wiki_varchar").WithDataType(entity.FieldTypeVarChar).WithMaxLength(65535)).
		WithField(entity.NewField().WithName("wiki_resource").WithIsPartitionKey(true).WithDataType(entity.FieldTypeVarChar).WithMaxLength(512)).
		WithField(entity.NewField().WithName("file_path").WithDataType(entity.FieldTypeVarChar).WithMaxLength(512)).
		WithField(entity.NewField().WithName("file_hash").WithDataType(entity.FieldTypeVarChar).WithMaxLength(512)).
		WithField(entity.NewField().WithName("start_line").WithDataType(entity.FieldTypeInt64)).
		WithField(entity.NewField().WithName("end_line").WithDataType(entity.FieldTypeInt64)).
		WithField(entity.NewField().WithName("language").WithDataType(entity.FieldTypeVarChar).WithMaxLength(64))
	err = p.MilvusClient.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(p.CollectionName, schema))
	if err != nil {
		return fmt.Errorf("rag:fail to create collection %w", err)
	}
	return nil
}
func (p *Pipeline) CreateIndex(ctx context.Context, overwrite bool, autoIndex bool) error {
	if overwrite {
		_ = p.MilvusClient.ReleaseCollection(ctx, milvusclient.NewReleaseCollectionOption(p.CollectionName))
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
	err := p.MilvusClient.DropCollection(ctx, milvusclient.NewDropCollectionOption(p.CollectionName))
	if err != nil {
		return fmt.Errorf("rag:fail to drop collections %w", err)
	}
	return nil
}
func (p *Pipeline) LoadCollections(ctx context.Context) error {
	exists, err := p.MilvusClient.HasCollection(ctx, milvusclient.NewHasCollectionOption(p.CollectionName))
	if err != nil {
		return fmt.Errorf("rag:check collection existence fail: %w", err)
	}
	if !exists {
		return fmt.Errorf("rag:collection does not exist")
	}

	loadCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	if loadCtx == nil {
		return fmt.Errorf("loadCtx is nil")
	}

	loadTask, err := p.MilvusClient.LoadCollection(loadCtx,
		milvusclient.NewLoadCollectionOption(p.CollectionName))
	if err != nil {
		return fmt.Errorf("rag:fail to load collections %w", err)
	}
	return loadTask.Await(loadCtx)
}
func (p *Pipeline) GetFileRecords(ctx context.Context, repoUrl string) (map[string]string, error) {
	queryOption := milvusclient.NewQueryOption(p.CollectionName).
		WithFilter(fmt.Sprintf("wiki_resource == '%s'", strings.ReplaceAll(repoUrl, "'", "\\'"))).
		WithOutputFields("file_path", "file_hash")
	queryResult, err := p.MilvusClient.Query(ctx, queryOption)
	if err != nil {
		return nil, fmt.Errorf("rag: query file records failed: %w", err)
	}
	// 从 Fields中获取列
	pathCol := queryResult.GetColumn("file_path")
	hashCol := queryResult.GetColumn("file_hash")

	//类型断言
	pathColTyped, ok := pathCol.(*column.ColumnVarChar)
	if !ok {
		return nil, fmt.Errorf("rag: file_path column type mismatch")
	}
	hashColTyped, ok := hashCol.(*column.ColumnVarChar)
	if !ok {
		return nil, fmt.Errorf("rag: file_hash column type mismatch")
	}

	// 提取数据
	paths := pathColTyped.Data()
	hashes := hashColTyped.Data()
	fileMap := make(map[string]string, len(paths))
	for i := 0; i < len(paths); i++ {
		fileMap[paths[i]] = hashes[i]
	}
	return fileMap, nil
}
func (p *Pipeline) DeleteFile(ctx context.Context, repoURL, filePath string) error {
	filter := fmt.Sprintf(
		"wiki_resource == '%s' and file_path == '%s'",
		strings.ReplaceAll(repoURL, "'", "\\'"),
		strings.ReplaceAll(filePath, "'", "\\'"),
	)
	deleteOption := milvusclient.NewDeleteOption(p.CollectionName).
		WithExpr(filter)
	_, err := p.MilvusClient.Delete(ctx, deleteOption)
	if err != nil {
		return fmt.Errorf("rag: delete file records failed: %w", err)
	}
	return nil
}
