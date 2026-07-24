package rag

import (
	"Main/utils"
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func CreateCollection() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	milvusAddr := "localhost:19530"
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: milvusAddr,
	})
	if err != nil {
		fmt.Println("rag:连接milvus错误！" + err.Error())
	}

	schema := entity.NewSchema().WithDynamicFieldEnabled(true).
		WithField(entity.NewField().WithName("wiki_id").WithIsAutoID(true).WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName("wiki_vector").WithDataType(entity.FieldTypeFloatVector).WithDim(utils.GetEmbedDim())).
		WithField(entity.NewField().WithName("wiki_varchar").WithDataType(entity.FieldTypeVarChar).WithMaxLength(65535)).
		WithField(entity.NewField().WithName("wiki_resource").WithDataType(entity.FieldTypeVarChar).WithMaxLength(256))

	err = client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(utils.GetCollectionsName(), schema))
	if err != nil {
		fmt.Println("rag:创建milvus错误！" + err.Error())
	}

	defer client.Close(ctx)

}
func CreateIndex(overwrite bool, autoIndex bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	milvusAddr := "localhost:19530"
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: milvusAddr,
	})
	if err != nil {
		fmt.Println("rag: 连接 Milvus 错误！" + err.Error())
		return
	}
	defer client.Close(ctx)

	if overwrite {
		err = client.DropIndex(ctx, milvusclient.NewDropIndexOption(utils.GetCollectionsName(), "wiki_vector"))
		if err != nil {
			fmt.Println("rag: 尝试删除旧索引: " + err.Error())
			return
		}
	}
	var idx index.Index
	if autoIndex {
		idx = index.NewAutoIndex(entity.COSINE)
	} else {
		idx = index.NewIvfFlatIndex(entity.COSINE, 128)
	}
	opt := milvusclient.NewCreateIndexOption(utils.GetCollectionsName(), "wiki_vector", idx)
	if _, err = client.CreateIndex(ctx, opt); err != nil {
		fmt.Println("rag: 创建索引失败！" + err.Error())
		return
	}
	fmt.Printf("rag:索引创建成功！\n")
}
